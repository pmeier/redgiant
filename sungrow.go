package redgiant

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pmeier/redgiant/internal/errors"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/google/uuid"
)

type Response struct {
	Code    int             `json:"result_code"`
	Message string          `json:"result_msg"`
	Data    json.RawMessage `json:"result_data"`
}

func (r Response) MarshalZerologObject(e *zerolog.Event) {
	e.Int("code", r.Code).Str("message", r.Message)
	if len(r.Data) > 0 {
		e.RawJSON("data", r.Data)
	}
}

type SungrowDisconnectedError struct {
	*errors.RedgiantError
}

func newSungrowDisconnectedError(msg string) error {
	return &SungrowDisconnectedError{RedgiantError: errors.New(msg, errors.WithHiddenFrames(2))}
}

type Sungrow struct {
	Host            string
	Username        string
	Password        string
	log             zerolog.Logger
	c               *http.Client
	mu              sync.Mutex
	ws              *websocket.Conn
	connected       bool
	token           string
	cancelHeartbeat context.CancelFunc
	reconnectTries  uint
}

func NewSungrow(host string, username string, password string, opts ...OptFunc) *Sungrow {
	o := ResolveOptions(append([]OptFunc{
		WithLogger(log.Logger),
		WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: time.Second * 60,
		}),
		WithReconnect(3),
	}, opts...)...)
	return &Sungrow{Host: host, Username: username, Password: password, c: o.HTTPClient, log: o.Logger, reconnectTries: o.ReconnectTries}
}

func (s *Sungrow) Connect() error {
	s.log.Trace().Msg("Redgiant.Connect()")

	log := s.log.With().Str("host", s.Host).Logger()

	if s.connected {
		log.Debug().Msg("already connected")
		return nil
	}
	log.Info().Msg("connecting")

	var tcc *tls.Config
	if _, ok := s.c.Transport.(*http.Transport); ok {
		tcc = s.c.Transport.(*http.Transport).TLSClientConfig
	} else {
		// FIXME: this also needs to be configurable
		tcc = &tls.Config{}
	}
	dialer := websocket.Dialer{TLSClientConfig: tcc}
	u := url.URL{Scheme: "wss", Host: s.Host, Path: "/ws/home/overview"}
	ws, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return errors.Wrap(err)
	}
	s.ws = ws

	type data struct {
		Token string `json:"token"`
	}

	token := make([]byte, 32)
	rand.Read(token)
	var d data
	err = s.Send("connect", map[string]any{"token": hex.EncodeToString(token), "id": uuid.NewString()}, &d)
	if err != nil {
		return err
	}
	s.connected = true

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelHeartbeat = cancel
	go s.heartbeat(ctx)

	err = s.Send("login", map[string]any{"token": d.Token, "username": "user", "passwd": s.Password}, &d)
	if err != nil {
		return err
	}
	s.token = d.Token

	log.Info().Msg("connected")
	return nil
}

func (s *Sungrow) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 3)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.log.Debug().Msg("heartbeat")
			if err := s.Send("ping", map[string]any{"token": ":", "id": uuid.NewString()}, nil); err != nil {
				s.log.Error().Err(err).Send()
			}
		}
	}
}

