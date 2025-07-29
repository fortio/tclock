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

```sh
tclock help
```
```
flags:
  -24
         Use 24-hour time format
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
