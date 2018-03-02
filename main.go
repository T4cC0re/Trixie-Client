package main

import (
	"./Trixie"
	"./Trixie/Tracer"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var exitcode = 0

func main() {
	defer Tracer.Un(Tracer.Track("main"))
	defer exitWithCode()

	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	configPath := fmt.Sprintf("%s%s.Trixie.json", usr.HomeDir, string(os.PathSeparator))

	// Create default configPath if not existent.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, ".Trixie.json does not exist. Creating a default config for you")
		defaultConf := Trixie.Config{
			TrixieURL:         Trixie.DefaultURL,
			AuthToken:         "",
			AuthTokenValidity: 0}
		jsonConf, err := json.Marshal(defaultConf)
		if err != nil {
			panic(Trixie.E_JSON_MARSHAL)
		}
		fmt.Println(string(jsonConf))
		err = ioutil.WriteFile(configPath, jsonConf, 0600)
		if err != nil {
			panic(Trixie.E_FILE_RW)
		}
	}

	rawConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(Trixie.E_FILE_RW)
	}

	config := new(Trixie.Config)
	err = json.Unmarshal(rawConfig, &config)
	if err != nil {
		panic(Trixie.E_JSON_MARSHAL)
	}

	binaryName, actionPrefix := detectBinary(os.Args[0])

	action := "help"
	params := []string{}

	if len(os.Args) >= 2 {
		action = os.Args[1]
		if len(os.Args) >= 3 {
			params = os.Args[2:]
		}
	}

	currentTime := uint32(time.Now().Unix())

	if currentTime > config.AuthTokenValidity {
		fmt.Fprintf(os.Stderr, "Your auth-token is expired. Please renew with '%s login'\n", binaryName)
	}

	trixie := Trixie.NewExecutor(configPath, config.TrixieURL, config.AuthToken, binaryName)

	code, err := trixie.Execute(action, actionPrefix, params...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	exitcode = code
	runtime.Goexit()
}
func exitWithCode() {
	Tracer.Log("Exiting with Code %d", exitcode)
	os.Exit(exitcode)
}

func detectBinary(p string) (string, string) {
	defer Tracer.Un(Tracer.Track("detectBinary"))
	strippedName := strings.Replace(filepath.Base(p), ".exe", "", -1)

	regex := regexp.MustCompile(`(?i)(^t_?|trixie-|___)?(?P<namespace>[a-z]*)$`)
	var namespace string
	if string_ := regex.FindStringSubmatch(strippedName); len(string_) == 3 {
		namespace = string_[2] + "."
		Tracer.Log("parsed namespace: '%s'", namespace)
		switch namespace {
		case "trixie.", "rixie.", "trace.":
			Tracer.Log("rewriting namespace '%s' to 'internal.'", namespace)
			namespace = "internal."
		}
	} else {
		namespace = "internal."
	}

	Tracer.Log("binary name: '%s', namespace: '%s'", strippedName, namespace)

	return strippedName, namespace
}