func (s *Sungrow) Close() {
	s.log.Trace().Msg("Sungrow.Close()")

	if s.ws == nil {
		s.log.Debug().Msg("already disconnected")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.token = ""
	if s.cancelHeartbeat != nil {
		s.cancelHeartbeat()
	}
	s.connected = false

	wmt := websocket.CloseMessage
	if err := s.ws.WriteMessage(wmt, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
		s.log.Debug().Msg("connection closed by server")
		return
	}
	rmt, _, err := s.ws.ReadMessage()
	if err != nil {
		s.log.Debug().Err(err).Msg("no closing message from server")
	} else if rmt != wmt {
		s.log.Debug().Int("write", wmt).Int("read", rmt).Msg("closing handshake message type mismatch")
	}

	s.log.Info().Str("host", s.Host).Msg("disconnected")
}

func (s *Sungrow) reconnect() error {
	s.Close()

	var err error
	for try := range s.reconnectTries {
		s.log.Info().Uint("try", try).Msg("reconnecting")
		if err = s.Connect(); err == nil {
			return nil
		}
		// FIXME: implement proper backoff here
		time.Sleep(time.Second * 20)
	}

	return newSungrowDisconnectedError("unable to reconnect")
}

func (s *Sungrow) Get(path string, params map[string]string, v any) error {
	s.log.Trace().Str("path", path).Any("params", params).Any("v", v).Msg("Sungrow.Get()")

	if s.token == "" {
		return errors.New("not connected")
	}

	u := url.URL{Scheme: "https", Host: s.Host, Path: path}
	q := u.Query()
	q.Set("lang", "zh_cn")
	q.Set("token", s.token)
	q.Set("page", "1")
	q.Set("limit", "10")
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	for {
		r, err := s.get(u)
		switch err.(type) {
		case SungrowDisconnectedError:
			if err := s.reconnect(); err != nil {
				return err
			}
			continue
		case error:
			return err
		}

		return json.Unmarshal(r.Data, v)
	}
}

func (s *Sungrow) get(u url.URL) (*Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Trace().Str("u", u.String()).Msg("Sungrow.get()")

	r, err := s.c.Get(u.String())
	if err != nil {
		return nil, newSungrowDisconnectedError(err.Error())
	}
	defer r.Body.Close()

	var resp Response
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return nil, errors.Wrap(err)
	}

	s.log.Trace().EmbedObject(resp).Msg("response")

	return &resp, nil
}

func (s *Sungrow) Send(service string, params map[string]any, v any) error {
	s.log.Trace().Str("service", service).Any("params", params).Msg("Sungrow.Send()")

	if (!s.connected && service != "connect") || (s.connected && s.token == "" && service != "login") {
		return errors.New("not connected")
	}
	reconnect := func() error {
		if service == "connect" || service == "login" {
			return errors.New("unable to connect")
		}
		return s.reconnect()
	}

	m := map[string]any{
		"lang":    "zh_cn",
		"token":   s.token,
		"service": service,
	}
	for k, v := range params {
		m[k] = v
	}

	for {
		resp, err := s.send(service, m)
		switch err.(type) {
		case *SungrowDisconnectedError:
			if err := reconnect(); err != nil {
				return err
			}
			continue
		case error:
			return errors.Wrap(err)
		}

		var d any
		if err := json.Unmarshal(resp.Data, &d); err != nil {
			d = string(resp.Data)
		}

		switch resp.Code {
		case 1:
			if service == "ping" {
				return nil
			}

			return json.Unmarshal(resp.Data, v)
		case 100, 104, 106:
			if err := reconnect(); err != nil {
				return err
			}
			continue
		default:
			return errors.New("unknown server error")
		}
	}
}

var responseCodesToBeDropped = []int{
	// The session of the web UI timed out
	103,
}

func (s *Sungrow) send(service string, m map[string]any) (*Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Trace().Str("service", service).Any("m", m).Msg("Sungrow.send()")

	if err := s.ws.WriteJSON(m); err != nil {
		return nil, newSungrowDisconnectedError(err.Error())
	}

	var r Response
	for {
		if err := s.ws.ReadJSON(&r); err != nil {
			return nil, newSungrowDisconnectedError(err.Error())
		}
		s.log.Trace().EmbedObject(r).Msg("read message")

		// Generally, there is a 1-to-1 correspondence between sent and received messages.
		// However, some messages are produced by the inverter without a corresponding one.
		// These messages have to be dropped.
		if slices.Contains(responseCodesToBeDropped, r.Code) {
			s.log.Debug().Str("reason", "code").Int("code", r.Code).Msg("message dropped")
			continue
		}

		if service != "ping" {
			var sd struct {
				Service string `json:"service"`
			}
			if err := json.Unmarshal(r.Data, &sd); err != nil {
				return nil, errors.Wrap(err)
			} else if sd.Service != service {
				s.log.Debug().Str("reason", "service mismatch").Str("write", service).Str("read", sd.Service).Msg("response dropped due to service mismatch")
				continue
			}
		}

		return &r, nil
	}
}
