package Trixie

import (
	"./Tracer"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Http struct {
	Url    string
	Auth   string
	Client *http.Client
}

func NewHTTP(url string, auth string) *Http {
	defer Tracer.Un(Tracer.Track("NewHTTP"))
	executor := new(Http)
	executor.Url = url
	executor.Auth = auth
	executor.Client = &http.Client{}
	return executor
}

func (t Http) makeRequest(method string, endpoint string, payload RemotePayload, includeAuth bool) (string, uint16, error) {
	defer Tracer.Un(Tracer.Track("makeRequest"))
	var body []byte
	var err error
	if method == "POST" {
		body, err = json.Marshal(&payload)
		if err != nil {
			return "", 0, err
		}
	}

	endpoint = fmt.Sprintf("%s%s", t.Url, endpoint)

	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println(err.Error())
		return "", 0, err
	}

	req.Header.Set("Connection", "close")

	if includeAuth {
		req.Header.Set("X-Trixie-Auth", t.Auth)
	}
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}

	Tracer.Log("Calling %s on '%s'", method, endpoint)

	resp, err := t.Client.Do(req)
	if err != nil {
		return "", 0, err
	}

	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}

	return string(body), uint16(resp.StatusCode), nil
}
