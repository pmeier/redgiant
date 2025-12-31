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

	"golang.org/x/sync/singleflight"

	"github.com/gorilla/websocket"
	"github.com/pmeier/redgiant/internal/errors"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/google/uuid"
)

type ConnectionStatus int

const (
	StatusDisconnected ConnectionStatus = iota
	StatusConnecting
	StatusConnected
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
	sfg             singleflight.Group
	status          ConnectionStatus
	ws              *websocket.Conn
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

func (s *Sungrow) reconnect() error {
	_, err, _ := s.sfg.Do("reconnect", func() (any, error) {
		for try := range s.reconnectTries {
			s.mu.Lock()

			s.log.Info().Uint("try", try).Msg("reconnecting")

			s.closeLocked()
			err := s.connectLocked()

			s.mu.Unlock()

			if err == nil {
				return nil, nil
			}

			// FIXME: implement proper backoff here
			time.Sleep(20 * time.Second)
		}
		return nil, newSungrowDisconnectedError("unable to reconnect")
	})
	return err
}

func (s *Sungrow) Connect() error {
	s.log.Trace().Msg("Sungrow.Connect()")

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status == StatusConnected {
		log.Debug().Str("host", s.Host).Msg("already connected")
		return nil
	}

	return s.connectLocked()
}

func (s *Sungrow) connectLocked() error {
	s.log.Trace().Msg("Sungrow.connectLocked()")

	log := s.log.With().Str("host", s.Host).Logger()
	log.Info().Msg("connecting")

	success := false
	defer func() {
		if !success {
			s.closeLocked()
		}
	}()

	s.status = StatusConnecting
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
	var d data

	token := make([]byte, 32)
	rand.Read(token)
	r, err := s.sendLocked("connect", map[string]any{"token": hex.EncodeToString(token), "id": uuid.NewString()})
	if err != nil {
		return err
	}
	if err = json.Unmarshal(r.Data, &d); err != nil {
		return err
	}

	r, err = s.sendLocked("login", map[string]any{"token": d.Token, "username": "user", "passwd": s.Password})
	if err != nil {
		return err
	}
	if err = json.Unmarshal(r.Data, &d); err != nil {
		return err
	}
	s.token = d.Token
	s.status = StatusConnected

	log.Info().Msg("connected")

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelHeartbeat = cancel
	go s.heartbeat(ctx)

	success = true
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

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status == StatusDisconnected {
		s.log.Debug().Msg("already disconnected")
		return
	}

	s.closeLocked()
}

func (s *Sungrow) closeLocked() {
	s.log.Trace().Msg("Sungrow.closeLocked()")

	if s.cancelHeartbeat != nil {
		s.cancelHeartbeat()
	}

	s.status = StatusDisconnected
	s.token = ""

	if s.ws != nil {
		closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
		s.ws.SetWriteDeadline(time.Now().Add(time.Second))
		if err := s.ws.WriteMessage(websocket.CloseMessage, closeMsg); err != nil {
			s.log.Debug().Msg("failed to send websocket close message")
		} else {
			s.ws.SetReadDeadline(time.Now().Add(time.Second))
			_, _, err := s.ws.ReadMessage()

			if err != nil && !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				s.log.Debug().Err(err).Msg("websocket close handshake incomplete")
			}
		}

		s.ws.Close()
		s.ws = nil
	}

	s.log.Info().Str("host", s.Host).Msg("disconnected")
}

func (s *Sungrow) Get(path string, params map[string]string, v any) error {
	s.log.Trace().Str("path", path).Any("params", params).Any("v", v).Msg("Sungrow.Get()")

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status != StatusConnected {
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

	reconnect := func() error {
		s.mu.Unlock()
		err := s.reconnect()
		s.mu.Lock()
		return err
	}

	for {
		r, err := s.getLocked(u)
		switch err.(type) {
		case *SungrowDisconnectedError:
			if err := reconnect(); err != nil {
				return err
			}
			continue
		case error:
			return err
		}

		return json.Unmarshal(r.Data, v)
	}
}

func (s *Sungrow) getLocked(u url.URL) (*Response, error) {
	s.log.Trace().Str("u", u.String()).Msg("Sungrow.getLocked()")

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

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status != StatusConnected {
		return errors.New("not connected")
	}

	reconnect := func() error {
		s.mu.Unlock()
		err := s.reconnect()
		s.mu.Lock()
		return err
	}

	for {
		resp, err := s.sendLocked(service, params)
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

func (s *Sungrow) sendLocked(service string, params map[string]any) (*Response, error) {
	s.log.Trace().Str("service", service).Any("params", params).Msg("Sungrow.sendLocked()")

	m := map[string]any{
		"lang":    "zh_cn",
		"token":   s.token,
		"service": service,
	}
	for k, v := range params {
		m[k] = v
	}

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
