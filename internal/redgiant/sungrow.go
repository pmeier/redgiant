package redgiant

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	log "github.com/sirupsen/logrus"

	"github.com/google/uuid"
)

type Response struct {
	Code    int             `json:"result_code"`
	Message string          `json:"result_msg"`
	Data    json.RawMessage `json:"result_data"`
}

type Sungrow struct {
	Host         string
	Password     string
	PingInterval time.Duration
	mu           sync.Mutex
	ws           *websocket.Conn
	connected    bool
	token        string
	lastMessage  time.Time
	cancel       context.CancelFunc
}

func NewSungrow(host string, password string) *Sungrow {
	return &Sungrow{Host: host, Password: password, PingInterval: time.Second * 10}
}

func (s *Sungrow) Connect() error {
	log.Trace("Redgiant.Connect()")

	if s.connected {
		log.Debug("already connected")
		return nil
	}

	dialer := websocket.Dialer{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	u := url.URL{Scheme: "wss", Host: s.Host, Path: "/ws/home/overview"}
	ws, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return err
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

	err = s.Send("login", map[string]any{"token": d.Token, "username": "user", "passwd": s.Password}, &d)
	if err != nil {
		return err
	}

	s.connected = true
	s.token = d.Token

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go s.heartbeat(ctx)

	return nil
}

func (s *Sungrow) heartbeat(ctx context.Context) {
	w := s.PingInterval / 10
	if w > time.Second {
		w = time.Second
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(w):
			if time.Since(s.lastMessage) < s.PingInterval {
				continue
			}
			if err := s.Send("ping", map[string]any{"token": ":", "id": uuid.NewString()}, nil); err != nil {
				log.Error(err.Error())
			}
		}
	}
}

func (s *Sungrow) Close() {
	log.Trace("Sungrow.Close()")

	if s.ws == nil {
		log.Debug("cannot close, a connection was never established")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.connected = false
	if s.cancel != nil {
		s.cancel()
	}

	wmt := websocket.CloseMessage
	if err := s.ws.WriteMessage(wmt, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
		log.Debug("connection closed by server")
		return
	}
	rmt, _, err := s.ws.ReadMessage()
	if err != nil {
		log.Debug("no closing message from server")
	} else if rmt != wmt {
		log.WithFields(log.Fields{"write": wmt, "read": rmt}).Debug("closing handshake message type mismatch")
	}
}

func (s *Sungrow) Get(path string, params map[string]string, v any) error {
	log.WithFields(log.Fields{"path": path, "params": params, "v": fmt.Sprintf("%#v", v)}).Trace("Sungrow.Get()")

	u := url.URL{Scheme: "https", Host: s.Host, Path: path}
	q := u.Query()
	q.Set("lang", "zh_cn")
	q.Set("token", s.token)
	q.Set("page", "1")
	q.Set("limit", "10")
	for k, v := range params {
		q.Set(k, v)
	}

	r, err := http.Get(u.String())
	if err != nil {
		return err
	}

	defer r.Body.Close()
	var resp Response
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return err
	}

	return json.Unmarshal(resp.Data, v)
}

func (s *Sungrow) Send(service string, params map[string]any, v any) error {
	log.WithFields(log.Fields{"service": service, "params": params}).Trace("Sungrow.Send()")

	if !s.connected && !(service == "connect" || service == "login") {
		return errors.New("not connected")
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
		log.WithFields(m).Trace("message")
		resp, err := s.send(m)
		if err != nil {
			return err
		}

		var d any
		if err := json.Unmarshal(resp.Data, &d); err != nil {
			d = string(resp.Data)
		}
		log.WithFields(log.Fields{"code": resp.Code, "message": resp.Message, "data": d}).Trace("response")

		if resp.Code == 1 {
			if v == nil {
				return nil
			}

			return json.Unmarshal(resp.Data, v)
		}

		log.WithFields(log.Fields{"code": resp.Code, "message": resp.Message}).Error("message unsuccessful")
		s.Close()

		switch resp.Code {
		case 100, 104, 106:
			// add a reconnect function with back-off
			log.WithField("host", s.Host).Info("reconnecting")
			err = s.Connect()
		default:
			err = errors.New("unknown server error")
		}

		if err != nil {
			return err
		}
	}
}

// Generally, there is a 1-to-1 correspondence between sent and received messages.
// However, some messages are produced by the inverter without a corresponding one.
// These messages have to be dropped.
var responseCodesToBeDropped = []int{
	// This response code is sent to indicate that the session timed out,
	// but this only applies to the web UI
	103,
}

func (s *Sungrow) send(m map[string]any) (Response, error) {
	log.Trace("Sungrow.send()")

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ws.WriteJSON(m); err != nil {
		return Response{}, err
	}

	var r Response
	for {
		if err := s.ws.ReadJSON(&r); err != nil {
			return Response{}, err
		}
		s.lastMessage = time.Now()

		if !slices.Contains(responseCodesToBeDropped, r.Code) {
			return r, nil
		}

		log.WithFields(log.Fields{"code": r.Code, "message": r.Message}).Debug("message dropped")
	}
}
