package commands

import (
	"os"
	"path/filepath"
	"fmt"
	"regexp"
	"strings"
	"runtime"
)

func CreateLinks () int {
	currentPath := os.Args[0]
    strippedName := filepath.Base(strings.Replace(currentPath, ".exe", "", -1))
    baseDir := filepath.Dir(currentPath)
    extension := ""

    if runtime.GOOS == "windows" {
    	extension = ".exe"
	}

	regex := regexp.MustCompile(`(?i)(t_?|trixie-|___)?(?P<namespace>[a-z]*)$`)
	var linkPrefix string
	if string_ := regex.FindStringSubmatch(strippedName); len(string_) == 3 {
		linkPrefix = string_[1]
	} else {
		panic("Unknown prefix used. Not creating links")
	}

	linker := Linker{currentPath, baseDir, linkPrefix, extension}

	errors := 0
	errors += linker.link("srvdb")
	errors += linker.link("srvcfg")
	errors += linker.link("fix")
	errors += linker.link("vmware")

	return errors
}

type Linker struct {
	CurrentPath string
	BaseDir string
	LinkPrefix string
	Extension string
}

func (t Linker) link(binaryName string) int {
	if err := os.Symlink(t.CurrentPath, fmt.Sprintf("%s/%s%s%s", t.BaseDir, t.LinkPrefix, binaryName, t.Extension)); err != nil {
		fmt.Fprintf(os.Stderr, "Could not create %s link (%s)", binaryName, err.Error())
		return 1
	}

	return 0
}