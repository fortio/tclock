package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"fortio.org/cli"
	"fortio.org/log"
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

type Config struct {
	ap       *ansipixels.AnsiPixels
	boxed    bool
	color    string
	colorBox string
	inverse  bool
	debug    bool
	bounce   int // bounce counter, 0 means no bouncing
}

func bounce(frame, maximum int) int {
	m := frame % (2 * maximum)
	if m < maximum {
		return m
	}
	return 2*maximum - 1 - m
}

func (c *Config) DrawAt(x, y int, str string) {
	if c.debug {
		c.ap.DrawSquareBox(0, 0, c.ap.W, c.ap.H)
		c.ap.WriteAt(0, c.ap.H-1, "Mouse %d, %d [%dx%d]", c.ap.Mx, c.ap.My, c.ap.W, c.ap.H)
	}
	lines := strings.Split(str, "\n")
	// Assume all lines are the same width (which is the case here with bignum padding).
	width := c.ap.ScreenWidth(lines[0])
	if c.boxed {
		width += 2 // add box padding
	}
	height := len(lines)
	if c.boxed {
		height += 2 // add box padding
	}
	if x < 0 && y < 0 {
		// center
		x = c.ap.W/2 + width/2
		y = c.ap.H/2 + height/2 + 1
	}
	// We are in 0.0 coordinates, but on apple terminal for instance the mouse can go past the width (!)
	// so clamp to valid screen dimensions.
	x = min(x, c.ap.W-1)
	y = min(y, c.ap.H-1)
	if c.bounce != 0 {
		x = width - 1 + bounce(c.bounce, c.ap.W-width+1)
		y = height - 1 + bounce(c.bounce, c.ap.H-height+1)
	}
	// draw from bottom right corner
	x++
	y++
	x = max(x, width)
	y = max(y, height)
	if c.boxed {
		if c.colorBox != "" {
			// draw box
			c.ap.DrawColoredBox(x-width, y-height, width, height, c.colorBox, false)
		} else {
			// draw box around the time
			c.ap.DrawRoundBox(x-width, y-height, width, height)
		}
		x--
		y--
		width -= 2
		height -= 2
	}
	// draw the lines
	prefix := c.color
	if c.inverse {
		prefix = ansipixels.Inverse + c.color
	}
	suffix := ansipixels.Reset
	for i, line := range lines {
		c.ap.WriteAtStr(x-width, y-height+i, prefix+line+suffix)
	}
	// ap.MoveCursor(x-1, y-1)
}

var colorMap = map[string]string{
	"red":       ansipixels.Red,
	"brightred": ansipixels.BrightRed,
	"green":     ansipixels.Green,
	"blue":      ansipixels.Blue,
	"yellow":    ansipixels.Yellow,
	"cyan":      ansipixels.Cyan,
	"white":     ansipixels.White,
	"black":     ansipixels.Black,
}

func main() {
	os.Exit(Main())
}

func ColorFromString(color string) (string, error) {
	if c, ok := colorMap[color]; ok {
		return c, nil
	}
	if len(color) == 6 {
		var i int
		_, err := fmt.Sscanf(color, "%x", &i)
		if err != nil {
			return "", fmt.Errorf("invalid hex color '%s', must be hex RRGGBB: %w", color, err)
		}
		r := (i >> 16) & 0xFF
		g := (i >> 8) & 0xFF
		b := i & 0xFF
		return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b), nil
	}
	return "", fmt.Errorf("invalid color '%s',"+
		" must be RRGGBB or one of: red, brightred, green, blue, yellow, cyan, white, black", color)
}

func Main() int { //nolint:funlen // we could split the flags and rest.
	cli.MinArgs = 0
	cli.MaxArgs = 1
	cli.ArgsHelp = " [digits:digits...]\npass only flags will display current time; move mouse and click to place on screen"
	fBounce := flag.Int("bounce", 0, "Bounce speed (0 is no bounce and normal mouse mode); 1 is fastest, 2 is slower, etc.")
	f24 := flag.Bool("24", false, "Use 24-hour time format")
	fNoSeconds := flag.Bool("no-seconds", false, "Don't show seconds")
	fNoBlink := flag.Bool("no-blink", false, "Don't blink the colon")
	fBox := flag.Bool("box", false, "Draw a simple box around the time")
	fColorBox := flag.String("color-box", "", "RGB color box around the time")
	fColor := flag.String("color", "red", "Color to use RRGGBB or one of: red, brightred, green, blue, yellow, cyan, white, black")
	fInverse := flag.Bool("inverse", false, "Inverse the foreground and background")
	fDebug := flag.Bool("debug", false, "Debug mode, display mouse position and screen borders")
	cli.Main()
	var numStr string
	if flag.NArg() == 1 {
		numStr = flag.Arg(0)
		fmt.Println(TimeString(numStr, false))
		return 0
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
	cfg := &Config{
		ap:      ap,
		boxed:   *fBox,
		inverse: *fInverse,
		debug:   *fDebug,
	}
	cfg.color, err = ColorFromString(*fColor)
	if err != nil {
		return log.FErrf("Color error: %v", err)
	}
	if *fColorBox != "" {
		cfg.colorBox, err = ColorFromString(*fColorBox)
		if err != nil {
			return log.FErrf("Color box error: %v", err)
		}
		cfg.boxed = true // color box implies boxed
	}
	ap.HideCursor()
	ap.ClearScreen()
	trackMouse := false
	bounceSpeed := *fBounce
	bounce := (bounceSpeed > 0)
	if !bounce {
		ap.MouseTrackingOn()
		trackMouse = true
	}
	_ = ap.GetSize()
	var prevNow time.Time
	prev := ""
	ap.OnResize = func() error {
		ap.ClearScreen()
		cfg.DrawAt(-1, -1, TimeString(prev, false))
		return nil
	}
	blinkEnabled := !*fNoBlink
	blink := false
	// TODO: how to get initial mouse position?
	x, y := ap.Mx, ap.My
	frame := 0
	for {
		_, err := ap.ReadOrResizeOrSignalOnce()
		if err != nil {
			return 1
		}
		if len(ap.Data) > 0 && (ap.Data[0] == 'q' || ap.Data[0] == 3) {
			return 0 // exit on 'q' or Ctrl-C
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
		switch {
		case bounce:
			if frame%bounceSpeed == 0 {
				cfg.bounce++
				doDraw = true
			}
			frame++
		case trackMouse && (ap.Mx != x || ap.My != y):
			x, y = ap.Mx, ap.My
			doDraw = true
		}
		if doDraw {
			ap.StartSyncMode()
			ap.ClearScreen()
			// -1 to switch to ansipixels 0,0 origin (from 1,1 terminal origin)
			// also means 0,0 is now -1,-1 and will center the time until the mouse is moved.
			cfg.DrawAt(x-1, y-1, TimeString(numStr, blink))
			ap.EndSyncMode()
		}
	}
}
