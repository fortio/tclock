package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"fortio.org/cli"
	"fortio.org/log"
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
	bounce      int             // bounce counter, 0 means no bouncing
	frame       int             // frame counter for breathing effect
	breath      bool            // whether to pulse the color
	bcolor      tcolor.RGBColor // color to use for breathing effect
	colorOutput tcolor.ColorOutput
	colorDisc   tcolor.RGBColor // color disc around the time, if set
	radius      float64         // radius of the disc around the time in proportion of the time width
	fillBlack   bool            // whether to fill the screen with black before drawing discs
	aliasing    float64         // aliasing factor for the disc drawing
	blackBG     string          // ANSI sequence for the black background: either 16 basic (color 0) or RGB black (truecolor).
	// whether to use linear blending for the color disc (instead of SRGB)
	blendingFunction func(tcolor.RGBColor, tcolor.RGBColor, float64) tcolor.RGBColor
}

func bounce(frame, maximum int) int {
	m := frame % (2 * maximum)
	if m < maximum {
		return m
	}
	return 2*maximum - 1 - m
}

func (c *Config) breathColor() tcolor.Color {
	spread := 100
	alpha := 0.15 + 0.85*float64(bounce(c.frame, spread))/float64(spread)
	return c.blendingFunction(c.ap.Background, c.bcolor, alpha).Color()
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
	if c.colorDisc != (tcolor.RGBColor{}) {
		// even radius is more symmetric
		mult := c.radius
		if c.breath {
			mult *= (1 + float64(bounce(c.frame/7, 10))/15.)
		}
		radius := 2 * int(math.Round(mult*float64(width)/4.))
		if radius <= height { // so something is visible
			radius = (2 * (height + 1)) / 2
		}
		c.ap.DiscBlendFN(x-width/2-1, y-height/2-1, radius, c.ap.Background, c.colorDisc, c.aliasing, c.blendingFunction)
	}
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
		prefix = c.colorOutput.Foreground(c.breathColor())
	}
	if c.inverse {
		prefix = ansipixels.Inverse + c.color
	}
	suffix := ""
	if c.fillBlack {
		prefix += c.blackBG
	} else {
		suffix = ansipixels.Reset
	}
	for i, line := range lines {
		c.ap.WriteAtStr(x-width, y-height+i, prefix+line+suffix)
	}
	// ap.MoveCursor(x-1, y-1)
}

func main() {
	os.Exit(Main())
}

func (c *Config) ClearScreen() {
	if c.fillBlack {
		c.ap.WriteString(c.blackBG)
	}
	c.ap.ClearScreen()
}

func RGBColor(color tcolor.Color) tcolor.RGBColor {
	t, v := color.Decode()
	if t == tcolor.ColorTypeBasic || t == tcolor.ColorType256 {
		return tcolor.RGBColor{R: 255, G: 20, B: 30}
	}
	return tcolor.ToRGB(t, v)
}

func Main() int { //nolint:funlen,gocognit,gocyclo // we could split the flags and rest.
	truecolorDefault := ansipixels.DetectColorMode().TrueColor
	discDefault := "E0C020"
	if !truecolorDefault {
		discDefault = "FFFFFF"
	}
	cli.MinArgs = 0
	cli.MaxArgs = 1
	cli.ArgsHelp = " [digits:digits...]\npass only flags will display current time; move mouse and click to place on screen"
	fBounce := flag.Int("bounce", 0, "Bounce speed (0 is no bounce and normal mouse mode); 1 is fastest, 2 is slower, etc.")
	f24 := flag.Bool("24", false, "Use 24-hour time format")
	fNoSeconds := flag.Bool("no-seconds", false, "Don't show seconds")
	fNoBlink := flag.Bool("no-blink", false, "Don't blink the colon")
	fBox := flag.Bool("box", false, "Draw a simple rounded corner outline around the time")
	fColorDisc := flag.String("color-disc", discDefault, "Color disc around the time, use \"\" to remove")
	fRadius := flag.Float64("radius", 1.2, "Radius of the disc around the time in proportion of the time width")
	fFillBlack := flag.Bool("black-bg", false, "Set a black background instead of using the terminal's background")
	fAliasing := flag.Float64("aliasing", 0.8, "Aliasing factor for the disc drawing (0.0 sharpest edge to 1.0 sphere effect)")
	fColorBox := flag.String("color-box", "", "Color box around the time")
	fColor := flag.String("color", "red",
		"Color to use: RRGGBB, hue,sat,lum ([0,1]) or one of: "+tcolor.ColorHelp)
	fBreath := flag.Bool("breath", false, "Pulse the color (only works for RGB)")
	fInverse := flag.Bool("inverse", false, "Inverse the foreground and background")
	fDebug := flag.Bool("debug", false, "Debug mode, display mouse position and screen borders")
	fTrueColor := flag.Bool("truecolor", truecolorDefault,
		"Use true color (24-bit RGB) instead of 8-bit ANSI colors (default is true if COLORTERM is set)")
	fLinearBlending := flag.Bool("linear", false, "Use linear blending for the color disc (more sphere like)")
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
		radius:      *fRadius,
		fillBlack:   *fFillBlack,
		aliasing:    *fAliasing,
	}
	if *fLinearBlending {
		cfg.blendingFunction = ansipixels.BlendLinear
	} else {
		cfg.blendingFunction = ansipixels.BlendSRGB
	}
	if cfg.ap.TrueColor {
		cfg.blackBG = tcolor.RGBColor{}.Background()
	} else {
		cfg.blackBG = tcolor.Black.Background()
	}
	if cfg.breath {
		color, _ := tcolor.FromString(*fColor)
		cfg.bcolor = RGBColor(color)
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
	if *fColorDisc != "" {
		color, err := tcolor.FromString(*fColorDisc)
		if err != nil {
			return log.FErrf("Color disc error: %v", err)
		}
		cfg.colorDisc = RGBColor(color)
	}
	ap.HideCursor()
	cfg.ClearScreen()
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
		cfg.ClearScreen()
		cfg.DrawAt(-1, -1, TimeString(prev, false))
		return nil
	}
	blinkEnabled := !*fNoBlink
	blink := false
	// TODO: how to get initial mouse position?
	x, y := ap.Mx, ap.My
	frame := 0
	if *fFillBlack {
		ap.Background = tcolor.RGBColor{}
	} else {
		ap.SyncBackgroundColor()
	}
	for {
		_, err := ap.ReadOrResizeOrSignalOnce()
		if err != nil {
			return 1
		}
		if len(ap.Data) > 0 && (ap.Data[0] == 'q' || ap.Data[0] == 3) {
			return 0 // exit on 'q' or Ctrl-C
		}
		// Click to place the time at the mouse position (or switch back to move with mouse).
		if ap.LeftClick() && ap.MouseRelease() {
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
			cfg.ClearScreen()
			// -1 to switch to ansipixels 0,0 origin (from 1,1 terminal origin)
			// also means 0,0 is now -1,-1 and will center the time until the mouse is moved.
			cfg.DrawAt(x-1, y-1, TimeString(numStr, blink))
			ap.EndSyncMode()
		}
	}
}
