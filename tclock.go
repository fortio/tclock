package main

import (
	"flag"
	"fmt"
	"strings"
	"unicode/utf8"

	"fortio.org/cli"
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

var (
	NumberLines []string
)

func AddTrailingSpaces(s string) string {
	s += strings.Repeat(" ", Width+1-utf8.RuneCountInString(s))
	return s
}

func InitNumberLines() {
	NumberLines = strings.Split(Numbers, "\n")[1:]
	for i := range NumberLines {
		NumberLines[i] = AddTrailingSpaces(NumberLines[i])
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

func main() {
	cli.MinArgs = 1
	cli.MaxArgs = 1
	cli.ArgsHelp = " number"
	cli.Main()
	InitNumberLines()
	numStr := flag.Arg(0)
	d := &Display{}
	for _, c := range numStr {
		d.PlaceDigit(c)
	}
	fmt.Println(d.String())
}
