package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"fortio.org/cli"
	"fortio.org/log"
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
	fNoBlink := flag.Bool("no-blink", false, "Don't blink the colon")
	// fNoBox := flag.Bool("no-box", false, "Don't draw a box around the time")
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
	hoffset := bignum.Width + 2
	if seconds {
		format += ":05"
		hoffset = 0
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
	frame := 0
	prev := ""
	ap.OnResize = func() error {
		ap.ClearScreen()
		ap.WriteBoxed(ap.H/2-bignum.Height/2, "%s", TimeString(prev))
		return nil
	}
	for {
		ap.StartSyncMode()
		now := time.Now()
		numStr = now.Format(format)
		if numStr != prev {
			ap.WriteBoxed(ap.H/2-bignum.Height/2, "%s", TimeString(numStr))
		}
		prev = numStr
		now = now.Truncate(time.Second) // change only when seconds change
		if now != prevNow && !*fNoBlink {
			log.LogVf("frame %d now %v vs prev %v", frame, now, prevNow)
			prevNow = now
			frame++
			what := "::"
			if frame%2 == 0 {
				what = ".."
			}
			ap.WriteAtStr(ap.W/2+bignum.Width+1-hoffset, ap.H/2-bignum.Height/2+2, what)
			if seconds {
				ap.WriteAtStr(ap.W/2-2*bignum.Width, ap.H/2-bignum.Height/2+2, what)
			}
		}
		ap.EndSyncMode()
		_, err := ap.ReadOrResizeOrSignalOnce()
		if err != nil {
			return
		}
		if len(ap.Data) > 0 && (ap.Data[0] == 'q' || ap.Data[0] == 3) {
			return // exit on 'q' or Ctrl-C
		}
	}
}
