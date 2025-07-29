// Package bignum implements a display for large numbers using 7 segments style
// unicode for terminal output. In the style of early digital clocks.
package bignum

import (
	"strings"
	"unicode/utf8"
)

const (
	Numbers = `
 ━━
┃  ┃

┃  ┃
 ━━


   ┃

   ┃


 ━━
   ┃
 ━━
┃
 ━━

 ━━
   ┃
 ━━
   ┃
 ━━


┃  ┃
 ━━
   ┃


 ━━
┃
 ━━
   ┃
 ━━

 ━━
┃
 ━━
┃  ┃
 ━━

 ━━
   ┃

   ┃


 ━━
┃  ┃
 ━━
┃  ┃
 ━━

 ━━
┃  ┃
 ━━
   ┃
 ━━



::


`
	Height = 5
	Width  = 4
)

// Line based version of Numbers with padding to fixed width.
var NumberLines []string

func AddTrailingSpaces(s string, extra int) string {
	s += strings.Repeat(" ", Width+extra-utf8.RuneCountInString(s))
	return s
}

func init() {
	NumberLines = strings.Split(Numbers, "\n")[1:]
	for i := range NumberLines {
		extra := 1
		if i >= 10*(Height+1) {
			extra = -1 // no trailing space for colon
		}
		NumberLines[i] = AddTrailingSpaces(NumberLines[i], extra)
	}
}

type Display struct {
	lines [Height]string
	col   int
}

func (d *Display) String() string {
	return strings.Join(d.lines[:], "\n")
}

func (d *Display) PlaceDigit(r rune) {
	digit := int(r - '0')
	if digit < 0 || digit > 9 {
		digit = 10 // treat as colon
	}
	start := digit * (Height + 1)
	for i := range Height {
		d.lines[i] += NumberLines[start+i]
	}
	d.col++
}
