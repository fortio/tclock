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
Change the color, draw box around, etc.. with flags.

```sh
tclock help
```
```
flags:
  -24
        Use 24-hour time format
  -box
        Draw a simple box around the time
  -color string
        Color to use RRGGBB or one of: red, brightred, green, blue, yellow, cyan, white, black (default "red")
  -color-box string
        RGB color box around the time
  -inverse
        Inverse the foreground and background
  -no-blink
        Don't blink the colon
  -no-seconds
        Don't show seconds
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
