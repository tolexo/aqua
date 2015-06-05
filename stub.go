package aqua

import (
	"errors"
	"github.com/kardianos/osext"
	"io/ioutil"
	"os"
	"strings"
)

func getContent(path string) (string, error) {

	var absPath string
	var exists bool

	if strings.HasPrefix(path, "/") {
		absPath = path
		_, err := os.Stat(absPath)
		exists = err == nil
	}
	if !exists {
		// try working directory
		wdir, ferr := os.Getwd()
		if ferr == nil {
			absPath = removeMultSlashes(wdir + "/" + path)
			_, err := os.Stat(absPath)
			exists = err == nil
		}
	}
	if !exists {
		// try executable directory
		edir, ferr := osext.ExecutableFolder()
		if ferr == nil {
			absPath = removeMultSlashes(edir + "/" + path)
			_, err := os.Stat(absPath)
			exists = err == nil
		}
	}

	if !exists {
		return "", errors.New("File not found in working/exe path: " + path)
	}

	b, err := ioutil.ReadFile(absPath)
	if err != nil {
		return "", err
	} else {
		return string(b), nil
	}
}
