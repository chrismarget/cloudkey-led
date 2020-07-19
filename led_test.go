package cloudkeyled

import "testing"

func TestLedMgr(t *testing.T){
	setLed, err := NewCloudKeyLed("/tmp")
	if err != nil {
		t.Fatal(err)
	}
	setLed <- LedSetting{quit: true}
	return
}
