package main

import (
	"github.com/chrismarget/cloudkey-led"
	"log"
	"time"
)

func main() {
	whiteDir := "/sys/devices/platform/leds-mt65xx/leds/white"
	white, err := cloudkeyled.New(whiteDir)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i <= 100; i++ {
		white <- cloudkeyled.Command{Percent: i}
		time.Sleep(20*time.Millisecond)
	}
	for i := 99; i >= 0; i-- {
		white <- cloudkeyled.Command{Percent: i}
		time.Sleep(10*time.Millisecond)
	}
	white <- cloudkeyled.Command{Percent: 20}
	time.Sleep(1000*time.Millisecond)
	white <- cloudkeyled.Command{Percent: 40}
	time.Sleep(1000*time.Millisecond)
	white <- cloudkeyled.Command{Percent: 60}
	time.Sleep(1000*time.Millisecond)
	white <- cloudkeyled.Command{Percent: 80}
	time.Sleep(1000*time.Millisecond)
	white <- cloudkeyled.Command{Percent: 100}
	time.Sleep(1000*time.Millisecond)

	log.Println("on")
	white <- cloudkeyled.Command{On: true}
	time.Sleep(time.Second)
	log.Println("off")
	white <- cloudkeyled.Command{Off: true}
	time.Sleep(time.Second)
	log.Println("on")
	white <- cloudkeyled.Command{On: true}
	time.Sleep(time.Second)
	log.Println("off")
	white <- cloudkeyled.Command{Off: true}
	time.Sleep(time.Second)
	log.Println("500 1500")
	white <- cloudkeyled.Command{OnOffDelay: []int{500,1500}}
	time.Sleep(8*time.Second)
	log.Println("1500 500")
	white <- cloudkeyled.Command{OnOffDelay: []int{1500,500}}
	time.Sleep(8*time.Second)
	log.Println("ramp up")
	for i := 0; i <=100; i++ {
		white <- cloudkeyled.Command{Percent: i}
		time.Sleep(5*time.Millisecond)
	}
	log.Println("ramp down")
	for i := 99; i >=0; i-- {
		white <- cloudkeyled.Command{Percent: i}
		time.Sleep(5*time.Millisecond)
	}
	//log.Println("five times")
	//white <- cloudkeyled.Command{Count: 5}
	log.Println("three times")
	white <- cloudkeyled.Command{Count: 3}
	log.Println("two times repeating")
	white <- cloudkeyled.Command{Count: 2, Repeat: true}
	time.Sleep(5 * time.Second)
	log.Println("3,2,1")
	white <- cloudkeyled.Command{Pattern: []int{3,2,1}}
	time.Sleep(15*time.Second)
	log.Println("1,2,3 ...")
	white <- cloudkeyled.Command{Pattern: []int{1,2,3}, Repeat: true}
	time.Sleep(15*time.Second)
	log.Println("calling quit")
	white <- cloudkeyled.Command{Quit: true}

	time.Sleep(5*time.Second)
	log.Println("main exit")
}
