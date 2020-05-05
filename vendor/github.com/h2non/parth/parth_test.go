package parth

import (
	"fmt"
	"reflect"
	"testing"
)

func TestBhvrSegment(t *testing.T) {
	path := "/junk/4/key/true/other/3.3/"
	key := ""

	t.Run("bool", applyToBoolTFunc(path, key, pti(3), true))
	t.Run("float32", applyToFloat32TFunc(path, key, pti(5), 3.3))
	t.Run("float64", applyToFloat64TFunc(path, key, pti(5), 3.3))
	t.Run("int", applyToIntTFunc(path, key, pti(1), 4))
	t.Run("int16", applyToInt16TFunc(path, key, pti(1), 4))
	t.Run("int32", applyToInt32TFunc(path, key, pti(1), 4))
	t.Run("int64", applyToInt64TFunc(path, key, pti(1), 4))
	t.Run("int8", applyToInt8TFunc(path, key, pti(1), 4))
	t.Run("string", applyToStringTFunc(path, key, pti(0), "junk"))
	t.Run("uint", applyToUintTFunc(path, key, pti(-1), 3))
	t.Run("uint16", applyToUint16TFunc(path, key, pti(-1), 3))
	t.Run("uint32", applyToUint32TFunc(path, key, pti(-1), 3))
	t.Run("uint64", applyToUint64TFunc(path, key, pti(-1), 3))
	t.Run("uint8", applyToUint8TFunc(path, key, pti(-1), 3))
	t.Run("unmarsaler", applyToUnmarshalerTFunc(path, key, pti(2), []byte("key")))

	t.Run("badType", func(t *testing.T) {
		var x uintptr
		err := Segment(path, 3, &x)
		exp(t, t.Name(), err)
	})
}

func TestBhvrSequent(t *testing.T) {
	path := "/junk/4/key/true/other/3.3/"
	var i *int

	t.Run("bool", applyToBoolTFunc(path, "key", i, true))
	t.Run("float32", applyToFloat32TFunc(path, "other", i, 3.3))
	t.Run("float64", applyToFloat64TFunc(path, "other", i, 3.3))
	t.Run("int", applyToIntTFunc(path, "junk", i, 4))
	t.Run("int16", applyToInt16TFunc(path, "junk", i, 4))
	t.Run("int32", applyToInt32TFunc(path, "junk", i, 4))
	t.Run("int64", applyToInt64TFunc(path, "junk", i, 4))
	t.Run("int8", applyToInt8TFunc(path, "junk", i, 4))
	t.Run("string", applyToStringTFunc(path, "key", i, "true"))
	t.Run("uint", applyToUintTFunc(path, "junk", i, 4))
	t.Run("uint16", applyToUint16TFunc(path, "junk", i, 4))
	t.Run("uint32", applyToUint32TFunc(path, "junk", i, 4))
	t.Run("uint64", applyToUint64TFunc(path, "junk", i, 4))
	t.Run("uint8", applyToUint8TFunc(path, "junk", i, 4))
	t.Run("unmarsaler", applyToUnmarshalerTFunc(path, "key", i, []byte("true")))

	t.Run("badType", func(t *testing.T) {
		var x uintptr
		err := Sequent(path, "key", &x)
		exp(t, t.Name(), err)
	})
}

func TestBhvrSpan(t *testing.T) {
	path := "/zero/one/two/three/four"

	tests := []struct {
		name string
		path string
		i, j int
		want string
		ck   checkFunc
	}{
		{"5 segs: +1,+3", path, 1, 3, "/one/two", unx},
		{"5 segs: +1,-2", path, 1, -2, "/one/two", unx},
		{"5 segs: +1,00", path, 1, 0, "/one/two/three/four", unx},
		{"5 segs: -3,-1", path, -3, -1, "/two/three", unx},
		{"5 segs: -3,00", path, -3, 0, "/two/three/four", unx},
		{"5 segs: -9,00", path, -9, 0, "", exp},
		{"5 segs: 00,+9", path, 0, 9, "", exp},
		{"3 no /: 00,+9", "zero/one/two", 0, 2, "zero/one", unx},
	}

	for _, tt := range tests {
		got, err := Span(tt.path, tt.i, tt.j)
		if tt.ck(t, tt.name, err) {
			continue
		}

		if got != tt.want {
			t.Errorf(gwxFmt, tt.name, got, tt.want)
		}
	}
}

