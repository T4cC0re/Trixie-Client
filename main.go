package main

import (
	"os/user"
	"fmt"
	"log"
	"os"
	"encoding/json"
	"io/ioutil"
	"time"
	"./Trixie"
	"path/filepath"
	"strings"
)

const (
	defaultTrixieURL = "https://trixie.bigpoint.net"
)

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal("Could not get user", err)
	}
	var configPath = fmt.Sprintf("%s%s.Trixie.json", usr.HomeDir, string(os.PathSeparator))
	// Create default configPath if not existent.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println(".Trixie.json does not exist")
		defaultConf := Trixie.Config{defaultTrixieURL, "", 0};
		jsonConf, err := json.Marshal(defaultConf)
		if err != nil {
			log.Fatal("Could not create default JSON", err)
		}
		fmt.Println(string(jsonConf))
		err = ioutil.WriteFile(configPath, jsonConf, 0600)
		if err != nil {
			log.Fatal("Could not write default configPath", err)
		}
	}
	rawConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatal("Could not read config", err)
	}

	var config Trixie.Config
	err = json.Unmarshal(rawConfig, &config)
	if err != nil {
		log.Fatal("Could not parse Config", err)
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
		fmt.Printf("Your auth-token is expired. Please renew with '%s refresh'\n", binaryName)
	}

	trixie := Trixie.NewExecutor(configPath, config.TrixieURL, config.AuthToken, binaryName)

	trixie.Execute(action, actionPrefix, params...)

	fmt.Println(configPath)
}

func detectBinary(p string) (string, string) {
	strippedName := strings.Replace(filepath.Base(p), ".exe", "", -1)
	switch strippedName {
	case "tsrvdb", "trixie-srvdb", "srvdb", "___srvdb":
		return strippedName, "srvdb."
	case "tvmware", "trixie-vmware", "vmware", "___vmware":
		return strippedName, "vmware."
	case "tfix", "trixie-fix", "fix", "___fix":
		return strippedName, "fix."
	case "magic", "trixie", "___trixie":
		return strippedName, "internal."
	}
	return strippedName, "internal."
}
