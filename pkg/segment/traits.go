package segment

import (
	"gopkg.in/segmentio/analytics-go.v3"
	"runtime"
)

func traits() analytics.Traits{
	base := analytics.NewTraits().Set("os", runtime.GOOS)
	return base
}