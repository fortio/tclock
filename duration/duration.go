// Deprecated: package moved to standalone 0 dependencies [fortio.org/duration] to
// allows duration parsing with "d" for days (24 hours) and "w" for week (7 days)
// and more.
package duration

import (
	"time"

	"fortio.org/duration"
)

const (
	// Deprecated: use [fortio.org/duration.Day] directly.
	Day = duration.Day
	// Deprecated: use [fortio.org/duration.Week] directly.
	Week = duration.Week
)

// Parse parses a duration string with "d" for days (24 hours)
// and "w" for weeks (7 days) in addition to what stdlib [time.ParseDuration] supports.
//
// Deprecated: use [fortio.org/duration.Parse] directly.
func Parse(s string) (time.Duration, error) {
	return duration.Parse(s)
}

// Deprecated: use [fortio.org/duration.Duration] directly.
type Duration = duration.Duration

// Flag defines a duration flag with the specified name, default value, and usage string, like
// [flag.Duration] but supporting durations in days (24 hours) and weeks (7 days)
// in addition to the other stdlib units.
//
// Deprecated: use [fortio.org/duration.Flag] directly.
func Flag(name string, value time.Duration, usage string) *time.Duration {
	return duration.Flag(name, value, usage)
}

// NextTime takes a partially parsed time.Time (without date) and returns the time in the future
// relative to now with same HH:MM:SS. It will adjust to daylight savings so that time maybe 25h or 23h
// in the future around daylight savings transitions. If it's a double 3a-4a transition, and now is the
// first 3:10am asking for 3:05a will not give the very next 3:05a (same day) but instead the one next day
// ie at minimum 23h later. The timezone of now is the one used and timezone of d should be the same.
//
// Deprecated: use [fortio.org/duration.NextTime] directly.
func NextTime(now, d time.Time) time.Time {
	return duration.NextTime(now, d)
}

// Deprecated: use [fortio.org/duration.ErrDateTimeParsing] directly.
var ErrDateTimeParsing = duration.ErrDateTimeParsing

// ParseDateTime parses date/times in one of the following format:
//   - Date and 24h time: YYYY-MM-DD HH:MM:SS
//   - Just a date: YYYY-MM-DD
//   - Just a 24h time: HH:MM:SS
//   - Just a time (12-hour, 'kitchen' style): H:MM AM/PM
//
// When date is missing next same time from now is used (ie later that day or next, up to 25h from now).
// When the time is missing 00:00 is assumed.
// The time/date is interpreted in relation to the location and timezone of the `now` parameter (local time
// per TZ environment when using time.Now() for instance).
//
// Deprecated: use [fortio.org/duration.ParseDateTime] directly.
func ParseDateTime(now time.Time, s string) (time.Time, error) {
	return duration.ParseDateTime(now, s)
}
