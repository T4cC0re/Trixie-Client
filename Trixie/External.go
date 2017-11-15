package Trixie

import (
	"fmt"
	"encoding/json"
	"os"
	"time"
)

func execHTTP(http *Http, action string, args ...string) int {
	payload := RemotePayload{args}
	resp, code, err := http.makeRequest(
		"POST",
		fmt.Sprintf("/action/%s", action),
		payload,
		true)
	if err != nil {
		panic(err)
	}

	var response RemoteResponse
	bResp := []byte(resp)
	if err := json.Unmarshal(bResp, &response); err != nil {
		panic(err)
	}

	printOutput(&response.Output)

	if len(response.Error) > 0 {
		fmt.Fprintf(
			os.Stderr,
			"Error:\n%s\n\n Your token may be expired/blacklisted or your command was invalid",
			response.Error)
	}

	if code > 399 {
		return 1
	}

	return 0
}

func execWebSocket(ws *WebSocket, auth string, action string, args ...string) int {
	if ! ws.Open{
		ws = NewWebSocket(ws.Addr)
	}

	payload := RemotePayload{args}
	time.Sleep(10)
	code, err := ws.makeRequest(action,  payload, auth)
	if err != nil {
		panic(err)
	}

	return int(code)
}
//
