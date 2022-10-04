package context

import (
	"context"
	"testing"

	"github.com/redhat-developer/odo/pkg/odo/commonflags"
)

func TestOutput(t *testing.T) {
	ctx := context.TODO()
	ctx = WithOutput(ctx, true)
	res := GetOutput(ctx)
	if res != true {
		t.Errorf("GetOutput should return true but returns %v", res)
	}

	ctx = context.TODO()
	res = GetOutput(ctx)
	if res != false {
		t.Errorf("GetOutput should return false but returns %v", res)
	}

	ctx = context.TODO()
	ctx = WithOutput(ctx, false)
	res = GetOutput(ctx)
	if res != false {
		t.Errorf("GetOutput should return false but returns %v", res)
	}
}

func TestRunOn(t *testing.T) {
	ctx := context.TODO()
	ctx = WithRunOn(ctx, commonflags.RunOnCluster)
	res := GetRunOn(ctx)
	if res != commonflags.RunOnCluster {
		t.Errorf("GetOutput should return %q but returns %q", commonflags.RunOnCluster, res)
	}

	ctx = context.TODO()
	ctx = WithRunOn(ctx, commonflags.RunOnPodman)
	res = GetRunOn(ctx)
	if res != commonflags.RunOnPodman {
		t.Errorf("GetOutput should return %q but returns %q", commonflags.RunOnPodman, res)
	}

	ctx = context.TODO()
	res = GetRunOn(ctx)
	if res != commonflags.RunOnDefault {
		t.Errorf("GetOutput should return %q (default) but returns %q", commonflags.RunOnDefault, res)
	}
}
