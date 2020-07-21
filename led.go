package cloudkeyled

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	onDelayFile  = "delay_on"
	offDelayFile = "delay_off"

	brightnessFile = "brightness"
	triggerFile    = "trigger"
	maxLevelFile   = "max_brightness"

	triggerNone  = "none"
	triggerTimer = "timer"
)

type Command struct {
	Quit       bool
	Percent    int
	On         bool
	Off        bool
	OnOffDelay []int
	Pattern    []int
	Count      int
	Repeat     bool
}

type trigger struct {
	file    string
	modes   []string
	current int
}

type led struct {
	dir            string
	max            int
	trigger        trigger
	cmdInChan      chan Command
	command        Command
	cmdStartedChan chan struct{}
	cmdQuitChan    chan struct{}
}

func New(dir string) (chan Command, error) {
	// validate specified path
	err := ensureDir(dir)
	if err != nil {
		return nil, err
	}

	led := led{
		dir:            dir,
		cmdInChan:      make(chan Command),
		cmdStartedChan: make(chan struct{}),
		cmdQuitChan:    make(chan struct{}),
	}
	err = led.readTrigger()
	if err != nil {
		return nil, err
	}

	maxBrightness, err := readNumFromFile(path.Join(dir, maxLevelFile))
	if err != nil {
		led.max = 1
	} else {
		led.max = maxBrightness
	}

	go led.startReceiver()

	return led.cmdInChan, nil
}

func (o *led) readTrigger() error {
	// trigger file exists?
	tf := path.Join(o.dir, triggerFile)
	stat, err := os.Stat(tf)
	if err != nil {
		return err
	}

	// trigger file is file?
	mode := stat.Mode()
	if !mode.IsRegular() {
		return fmt.Errorf("%s is not a regular file", tf)
	}

	// file looks good. save it.
	o.trigger.file = tf

	data, err := ioutil.ReadFile(tf)
	if err != nil {
		return err
	}

	// construct the trigger structure. Doing so requires finding the trigger value
	// in square brackets, removing those brackets, and storing its index value.
	modes := strings.Split(strings.TrimRight(string(data), "\n"), " ")
	for i, mode := range modes {
		if strings.HasPrefix(mode, "[") && strings.HasSuffix(mode, "]") {
			mode = mode[1 : len(mode)-1]
			o.trigger.current = i
		}
		if len(mode) > 0 {
			o.trigger.modes = append(o.trigger.modes, mode)
		}
	}

	return nil
}

func (o *led) setTrigger(new string) error {
	// validate input
	var validTrigger bool
	var newIndex int
	for i, mode := range o.trigger.modes {
		if new == mode {
			validTrigger = true
			newIndex = i
			break
		}
	}
	if !validTrigger {
		return fmt.Errorf("invalid trigger %s - valid triggers are %s", new, strings.Join(o.trigger.modes, "/"))
	}

	// previously set?
	if o.trigger.current == newIndex {
		return nil
	} else {
		o.trigger.current = newIndex
		err := ioutil.WriteFile(o.trigger.file, []byte(new), 0600)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o led) getTrigger() string {
	return o.trigger.modes[o.trigger.current]
}

func (o *led) off() {
	err := o.setBrightness(0)
	if err != nil {
		log.Println(err)
	}
	o.cmdStartedChan <- struct{}{}
	<-o.cmdQuitChan
}

func (o *led) on() {
	err := o.setBrightness(o.max)
	if err != nil {
		log.Println(err)
	}
	o.cmdStartedChan <- struct{}{}
	<-o.cmdQuitChan
}

func (o *led) delayOnOff() {
	err := o.setDelayOnOff(o.command.OnOffDelay[0], o.command.OnOffDelay[1])
	if err != nil {
		log.Println(err)
	}
	o.cmdStartedChan <- struct{}{}
	<-o.cmdQuitChan
}

func (o *led) percent() {
	percent := o.command.Percent
	maxVal := 499
	if percent > 100 {
		percent = 100
	}
	if percent < 0 {
		percent = 0
	}

	onVal := maxVal * percent / 100
	offVal := maxVal * (100 - percent) / 100

	err := o.setDelayOnOff(onVal, offVal)
	if err != nil {
		log.Println(err)
	}
	o.cmdStartedChan <- struct{}{}
	<-o.cmdQuitChan
}

func (o *led) count() {
	err := o.flashXtimes(o.command.Count)
	o.cmdStartedChan <- struct{}{}
	if err != nil {
		log.Println(err)
		<-o.cmdQuitChan
		return
	}
	for o.command.Repeat {
		select {
		case <-o.cmdQuitChan:
			return
		default:
			err := o.flashXtimes(o.command.Count)
			if err != nil {
				log.Println(err)
				<-o.cmdQuitChan
				return
			}

		}
	}
	<-o.cmdQuitChan
}

func (o *led) pattern() {
	err := o.flashPattern(o.command.Pattern)
	o.cmdStartedChan <- struct{}{}
	if err != nil {
		log.Println(err)
		<-o.cmdQuitChan
		return
	}
	for o.command.Repeat {
		select {
		case <-o.cmdQuitChan:
			return
		default:
			err := o.flashPattern(o.command.Pattern)
			if err != nil {
				log.Println(err)
				<-o.cmdQuitChan
				return
			}
		}
	}
	<-o.cmdQuitChan
}

func (o *led) startReceiver() {
	// start dummy "previous command" to indicate the loop can proceed
	go func() { <-o.cmdQuitChan }()

	// loop over incoming commands.
CMDLOOP:
	for cmd := range o.cmdInChan {
		// stop the previous command
		o.cmdQuitChan <- struct{}{}
		o.command = cmd
		switch {
		case cmd.Quit:
			break CMDLOOP
		case cmd.Off:
			go o.off()
			<-o.cmdStartedChan
		case cmd.On:
			go o.on()
			<-o.cmdStartedChan
		case len(cmd.OnOffDelay) == 2:
			go o.delayOnOff()
			<-o.cmdStartedChan
		case cmd.Count > 0:
			go o.count()
			<-o.cmdStartedChan
		case len(cmd.Pattern) > 0:
			go o.pattern()
			<-o.cmdStartedChan
		default:
			go o.percent()
			<-o.cmdStartedChan
		}
	}
}

func (o *led) setDelayOnOff(on int, off int) error {
	err := o.setTrigger(triggerTimer)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(o.dir, onDelayFile), []byte(strconv.Itoa(on)), 0000)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(o.dir, offDelayFile), []byte(strconv.Itoa(off)), 0000)
	if err != nil {
		return err
	}

	return nil
}

func (o *led) setBrightness(b int) error {
	err := o.setTrigger(triggerNone)
	if err != nil {
		return err
	}

	bf := path.Join(o.dir, brightnessFile)
	bs := strconv.Itoa(b) + "\n"
	return ioutil.WriteFile(bf, []byte(bs), 0600)
}

func (o *led) flashXtimes(times int) error {
	err := o.setBrightness(0)
	if err != nil {
		return err
	}

	time.Sleep(500 * time.Millisecond)
	for i := 0; i < times; i++ {
		err := o.setBrightness(o.max)
		if err != nil {
			return err
		}
		time.Sleep(250 * time.Millisecond)
		err = o.setBrightness(0)
		if err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}

func (o *led) flashPattern(pattern []int) error {
	time.Sleep(time.Second)
	for _, i := range pattern {
		err := o.flashXtimes(i)
		if err != nil {
			return err
		}
		time.Sleep(time.Second)
	}
	return nil
}
