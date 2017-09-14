package Trixie

import (
	"fmt"
	"strings"
	"os"
	"encoding/json"
	"log"
	"io/ioutil"
	"net/http"
	"golang.org/x/crypto/ssh/terminal"
	"bufio"
	"errors"
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

Each command is also matched for '*', 'trixie-*' and '___*', where * = srvdb, etc.

To execute namespaced commands with 'trixie' use 'trixie action <your command> [<params> ...]'
 - e.g. 'trixie action srvdb.search svc.trixie%.%'

Commands you can use everywhere:
  - 'login' [opt. max. time in minutes]: Login to the Trixie-Server/refresh the current session if still valid
  - '--help'

Examples for srvdb:
  - tsrvdb get srv123456 svc.%
  - tsrvdb propsearch svc.trixie%.ip=10.22.33.44

Examples for fix:
  - tfix publicNetwork srv123456

Examples for vmware:
   - tvmware createMemcache pinf600 trixie-sessions-01 12G

More questions? Ask my master @ h.meyer@bigpoint.net!
`)
}

func (t Executor) Execute(action string, actionPrefix string, args ...string) {
	switch action {
	case "--help", "-h", "help":
		t.printHelp()
		os.Exit(0)
	case "login":
		err := t.login()
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	default:
		if actionPrefix == "internal." {
			os.Exit(t.execInternal(action, args...))
		} else {
			fmt.Printf(
				"Calling '%s' with '%s' as %s: '%s%s %s'\n",
				t.Url,
				t.Auth,
				t.BinaryName,
				actionPrefix,
				action,
				strings.Join(args, " "))

			os.Exit(t.execExternal(action, args...))
		}
	}

}
func (t Executor) execInternal(action string, args ...string) int {
	//TODO: Handle internal actions
	//CreateLinks
	//Cleanup > delete config etc.
	return 1
}

func (t Executor) execExternal(action string, args ...string) int {
	//TODO: Handle external actions
	return 1
}

func readPassword(prompt string) string {
	fmt.Print(prompt)
	pass, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintln(os.Stderr, "\n\nReading password failed. If you use MSYS/Cygwin/Git-Bash, please use winpty.")
		panic(err.Error())
	}
	println()
	return string(pass)
}

func (t Executor) login() error {
	//TODO: Actual login
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
		fmt.Println(resp.Body)
		//TODO Save new auth.
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

	var dat map[string]interface{}

	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(body, &dat)
		if err != nil {
			return err
		}
	} else {
		return errors.New("login failed!")
	}
	/// Save auth:

	fmt.Printf("Token: %s\nVaild until: %u", dat["token"], dat["tokenValidity"])

	token := string(dat["token"].(string))
	validity := uint32(dat["tokenValidity"].(float64))

	newConf := Config{t.Url, token, validity}
	jsonConf, err := json.Marshal(newConf)
	if err != nil {
		return err
	}
	fmt.Println(string(jsonConf))
	err = ioutil.WriteFile(t.ConfigPath, jsonConf, 0600)
	if err != nil {
		return err
	}

	t.Auth = token

	return nil
}
