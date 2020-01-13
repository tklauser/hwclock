// Copyright 2019 Tobias Klauser. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

type rtc struct {
	file *os.File
}

func openRTC() (*rtc, error) {
	devs := []string{
		"/dev/rtc",
		"/dev/rtc0",
		"/dev/misc/rtc0",
	}

	for _, dev := range devs {
		f, err := os.Open(dev)
		if err == nil || !os.IsNotExist(err) {
			return &rtc{f}, err
		}
	}

	return nil, errors.New("No RTC device found")
}

func (r *rtc) read() (time.Time, error) {
	rt, err := unix.IoctlGetRTCTime(int(r.file.Fd()))
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(int(rt.Year)+1900,
		time.Month(rt.Mon+1),
		int(rt.Mday),
		int(rt.Hour),
		int(rt.Min),
		int(rt.Sec),
		0,
		time.UTC), nil
}

func (r *rtc) write(tu time.Time) error {
	rt := unix.RTCTime{
		Sec:   int32(tu.Second()),
		Min:   int32(tu.Minute()),
		Hour:  int32(tu.Hour()),
		Mday:  int32(tu.Day()),
		Mon:   int32(tu.Month() - 1),
		Year:  int32(tu.Year() - 1900),
		Wday:  int32(0),
		Yday:  int32(0),
		Isdst: int32(0)}

	return unix.IoctlSetRTCTime(int(r.file.Fd()), &rt)
}

func main() {
	hctosys := flag.Bool("s", false, "Set the System Clock from the Hardware Clock.")
	systohc := flag.Bool("w", false, "Set the Hardware Clock from the System Clock.")

	flag.Parse()

	if *hctosys && *systohc {
		log.Fatal("Options -s and -w are mutually exclusive")
	}

	rtc, err := openRTC()
	if err != nil {
		log.Fatal(err)
	}

	if *systohc {
		tu := time.Now().UTC()
		if err := rtc.write(tu); err != nil {
			log.Fatal(err)
		}
	}

	// Read current RTC value
	t, err := rtc.read()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(t.Local().Format("Mon Jan 2 2006 15:04:05 -0700 MST"))
}
