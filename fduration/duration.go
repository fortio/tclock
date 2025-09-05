// Package fduration allows duration parsing with "d" for days (24 hours).
package fduration

import (
	"errors"
	"flag"
	"strconv"
	"time"
	"unicode"
)

func ParseDuration(s string) (time.Duration, error) {
	// copy of time.ParseDuration, with added "d"
	// grammar: [0-9]+(ns|us|µs|ms|s|m|h|d)
	orig := s
	var d time.Duration
	for s != "" {
		// consume leading spaces
		for len(s) > 0 && unicode.IsSpace(rune(s[0])) {
			s = s[1:]
		}
		if s == "" {
			break
		}
		// find number part
		i := 0
		for i < len(s) && (('0' <= s[i] && s[i] <= '9') || s[i] == '.') {
			i++
		}
		if i == 0 {
			return 0, errors.New("time: invalid duration " + orig)
		}
		num := s[:i]
		s = s[i:]

		// find unit
		j := 0
		for j < len(s) && unicode.IsLetter(rune(s[j])) {
			j++
		}
		if j == 0 {
			return 0, errors.New("time: missing unit in duration " + orig)
		}
		unit := s[:j]
		s = s[j:]

		// parse number
		v, err := strconv.ParseFloat(num, 64)
		if err != nil {
			return 0, err
		}

		var mult time.Duration
		switch unit {
		case "ns":
			mult = time.Nanosecond
		case "us", "µs":
			mult = time.Microsecond
		case "ms":
			mult = time.Millisecond
		case "s":
			mult = time.Second
		case "m":
			mult = time.Minute
		case "h":
			mult = time.Hour
		case "d":
			mult = 24 * time.Hour
		default:
			return 0, errors.New("time: unknown unit " + unit + " in duration " + orig)
		}
		d += time.Duration(v * float64(mult))
	}
	return d, nil
}

type Duration time.Duration

func (d *Duration) String() string {
	return time.Duration(*d).String()
}

func (d *Duration) Set(s string) error {
	dd, err := ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(dd)
	return nil
}

func DurationFlag(name string, value time.Duration, usage string) *time.Duration {
	d := Duration(value)
	flag.Var(&d, name, usage)
	return (*time.Duration)(&d)
}
