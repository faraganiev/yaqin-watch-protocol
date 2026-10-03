package main

import (
	"fmt"
	watch "github.com/faraganiev/yaqin-watch-protocol"
	"time"
)

func main() {
	packet := watch.Build("SG", "1234567890", []byte("LK,120,2,87"))
	frame, _, err := watch.Parse(packet)
	if err != nil {
		panic(err)
	}
	msg := watch.Split(frame.Content)
	hb := watch.DecodeHeartbeat(frame, msg, time.Now().UTC(), "example:5093")
	fmt.Printf("device=%s battery=%d%% steps=%d\n", hb.DeviceID, hb.Battery, hb.Steps)
}
