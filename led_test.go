package cloudkeyled

import (
	"crypto/rand"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"testing"
)

func TestNew(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}

	// write the trigger file to the temp dir
	liveTriggerIndex := 2 // index of item in square brackets below
	triggerString := "none nand-disk [timer] mmc0 mmc1 rfkill0 \n"
	tf := path.Join(dir, triggerFile)
	err = ioutil.WriteFile(tf, []byte(triggerString), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// generate a random value, write the max_brightness file
	r := make([]byte, 1)
	_, err = rand.Read(r)
	if err != nil {
		t.Fatal(err)
	}
	max :=int(r[0])
	randomValString := strconv.Itoa(max) + "\n"
	err = ioutil.WriteFile(path.Join(dir,maxLevelFile),[]byte(randomValString),0600)
	if err != nil {
		t.Fatal(err)
	}

	// create the new led
	led, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	if liveTriggerIndex != led.trigger.current {
		log.Fatalf("bad preset trigger %d, expected %d", led.trigger.current, liveTriggerIndex)
	}

	if max != led.max {
		log.Fatalf("bad max value %d, expected %d", led.max, max)
	}

	err = os.RemoveAll(dir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetTrigger(t *testing.T) {
	testVal := "one"
	tf := "/tmp/trigger"
	modes := []string{"zero", "one", "two"}
	led := led{
		trigger: trigger{
			file: tf,
			modes: modes,
			current: 2,
		},
	}

	err := led.setTrigger("one")
	if err != nil {
		t.Fatal(err)
	}

	result, err := ioutil.ReadFile(tf)
	if err != nil {
		t.Fatal(err)
	}

	if string(result) != testVal {
		t.Fatalf("expected %s, got %s", testVal, string(result))
	}

	log.Printf("result: %s, index %d\n", result, led.trigger.current)

	err = os.Remove(tf)
	if err != nil {
		t.Fatal(err)
	}
}
