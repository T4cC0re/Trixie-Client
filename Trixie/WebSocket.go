package Trixie

import (
	"./Tracer"
	"encoding/json"
	"github.com/gorilla/websocket"
	"os"
	"os/signal"
	"strings"
	"time"
)

type WebSocket struct {
	Addr          string
	Socket        *websocket.Conn
	Open          bool
	Finished      bool
	Authenticated bool
	Code          uint8
	CodeChan      chan uint8
	ErrorChan     chan error
}

type WSAuth struct {
	Action string `json:"action"`
	Token  string `json:"token"`
}
type WSDo struct {
	Action string   `json:"action"`
	Params []string `json:"params"`
}

type WSMsg struct {
	Fd       uint8  `json:"fd"`
	Log      string `json:"log"`
	Auth     bool   `json:"auth"`
	Finished bool   `json:"finished"`
	Code     uint8  `json:"code"`
}

func NewWebSocket(addr string, ignoreSignals bool) (*WebSocket, error) {
	defer Tracer.Un(Tracer.Track("NewWebSocket"))

	ws := WebSocket{}
	ws.Addr = addr
	ws.Open = true
	ws.Authenticated = false
	ws.Code = 0

	interrupt := make(chan os.Signal, 1)
	if !ignoreSignals {
		signal.Notify(interrupt, os.Interrupt)
	}
	url := strings.Replace(addr, "http", "ws", 1)

	socket, _, err := websocket.DefaultDialer.Dial(url, nil)
	ws.Socket = socket
	if err != nil {
		return nil, err
	}

	socket.SetCloseHandler(func(code int, text string) error {
		defer Tracer.Un(Tracer.Track("Socket Close Handler"))
		//signal.Notify(interrupt, syscall.SIGHUP)
		ws.Open = false
		return nil
	})

	done := make(chan struct{})
	ws.CodeChan = make(chan uint8)
	ws.ErrorChan = make(chan error)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	go func() {
		defer Tracer.Un(Tracer.Track("NewWebSocket::go func 1"))
		defer socket.Close()
		defer close(done)
		for {
			if !ws.Open {
				return
			}
			_, message, err := socket.ReadMessage()
			b := WSMsg{}
			err = json.Unmarshal(message, &b)
			if err != nil {
				ws.ErrorChan <- err
			}

			Tracer.Log("Received Message: fd: %d, text: '%s', auth: %b, code: %d, finished: %b", b.Fd, b.Log, b.Auth, b.Code, b.Finished)

			printWSOutLine(&b)
			if b.Auth {
				ws.Authenticated = true
				continue // API will send a 'finished' upon auth.
			}
			if b.Code != 0 {
				ws.CodeChan <- b.Code
			}
			if b.Finished {
				ws.Finished = true
				// Make sure we have a code at the end
				if b.Code == 0 {
					ws.CodeChan <- 0
				}
			}
			Tracer.Log("EOM")
		}
	}()

	go func() {
		defer Tracer.Un(Tracer.Track("NewWebSocket::go func 2"))
		for {
			if !ws.Open {
				break
			}
			select {
			case <-ticker.C:
				err := socket.WriteMessage(websocket.PingMessage, nil)
				if err != nil {
					ws.ErrorChan <- err
				}
			case <-interrupt:
				// To cleanly close a connection, a client should send a close
				// frame and wait for the server to close the connection.
				err := socket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					ws.ErrorChan <- err
				}
				select {
				case <-done:
				case <-time.After(time.Second):
				}
				socket.Close()
			}
		}
	}()

	return &ws, nil
}

func (t *WebSocket) makeRequest(endpoint string, payload RemotePayload, auth string) (uint8, error) {
	defer Tracer.Un(Tracer.Track("makeRequest"))
	authPayload, err := json.Marshal(WSAuth{"auth", auth})
	if err != nil {
		return 1, err
	}
	t.Socket.WriteMessage(websocket.TextMessage, authPayload)

	for {
		if t.Authenticated || t.Finished || !t.Open {
			break
		}
	}

	commandPayload, err := json.Marshal(WSDo{endpoint, payload.Params})
	if err != nil {
		return 1, err
	}

	t.Code = 1
	t.Socket.WriteMessage(websocket.TextMessage, commandPayload)
	for {
		if t.Finished || !t.Open {
			break
		}
		select {
		case t.Code = <-t.CodeChan:
			break
		case err = <-t.ErrorChan:
			return t.Code, err
		}
	}

	t.Finished = false

	Tracer.Log("Got Code %d", t.Code)

	return t.Code, nil
}
