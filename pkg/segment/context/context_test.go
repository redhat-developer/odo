package context

import (
	"context"
	"reflect"
	"testing"
)

func TestSegmentContext(t *testing.T) {
	key, value := "componentType", "java"
	ctx := NewContext(context.Background())
	SetContextProperty(ctx, key, value)

	got := GetContextProperties(ctx)
	want := map[string]interface{}{key: value}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: %q got: %q", want, got)
	}
}
