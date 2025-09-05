package duration_test

import (
	"flag"
	"fmt"
	"testing"
	"time"

	"fortio.org/tclock/duration"
)

func TestDurationParseErrors(t *testing.T) {
	tests := []struct {
		input string
	}{
		{" 23 "},
		{"s"},
		{"23.5.7s"},
		{"5x"},
		{"-3d-7h"},
		{"3d -7h 10m"},
	}
	for _, test := range tests {
		var d duration.Duration
		err := d.Set(test.input) // Parse errors through Set.
		t.Logf("Parsing %q: %v", test.input, err)
		if err == nil {
			t.Error("Expected error but got none, d=", d)
		}
	}
}

func TestDurationString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1d 2h 3m 1ns", "1d2h3m0.000000001s"},
		{"1d 2h 3m 0.5s", "1d2h3m0.5s"},
		{"1d 2h 3m", "1d2h3m"},
		{"1d 2h", "1d2h"},
		{"2h 3s", "2h3s"},
		{"1d", "1d"},
		{"0s", "0s"},
		{"1.5s", "1.5s"},
		{"1.5ms", "1.5ms"},
		{"1.5h", "1h30m"},
		{"1.5d", "1d12h"},
		{"6d23h59m59.999s", "6d23h59m59.999s"},
		{"7d", "1w"},
		{"8d", "1w1d"},
		{"1w2d3h4m5s", "1w2d3h4m5s"},
		{"99h", "4d3h"},
		{"   ", "0s"},
		{"10us", "10µs"},
		{"-1d7h", "-1d7h"},
	}
	for _, test := range tests {
		var d duration.Duration
		err := d.Set(test.input) // so we exercise both Set and Parse.
		if err != nil {
			t.Error("Error:", err)
			continue
		}
		t.Log("Parsed duration (std) :", time.Duration(d))
		str := d.String()
		t.Log("Parsed duration (ours):", str)
		if str != test.expected {
			t.Errorf("Expected %q but got %q", test.expected, str)
		}
	}
}

func TestFlag(t *testing.T) {
	defaultValue, _ := duration.Parse("1d3m")
	f := duration.Flag("test", defaultValue, "test `duration`")
	flag.Lookup("test").Value.Set("1d1h")
	if *f != 25*time.Hour {
		t.Errorf("Expected 25h but got %v", *f)
	}
}

func TestParseDateTime(t *testing.T) {
	l, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal("Error:", err)
	}
	now := time.Date(2025, 7, 31, 10, 30, 12, 0, time.Local)
	// dst transition days
	dst1 := time.Date(2025, 11, 1, 10, 30, 42, 0, l)
	dst2 := time.Date(2025, 3, 8, 10, 30, 42, 0, l)
	tests := []struct {
		now      time.Time
		input    string
		expected string
	}{
		{now, "1990-12-07 15:33:07", "1990-12-07 15:33:07"},
		{now, "3:07pm", "2025-07-31 15:07:00"},
		{now, "3:07 PM", "2025-07-31 15:07:00"},
		{now, "10:29am", "2025-08-01 10:29:00"},
		{now, "10:31:47", "2025-07-31 10:31:47"},
		{now, "2021-12-31", "2021-12-31 00:00:00"},
		// Daylight savings checks
		// dst1 transition has 25h to same time next day:
		{dst1, "3:07pm", "2025-11-01 15:07:00"},
		{dst1, "2:07am", "2025-11-02 02:07:00"},
		{dst1, "3:07am", "2025-11-02 03:07:00"},
		{dst1, "9:32am", "2025-11-02 09:32:00"},
		// dst2 transition has 23h to same time next day:
		{dst2, "3:07am", "2025-03-09 03:07:00"},
		{dst2, "9:07am", "2025-03-09 09:07:00"},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			d, err := duration.ParseDateTime(test.now, test.input)
			if err != nil {
				t.Fatal("Error:", err)
			}
			str := d.Format(time.DateTime)
			if str != test.expected {
				t.Errorf("Expected %v but got %v", test.expected, str)
			}
		})
	}
	d, err := duration.ParseDateTime(now, "23:00") // consider making this work instead of error
	if err == nil {
		t.Error("Expected error but got none:", d)
	}
}

func Example() {
	d, err := duration.Parse("1w 2d 3h 4m")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Parsed duration (std):", d)
	fmt.Println("Parsed duration (new):", duration.Duration(d))
	// Output:
	// Parsed duration (std): 219h4m0s
	// Parsed duration (new): 1w2d3h4m
}
