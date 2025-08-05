package main

import (
	"math"

	"fortio.org/terminal/ansipixels"
	"fortio.org/terminal/ansipixels/tcolor"
)

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func intensity(x, y, radius int) float64 {
	x = abs(x)
	y = abs(y)
	r := radius * radius
	d := x*x + y*y
	if d > r {
		return 0
	}
	if (x-1)*(x-1)+(y-1)*(y-1) < r {
		return 1
	}
	return (float64(radius) - math.Sqrt(float64(d))) / float64(2*x+2*y+2)
}

func DrawDisc(ap *ansipixels.AnsiPixels, x, y, radius int, color string) {
	ap.WriteString(color)
	for i := -radius; i <= radius; i++ {
		for j := -radius; j <= radius; j += 2 {
			xx := x + i
			yy := y + j/2
			if xx < 0 || yy < 0 || xx >= ap.W || yy >= ap.H {
				continue // skip out of bounds
			}
			intTop := intensity(i, j, radius)
			intBottom := intensity(i, j+1, radius)
			if intTop == 0 && intBottom == 0 {
				continue // skip if not in the disc
			}
			ap.MoveCursor(xx, yy)
			switch {
			case intTop == 1 && intBottom == 1:
				ap.WriteRune(ansipixels.FullPixel)
			case intTop > intBottom:
				ap.WriteRune(ansipixels.TopHalfPixel)
			default: // bottom
				ap.WriteRune(ansipixels.BottomHalfPixel)
			}
		}
	}
	ap.WriteString(tcolor.Reset)
}
