package main

import (
	"flag"
	"fmt"

	"fortio.org/cli"
	"fortio.org/tclock/bignum"
)

func main() {
	cli.MinArgs = 1
	cli.MaxArgs = 1
	cli.ArgsHelp = " number"
	cli.Main()
	numStr := flag.Arg(0)
	d := &bignum.Display{}
	for _, c := range numStr {
		d.PlaceDigit(c)
	}
	fmt.Println(d.String())
}
