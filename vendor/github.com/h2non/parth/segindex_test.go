package parth

import "testing"

func TestUnitSegStartIndexFromEnd(t *testing.T) {
	tests := []struct {
		i      int
		s      string
		want   int
		okWant bool
	}{
		{-1, "/test1", 0, true},
		{-1, "/test1/test-2", 6, true},
		{-1, "/test1/test-2/test_3", 13, true},
		{-3, "test3/t3/", 0, true},
		{-1, "test4/t4/", 8, true},
		{-2, "/t5/f/fiv/55/5/fi/ve", 14, true},
		{-1, "/", 0, true},
		{-2, "/", 0, false},
		{-4, "/test/out", 0, false},
		{1, "/test/out", 0, false},
	}

	for _, tt := range tests {
		got, okGot := segStartIndexFromEnd(tt.s, tt.i)
		if okGot != tt.okWant {
			t.Errorf(gwxFmt, tt.s, okGot, tt.okWant)
			continue
		}

		if got != tt.want {
			t.Errorf(gwxFmt, tt.s, got, tt.want)
		}
	}
}

func TestUnitSegStartIndexFromStart(t *testing.T) {
	tests := []struct {
		i      int
		s      string
		want   int
		okWant bool
	}{
		{0, "/test1", 0, true},
		{1, "/test1/test-2", 6, true},
		{2, "/t1-2/fd", 0, false},
		{2, "/test1/test-2/test_3", 13, true},
		{0, "test3/t3/", 0, true},
		{1, "test4/t4/", 5, true},
		{6, "/t5/f/fiv/55/5/fi/ve", 17, true},
		{0, "/", 0, true},
		{1, "/", 0, false},
		{2, "/", 0, false},
		{4, "/test/out", 0, false},
		{-1, "/test/out", 0, false},
		{2, "/0/1//", 4, true},
		{3, "/0/1//", 5, true},
	}

	for _, tt := range tests {
		got, okGot := segStartIndexFromStart(tt.s, tt.i)
		if okGot != tt.okWant {
			t.Errorf(gwxFmt, tt.s, okGot, tt.okWant)
			continue
		}

		if got != tt.want {
			t.Errorf(gwxFmt, tt.s, got, tt.want)
		}
	}
}

func TestUnitSegEndIndexFromEnd(t *testing.T) {
	tests := []struct {
		i      int
		s      string
		want   int
		okWant bool
	}{
		{0, "/test1", 6, true},
		{-1, "/t1", 0, true},
		{-1, "/test1/test-2", 6, true},
		{-2, "/test1/t-2/t_3", 6, true},
		{-3, "test3/t3/", 0, true},
		{-1, "test4/t4/", 8, true},
		{-2, "/t5/f/fiv/55/5/fi/ve", 14, true},
		{-1, "/", 0, true},
		{-4, "/test/out", 0, false},
		{4, "/test/out", 0, false},
	}

	for _, tt := range tests {
		got, okGot := segEndIndexFromEnd(tt.s, tt.i)
		if okGot != tt.okWant {
			t.Errorf(gwxFmt, tt.s, okGot, tt.okWant)
			continue
		}

		if got != tt.want {
			t.Errorf(gwxFmt, tt.s, got, tt.want)
		}
	}
}

func TestUnitSegEndIndexFromStart(t *testing.T) {
	tests := []struct {
		i      int
		s      string
		want   int
		okWant bool
	}{
		{1, "/test1", 6, true},
		{2, "/test1/test-2", 13, true},
		{2, "/test1/test-2/test_3", 13, true},
		{1, "test3/t3/", 5, true},
		{2, "test4/t4/", 8, true},
		{5, "/t5/f/fiv/55/5/fi/ve", 14, true},
		{1, "/", 1, true},
		{-4, "/test/out", 0, false},
		{4, "/test/out", 0, false},
	}

	for _, tt := range tests {
		got, okGot := segEndIndexFromStart(tt.s, tt.i)
		if okGot != tt.okWant {
			t.Errorf(gwxFmt, tt.s, okGot, tt.okWant)
			continue
		}

		if got != tt.want {
			t.Errorf(gwxFmt, tt.s, got, tt.want)
		}
	}
}

func TestUnitSegIndexByKey(t *testing.T) {
	tests := []struct {
		k      string
		s      string
		want   int
		okWant bool
	}{
		{"test", "/1/test/3", 2, true},
		{"test", "/0/test/1/test/3", 2, true},
		{"2", "/2/t/3", 0, true},
		{"3", "/1/test/3", 7, true},
		{"4", "/44/44/33", 0, false},
		{"best", "12/best/3", 2, true},
		{"6", "6/tt/66", 0, true},
		{"7", "1/test/7", 6, true},
		{"first", "first/2/three", 0, true},
		{"bad", "/ba/d/", 0, false},
		{"11", "/4/56/11/", 5, true},
		{"", "/4/56/11/", 0, false},
		{"t", "", 0, false},
	}

	for _, tt := range tests {
		got, okGot := segIndexByKey(tt.s, tt.k)
		if okGot != tt.okWant {
			t.Errorf(gwxFmt, tt.s, okGot, tt.okWant)
			continue
		}

		if got != tt.want {
			t.Errorf(gwxFmt, tt.s, got, tt.want)
		}
	}
}
