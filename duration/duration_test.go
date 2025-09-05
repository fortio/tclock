package duration_test

import (
	"testing"

	"fortio.org/tclock/duration"
)

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
	}
	for _, test := range tests {
		d, err := duration.Parse(test.input)
		if err != nil {
			t.Error("Error:", err)
			continue
		}
		t.Log("Parsed duration (std) :", d)
		str := duration.Duration(d).String()
		t.Log("Parsed duration (ours):", str)
		if str != test.expected {
			t.Errorf("Expected %q but got %q", test.expected, str)
		}
	}
}