func TestBhvrSubSeg(t *testing.T) {
	path := "/junk/4/key/true/other/3.3/"

	t.Run("bool", applyToBoolTFunc(path, "junk", pti(2), true))
	t.Run("float32", applyToFloat32TFunc(path, "true", pti(1), 3.3))
	t.Run("float64", applyToFloat64TFunc(path, "true", pti(1), 3.3))
	t.Run("int", applyToIntTFunc(path, "junk", pti(0), 4))
	t.Run("int16", applyToInt16TFunc(path, "junk", pti(0), 4))
	t.Run("int32", applyToInt32TFunc(path, "junk", pti(0), 4))
	t.Run("int64", applyToInt64TFunc(path, "junk", pti(0), 4))
	t.Run("int8", applyToInt8TFunc(path, "junk", pti(0), 4))
	t.Run("string", applyToStringTFunc(path, "junk", pti(3), "other"))
	t.Run("uint", applyToUintTFunc(path, "junk", pti(0), 4))
	t.Run("uint16", applyToUint16TFunc(path, "junk", pti(0), 4))
	t.Run("uint32", applyToUint32TFunc(path, "junk", pti(0), 4))
	t.Run("uint64", applyToUint64TFunc(path, "junk", pti(0), 4))
	t.Run("uint8", applyToUint8TFunc(path, "junk", pti(0), 4))
	t.Run("unmarsaler", applyToUnmarshalerTFunc(path, "key", pti(1), []byte("other")))

	t.Run("badType", func(t *testing.T) {
		var x uintptr
		err := SubSeg(path, "key", 2, &x)
		exp(t, t.Name(), err)
	})
}

func TestBhvrSubSpan(t *testing.T) {
	path := "/zero/one/two/key/four/five/six"

	tests := []struct {
		name string
		path string
		key  string
		i, j int
		want string
		ck   checkFunc
	}{
		{"7 segs: 00,+2", path, "key", 0, 2, "/four/five", unx},
		{"7 segs: +1,-1", path, "key", 1, -1, "/five", unx},
		{"7 segs: +1,00", path, "key", 1, 0, "/five/six", unx},
		{"7 segs: -3,-1", path, "key", -3, -1, "/four/five", unx},
		{"7 segs: -3,00", path, "key", -3, 0, "/four/five/six", unx},
		{"7 segs: -9,00", path, "key", -9, 0, "", exp},
		{"7 segs: 00,+9", path, "key", 0, 9, "", exp},
	}

	for _, tt := range tests {
		got, err := SubSpan(tt.path, tt.key, tt.i, tt.j)
		if tt.ck(t, tt.name, err) {
			continue
		}

		if got != tt.want {
			t.Errorf(gwxFmt, tt.name, got, tt.want)
		}
	}
}

func TestBhvrParth(t *testing.T) {
	t.Run("bySpan/segment", func(t *testing.T) {
		p := NewBySpan("/zero/one/two/three", 1, 3)

		var got string
		p.Segment(1, &got)
		if unx(t, t.Name(), p.Err()) {
			return
		}

		want := "two"
		if got != want {
			t.Errorf(gwFmt, got, want)
		}
	})

	t.Run("bySubSpan/sequent", func(t *testing.T) {
		p := NewBySubSpan("/zero/one/two/three/four", "one", 1, 0)

		var got string
		p.Sequent("three", &got)
		if unx(t, t.Name(), p.Err()) {
			return
		}

		want := "four"
		if got != want {
			t.Errorf(gwFmt, got, want)
		}
	})

	t.Run("basic", func(t *testing.T) {
		p := New("/zero/one/two/three")

		t.Run("subSeg", func(t *testing.T) {
			var got string
			p.SubSeg("one", 1, &got)

			want := "three"
			if got != want {
				t.Errorf(gwFmt, got, want)
			}
		})

		t.Run("span", func(t *testing.T) {
			got := p.Span(1, 3)

			want := "/one/two"
			if got != want {
				t.Errorf(gwFmt, got, want)
			}
		})

		t.Run("subSpan", func(t *testing.T) {
			got := p.SubSpan("one", 0, 0)

			want := "/two/three"
			if got != want {
				t.Errorf(gwFmt, got, want)
			}
		})

		unx(t, t.Name(), p.Err())
	})
}

