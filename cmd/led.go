package main

import (
	"github.com/chrismarget/cloudkey-led"
	"log"
	"time"
)

func main() {
	//on := cloudkeyled.LedSetting{Percent: 75}
	//off := cloudkeyled.LedSetting{Percent: 25}

	whiteDir := "/sys/devices/platform/leds-mt65xx/leds/white"
	//blueDir := "/sys/devices/platform/leds-mt65xx/leds/blue"

	whiteLed, err := cloudkeyled.NewCloudKeyLed(whiteDir)
	if err != nil {
		log.Fatal(err)
	}

	//blueLed, err := cloudkeyled.NewCloudKeyLed(blueDir)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//whiteLed<-cloudkeyled.LedSetting{OnOff: []int{50,50}}
	//whiteLed<-cloudkeyled.LedSetting{OnOff: []int{25,75}}
	//for i := 0; i <= 100; i += 3 {
	//	on := i
	//	off := 100 - i
	//	log.Println(on, off)
	//	whiteLed <- cloudkeyled.LedSetting{OnOff: []int{on, off}}
	//	time.Sleep(10 * time.Millisecond)
	//}

	whiteLed<-cloudkeyled.LedSetting{Pattern: []int{2,4,6}}
	//time.Sleep(2*time.Second)
	//whiteLed<-cloudkeyled.LedSetting{Count: 4}
	//whiteLed<-off
	for {
		time.Sleep(1000 * time.Millisecond)
	}
}
