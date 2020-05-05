package parth

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"testing"
)

var (
	x interface{}
)

func BenchmarkSegmentString(b *testing.B) {
	p := "/zero/1/2"
	var r string

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = Segment(p, 1, &r)
	}

	x = r
}

func BenchmarkSegmentInt(b *testing.B) {
	p := "/zero/1"
	var r int

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = Segment(p, 1, &r)
	}

	x = r
}

func BenchmarkSegmentIntNegIndex(b *testing.B) {
	p := "/zero/1"
	var r int

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = Segment(p, -1, &r)
	}

	x = r
}

func BenchmarkSpan(b *testing.B) {
	p := "/zero/1/2"
	var r string

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r, _ = Span(p, 0, 1)
	}

	x = r
}

func BenchmarkStdlibSegmentString(b *testing.B) {
	p := "/zero/1"
	var r string

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		r, _ = stdlibSegmentString(p, 1)
	}

	x = r
}

func BenchmarkStdlibSegmentInt(b *testing.B) {
	p := "/zero/1"
	var r int

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		r, _ = stdlibSegmentInt(p, 1)
	}

	x = r
}

func BenchmarkStdlibSpan(b *testing.B) {
	p := "/zero/1/2"
	var r string

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		r, _ = stdlibSpan(p, 0, 1)
	}

	x = r
}

type testkey string

var idKey testkey = "id"

type param struct {
	key   string
	value string
}

type params []param

func (ps params) byName(name string) string {
	for i := range ps {
		if ps[i].key == name {
			return ps[i].value
		}
	}
	return ""
}

func makeParams(val string) params {
	return params{
		{key: string(idKey), value: val},
	}
}

func BenchmarkContextLookupSetGet(b *testing.B) {
	req, _ := http.NewRequest("GET", "", nil)
	ps := makeParams("123")
	var v, r0 string

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		v = ps.byName(string(idKey))

		ctx := context.WithValue(req.Context(), idKey, v)
		req = req.WithContext(ctx)

		r0 = req.Context().Value(idKey).(string)
	}

	x = r0
}

func stdlibSegmentString(p string, i int) (string, error) {
	s, err := stdlibSpan(p, i, i+1)
	if err != nil {
		return "", err
	}

	if s[0] == '/' {
		s = s[1:]
	}

	return s, nil
}

func stdlibSegmentInt(p string, i int) (int, error) {
	s, err := stdlibSegmentString(p, i)
	if err != nil {
		return 0, err
	}

	v, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		return 0, err
	}

	return int(v), nil
}

func stdlibSpan(p string, i, j int) (string, error) {
	cs := strings.Split(p, "/")

	pfx := "/"
	start := 1
	if p[0] != '/' {
		start = 0
		if i == 0 {
			pfx = ""
		}
	}
	cs = cs[start:]

	if len(cs) == 0 || i >= len(cs) || j > len(cs) || i < 0 || j <= 0 {
		return "", fmt.Errorf("segment out of bounds")
	}

	if i > j {
		return "", fmt.Errorf("segments reversed")
	}

	return pfx + path.Join(cs[i:j]...), nil
}
