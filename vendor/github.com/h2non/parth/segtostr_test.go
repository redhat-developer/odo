package parth

import "testing"

func TestUnitFirstFloatFromString(t *testing.T) {
	tests := []struct {
		s      string
		want   string
		okWant bool
	}{
		{"/0.1", "0.1", true},
		{"/0.2a", "0.2", true},
		{"/aaaa1.3", "1.3", true},
		{"/4", "4", true},
		{"/5aaaa", "5", true},
		{"/aaa6aa", "6", true},
		{"/.7.aaaa", ".7", true},
		{"/.8aa", ".8", true},
		{"/-9", "-9", true},
		{"/10-", "10", true},
		{"/3.14e+11", "3.14e+11", true},
		{"/3.14e.+12", "3.14", true},
		{"/3.14e+.13", "3.14", true},
		{"/3.14e+.13", "3.14", true},
		{"/error", "", false},
		{"/.", "", false},
	}

	for _, tt := range tests {
		got, okGot := firstFloatFromString(tt.s)
		if okGot != tt.okWant {
			t.Errorf(gwxFmt, tt.s, okGot, tt.okWant)
			continue
		}

		if got != tt.want {
			t.Errorf(gwxFmt, tt.s, got, tt.want)
		}
	}
}

func TestUnitFirstIntFromString(t *testing.T) {
	var tests = []struct {
		s      string
		want   string
		okWant bool
	}{
		{"0.1", "0", true},
		{"0.2a", "0", true},
		{"aaaa1.3", "1", true},
		{"4", "4", true},
		{"5aaaa", "5", true},
		{"aaa6aa", "6", true},
		{".7.aaaa", "0", true},
		{".8aa", "0", true},
		{"-9", "-9", true},
		{"10-", "10", true},
		{"3.14e+11", "3", true},
		{"3.14e.+12", "3", true},
		{"3.14e+.13", "3", true},
		{"18446744073709551615", "18446744073709551615", true},
		{".", "", false},
		{"error", "", false},
	}

	for _, tt := range tests {
		got, okGot := firstIntFromString(tt.s)
		if okGot != tt.okWant {
			t.Errorf(gwxFmt, tt.s, okGot, tt.okWant)
			continue
		}

		if got != tt.want {
			t.Errorf(gwxFmt, tt.s, got, tt.want)
		}
	}
}

func TestUnitFirstUintFromString(t *testing.T) {
	var tests = []struct {
		s      string
		want   string
		okWant bool
	}{
		{"0.1", "0", true},
		{"0.2a", "0", true},
		{"aaaa1.3", "1", true},
		{"4", "4", true},
		{"5aaaa", "5", true},
		{"aaa6aa", "6", true},
		{".7.aaaa", "0", true},
		{".8aa", "0", true},
		{"-9", "9", true},
		{"10-", "10", true},
		{"3.14e+11", "3", true},
		{"3.14e.+12", "3", true},
		{"3.14e+.13", "3", true},
		{"18446744073709551615", "18446744073709551615", true},
		{".", "", false},
		{"error", "", false},
	}

	for _, tt := range tests {
		got, okGot := firstUintFromString(tt.s)
		if okGot != tt.okWant {
			t.Errorf(gwxFmt, tt.s, okGot, tt.okWant)
			continue
		}

		if got != tt.want {
			t.Errorf(gwxFmt, tt.s, got, tt.want)
		}
	}
}
