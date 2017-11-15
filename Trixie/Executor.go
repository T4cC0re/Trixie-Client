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
	"./Commands"
	"flag"
)

type Executor struct {
	ConfigPath    string
	Url           string
	Auth          string
	BinaryName    string
	HTTP          Http
	UseWebSockets bool
}

type APIAuthResponse struct {
	token         string
	tokenValidity uint32
}

type RemotePayload struct {
	Params []string `json:"params"`
}

type OutLine struct {
	Fd  uint8  `json:"fd"`
	Log string `json:"log"`
}

type RemoteResponse struct {
	Stdout string    `json:"stdout"`
	Stderr string    `json:"stderr"`
	Output []OutLine `json:"output"`
	Error  string    `json:"error"`
}

func NewExecutor(configPath string, url string, auth string, binaryName string) *Executor {
	executor := new(Executor)
	executor.ConfigPath = configPath
	executor.Url = url
	executor.Auth = auth
	executor.BinaryName = binaryName
	executor.HTTP = *NewHTTP(url, auth)
	ws := flag.Bool("ws", false, "Use WebSocket Connection")
	flag.Parse()
	executor.UseWebSockets = *ws
	return executor
}

func (t Executor) printHelp() {
	fmt.Print("                                       [35m___...----..                \n" +
		"                                 [35m..''``         [94mx  [35m`.              \n" +
		"                                [35m/   [94m*       [96mx        [35m`.            \n" +
		"                               [35m|   \\_\\_\\         [97m.  [96mX  [35m\\           \n" +
		"                               [35m:  .'    '.  [94m*   [97m`X'     [35m|          \n" +
		"                               [35ml /       '.     [97m' `     [35m.          \n" +
		"                                [35mV         |   [96m*     .   [35m|          \n" +
		"                                          [35m|        [96m`X'  [35m|.         \n" +
		"                                     [35m_..--j   [94mX    [96m' `  [35m` `''--... \n" +
		"                                   [35m<'__  [96m*           [97mx      *    [35m.>\n" +
		"                                      [96m/[35m`[94m\\[35m'''----____      ___..''  \n" +
		"                                     [96m(   [94m\\_  [37m(   [35m/[90m.-[35m`[37m|[35m'[37m|[35m`'[37m| [96m.'   | \n" +
		"                                     [96m(     \\[37m,'\\ [35m([90m(WW [37m| \\[90mW[35m)[37mj [96m|   / .\n" +
		"          [96m..---'''''---              (      |  [37m\\_[35m\\[37m_ /   [94m``-.[96m:  :_/|\n" +
		"        [96m,'             `'.           (      |          [94m\\__/  [96m`---' \n" +
		"       [96m/   _              '.          \\     '. [94m-,______.-'         \n" +
		"      [96m| .-'/                :[94m__________[96m`.    |    [94m/                \n" +
		"      [96m'`  -          .-''>-'             \\   '.  [94m(                 \n" +
		"          [96m|         /   [94m/  [96m. [97m.            [96m'.  '.  [94m\\                \n" +
		"          [96m|         |  [94m|  [96m/|[97m`X'        [96m':-._)  |   [94m|               \n" +
		"         [96m.'         |  [94m| [96m( :[97m'[37m|[97m`          [96m`-___.'   [94m|               \n" +
		"         [96m|          |  [94m|  [96m\\ `[37m|[96m>                    [94m|               \n" +
		"         [96m|          | [94m/ \\  [96m``[37m'   [94m/             \\__/                \n" +
		"         [96m'.        .'[94m'   |      /-,_______\\       \\                \n" +
		"          [96m|        |   [94m_/      /     |    |\\       \\               \n" +
		"          [96m|        |  [94m/       /     |     | `--,    \\              \n" +
		"         [96m.'       :   [94m|      |      |     |   /      )             \n" +
		"   [96m.__..'    ,   :[94m\\__/|      |      |      | (       |             \n" +
		"    [96m`-.___.-`;  /     [94m|      |      |      |  \\      |             \n" +
		"           [96m.:_-'      [94m|       \\     |       \\  `.___/              \n" +
		"                       [94m\\_______)     \\_______)                     \n" +
		"[96m\n" +
		"\n\n" +
		`the great and powerful Trixie! (v` + version + `)
... is here to help :)

Since Trixie is pure magic, the behaviour will depend on the filename.
The appropriate links should have been created for you :)
For Trixie native commands use 'trixie'
For 'srvdb.' commands use 'tsrvdb'
For 'fix.' commands use 'tfix'
For 'vmware.' commands use 'tvmware'

Each command is also matched for '*', 'trixie-*' and '___*', where * = srvdb, fix, vmware, etc.

To execute namespaced commands with 'trixie' use 'trixie [do|action] <your command> [<params> ...]'
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
	case "showvm":
		os.Exit(t.execInternal(action, args...))
	default:
		if actionPrefix == "internal." {
			os.Exit(t.execInternal(action, args...))
		} else {
			os.Exit(t.execExternal(fmt.Sprintf("%s%s", actionPrefix, action), args...))
		}
	}

}

func printOutput(output *[]OutLine) {
	for _, line := range *output {
		printOutLine(&line)
	}
}

func printWSOutLine(line *WSMsg) {
	switch line.Fd {
	case 1:
		fmt.Fprint(os.Stdout, line.Log)
		break
	case 2:
		fmt.Fprint(os.Stderr, line.Log)
		break
	default:
		break
	}
}

func printOutLine(line *OutLine) {
	switch line.Fd {
	case 1:
		fmt.Fprint(os.Stdout, line.Log)
		break
	case 2:
		fmt.Fprint(os.Stderr, line.Log)
		break
	default:
		break
	}
}

func (t Executor) execInternal(action string, args ...string) int {
	switch action {
	case "createlinks":
		return commands.CreateLinks()
	case "showvm":
		_args := []string{"vm.console", args[0]}
		payload := RemotePayload{_args}
		resp, code, err := t.HTTP.makeRequest("POST", "/action/vmware.govc", payload, true)
		if err != nil {
			panic(err.Error())
		}

		var response RemoteResponse
		bResp := []byte(resp)
		if err := json.Unmarshal(bResp, &response); err != nil {
			panic(err)
		}

		if code > 399 {
			printOutput(&response.Output)
		}

		if len(response.Output[0].Log) == 0 {
			panic("Did not receive a VMRC link. Does the VM exist?")
		}

		return commands.Open(response.Output[0].Log)
	case "do", "action":
		return t.execExternal(args[0], args[1:]...)
	}
	return 1
}

func (t Executor) execExternal(action string, args ...string) int {
	//if t.UseWebSockets {
	ws := NewWebSocket(t.Url)
	return execWebSocket(ws, t.Auth, action, args...)
	//}
	//return  execHTTP(&t.HTTP, action, args...)
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
	resp, err := t.HTTP.Client.Do(req)
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

	resp, err = t.HTTP.Client.Do(req)
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
