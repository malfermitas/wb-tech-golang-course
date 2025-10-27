package main

import (
	"fmt"
	"os"
	"taskl2_8/ntptime"
	"time"
)

const NTPServer = "ntp0.ntp-servers.net"

func main() {
	currentTime, err := ntptime.GetTime(NTPServer, time.DateTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(-1)
	} else {
		fmt.Println(currentTime.CurrentTimeString)
		os.Exit(0)
	}
}
