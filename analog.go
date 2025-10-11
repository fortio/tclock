package main

import (
	"math"

	"fortio.org/terminal/ansipixels"
	"fortio.org/terminal/ansipixels/tcolor"
)

func drawLineHigh(ap *ansipixels.AnsiPixels, x1, y1, x2, y2 int) map[[2]int]struct{} {
	//bresenham's algorithm
	ap.StartSyncMode()
	defer ap.EndSyncMode()
	dx, dy := x2-x1, y2-y1
	x, y := x1, y1
	dy1 := 1
	dx1 := 1
	if y2 < y1 {
		dy1 = -1
	}
	if x2 < x1 {
		dx1 = -1
		dx = -dx
	}
	pixels := make(map[[2]int]struct{})

	p := 2*dx - dy
	for y = y1; y != y2; y += dy1 {
		if p > 0 {
			x += dx1
			p += 2 * (dx - dy)
			// ap.WriteAt(x, y, "%s ", color.Background())
			pixels[[2]int{x, y}] = struct{}{}
			continue
		}
		p += 2 * dx
		// ap.WriteAt(x, y, "%s ", color.Background())
		pixels[[2]int{x, y}] = struct{}{}
	}
	return pixels
}
func drawLineLow(ap *ansipixels.AnsiPixels, x1, y1, x2, y2 int) map[[2]int]struct{} {
	//bresenham's algorithm
	ap.StartSyncMode()
	defer ap.EndSyncMode()
	pixels := make(map[[2]int]struct{})
	dx, dy := x2-x1, y2-y1
	y := y1
	dy1 := 1
	dx1 := 1
	if y2 < y1 {
		dy1 = -1
		dy = -dy
	}
	if x2 < x1 {
		dx1 = -1
		dx = -dx
	}

	p := 2*dy - dx
	for x := x1; x != x2; x += dx1 {
		if p > 0 {
			y += dy1
			p += 2 * (dy - dx)
			// ap.WriteAt(x, y, "%s ", color.Background())
			pixels[[2]int{x, y}] = struct{}{}
			continue
		}
		p += 2 * dy
		// ap.WriteAt(x, y, "%s ", color.Background())
		pixels[[2]int{x, y}] = struct{}{}
	}
	return pixels
}

func drawLine(ap *ansipixels.AnsiPixels, sx, sy float64, cx, cy, radius int, color tcolor.Color) {

	x0, y0 := cx, cy
	x1 := x0 + int(sx*float64(radius)*2-1)
	y1 := y0 + int(sy*float64(radius)-1)
	var pix map[[2]int]struct{}
	if max(y0-y1, y1-y0) < max(x0-x1, x1-x0) {
		if x1 > x0 {
			pix = drawLineLow(ap, x0, y0, x1, y1)
		} else {
			pix = drawLineLow(ap, x1, y1, x0, y0)
		}
	} else {
		if y1 > y0 {
			pix = drawLineHigh(ap, x0, y0, x1, y1)
		} else {
			pix = drawLineHigh(ap, x1, y1, x0, y0)
		}
	}
	drawPixels(ap, pix, color)
}

func drawPixels(ap *ansipixels.AnsiPixels, pixels map[[2]int]struct{}, color tcolor.Color) {
	for coordAry := range pixels {
		x, y := coordAry[0], coordAry[1]
		switch y % 2 {
		case 0:

			lower := [2]int{x, y + 1}
			_, ok := pixels[lower]
			if ok {
				ap.WriteAt(x, y/2, "%s%s%c", color.Foreground(), tcolor.RGBColor{R: 255, G: 255, B: 255}.Background(), ansipixels.FullPixel)
			} else {
				ap.WriteAt(x, y/2, "%s%s%c", color.Foreground(), tcolor.RGBColor{R: 255, G: 255, B: 255}.Background(), ansipixels.TopHalfPixel)

			}
		case 1:
			upper := [2]int{x, y - 1}
			_, ok := pixels[upper]
			if !ok {
				ap.WriteAt(x, y/2, "%s%s%c", color.Foreground(), tcolor.RGBColor{R: 255, G: 255, B: 255}.Background(), ansipixels.BottomHalfPixel)
			}
		}
	}

}

func rotateFrom12(theta float64) (float64, float64) {
	return -math.Sin(theta), -math.Cos(theta)
}

func unitCircleCoordsForClockHands(seconds, minutes, hours int) (float64, float64, float64, float64, float64, float64) {
	secondsAngleFrom12 := 2. * math.Pi * (60. - float64(seconds)) / 60.
	minutesAngleFrom12 := 2. * math.Pi * (60. - float64(minutes)) / 60.
	hoursAngleFrom12 := 2. * math.Pi * (12. - float64(hours)) / 12.
	x1, y1 := rotateFrom12(secondsAngleFrom12)
	x2, y2 := rotateFrom12(minutesAngleFrom12)
	x3, y3 := rotateFrom12(hoursAngleFrom12)
	return x1, y1, x2, y2, x3, y3
}
