package cloudkeyled

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"time"
)

const (
	triggerFile  = "trigger"
	onDelayFile  = "delay_on"
	offDelayFile = "delay_off"
)

type LedSetting struct {
	Quit    bool
	Percent int
	OnOff   []int
	Pattern []int
	Count   int
}

func timerTrigger(dir string, trigger string) error {
	trigF := path.Join(dir, triggerFile)
	onF := path.Join(dir, onDelayFile)
	offF := path.Join(dir, offDelayFile)

	_, onFerr := os.Stat(onF)
	_, offFerr := os.Stat(offF)
	if onFerr != nil || offFerr != nil {
		// delay_on / delay_off files not found ... Need trigger file to create them
		_, err := os.Stat(trigF)
		if err != nil {
			return err
		}

		// cause delay_on / delay_off files to be created
		err = ioutil.WriteFile(trigF, []byte(trigger), 0000)
		if err != nil {
			return err
		}

		timer := time.NewTimer(5000 * time.Millisecond)
		var timeout bool
		go func() {
			<-timer.C
			timeout = true
		}()

		var on_found, off_found bool
		for !on_found || !off_found || !timeout {
			if !on_found {
				_, err = os.Stat(onF)
				if err != nil {
					continue
				} else {
					on_found = true
					continue
				}
			}

			if !off_found {
				_, err = os.Stat(offF)
				if err != nil {
					continue
				} else {
					off_found = true
					continue
				}
			}

			if on_found && off_found {
				break
			}

			time.Sleep(10 * time.Millisecond)
		}

		if timeout {
			return fmt.Errorf("found on: %t, found off: %t", on_found, off_found)
		}
	}

	return nil
}

func onOff(dir string, on int, off int, done chan struct{}) {
	err := timerTrigger(dir, "timer")
	if err != nil {
		log.Println(err)
		<-done
		return
	}

	err = ioutil.WriteFile(path.Join(dir, onDelayFile), []byte(strconv.Itoa(int(on))), 0000)
	if err != nil {
		log.Println(err)
		<-done
		return
	}

	err = ioutil.WriteFile(path.Join(dir, offDelayFile), []byte(strconv.Itoa(int(off))), 0000)
	if err != nil {
		log.Println(err)
		<-done
		return
	}

	<-done
}

func count(dir string, count int, done chan struct{}) {
	localDone := make(chan struct{})

	go percent(dir, 0, localDone)
	time.Sleep(500 * time.Millisecond)

	for i := count; i > 0; i-- {
		go percent(dir, 100, localDone)
		time.Sleep(250 * time.Millisecond)
		go percent(dir, 0, localDone)
		time.Sleep(500 * time.Millisecond)
	}
	done <- struct{}{}
	<-done
}

func pattern(dir string, pattern []int, done chan struct{}) {
	localDone := make(chan struct{})
	time.Sleep(1*time.Second)
	for _, i := range pattern {
		go count(dir, i, localDone)
		<- localDone
		localDone <- struct{}{}
		time.Sleep(1*time.Second)
	}
	done <- struct{}{}
	<-done
}

func percent(dir string, percent int, done chan struct{}) {
	if percent > 100 {
		percent = 100
	}
	if percent < 0 {
		percent = 0
	}

	onOff(dir, percent, 100-percent, done)
}

func NewCloudKeyLed(dir string) (chan LedSetting, error) {
	// ensure specified LED path exists
	stat, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", dir)
	}

	settingChan := make(chan LedSetting)
	go cloudKeyLedReceiver(dir, settingChan)

	return settingChan, nil
}

func dummy(doneChan chan struct{}) {
	<-doneChan
}

func cloudKeyLedReceiver(dir string, in chan LedSetting) {
	doneChan := make(chan struct{})
	//go percent(dir, 0, doneChan)
	go dummy(doneChan)

	for instruction := range in {
		doneChan <- struct{}{}
		switch {
		case instruction.Quit:
			log.Println("quitting")
			return
		case instruction.OnOff != nil:
			go onOff(dir, instruction.OnOff[0], instruction.OnOff[1], doneChan)
		case instruction.Count > 0:
			log.Println("count: ", instruction.Count)
			go count(dir, instruction.Count, doneChan)
			<-doneChan
		case instruction.Pattern != nil:
			go pattern(dir, instruction.Pattern, doneChan)
			<-doneChan
		default:
			go percent(dir, instruction.Percent, doneChan)
		}
	}
}
