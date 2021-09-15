package service

import (
	"fmt"
	"testing"
)

func TestBuildCRDFromParams(t *testing.T) {
	params := map[string]string{
		"u":     "1",
		"a.b.c": "2",
		"a.b.d": "3",
		"a.B":   "4",
	}
	res, _ := BuildCRDFromParams(params, "a group", "a version", "a kind")
	fmt.Printf("%+v\n", res)
}
