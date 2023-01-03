package context

import (
	"context"
	"testing"

	"github.com/redhat-developer/odo/pkg/odo/commonflags"
)

func TestOutput(t *testing.T) {
	ctx := context.TODO()
	ctx = WithJsonOutput(ctx, true)
	res := IsJsonOutput(ctx)
	if res != true {
		t.Errorf("GetOutput should return true but returns %v", res)
	}

	ctx = context.TODO()
	res = IsJsonOutput(ctx)
	if res != false {
		t.Errorf("GetOutput should return false but returns %v", res)
	}

	ctx = context.TODO()
	ctx = WithJsonOutput(ctx, false)
	res = IsJsonOutput(ctx)
	if res != false {
		t.Errorf("GetOutput should return false but returns %v", res)
	}
}

func TestPlatform(t *testing.T) {
	ctx := context.TODO()
	ctx = WithPlatform(ctx, commonflags.PlatformCluster)
	res := GetPlatform(ctx, commonflags.PlatformCluster)
	if res != commonflags.PlatformCluster {
		t.Errorf("GetOutput should return %q but returns %q", commonflags.PlatformCluster, res)
	}

	ctx = context.TODO()
	ctx = WithPlatform(ctx, commonflags.PlatformPodman)
	res = GetPlatform(ctx, commonflags.PlatformCluster)
	if res != commonflags.PlatformPodman {
		t.Errorf("GetOutput should return %q but returns %q", commonflags.PlatformPodman, res)
	}

	ctx = context.TODO()
	res = GetPlatform(ctx, commonflags.PlatformCluster)
	if res != commonflags.PlatformCluster {
		t.Errorf("GetOutput should return %q (default) but returns %q", commonflags.PlatformCluster, res)
	}
}
