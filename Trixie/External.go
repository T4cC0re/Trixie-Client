package Trixie

import (
	"./Tracer"
	"time"
)

func execWebSocket(url string, auth string, action string, args ...string) (int, error) {
	defer Tracer.Un(Tracer.Track("execWebSocket"))
	ws, err := NewWebSocket(url, false)
	payload := RemotePayload{args}
	time.Sleep(10)
	code, err := ws.makeRequest(action, payload, auth)

	return int(code), err
}