func segSeqSubSeg(path, key string, i *int, v interface{}) error {
	if path != "" && key != "" && i != nil {
		return SubSeg(path, key, *i, v)
	}

	if path != "" && key != "" {
		return Sequent(path, key, v)
	}

	if path != "" && i != nil {
		return Segment(path, *i, v)
	}

	return fmt.Errorf("see segSeqSubSeg for missing requirements")
}

func applyToBoolTFunc(path, key string, i *int, want bool) func(*testing.T) {
	return func(t *testing.T) {
		subj := subject(path, key)
		if i != nil {
			subj = subject(path, key, *i)
		}

		var got bool
		err := segSeqSubSeg(path, key, i, &got)
		if unx(t, subj, err) {
			return
		}

		if got != want {
			t.Errorf(gwFmt, got, want)
		}
	}
}

func applyToFloat32TFunc(path, key string, i *int, want float32) func(*testing.T) {
	return func(t *testing.T) {
		subj := subject(path, key)
		if i != nil {
			subj = subject(path, key, *i)
		}

		var got float32
		err := segSeqSubSeg(path, key, i, &got)
		if unx(t, subj, err) {
			return
		}

		if got != want {
			t.Errorf(gwFmt, got, want)
		}
	}
}

func applyToFloat64TFunc(path, key string, i *int, want float64) func(*testing.T) {
	return func(t *testing.T) {
		subj := subject(path, key)
		if i != nil {
			subj = subject(path, key, *i)
		}

		var got float64
		err := segSeqSubSeg(path, key, i, &got)
		if unx(t, subj, err) {
			return
		}

		if got != want {
			t.Errorf(gwFmt, got, want)
		}
	}
}

func applyToIntTFunc(path, key string, i *int, want int) func(*testing.T) {
	return func(t *testing.T) {
		subj := subject(path, key)
		if i != nil {
			subj = subject(path, key, *i)
		}

		var got int
		err := segSeqSubSeg(path, key, i, &got)
		if unx(t, subj, err) {
			return
		}

		if got != want {
			t.Errorf(gwFmt, got, want)
		}
	}
}

func applyToInt16TFunc(path, key string, i *int, want int16) func(*testing.T) {
	return func(t *testing.T) {
		subj := subject(path, key)
		if i != nil {
			subj = subject(path, key, *i)
		}

		var got int16
		err := segSeqSubSeg(path, key, i, &got)
		if unx(t, subj, err) {
			return
		}

		if got != want {
			t.Errorf(gwFmt, got, want)
		}
	}
}

func applyToInt32TFunc(path, key string, i *int, want int32) func(*testing.T) {
	return func(t *testing.T) {
		subj := subject(path, key)
		if i != nil {
			subj = subject(path, key, *i)
		}

		var got int32
		err := segSeqSubSeg(path, key, i, &got)
		if unx(t, subj, err) {
			return
		}

		if got != want {
			t.Errorf(gwFmt, got, want)
		}
	}
}

func applyToInt64TFunc(path, key string, i *int, want int64) func(*testing.T) {
	return func(t *testing.T) {
		subj := subject(path, key)
		if i != nil {
			subj = subject(path, key, *i)
		}

		var got int64
		err := segSeqSubSeg(path, key, i, &got)
		if unx(t, subj, err) {
			return
		}

		if got != want {
			t.Errorf(gwFmt, got, want)
		}
	}
}

func applyToInt8TFunc(path, key string, i *int, want int8) func(*testing.T) {
	return func(t *testing.T) {
		subj := subject(path, key)
		if i != nil {
			subj = subject(path, key, *i)
		}

		var got int8
		err := segSeqSubSeg(path, key, i, &got)
		if unx(t, subj, err) {
			return
		}

		if got != want {
			t.Errorf(gwFmt, got, want)
		}
	}
}

