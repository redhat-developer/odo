package scm

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

func TestCreateListenerBinding(t *testing.T) {
	validListenerBinding := triggersv1.EventListenerBinding{
		Name: "sample",
	}
	listenerBinding := createListenerBinding("sample")
	if diff := cmp.Diff(validListenerBinding, *listenerBinding); diff != "" {
		t.Fatalf("createListenerBinding() failed:\n%s", diff)
	}
}

func TestCreateListenerTemplate(t *testing.T) {
	validListenerTemplate := triggersv1.EventListenerTemplate{
		Name: "sample",
	}
	listenerTemplate := createListenerTemplate("sample")
	if diff := cmp.Diff(validListenerTemplate, listenerTemplate); diff != "" {
		t.Fatalf("createListenerTemplate() failed:\n%s", diff)
	}
}

func TestCreateEventInterceptor(t *testing.T) {
	validEventInterceptor := triggersv1.EventInterceptor{
		CEL: &triggersv1.CELInterceptor{
			Filter: "sampleFilter sample",
		},
	}
	eventInterceptor := createEventInterceptor("sampleFilter %s", "sample")
	if diff := cmp.Diff(validEventInterceptor, *eventInterceptor); diff != "" {
		t.Fatalf("createEventInterceptor() failed:\n%s", diff)
	}
}

func TestGetDriverName(t *testing.T) {

	tests := []struct {
		url          string
		driver       string
		driverErrMsg string
	}{
		{
			"http://github.com/",
			"github",
			"",
		},
		{
			"http://github.com/foo/bar",
			"github",
			"",
		},
		{
			"https://githuB.com/foo/bar.git",
			"github",
			"",
		},
		{
			"http://gitlab.com/foo/bar.git2",
			"gitlab",
			"",
		},
		{
			"http://gitlab/foo/bar/",
			"",
			"unable to determine type of Git host from: http://gitlab/foo/bar/",
		},
		{
			"https://gitlab.a.b/foo/bar/bar",
			"",
			"unable to determine type of Git host from: https://gitlab.a.b/foo/bar/bar",
		},
		{
			"https://gitlab.org2/f.b/bar.git",
			"",
			"unable to determine type of Git host from: https://gitlab.org2/f.b/bar.git",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(rt *testing.T) {
			gotDriver, err := getDriverName(tt.url)
			if err != nil {
				if diff := cmp.Diff(tt.driverErrMsg, err.Error()); diff != "" {
					rt.Errorf("driver errMsg mismatch: \n%s", diff)
				}
			}
			if diff := cmp.Diff(tt.driver, gotDriver); diff != "" {
				rt.Errorf("driver mismatch: \n%s", diff)
			}
		})
	}
}
