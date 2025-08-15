package main

import (
	"math"
	"strings"

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
	r := float64(radius * radius)
	fx := float64(abs(x)) + 0.5
	fy := float64(abs(y)) + 0.5
	d := fx*fx + fy*fy
	if d > r {
		return 0
	}
	edgeDistance := math.Sqrt(r - d)
	if edgeDistance > 0.8*float64(radius) {
		return 1 // full intensity
	}
	return edgeDistance / float64(radius) / 0.8
}

func HSLColor(color tcolor.Color) tcolor.HSLColor {
	t, v := color.Decode()
	if t == tcolor.ColorTypeBasic {
		return tcolor.RGBColor{R: 255, G: 20, B: 30}.HSL()
	}
	return tcolor.ToHSL(t, v)
}

func DrawDisc(ap *ansipixels.AnsiPixels, x, y, radius int, hsl tcolor.HSLColor, fillBlack bool) {
	if fillBlack {
		// black background on all lines
		for j := range ap.H {
			ap.MoveCursor(0, j)
			ap.WriteString(tcolor.Black.Background())
			ap.WriteString(strings.Repeat(" ", ap.W))
		}
	}
	tcolOut := tcolor.ColorOutput{TrueColor: ap.TrueColor}
	for j := -radius; j <= radius; j += 2 {
		first := true
		inside := false
		for i := -radius; i <= radius; i++ {
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
			if first {
				ap.MoveCursor(xx, yy)
				first = false
			}
			if intTop == 1 && intBottom == 1 {
				if !inside {
					ap.WriteString(tcolOut.Foreground(hsl.Color()))
					inside = true
				}
				ap.WriteRune(ansipixels.FullPixel)
				continue
			}
			newTopL := float64(hsl.L) * intTop
			newBottomL := float64(hsl.L) * intBottom
			ncTop := hsl
			ncTop.L = tcolor.Uint10(math.Round(newTopL))
			ncBottom := hsl
			ncBottom.L = tcolor.Uint10(math.Round(newBottomL))
			ap.WriteString(tcolOut.Background(ncTop.Color()))
			ap.WriteString(tcolOut.Foreground(ncBottom.Color()))
			ap.WriteRune(ansipixels.BottomHalfPixel)
		}
	}
	ap.WriteString(tcolor.Reset)
}
