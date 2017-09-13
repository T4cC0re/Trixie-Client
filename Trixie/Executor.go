package Trixie

import (
	"fmt"
	"strings"
	"os"
	"time"
	"encoding/json"
	"log"
	"io/ioutil"
)

type Executor struct {
	ConfigPath string
	Url        string
	Auth       string
	BinaryName string
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
  - 'refresh'/'login' [opt. max. time in minutes]: Login to the Trixie-Server/refresh the current session if still valid
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
	fmt.Printf(
		"Calling '%s' with '%s' as %s: '%s%s %s'\n",
		t.Url,
		t.Auth,
		t.BinaryName,
		actionPrefix,
		action,
		strings.Join(args, " "))

	switch action {
	case "--help", "-h", "help":
		t.printHelp()
		os.Exit(0)
	case "refresh", "login":
		err := t.login()
		if err != nil {
			log.Fatal("Login to Trixie", err)
		}
		os.Exit(0)
	default:
		if actionPrefix == "internal." {
			os.Exit(t.execInternal(action, args...))
		} else {
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

func (t Executor) login() error {
	//TODO: Actual login
	// First check renewal. If fails, use username and password (prompt user)

	token := "FoobarToken"
	validity := uint32(time.Now().Unix()) + 15

	newConf := Config{t.Url, token, validity};
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
