package main

import (
	"math"
	"time"

	"fortio.org/terminal/ansipixels"
	"fortio.org/terminal/ansipixels/tcolor"
)

type Point [2]int

type Pixels map[Point]tcolor.RGBColor

// Same as ansipixels.DrawLine but without the image and collecting pixels instead
// (and handling the 1/2 height of terminal characters).
func drawLine(pix Pixels, sx, sy, x0i, y0i int, color tcolor.RGBColor) {
	x1i := x0i + sx
	y0i *= 2
	y1i := y0i + sy

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
			pix[Point{y, x}] = color
		} else {
			pix[Point{x, y}] = color
		}
		err -= dy
		if err < 0 {
			y += yStep
			err += float64(dx)
		}
	}
}

func drawPixels(ap *ansipixels.AnsiPixels, pixels Pixels, background tcolor.RGBColor) {
	for coordAry, color := range pixels {
		x, y := coordAry[0], coordAry[1]
		switch y % 2 {
		case 0:
			ap.MoveCursor(x, y/2)
			lower := Point{x, y + 1}
			if v, ok := pixels[lower]; ok {
				if v == color {
					ap.WriteString(color.Foreground())
					ap.WriteRune(ansipixels.FullPixel)
					continue
				}
				ap.WriteString(v.Foreground())
				ap.WriteString(color.Background())
				delete(pixels, lower) // drawn together
			} else {
				ap.WriteString(color.Background())
				ap.WriteString(background.Foreground())
			}
			ap.WriteRune(ansipixels.BottomHalfPixel)
		case 1:
			upper := Point{x, y - 1}
			if _, ok := pixels[upper]; !ok {
				ap.MoveCursor(x, y/2)
				ap.WriteString(background.Background())
				ap.WriteString(color.Foreground())
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
	sx, sy := angleCoords(60, sec, .9*r)
	mx, my := angleCoords(60, minute+sec/60., .80*r)
	hx, hy := angleCoords(12, float64(hour%12)+minute/60., .47*r)
	pix := make(Pixels)
	drawLine(pix, sx, sy, cx, cy, tcolor.RGBColor{R: 0x50, G: 0x80, B: 0x50})
	drawLine(pix, mx, my, cx, cy, tcolor.RGBColor{R: 0x2C, G: 0x59, B: 0xD4}) // #2C59D4
	drawLine(pix, hx, hy, cx, cy, tcolor.RGBColor{R: 255, G: 0xA7, B: 10})
	drawPixels(c.ap, pix, background)
	c.ap.WriteString(tcolor.Reset)
	for n := 1; n <= 60; n++ {
		nx, ny := angleCoords(60, float64(n%60), r)
		if n%5 == 0 {
			m := n / 5
			if m >= 10 {
				nx--
			}
			c.ap.WriteAt(cx+nx, cy+(ny-1)/2, "%d", m)
		} else {
			c.ap.WriteAt(cx+nx, cy+(ny-1)/2, "•") // middle dot.
		}
	}
}