func applyToStringTFunc(path, key string, i *int, want string) func(*testing.T) {
	return func(t *testing.T) {
		subj := subject(path, key)
		if i != nil {
			subj = subject(path, key, *i)
		}

		var got string
		err := segSeqSubSeg(path, key, i, &got)
		if unx(t, subj, err) {
			return
		}

		if got != want {
			t.Errorf(gwFmt, got, want)
		}
	}
}

func applyToUintTFunc(path, key string, i *int, want uint) func(*testing.T) {
	return func(t *testing.T) {
		subj := subject(path, key)
		if i != nil {
			subj = subject(path, key, *i)
		}

		var got uint
		err := segSeqSubSeg(path, key, i, &got)
		if unx(t, subj, err) {
			return
		}

		if got != want {
			t.Errorf(gwFmt, got, want)
		}
	}
}

func applyToUint16TFunc(path, key string, i *int, want uint16) func(*testing.T) {
	return func(t *testing.T) {
		subj := subject(path, key)
		if i != nil {
			subj = subject(path, key, *i)
		}

		var got uint16
		err := segSeqSubSeg(path, key, i, &got)
		if unx(t, subj, err) {
			return
		}

		if got != want {
			t.Errorf(gwFmt, got, want)
		}
	}
}

func applyToUint32TFunc(path, key string, i *int, want uint32) func(*testing.T) {
	return func(t *testing.T) {
		subj := subject(path, key)
		if i != nil {
			subj = subject(path, key, *i)
		}

		var got uint32
		err := segSeqSubSeg(path, key, i, &got)
		if unx(t, subj, err) {
			return
		}

		if got != want {
			t.Errorf(gwFmt, got, want)
		}
	}
}

func applyToUint64TFunc(path, key string, i *int, want uint64) func(*testing.T) {
	return func(t *testing.T) {
		subj := subject(path, key)
		if i != nil {
			subj = subject(path, key, *i)
		}

		var got uint64
		err := segSeqSubSeg(path, key, i, &got)
		if unx(t, subj, err) {
			return
		}

		if got != want {
			t.Errorf(gwFmt, got, want)
		}
	}
}

func applyToUint8TFunc(path, key string, i *int, want uint8) func(*testing.T) {
	return func(t *testing.T) {
		subj := subject(path, key)
		if i != nil {
			subj = subject(path, key, *i)
		}

		var got uint8
		err := segSeqSubSeg(path, key, i, &got)
		if unx(t, subj, err) {
			return
		}

		if got != want {
			t.Errorf(gwFmt, got, want)
		}
	}
}

func applyToUnmarshalerTFunc(path, key string, i *int, want []byte) func(*testing.T) {
	return func(t *testing.T) {
		subj := subject(path, key)
		if i != nil {
			subj = subject(path, key, *i)
		}

		var got custom
		err := segSeqSubSeg(path, key, i, &got)
		if unx(t, subj, err) {
			return
		}

		if !reflect.DeepEqual([]byte(got), want) {
			t.Errorf(gwFmt, got, want)
		}
	}
}

type custom []byte

func (c *custom) UnmarshalSegment(d string) error {
	*c = []byte(d)
	return nil
}

func pti(i int) *int {
	return &i
}

var (
	gwFmt  = "got %v, want %v"
	gwxFmt = "subj '%v': got %v, want %v"
)

type checkFunc func(*testing.T, interface{}, error) bool

func unx(t *testing.T, subj interface{}, err error) bool {
	b := err != nil
	if b {
		t.Errorf(gwxFmt, subj, err, nil)
	}
	return b
}

func exp(t *testing.T, subj interface{}, err error) bool {
	b := err == nil
	if b {
		t.Errorf(gwxFmt, subj, nil, "{error}")
	}
	return b
}

func subject(path, key string, indexes ...int) string {
	s := "path " + path
	if key != "" {
		s += ", key " + key
	}

	if len(indexes) > 0 {
		s += ", indexes"
	}

	for _, i := range indexes {
		s += fmt.Sprintf(" %d", i)
	}

	return s
}
