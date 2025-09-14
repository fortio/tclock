package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"fortio.org/cli"
	"fortio.org/duration"
	"fortio.org/log"
	"fortio.org/tclock/bignum"
	"fortio.org/terminal"
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
	bounceSpeed int             // frame skip for bouncing, 0 means no bouncing
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
	// Extra text (countdown)
	text string
	// In tail mode we stick the clock at the top right of the screen.
	topRight bool
	tail     io.Reader
	// countdown mode
	countDown          bool
	end                time.Time
	extraNewLinesAtEnd bool
	// time format
	format string
	// Mouse tracking flip flop (on click toggle or just off when using bounce or tail mode)
	trackMouse bool
	// Blinking of the second
	blinkEnabled bool
	// Show seconds
	seconds bool
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
	if c.topRight {
		x = c.ap.W - 1
		y = height - 1
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
	if c.text != "" {
		center := x - width/2 - c.ap.ScreenWidth(c.text)/2 - 1
		c.ap.WriteAtStr(center, y+1, c.text)
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

func DurationString(duration time.Duration, withSeconds bool) string {
	str := DurationDDHHMM(duration)
	if withSeconds {
		str += fmt.Sprintf(":%02d", int(duration.Seconds())%60)
	}
	return str
}

func DurationDDHHMM(duration time.Duration) string {
	minutes := int(duration.Minutes()) % 60
	hours := int(duration.Hours()) % 24
	if duration >= 24*time.Hour {
		days := int(duration.Hours()) / 24
		return fmt.Sprintf("%02d:%02d:%02d", days, hours, minutes)
	}
	if duration >= time.Hour {
		return fmt.Sprintf("%02d:%02d", hours, minutes)
	}
	return fmt.Sprintf("%02d", minutes)
}

func (c *Config) Tail() *Config {
	c.topRight = true
	c.colorDisc = tcolor.RGBColor{}
	c.boxed = true
	return c
}

func Main() int { //nolint:funlen,gocognit,gocyclo,maintidx // we could split the flags and rest.
	truecolorDefault := ansipixels.DetectColorMode().TrueColor
	discDefault := "E0C020"
	if !truecolorDefault {
		discDefault = "FFFFFF"
	}
	cli.MinArgs = 0
	cli.MaxArgs = 1
	cli.ArgsHelp = " [digits:digits... or - for stdin tailing]\n" +
		"pass only flags will display current time; move mouse and click to place on screen"
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
	fCountdown := duration.Flag("countdown", 0, "If > 0, countdown from this `duration` instead of showing the time")
	fText := flag.String("text", "",
		"Text to display below the clock (during countdown will be the target time, use none for no extra text)")
	fUntil := flag.String("until", "",
		"If set, countdown until this `date/time` (\"YYYY-MM-DD HH:MM:SS\" or for instance \"3:05 pm\") instead of showing the time")
	fTail := flag.String("tail", "",
		"Tail the given `filename` while showing the clock, or `-` for stdin")
	cli.Main()
	colorOutput := tcolor.ColorOutput{TrueColor: *fTrueColor}
	format := "3:04"
	if *f24 {
		format = "15:04"
	}
	cfg := &Config{
		boxed:       *fBox,
		inverse:     *fInverse,
		debug:       *fDebug,
		breath:      *fBreath,
		colorOutput: colorOutput,
		radius:      *fRadius,
		fillBlack:   *fFillBlack,
		aliasing:    *fAliasing,
		format:      format,
		seconds:     !*fNoSeconds,
		bounceSpeed: *fBounce,
	}
	if cfg.seconds {
		cfg.format += ":05"
	}
	showText := *fText != "none"
	if showText {
		cfg.text = *fText
	}
	now := time.Now()
	if *fCountdown > 0 {
		cfg.countDown = true
		cfg.end = now.Add(*fCountdown)
	}
	if *fUntil != "" {
		cfg.countDown = true
		var err error
		cfg.end, err = duration.ParseDateTime(now, *fUntil)
		if err != nil {
			return log.FErrf("Invalid until time: %v", err)
		}
	}
	if cfg.countDown && showText && cfg.text == "" {
		toStr := cfg.end.Format(cfg.format)
		if cfg.end.Sub(now) >= 24*time.Hour {
			toStr = fmt.Sprintf("%s %s", cfg.end.Format("2006-01-02"), toStr)
		}
		extra := ""
		if !*f24 && cfg.end.Hour() >= 12 {
			extra = " pm"
		}
		cfg.text = "Countdown to " + toStr + extra
	}
	if *fLinearBlending {
		cfg.blendingFunction = ansipixels.BlendLinear
	} else {
		cfg.blendingFunction = ansipixels.BlendNSRGB
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
	ap := ansipixels.NewAnsiPixels(60)
	cfg.ap = ap

	if flag.NArg() == 1 {
		numStr := flag.Arg(0)
		if numStr == "-" {
			return StdinTail(cfg.Tail())
		}
		if len(numStr) == 0 || numStr[0] < '0' || numStr[0] > '9' {
			cli.ErrUsage("No arguments, or <digits> or -")
			return 1
		}
		fmt.Println(TimeString(numStr, false))
		return 0
	}

	if *fTail != "" {
		cfg.Tail()
		if *fTail == "-" {
			return StdinTail(cfg)
		}
		file, err := os.Open(*fTail)
		if err != nil {
			return log.FErrf("Error opening tail file: %v", err)
		}
		defer file.Close() // pointless in main but makes AI happy.
		cfg.tail = file
		ap.MoveCursor(0, 0)
		ap.SaveCursorPos()
	}
	if err := ap.Open(); err != nil {
		return log.FErrf("Error opening terminal: %v", err)
	}
	cfg.extraNewLinesAtEnd = true
	defer func() {
		if cfg.extraNewLinesAtEnd {
			fmt.Fprintf(ap.Out, "\r\n\n\n\n")
		}
		ap.ShowCursor()
		ap.MouseTrackingOff()
		ap.EndSyncMode()
		ap.Restore()
	}()
	if cfg.ap.TrueColor {
		cfg.blackBG = tcolor.RGBColor{}.Background()
	} else {
		cfg.blackBG = tcolor.Black.Background()
	}
	if !cfg.topRight {
		ap.HideCursor()
	}
	cfg.ClearScreen()
	if (cfg.bounceSpeed <= 0) && !cfg.topRight {
		ap.MouseTrackingOn()
		cfg.trackMouse = true
	}
	_ = ap.GetSize()
	cfg.blinkEnabled = !*fNoBlink
	if *fFillBlack {
		ap.Background = tcolor.RGBColor{}
	} else {
		ap.SyncBackgroundColor()
	}
	return RawModeLoop(now, cfg)
}

//nolint:gocognit // yeah
func RawModeLoop(now time.Time, cfg *Config) int {
	var numStr string
	ap := cfg.ap
	var buf [4096]byte
	writer := terminal.CRLFWriter{Out: ap.Out}
	blink := false
	var prevNow time.Time
	// TODO: how to get initial mouse position?
	x, y := ap.Mx, ap.My
	frame := 0
	prev := ""
	ap.OnResize = func() error {
		cfg.ClearScreen()
		cfg.DrawAt(-1, -1, TimeString(prev, false))
		return nil
	}
	for {
		_, err := ap.ReadOrResizeOrSignalOnce()
		if err != nil {
			return 1
		}
		// Exit on 'q' or Ctrl-C but with status error in countdown mode.
		if len(ap.Data) > 0 && (ap.Data[0] == 'q' || ap.Data[0] == 3) {
			if cfg.countDown {
				ap.WriteAt(0, ap.H-3, "Countdown aborted at %s\r\n", now.Format(cfg.format))
				return 1
			}
			return 0
		}
		// Click to place the time at the mouse position (or switch back to move with mouse).
		if ap.LeftClick() && ap.MouseRelease() {
			cfg.trackMouse = !cfg.trackMouse
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
		switch {
		case (cfg.bounceSpeed > 0):
			if frame%cfg.bounceSpeed == 0 {
				cfg.bounce++
				doDraw = true
			}
			frame++
		case cfg.trackMouse && (ap.Mx != x || ap.My != y):
			x, y = ap.Mx, ap.My
			doDraw = true
		}
		n := 0
		if cfg.tail != nil {
			n, err = cfg.tail.Read(buf[:])
			if err != nil && !errors.Is(err, io.EOF) {
				return log.FErrf("Error reading tail file: %v", err)
			}
		}
		if doDraw || n > 0 {
			cfg.frame++
			ap.StartSyncMode()
			if cfg.tail == nil {
				cfg.ClearScreen()
			}
			if n > 0 {
				_, _ = writer.Write(buf[:n])
				ap.SaveCursorPos()
			}
			// -1 to switch to ansipixels 0,0 origin (from 1,1 terminal origin)
			// also means 0,0 is now -1,-1 and will center the time until the mouse is moved.
			cfg.DrawAt(x-1, y-1, TimeString(numStr, blink))
			ap.RestoreCursorPos()
			ap.EndSyncMode()
		}
	}
}
