package Trixie

import (
	"./Commands"
	"./Tracer"
	"bufio"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type Executor struct {
	ConfigPath string
	Url        string
	Auth       string
	BinaryName string
	HTTP       Http
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
	defer Tracer.Un(Tracer.Track("NewExecutor"))
	executor := new(Executor)
	executor.ConfigPath = configPath
	executor.Url = url
	executor.Auth = auth
	executor.BinaryName = binaryName
	executor.HTTP = *NewHTTP(url, auth)
	return executor
}

func (t Executor) printInfo() {
	defer Tracer.Un(Tracer.Track("printInfo"))
	// This junky blob will render as a nice ASCII Art :)
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
`)
}

func (t Executor) Execute(action string, actionPrefix string, args ...string) (int, error) {
	defer Tracer.Un(Tracer.Track("Execute"))
	action = strings.ToLower(strings.Trim(action, "\r\n "))
	switch action {
	case "login":
		err := t.login()
		if err != nil {
			return 1, err
		}
		os.Exit(0)
	case "-h", "--help", "help":
		t.printInfo()
		return t.execExternal("list")
	case "info":
		t.printInfo()
		return t.execExternal("info")
	default:
		if actionPrefix == "internal." {
			return t.execInternal(action, args...)
		} else {
			return t.execExternal(fmt.Sprintf("%s%s", actionPrefix, action), args...)
		}
	}
	return 1, E_UNKNOWN
}

func printWSOutLine(line *WSMsg) {
	defer Tracer.Un(Tracer.Track("printWSOutLine"))
	if strings.HasPrefix(line.Log, "vmrc:") {
		commands.Open(line.Log)
		fmt.Fprintf(
			os.Stdout,
			"Opened VMRC. If it does not open after a few seconds, make sure you installed VMware Remote Console.\nURI used:\n%s",
			line.Log)
		return
	}
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

func (t Executor) execInternal(action string, args ...string) (int, error) {
	defer Tracer.Un(Tracer.Track("execInternal"))
	switch action {
	case "createlink":
		return commands.CreateLink(args)
	case "do", "action":
		return t.execExternal(args[0], args[1:]...)
	}
	return 1, E_UNKNOWN
}

func (t Executor) execExternal(action string, args ...string) (int, error) {
	defer Tracer.Un(Tracer.Track("execExternal"))
	t.renew()
	return execWebSocket(t.Url, t.Auth, action, args...)
}

func readPassword(prompt string) string {
	defer Tracer.Un(Tracer.Track("readPassword"))
	fmt.Print(prompt)
	pass, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintln(os.Stderr, "\n\nReading password failed. If you use MSYS2/MINGW/Cygwin/Git-Bash, please prefix the command with 'winpty'.")
		os.Exit(1)
	}
	println()
	return string(pass)
}

func (t Executor) saveAuth(resp *http.Response) {
	defer Tracer.Un(Tracer.Track("saveAuth"))

	body, _ := ioutil.ReadAll(resp.Body)
	t.saveAuthByJSON(body)
}

func (t Executor) saveAuthByJSON(body []byte) error {
	defer Tracer.Un(Tracer.Track("saveAuthByJSON"))
	var dat map[string]interface{}

	if err := json.Unmarshal(body, &dat); err != nil {
		return E_JSON_MARSHAL
	}

	token := string(dat["token"].(string))
	validity := uint32(dat["tokenValidity"].(float64))

	newConf := Config{t.Url, token, validity}
	jsonConf, err := json.Marshal(newConf)
	if err != nil {
		return E_JSON_MARSHAL
	}

	err = ioutil.WriteFile(t.ConfigPath, jsonConf, 0600)
	if err != nil {
		return E_FILE_RW
	}

	t.Auth = token
	return nil
}

func (t Executor) renew() error {
	//return errors.New("renewal failed") //TODO: Hack hack hack!
	defer Tracer.Un(Tracer.Track("renew"))

	client := Http{t.Url, t.Auth, &http.Client{}}
	body, code, err := client.makeRequest("GET", "/auth/3600", RemotePayload{}, true)
	if err != nil {
		return E_HTTP_REQUEST_FAILED
	}

	if code == 200 {
		t.saveAuthByJSON([]byte(body))
		return nil
	}

	return E_RENEW_FAILED
}

func (t Executor) login() error {
	defer Tracer.Un(Tracer.Track("login"))
	if err := t.renew(); err == nil {
		return nil
	}

	fmt.Print("Existing Auth-token (if existent) is expired\nEnter Username:\n")

	//password, _ := speakeasy.Ask("Password: ")
	reader := bufio.NewReader(os.Stdin)
	username, _ := reader.ReadString('\n')
	username = strings.Trim(username, "\r\n ")
	password := readPassword("Password:\n")

	if len(username) <= 0 || len(password) <= 0 {
		return E_USERNAME_PASSWORD_BLANK
	}

	fmt.Printf("User: '%s' Pass: <given>\n", username)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/auth", t.Url), nil)
	if err != nil {
		return E_HTTP_REQUEST_FAILED
	}

	req.SetBasicAuth(username, password)

	resp, err := t.HTTP.Client.Do(req)
	if err != nil {
		return E_HTTP_REQUEST_FAILED
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		t.saveAuth(resp)
		return nil
	} else {
		return E_LOGIN_FAILED
	}
}
