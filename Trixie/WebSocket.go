package Trixie

import (
	"log"
	"os"
	"os/signal"
	"github.com/gorilla/websocket"
	"time"
	"encoding/json"
	"strings"
)

type WebSocket struct {
	Addr string
	Socket *websocket.Conn
	Open bool
	Finished bool
	Authenticated bool
	Code uint8
}

type WSAuth struct {
	Action string `json:"action"`
	Token string `json:"token"`
}
type WSDo struct {
	Action string `json:"action"`
	Params []string `json:"params"`
}


type WSMsg struct {
	Fd uint8 `json:"fd"`
	Log string `json:"log"`
	Auth bool `json:"auth"`
	Finished bool `json:"finished"`
	Code uint8 `json:"code"`
}

func NewWebSocket(addr string) *WebSocket {

	ws := WebSocket{}
	ws.Addr = addr
	ws.Open = true
	ws.Authenticated = false
	ws.Code = 0

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	url := strings.Replace(addr, "http", "ws", 1)
	//u := url.URL{Scheme: "ws", Host: strings.Replace("http"), Path: "/action/test"}
	log.Printf("connecting to %s", url)

	socket, _, err := websocket.DefaultDialer.Dial(url, nil)
	ws.Socket = socket
	if err != nil {
		log.Fatal("dial:", err)
	}

	socket.SetCloseHandler(func (code int, text string) error {
		ws.Open = false
		return nil
	})

	done := make(chan struct{})

	go func() {
		defer socket.Close()
		defer close(done)
		for {
			if !ws.Open {
				return
			}
			_, message, err := socket.ReadMessage()
			if err != nil {
				panic(err)
			}
			b := WSMsg{}
			err = json.Unmarshal(message, &b)
			if err != nil {
				panic(err)
			}
			printWSOutLine(&b)
			if b.Auth {
				ws.Authenticated = true
			}
			if b.Finished {
				ws.Finished = true
			}
			if b.Code > 0 {
				ws.Code = b.Code
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	go func() {
		for {
			if !ws.Open {
				break
			}
			select {
			case <-ticker.C:
				err := socket.WriteMessage(websocket.PingMessage, nil)
				if err != nil {
					log.Println("write:", err)
				}
			case <-interrupt:
				log.Println("interrupt")
				// To cleanly close a connection, a client should send a close
				// frame and wait for the server to close the connection.
				err := socket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("write close:", err)
				}
				select {
				case <-done:
				case <-time.After(time.Second):
				}
				socket.Close()
			}
		}
	}()

	return  &ws
}

func (t *WebSocket) makeRequest(endpoint string, payload RemotePayload, auth string) (uint8, error) {
	authPayload, err := json.Marshal(WSAuth{"auth", auth})
	if err != nil {
		panic(err)
	}
	t.Socket.WriteMessage(websocket.TextMessage, authPayload)
	
	for {
		if t.Authenticated || t.Finished || !t.Open {
			break
		} 
	}

	commandPayload, err := json.Marshal(WSDo{endpoint, payload.Params})
	if err != nil {
		panic(err)
	}

	t.Socket.WriteMessage(websocket.TextMessage, commandPayload)
	for {
		if t.Finished || !t.Open {
			break
		}
	}

	t.Finished = false
	t.Code = 0

	return t.Code, nil
}
