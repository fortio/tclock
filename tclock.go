package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"fortio.org/cli"
	"fortio.org/log"
	"fortio.org/safecast"
	"fortio.org/tclock/bignum"
	"fortio.org/terminal/ansipixels"
	"fortio.org/terminal/ansipixels/tcolor"
)

func TimeString(numStr string, blink bool) string {
	d := &bignum.Display{}
	for _, c := range numStr {
		d.PlaceDigit(c, blink)
	}
	return d.String()
}

type Config struct {
	ap          *ansipixels.AnsiPixels
	boxed       bool
	color       string
	colorBox    string
	inverse     bool
	debug       bool
	bounce      int  // bounce counter, 0 means no bouncing
	frame       int  // frame counter for breathing effect
	breath      bool // whether to pulse the color
	bcolor      tcolor.RGBColor
	colorOutput tcolor.ColorOutput
}

func bounce(frame, maximum int) int {
	m := frame % (2 * maximum)
	if m < maximum {
		return m
	}
	return 2*maximum - 1 - m
}

func breath(frame int, c tcolor.RGBColor) tcolor.Color {
	maxi := int(max(c.R, c.G, c.B))
	mini := 2 * maxi / 5
	n := maxi - mini
	x := bounce(frame, n)
	c.R = safecast.MustConvert[uint8](max(0, int(c.R)-x))
	c.G = safecast.MustConvert[uint8](max(0, int(c.G)-x))
	c.B = safecast.MustConvert[uint8](max(0, int(c.B)-x))
	return tcolor.Color{RGBColor: c}
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
	// draw the digits
	prefix := c.color
	if c.breath {
		prefix = c.colorOutput.Foreground(breath(c.frame, c.bcolor))
	}
	if c.inverse {
		prefix = ansipixels.Inverse + c.color
	}
	suffix := ansipixels.Reset
	for i, line := range lines {
		c.ap.WriteAtStr(x-width, y-height+i, prefix+line+suffix)
	}
	// ap.MoveCursor(x-1, y-1)
}

func main() {
	os.Exit(Main())
}

func Main() int { //nolint:funlen // we could split the flags and rest.
	cli.MinArgs = 0
	cli.MaxArgs = 1
	cli.ArgsHelp = " [digits:digits...]\npass only flags will display current time; move mouse and click to place on screen"
	fBounce := flag.Int("bounce", 0, "Bounce speed (0 is no bounce and normal mouse mode); 1 is fastest, 2 is slower, etc.")
	f24 := flag.Bool("24", false, "Use 24-hour time format")
	fNoSeconds := flag.Bool("no-seconds", false, "Don't show seconds")
	fNoBlink := flag.Bool("no-blink", false, "Don't blink the colon")
	fBox := flag.Bool("box", false, "Draw a simple rounded corner outline around the time")
	fColorBox := flag.String("color-box", "", "Color box around the time")
	fColor := flag.String("color", "red",
		"Color to use: RRGGBB, hue,sat,lum ([0,1]) or one of: "+tcolor.ColorHelp)
	fBreath := flag.Bool("breath", false, "Pulse the color (only works for RGB)")
	fInverse := flag.Bool("inverse", false, "Inverse the foreground and background")
	fDebug := flag.Bool("debug", false, "Debug mode, display mouse position and screen borders")
	defaultTrueColor := false
	if os.Getenv("COLORTERM") != "" {
		defaultTrueColor = true
	}
	fTrueColor := flag.Bool("true-color", defaultTrueColor,
		"Use true color (24-bit RGB) instead of 8-bit ANSI colors (default is true if COLORTERM is set)")
	cli.Main()
	colorOutput := tcolor.ColorOutput{TrueColor: *fTrueColor}
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
	if err := ap.Open(); err != nil {
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
		ap:          ap,
		boxed:       *fBox,
		inverse:     *fInverse,
		debug:       *fDebug,
		breath:      *fBreath,
		colorOutput: colorOutput,
	}
	if cfg.breath {
		var black tcolor.RGBColor
		color, err := tcolor.FromString(*fColor)
		cfg.bcolor = color.RGBColor
		if err != nil || cfg.bcolor == black {
			log.Errf("Using red instead of color %v / %v", err, color)
			cfg.bcolor = tcolor.RGBColor{R: 255, G: 20, B: 30}
		}
	} else {
		color, err := tcolor.FromString(*fColor)
		if err != nil {
			return log.FErrf("Color error: %v", err)
		}
		cfg.color = colorOutput.Foreground(color)
	}
	if *fColorBox != "" {
		color, err := tcolor.FromString(*fColorBox)
		if err != nil {
			return log.FErrf("Color box error: %v", err)
		}
		cfg.colorBox = colorOutput.Foreground(color)
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
		doDraw := cfg.breath
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
			cfg.frame++
			ap.StartSyncMode()
			ap.ClearScreen()
			// -1 to switch to ansipixels 0,0 origin (from 1,1 terminal origin)
			// also means 0,0 is now -1,-1 and will center the time until the mouse is moved.
			cfg.DrawAt(x-1, y-1, TimeString(numStr, blink))
			ap.EndSyncMode()
		}
	}
}
