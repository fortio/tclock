package main

import (
	"math"
	"time"

	"fortio.org/sets"
	"fortio.org/terminal/ansipixels"
	"fortio.org/terminal/ansipixels/tcolor"
)

func drawLine(ap *ansipixels.AnsiPixels, sx, sy float64, cx, cy, radius int, color, background tcolor.RGBColor) {
	x0, y0 := cx, cy
	x1 := x0 + int(sx*float64(radius)*2-1)
	y1 := y0 + int(sy*float64(radius)-1)
	pix := sets.New[[2]int]()
	x0i, y0i := x0, y0
	x1i, y1i := x1, y1

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
			pix.Add([2]int{y, x})
		} else {
			pix.Add([2]int{x, y})
		}
		err -= dy
		if err < 0 {
			y += yStep
			err += float64(dx)
		}
	}

	drawPixels(ap, pix, color, background)
}

func drawPixels(ap *ansipixels.AnsiPixels, pixels sets.Set[[2]int], color, background tcolor.RGBColor) {
	for coordAry := range pixels {
		x, y := coordAry[0], coordAry[1]
		switch y % 2 {
		case 0:

			lower := [2]int{x, y + 1}
			if pixels.Has(lower) {
				ap.WriteAt(x, y/2, "%s%s%c", color.Foreground(), background.Background(), ansipixels.FullPixel)
			} else {
				ap.WriteAt(x, y/2, "%s%s%c", color.Foreground(), background.Background(), ansipixels.TopHalfPixel)

			}
		case 1:
			upper := [2]int{x, y - 1}
			if !pixels.Has(upper) {
				ap.WriteAt(x, y/2, "%s%s%c", color.Foreground(), background.Background(), ansipixels.BottomHalfPixel)
			}
		}
	}

}

func rotateFrom12(theta, radius float64) (float64, float64) {
	return -math.Sin(theta) * radius, -math.Cos(theta) * radius
}

func calculateAngle(maxV, timeValue int) float64 {
	return 2. * math.Pi * (float64(maxV) - float64(timeValue)) / float64(maxV)
}

func angleCoords(maxV, timeValue int, radius float64) (float64, float64) {
	return rotateFrom12(calculateAngle(maxV, timeValue), radius)
}

func (c *Config) DrawClock(radius, cx, cy int, background tcolor.RGBColor, now time.Time) {
	sec, minute, hour := now.Second(), now.Minute(), now.Hour()
	sx, sy := angleCoords(60, sec, .8)
	mx, my := angleCoords(60, minute, .7)
	hx, hy := angleCoords(12, hour%12, .4)
	cy *= 2
	drawLine(c.ap, sx, sy*2, cx, cy, radius/2, tcolor.RGBColor{R: 0, G: 0, B: 255}, background)
	drawLine(c.ap, mx, my*2, cx, cy, radius/2, tcolor.RGBColor{R: 0, G: 255, B: 0}, background)
	drawLine(c.ap, hx, hy*2, cx, cy, radius/2, tcolor.RGBColor{R: 255, G: 0, B: 0}, background)
}
