// Package duration allows duration parsing with "d" for days (24 hours) and "w" for week (7 days).
package duration

import (
	"errors"
	"flag"
	"strconv"
	"time"
	"unicode"
)

const (
	Day  = 24 * time.Hour
	Week = 7 * Day
)

// ParseDuration parses a duration string with "d" for days (24 hours)
// and "w" for weeks (7 days) in addition to what stdlib time.ParseDuration supports.
func ParseDuration(s string) (time.Duration, error) {
	orig := s
	var d time.Duration
	neg := false
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
		for i < len(s) && (('0' <= s[i] && s[i] <= '9') || s[i] == '.' || s[i] == '-') {
			i++
		}
		if i == 0 {
			return 0, errors.New("invalid duration " + orig)
		}
		num := s[:i]
		s = s[i:]

		// find unit
		j := 0
		for j < len(s) && unicode.IsLetter(rune(s[j])) {
			j++
		}
		if j == 0 {
			return 0, errors.New("missing unit in duration " + orig)
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
			mult = Day
		case "w":
			mult = Week
		default:
			return 0, errors.New("unknown unit " + unit + " in duration " + orig)
		}
		if v < 0 {
			if neg || d != 0 {
				return 0, errors.New("unexpected negative sign in middle of duration " + orig)
			}
			neg = true
			v = -v
		}
		d += time.Duration(v * float64(mult))
	}
	if neg {
		d = -d
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

// Flag defines a duration flag with the specified name, default value, and usage string, like
// [flag.Duration] but supporting durations in days (24 hours) in addition to the other stdlib units.
func Flag(name string, value time.Duration, usage string) *time.Duration {
	d := Duration(value)
	flag.Var(&d, name, usage)
	return (*time.Duration)(&d)
}
