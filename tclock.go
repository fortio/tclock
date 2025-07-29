package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"fortio.org/cli"
	"fortio.org/tclock/bignum"
	"fortio.org/terminal/ansipixels"
)

func TimeString(numStr string, blink bool) string {
	d := &bignum.Display{}
	for _, c := range numStr {
		d.PlaceDigit(c, blink)
	}
	return d.String()
}

func main() {
	cli.MinArgs = 0
	cli.MaxArgs = 1
	cli.ArgsHelp = " [digits:digits...] or current time"
	f24 := flag.Bool("24", false, "Use 24-hour time format")
	fNoSeconds := flag.Bool("no-seconds", false, "Don't show seconds")
	fNoBlink := flag.Bool("no-blink", false, "Don't blink the colon")
	// fNoBox := flag.Bool("no-box", false, "Don't draw a box around the time")
	cli.Main()
	var numStr string
	if flag.NArg() == 1 {
		numStr = flag.Arg(0)
		fmt.Println(TimeString(numStr, false))
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
	ap := ansipixels.NewAnsiPixels(60)
	err := ap.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening terminal: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		fmt.Fprintf(ap.Out, "\r\n\n\n\n")
		ap.ShowCursor()
		ap.EndSyncMode()
		ap.Restore()
	}()
	ap.HideCursor()
	ap.ClearScreen()
	_ = ap.GetSize()
	var prevNow time.Time
	prev := ""
	ap.OnResize = func() error {
		ap.ClearScreen()
		ap.WriteBoxed(ap.H/2-bignum.Height/2, "%s", TimeString(prev, false))
		return nil
	}
	blink := false
	doBlink := !*fNoBlink
	for {
		doDraw := false
		now := time.Now()
		numStr = now.Format(format)
		if numStr != prev {
			doDraw = true
		}
		prev = numStr
		now = now.Truncate(time.Second) // change only when seconds change
		if now != prevNow && doBlink {
			blink = !blink
			doDraw = true
		}
		prevNow = now
		if doDraw {
			ap.StartSyncMode()
			ap.WriteBoxed(ap.H/2-bignum.Height/2, "%s", TimeString(numStr, blink))
			ap.EndSyncMode()
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
