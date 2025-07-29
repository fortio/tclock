package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fortio.org/cli"
	"fortio.org/tclock/bignum"
	"fortio.org/terminal/ansipixels"
)

func TimeString(numStr string) string {
	d := &bignum.Display{}
	for _, c := range numStr {
		d.PlaceDigit(c)
	}
	return d.String()
}

func main() {
	cli.MinArgs = 0
	cli.MaxArgs = 1
	cli.ArgsHelp = " [digits:digits...] or current time"
	f24 := flag.Bool("24", false, "Use 24-hour time format")
	fNoSeconds := flag.Bool("no-seconds", false, "Don't show seconds")
	cli.Main()
	var numStr string
	if flag.NArg() == 1 {
		numStr = flag.Arg(0)
		fmt.Println(TimeString(numStr))
		return
	}
	format := "3:04"
	if *f24 {
		format = "15:04"
	}
	seconds := !*fNoSeconds
	if seconds {
		format += ":05"
	}
	prev := ""
	tick := time.NewTicker(time.Second)
	ap := ansipixels.NewAnsiPixels(60)
	err := ap.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening terminal: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		fmt.Fprintf(ap.Out, "\r\n\n")
		ap.ShowCursor()
		ap.EndSyncMode()
		tick.Stop()
		ap.Restore()
	}()
	ap.HideCursor()
	ap.ClearScreen()
	_ = ap.GetSize()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	frame := 0
	ap.OnResize = func() error {
		ap.ClearScreen()
		ap.WriteBoxed(ap.H/2-bignum.Height/2, TimeString(prev))
		return nil
	}
	for {
		numStr = time.Now().Format(format)
		if numStr != prev {
			if prev != "" {
				fmt.Fprintf(ap.Out, "\x1b[%dA\r", bignum.Height-1) // Move cursor up to overwrite previous output
			}
			ap.StartSyncMode()
			ap.WriteBoxed(ap.H/2-bignum.Height/2, TimeString(numStr))
			ap.EndSyncMode()
		}
		prev = numStr
		frame++
		if frame%2 == 0 {
			// invert the dots

		}
		_, err := ap.ReadOrResizeOrSignalOnce()
		if err != nil {
			return
		}
		if len(ap.Data) > 0 && (ap.Data[0] == 'q' || ap.Data[0] == 3) {
			return // exit on 'q' or Ctrl-C
		}
	}
}
