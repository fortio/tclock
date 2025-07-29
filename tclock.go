package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
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

func DrawAt(ap *ansipixels.AnsiPixels, x, y int, boxed bool, str string) {
	// ap.DrawSquareBox(0, 0, ap.W, ap.H)
	// ap.WriteAt(0, ap.H-1, "Mouse %d, %d", ap.Mx, ap.My)
	lines := strings.Split(str, "\n")
	// Assume all lines are the same width (which is the case here with bignum padding).
	width := ap.ScreenWidth(lines[0])
	if boxed {
		width += 2 // add box padding
	}
	height := len(lines)
	if boxed {
		height += 2 // add box padding
	}
	if x < 0 && y < 0 {
		// center
		x = ap.W/2 + width/2
		y = ap.H/2 + height/2 + 1
	}
	x = min(x, ap.W)
	y = min(y, ap.H)
	// draw from bottom right corner
	x++
	y++
	x = max(x, width)
	y = max(y, height)
	// ap.WriteAt(0, ap.H-3, "x, y, width, height: %d, %d, %d, %d", x, y, width, height)
	if boxed {
		// draw box
		ap.DrawRoundBox(x-width, y-height, width, height)
		x--
		y--
		width -= 2
		height -= 2
	}
	// draw the lines
	for i, line := range lines {
		ap.WriteAtStr(x-width, y-height+i, line)
	}
	// ap.MoveCursor(x-1, y-1)
}

func main() {
	cli.MinArgs = 0
	cli.MaxArgs = 1
	cli.ArgsHelp = " [digits:digits...] or current time"
	f24 := flag.Bool("24", false, "Use 24-hour time format")
	fNoSeconds := flag.Bool("no-seconds", false, "Don't show seconds")
	fNoBlink := flag.Bool("no-blink", false, "Don't blink the colon")
	fNoBox := flag.Bool("no-box", false, "Don't draw a box around the time")
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
		ap.MouseTrackingOff()
		ap.EndSyncMode()
		ap.Restore()
	}()
	ap.HideCursor()
	ap.ClearScreen()
	ap.MouseTrackingOn()
	_ = ap.GetSize()
	var prevNow time.Time
	prev := ""
	ap.OnResize = func() error {
		ap.ClearScreen()
		DrawAt(ap, -1, -1, !*fNoBox, TimeString(prev, false))
		return nil
	}
	blinkEnabled := !*fNoBlink
	blink := false
	// TODO: how to get initial mouse position?
	x, y := ap.Mx, ap.My
	trackMouse := true
	for {
		_, err := ap.ReadOrResizeOrSignalOnce()
		if err != nil {
			return
		}
		if len(ap.Data) > 0 && (ap.Data[0] == 'q' || ap.Data[0] == 3) {
			return // exit on 'q' or Ctrl-C
		}
		// Click to place the time at the mouse position (or switch back to move with mouse).
		if ap.LeftClick() {
			trackMouse = !trackMouse
		}
		doDraw := false
		now := time.Now()
		numStr = now.Format(format)
		if numStr != prev {
			doDraw = true
		}
		prev = numStr
		now = now.Truncate(time.Second) // change only when seconds change
		if now != prevNow && blinkEnabled {
			blink = !blink
			doDraw = true
		}
		prevNow = now
		if trackMouse && (ap.Mx != x || ap.My != y) {
			x, y = ap.Mx, ap.My
			doDraw = true
		}
		if doDraw {
			ap.StartSyncMode()
			ap.ClearScreen()
			// -1 to switch to ansipixels 0,0 origin (from 1,1 terminal origin)
			// also means 0,0 is now -1,-1 and will center the time until the mouse is moved.
			DrawAt(ap, x-1, y-1, !*fNoBox, TimeString(numStr, blink))
			ap.EndSyncMode()
		}
	}
}
