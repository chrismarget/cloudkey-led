package cloudkeyled

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestEnsureDir(t *testing.T) {
	dir, err := ioutil.TempDir("","")
	if err != nil {
		t.Fatal(err)
	}

	err = ensureDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Remove(dir)
	if err != nil {
		t.Fatal(err)
	}
}
