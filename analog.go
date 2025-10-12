package main

import (
	"math"
	"time"

	"fortio.org/sets"
	"fortio.org/terminal/ansipixels"
	"fortio.org/terminal/ansipixels/tcolor"
)

type Point [2]int

func drawLine(ap *ansipixels.AnsiPixels, sx, sy, x0i, y0i int, color, background tcolor.RGBColor) {
	x1i := x0i + sx
	y0i *= 2
	y1i := y0i + sy
	pix := sets.New[Point]()

	steep := math.Abs(float64(y1i-y0i)) > math.Abs(float64(x1i-x0i))
	if steep {
		x0i, y0i = y0i, x0i
		x1i, y1i = y1i, x1i
	}

	if x0i > x1i {
		x0i, x1i = x1i, x0i
		y0i, y1i = y1i, y0i
	}

	dx := x1i - x0i
	dy := math.Abs(float64(y1i - y0i))
	err := float64(dx) / 2.0
	yStep := 1
	if y0i > y1i {
		yStep = -1
	}

	y := y0i
	for x := x0i; x <= x1i; x++ {
		if steep {
			pix.Add(Point{y, x})
		} else {
			pix.Add(Point{x, y})
		}
		err -= dy
		if err < 0 {
			y += yStep
			err += float64(dx)
		}
	}
	drawPixels(ap, pix, color, background)
}

func drawPixels(ap *ansipixels.AnsiPixels, pixels sets.Set[Point], color, background tcolor.RGBColor) {
	ap.WriteString(color.Foreground())
	ap.WriteString(background.Background())
	for coordAry := range pixels {
		x, y := coordAry[0], coordAry[1]
		switch y % 2 {
		case 0:
			ap.MoveCursor(x, y/2)
			lower := Point{x, y + 1}
			if pixels.Has(lower) {
				ap.WriteRune(ansipixels.FullPixel)
				pixels.Remove(lower)
			} else {
				ap.WriteRune(ansipixels.TopHalfPixel)
			}
		case 1:
			upper := Point{x, y - 1}
			if !pixels.Has(upper) {
				ap.MoveCursor(x, y/2)
				ap.WriteRune(ansipixels.BottomHalfPixel)
			}
		}
	}
}

func rotateFrom12(theta, radius float64) (int, int) {
	return int(math.Round(-math.Sin(theta) * radius)), int(math.Round(-math.Cos(theta) * radius))
}

func calculateAngle(maxV, timeValue float64) float64 {
	return 2. * math.Pi * (maxV - timeValue) / maxV
}

func angleCoords(maxV, timeValue float64, radius float64) (int, int) {
	return rotateFrom12(calculateAngle(maxV, timeValue), radius)
}

func (c *Config) DrawHands(cx, cy, radius int, background tcolor.RGBColor, now time.Time) {
	sec, minute, hour := float64(now.Second()), float64(now.Minute()), now.Hour()
	r := float64(radius)
	sx, sy := angleCoords(60, sec, .85*r)
	mx, my := angleCoords(60, minute+sec/60., .8*r)
	hx, hy := angleCoords(12, float64(hour%12)+minute/60., .5*r)
	drawLine(c.ap, sx, sy, cx, cy, tcolor.RGBColor{R: 0x65, G: 0xb3, B: 0x37}, background) // #65B337
	drawLine(c.ap, mx, my, cx, cy, tcolor.RGBColor{R: 0x2C, G: 0x59, B: 0xD4}, background) // #2C59D4
	drawLine(c.ap, hx, hy, cx, cy, tcolor.RGBColor{R: 255, G: 0x18, B: 10}, background)
	c.ap.WriteString(tcolor.Reset)
}
