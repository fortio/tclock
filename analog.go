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

func calculateAngle(maxV, timeValue int) float64 {
	return 2. * math.Pi * (float64(maxV) - float64(timeValue)) / float64(maxV)
}

func angleCoords(maxV, timeValue int, radius float64) (int, int) {
	return rotateFrom12(calculateAngle(maxV, timeValue), radius)
}

func (c *Config) DrawHands(cx, cy, radius int, background tcolor.RGBColor, now time.Time) {
	sec, minute, hour := now.Second(), now.Minute(), now.Hour()
	r := float64(radius)
	sx, sy := angleCoords(60, sec, .8*r)
	mx, my := angleCoords(60, minute, .7*r)
	hx, hy := angleCoords(12, hour%12, .4*r)
	drawLine(c.ap, sx, sy, cx, cy, tcolor.RGBColor{R: 0, G: 0, B: 255}, background)
	drawLine(c.ap, mx, my, cx, cy, tcolor.RGBColor{R: 0, G: 255, B: 0}, background)
	drawLine(c.ap, hx, hy, cx, cy, tcolor.RGBColor{R: 255, G: 0, B: 0}, background)
}
