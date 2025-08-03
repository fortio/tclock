[![GoDoc](https://godoc.org/fortio.org/tclock/bignum?status.svg)](https://pkg.go.dev/fortio.org/tclock/bignum)
[![Go Report Card](https://goreportcard.com/badge/fortio.org/tclock)](https://goreportcard.com/report/fortio.org/tclock)
[![CI Checks](https://github.com/fortio/tclock/actions/workflows/include.yml/badge.svg)](https://github.com/fortio/tclock/actions/workflows/include.yml)
# tclock
Terminal Clock using Ansipixels library

`tclock` is a simple terminal/TUI clock (well for now just big numbers printer).

## Install
You can get the binary from [releases](https://github.com/fortio/tclock/releases)

Or just run
```
CGO_ENABLED=0 go install fortio.org/tclock@latest  # to install (in ~/go/bin typically) or just
CGO_ENABLED=0 go run fortio.org/tclock@latest  # to run without install
```

or even
```
docker run -ti fortio/tclock # but that's obviously slower
```

or
```
brew install fortio/tap/tclock
```

## Run

Move the mouse to place the clock, click to leave it there, click again to put it somewhere else.
Change the color (can be specified by name, RRGGBB or Hue,Sat,Luminance in [0,1] range), draw boxes around, etc.. with flags.

```sh
tclock help
```
```
flags:
  -24
        Use 24-hour time format
  -bounce int
        Bounce speed (0 is no bounce and normal mouse mode); 1 is fastest, 2 is slower, etc.
  -box
        Draw a simple rounded corner outline around the time
  -breath
        Pulse the color (only works for RGB)
  -color string
        Color to use: RRGGBB, hue,sat,lum ([0,1]) or one of: none, black, red, green,
yellow, blue, purple, cyan, gray, darkgray, brightred, brightgreen, brightyellow,
brightblue, brightpurple, brightcyan, white (default "red")
  -color-box string
        Color box around the time
  -debug
        Debug mode, display mouse position and screen borders
  -inverse
        Inverse the foreground and background
  -no-blink
        Don't blink the colon
  -no-seconds
        Don't show seconds
  -true-color
        Use true color (24-bit RGB) instead of 8-bit ANSI colors (default is true if
COLORTERM is set)
```

```sh
$ tclock
```

```
         ╭────────────────────────────────────╮
         │      ━━      ━━   ━━      ━━   ━━  │
         │   ┃ ┃  ┃       ┃    ┃    ┃  ┃    ┃ │
         │          ::  ━━   ━━  ::           │
         │   ┃ ┃  ┃    ┃       ┃    ┃  ┃    ┃ │
         │      ━━      ━━   ━━      ━━       │
         ╰────────────────────────────────────╯

```
