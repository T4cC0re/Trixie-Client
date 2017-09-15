package Trixie

import (
	"fmt"
	"strings"
	"os"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"golang.org/x/crypto/ssh/terminal"
	"bufio"
	"errors"
	"bytes"
	"./Commands"
)

type Executor struct {
	ConfigPath string
	Url        string
	Auth       string
	BinaryName string
	Client     *http.Client
}

type APIAuthResponse struct {
	token         string
	tokenValidity uint32
}

type RemotePayload struct {
	Params []string `json:"params"`
}

type RemoteResponse struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Error string  `json:"error"`
}

func NewExecutor(configPath string, url string, auth string, binaryName string) *Executor {
	executor := new(Executor)
	executor.ConfigPath = configPath
	executor.Url = url
	executor.Auth = auth
	executor.BinaryName = binaryName
	executor.Client = &http.Client{}
	return executor
}

func (t Executor) printHelp() {
	fmt.Print(
		`the great and powerful Trixie! (v` + version + `)
... is here to help :)

Since Trixie is pure magic, the behaviour will depend on the filename.
The appropriate links should have been created for you :)
For Trixie native commands use 'trixie'
For 'srvdb.' commands use 'tsrvdb'
For 'fix.' commands use 'tfix'
For 'vmware.' commands use 'tvmware'

Each command is also matched for '*', 'trixie-*' and '___*', where * = srvdb, fix, vmware, etc.

To execute namespaced commands with 'trixie' use 'trixie action <your command> [<params> ...]'
 - e.g. 'trixie do srvdb.search svc.trixie%.%'

Commands you can use everywhere:
  - 'login' [opt. max. time in minutes]: Login to the Trixie-Server/refresh the current session if still valid
  - '--help'

Examples for srvdb:
  - tsrvdb get srv123456 svc.%
  - tsrvdb propsearch svc.trixie%.ip=10.22.33.44

Examples for fix:
  - tfix publicnetwork srv123456

Examples for vmware:
   - tvmware create.memcache pinf600 trixie-sessions-01 12G

More questions? Ask my master @ h.meyer@bigpoint.net!
`)
}

func (t Executor) Execute(action string, actionPrefix string, args ...string) {
	action = strings.ToLower(strings.Trim(action, "\r\n "))
	switch action {
	case "--help", "-h", "help":
		t.printHelp()
		os.Exit(0)
	case "login":
		err := t.login()
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	default:
		if actionPrefix == "internal." {
			os.Exit(t.execInternal(action, args...))
		} else {
			fmt.Printf(
				"Calling '%s' with as %s: '%s%s %s'\n",
				t.Url,
				t.BinaryName,
				actionPrefix,
				action,
				strings.Join(args, " "))

			os.Exit(t.execExternal(fmt.Sprintf("%s%s", actionPrefix, action), args...))
		}
	}

}
func (t Executor) execInternal(action string, args ...string) int {
	//TODO: Handle internal actions
	switch action {
	case "createlinks":
		return commands.CreateLinks()
	}
	//CreateLinks
	//Cleanup > delete config etc.
	return 1
}

func (t Executor) execExternal(action string, args ...string) int {
	payload := RemotePayload{args}
	resp, code, err := t.makeRequest(
		"POST",
		fmt.Sprintf("%s/action/%s", t.Url, action),
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

	fmt.Fprint(os.Stdout, response.Stdout)
	fmt.Fprint(os.Stderr, response.Stderr)

	if len(response.Error) > 0 {
		panic(response.Error)
	}

	if code > 399 {
		return 1
	}
	return 0
}

func readPassword(prompt string) string {
	fmt.Print(prompt)
	pass, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintln(os.Stderr, "\n\nReading password failed. If you use MSYS2/MINGW/Cygwin/Git-Bash, please use winpty.")
		panic(err.Error())
	}
	println()
	return string(pass)
}

func (t Executor) makeRequest(method string, endpoint string, payload RemotePayload, includeAuth bool) (string, uint16, error) {
	var body []byte
	var err error
	if method == "POST" {
		body, err = json.Marshal(&payload)
		if err != nil {
			panic(err)
		}
	}

	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", 0, err
	}

	if includeAuth {
		req.Header.Add("X-Trixie-Auth", t.Auth)
	}
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")

	}
	resp, err := t.Client.Do(req)
	if err != nil {
		return "" ,0, err
	}

	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return string(body), uint16(resp.StatusCode), nil
}

func (t Executor) saveAuth(resp *http.Response) {
	var dat map[string]interface{}

	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &dat); err != nil {
		panic(err)
	}

	fmt.Printf("retrieved token is vaild until: %u", uint32(dat["tokenValidity"].(float64)))

	token := string(dat["token"].(string))
	validity := uint32(dat["tokenValidity"].(float64))

	newConf := Config{t.Url, token, validity}
	jsonConf, err := json.Marshal(newConf)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(t.ConfigPath, jsonConf, 0600)
	if err != nil {
		panic(err)
	}

	t.Auth = token
}

func (t Executor) login() error {
	// First check renewal. If fails, use username and password (prompt user)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/auth", t.Url), nil)
	if err != nil {
		return err
	}

	req.Header.Add("X-Trixie-Auth", t.Auth)
	resp, err := t.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		t.saveAuth(resp)
		return nil
	}

	fmt.Print("Existing Auth-token (if existent) is expired\nEnter Username:\n")

	//password, _ := speakeasy.Ask("Password: ")
	reader := bufio.NewReader(os.Stdin)
	username, _ := reader.ReadString('\n')
	username = strings.Trim(username, "\r\n ")
	password := readPassword("Password:\n")

	if len(username) <= 0 || len(password) <= 0 {
		return errors.New("username or password blank")
	}

	fmt.Printf("User: '%s' Pass: <given>\n", username)

	req, err = http.NewRequest("GET", fmt.Sprintf("%s/auth", t.Url), nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(username, password)

	resp, err = t.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		t.saveAuth(resp)
		return nil
	} else {
		return errors.New("login failed!")
	}
}
