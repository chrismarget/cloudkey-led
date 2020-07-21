package cloudkeyled

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func ensureDir(dir string) error {
	// validate path exists
	stat, err := os.Stat(dir)
	if err != nil {
		return err
	}

	// validate path is dir
	if !stat.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}

	return nil
}

func readNumFromFile(file string) (int, error) {
	fileData, err := ioutil.ReadFile(file)
	if err != nil {
		return 0, err
	}

	fileString := strings.TrimSuffix(string(fileData), "\n")
	return strconv.Atoi(fileString)
}
