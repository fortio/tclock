// Image based variant of the analog clock. With antialiasing.

package main

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"time"

	"fortio.org/terminal/ansipixels"
	"fortio.org/terminal/ansipixels/tcolor"
)

func point(a, r float64) (float64, float64) {
	return -r * math.Sin(a), -r * math.Cos(a)
}

func angle(maxV, timeValue float64) float64 {
	return 2. * math.Pi * (maxV - timeValue) / maxV
}

func coords(maxV, timeValue float64, radius float64) (float64, float64) {
	return point(angle(maxV, timeValue), radius)
}

func (c *Config) DrawImage(now time.Time, seconds bool) {
	r := min(float64(c.ap.W)/2, float64(c.ap.H)) - 1
	cxf := float64(c.ap.W) / 2
	cyf := float64(c.ap.H)
	cx := int(cxf)
	cy := int(cyf / 2)
	// new NRGBA image of the right size
	img := image.NewNRGBA(image.Rect(0, 0, c.ap.W, 2*c.ap.H))
	sec, minute, hour := float64(now.Second()), float64(now.Minute()), now.Hour()
	sx, sy := coords(60, sec, .94*r)
	m := minute + sec/60.
	mx, my := coords(60, m, .80*r)
	hx, hy := coords(12, float64(hour%12)+m/60., .47*r)
	minDotColor := color.NRGBA{R: 255, G: 255, B: 255, A: 200}
	hourDotColor := color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	if seconds {
		// Minutes/Seconds markers:
		for n := range 60 {
			color := minDotColor
			if n%5 == 0 {
				color = hourDotColor
			}
			nx1, ny1 := coords(60, float64(n), r-0.5)
			nx2, ny2 := coords(60, float64(n), r+0.5)
			ansipixels.DrawAALine(img, cxf+nx1, cyf+ny1, cxf+nx2, cyf+ny2, color)
		}
		ansipixels.DrawAALine(img, cxf, cyf, cxf+sx, cyf+sy, color.NRGBA{R: 0x50, G: 0x80, B: 0x50, A: 200})
	}
	ansipixels.DrawAALine(img, cxf, cyf, cxf+mx, cyf+my, color.NRGBA{R: 0x2C, G: 0x59, B: 0xD4, A: 200}) // #2C59D4
	ansipixels.DrawAALine(img, cxf, cyf, cxf+hx, cyf+hy, color.NRGBA{R: 255, G: 0xA7, B: 10, A: 200})

	// back to RGBA for drawing
	dst := image.NewRGBA(img.Bounds())
	draw.Draw(dst, img.Bounds(), img, image.Point{}, draw.Src)
	// draw the image at the center of the available space
	_ = c.ap.ShowScaledImage(dst)
	if !seconds {
		// Numbers for hours:
		c.ap.WriteString(tcolor.Reset)
		for n := 5; n <= 60; n += 5 {
			nx, ny := angleCoords(60, float64(n%60), r)
			m := n / 5
			if m >= 10 {
				nx--
			}
			c.ap.WriteAt(cx+nx, cy+(ny-1)/2, "%d", m)
		}
	}
}
