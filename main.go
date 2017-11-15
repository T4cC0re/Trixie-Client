package main

import (
	"os/user"
	"fmt"
	"os"
	"encoding/json"
	"io/ioutil"
	"time"
	"./Trixie"
	"path/filepath"
	"strings"
	"regexp"
)

const (
	defaultTrixieURL = "https://trixie.bigpoint.net"
)

func main() {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	configPath := fmt.Sprintf("%s%s.Trixie.json", usr.HomeDir, string(os.PathSeparator))

	// Create default configPath if not existent.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, ".Trixie.json does not exist. Creating a default config for you")
		defaultConf := Trixie.Config{defaultTrixieURL, "", 0};
		jsonConf, err := json.Marshal(defaultConf)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(jsonConf))
		err = ioutil.WriteFile(configPath, jsonConf, 0600)
		if err != nil {
			panic(err)
		}
	}

	rawConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	config := new(Trixie.Config)
	err = json.Unmarshal(rawConfig, &config)
	if err != nil {
		panic(err)
	}

	binaryName, actionPrefix := detectBinary(os.Args[0])

	action := "--help"
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

	trixie.Execute(action, actionPrefix, params...)

	fmt.Println(configPath)
}

func detectBinary(p string) (string, string) {
	strippedName := strings.Replace(filepath.Base(p), ".exe", "", -1)
	regex := regexp.MustCompile(`(?i)(t_?|trixie-|___)?(?P<namespace>[a-z]*)$`)
	var namespace string
	if string_ := regex.FindStringSubmatch(strippedName); len(string_) == 3 {
		namespace = string_[2] + "."

		if namespace == "trixie." || namespace == "rixie." {
			namespace = "internal."
		}
	} else {
		namespace = "internal."
	}

	// Enable for debugging:
	//fmt.Fprintf(os.Stderr, "Detected %s => %s\n", strippedName, namespace)

	return strippedName, namespace
}
