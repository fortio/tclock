package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fortio.org/cli"
	"fortio.org/tclock/bignum"
)

func ShowTime(numStr string) {
	d := &bignum.Display{}
	for _, c := range numStr {
		d.PlaceDigit(c)
	}
	fmt.Print(d.String())
}

func main() {
	cli.MinArgs = 0
	cli.MaxArgs = 1
	cli.ArgsHelp = " [digits:digits...] or current time"
	f24 := flag.Bool("24", false, "Use 24-hour time format")
	fSeconds := flag.Bool("seconds", false, "Show seconds (default is minutes only)")
	cli.Main()
	var numStr string
	if flag.NArg() == 1 {
		numStr = flag.Arg(0)
		ShowTime(numStr)
		fmt.Println() // Ensure newline after the number
		return
	}
	format := "3:04"
	if *f24 {
		format = "15:04"
	}
	if *fSeconds {
		format += ":05"
	}
	prev := ""
	tick := time.NewTicker(time.Second) // To avoid busy waiting
	defer func() {
		tick.Stop()
		fmt.Printf("\r\x1b[%dB\n", bignum.Height/2+1)
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	for {
		numStr = time.Now().Format(format)
		if numStr != prev {
			if prev != "" {
				fmt.Printf("\x1b[%dA\r", bignum.Height/2) // Move cursor up to overwrite previous output
			}
			ShowTime(numStr)
			fmt.Printf("\x1b[%dA\x1b[%dD", bignum.Height/2, bignum.Width*3)
		}
		prev = numStr
		select {
		case <-sig:
			return
		case <-tick.C:
		}
	}

}
