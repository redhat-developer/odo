package context

import (
	"context"
	"reflect"
	"testing"
)

func TestGetContextProperties(t *testing.T) {
	key, value := "preferenceKey", "consenttelemetry"
	ctx := NewContext(context.Background())
	setContextProperty(ctx, key, value)

	got := GetContextProperties(ctx)
	want := map[string]interface{}{key: value}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: %q got: %q", want, got)
	}
}

func TestSetComponentType(t *testing.T) {
	key, value := "componentType", "maven"
	ctx := NewContext(context.Background())
	SetComponentType(ctx, value)

	if _, contains := GetContextProperties(ctx)[key]; !contains {
		t.Errorf("component type was not set.")
	}
}
