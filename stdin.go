package main

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fortio.org/log"
	"fortio.org/terminal"
)

func StdinTail(cfg *Config) int {
	reader := terminal.NewTimeoutReader(os.Stdin, 100*time.Millisecond)
	var numStr string
	ap := cfg.ap
	var buf [4096]byte
	ap.Out = bufio.NewWriter(os.Stdout)
	_ = ap.GetSize()
	blink := false
	var prevNow time.Time
	frame := 0
	prev := ""
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-c:
			log.LogVf("Interrupted, exiting")
			return 0
		default:
		}
		doDraw := cfg.breath
		now := time.Now()
		if cfg.countDown {
			left := cfg.end.Sub(now).Round(time.Second)
			if left < 0 {
				ap.WriteAt(0, ap.H-2, "\aTime's up reached at %s\r\n", now.Format(cfg.format))
				cfg.extraNewLinesAtEnd = false
				return 0
			}
			numStr = DurationString(left, cfg.seconds)
		} else {
			numStr = now.Format(cfg.format)
		}
		if numStr != prev {
			doDraw = true
		}
		prev = numStr
		now = now.Truncate(time.Second) // change only when seconds change
		if now != prevNow && cfg.blinkEnabled {
			blink = !blink
			doDraw = true
		}
		prevNow = now
		n, err := reader.Read(buf[:])
		if err != nil && !errors.Is(err, io.EOF) {
			return log.FErrf("Error reading stdin: %v", err)
		}
		if cfg.bounceSpeed > 0 {
			if frame%cfg.bounceSpeed == 0 {
				cfg.bounce++
				doDraw = true
			}
			frame++
		}
		if doDraw || n > 0 {
			cfg.frame++
			ap.StartSyncMode()
			if n > 0 {
				_, _ = ap.Out.Write(buf[:n])
				ap.SaveCursorPos()
			}
			cfg.DrawAt(-1, -1, TimeString(numStr, blink))
			ap.RestoreCursorPos()
			ap.EndSyncMode()
		}
	}
}
