// Package duration allows duration parsing with "d" for days (24 hours) and "w" for week (7 days).
package duration

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"
	"unicode"
)

const (
	Day  = 24 * time.Hour
	Week = 7 * Day
)

// Parse parses a duration string with "d" for days (24 hours)
// and "w" for weeks (7 days) in addition to what stdlib [time.ParseDuration] supports.
func Parse(s string) (time.Duration, error) {
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

//nolint:recvcheck // need pointer receiver obviously for Set and for String avoids pointer.
type Duration time.Duration

// String formats the duration using weeks and days if applicable and omitting all 0 values and trailing zeroes
// for instance "1d3m" instead of "24h3m0s" (stdlib).
//
//nolint:durationcheck // yes that's correct here
func (d Duration) String() string {
	td := time.Duration(d)
	// Small durations use stdlib for ms, ns etc
	if td < 1*time.Second && td > -1*time.Second {
		return td.String()
	}
	res := &strings.Builder{}
	if td < 0 {
		td = -td
		res.WriteByte('-')
	}
	days := td / Day
	td -= days * Day
	if days > 0 {
		weeks := days / 7
		if weeks > 0 {
			fmt.Fprintf(res, "%dw", weeks)
			days -= weeks * 7
		}
		if days > 0 {
			fmt.Fprintf(res, "%dd", days)
		}
	}
	hours := td / time.Hour
	if hours > 0 {
		fmt.Fprintf(res, "%dh", hours)
	}
	td -= hours * time.Hour
	minutes := td / time.Minute
	if minutes > 0 {
		fmt.Fprintf(res, "%dm", minutes)
	}
	td -= minutes * time.Minute
	seconds := td / time.Second
	td -= seconds * time.Second
	roundSeconds := (td == 0)
	if roundSeconds && seconds == 0 {
		return res.String()
	}
	if roundSeconds {
		fmt.Fprintf(res, "%ds", seconds)
		return res.String()
	}
	// fractional seconds
	fmt.Fprintf(res, "%d.", seconds)
	res.Write(writeFrac(td))
	res.WriteByte('s')
	return res.String()
}

// writeFrac writes the fractional part of duration (once seconds have been removed)
// up to nanosecond if present but returns a string with leading but not trailing zeroes.
func writeFrac(td time.Duration) []byte {
	var buf [9]byte
	i := len(buf)
	notZeroes := -1
	for td > 0 && i > 0 {
		i--
		v := td % 10
		if v != 0 && notZeroes == -1 {
			notZeroes = i + 1
		}
		buf[i] = byte('0' + v)
		td /= 10
	}
	for i > 0 {
		i--
		buf[i] = '0'
	}
	return buf[:notZeroes]
}

func (d *Duration) Set(s string) error {
	dd, err := Parse(s)
	if err != nil {
		return err
	}
	*d = Duration(dd)
	return nil
}

// Flag defines a duration flag with the specified name, default value, and usage string, like
// [flag.Duration] but supporting durations in days (24 hours) and weeks (7 days)
// in addition to the other stdlib units.
func Flag(name string, value time.Duration, usage string) *time.Duration {
	d := Duration(value)
	flag.Var(&d, name, usage)
	return (*time.Duration)(&d)
}

// NextTime takes a partially parsed time.Time (without date) and returns the time in the future
// relative to now with same HH:MM:SS. It will adjust to daylight savings so that time maybe 25h or 23h
// in the future around daylight savings transitions. If it's a double 3a-4a transition, and now is the
// first 3:10am asking for 3:05a will not give the very next 3:05a (same day) but instead the one next day
// ie at minimum 23h later.
func NextTime(now, d time.Time) time.Time {
	h := d.Hour()
	d = time.Date(now.Year(), now.Month(), now.Day(), h, d.Minute(), d.Second(), d.Nanosecond(), now.Location())
	if d.Before(now) {
		d = d.Add(24 * time.Hour)
		// daylight savings madness (see tests)
		if d.Hour() < h {
			d = d.Add(1 * time.Hour)
		}
		if d.Hour() > h {
			d = d.Add(-1 * time.Hour)
		}
	}
	return d
}

var ErrDateTimeParsing = errors.New("expecting one of YYYY-MM-DD HH:MM:SS, YYYY-MM-DD, HH:MM:SS, or H:MM am/pm")

// ParseDateTime parses date/times in one of the following format:
//   - Date and 24h time: YYYY-MM-DD HH:MM:SS
//   - Just a date: YYYY-MM-DD
//   - Just a time (12-hour, 'kitchen' style): H:MM AM/PM
//   - Just a 24h time: HH:MM:SS
//
// When date is missing next same time from now is used. When the time is missing 00:00 is assumed.
func ParseDateTime(now time.Time, s string) (time.Time, error) {
	d, err := time.Parse(time.DateTime, s)
	if err == nil {
		return d, nil
	}
	d, err = time.Parse(time.DateOnly, s)
	if err == nil {
		return d, nil
	}
	d, err = time.Parse(time.TimeOnly, s)
	if err == nil {
		return NextTime(now, d), nil
	}
	d, err = time.Parse(time.Kitchen, strings.ToUpper(strings.ReplaceAll(s, " ", "")))
	if err == nil {
		return NextTime(now, d), nil
	}
	return d, ErrDateTimeParsing
}
