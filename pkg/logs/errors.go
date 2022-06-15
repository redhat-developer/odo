package logs

import (
	"fmt"

	odolabels "github.com/redhat-developer/odo/pkg/labels"
)

type InvalidModeError struct {
	mode string
}

func (e InvalidModeError) Error() string {
	return fmt.Sprintf("invalid mode %q; valid modes are %q, %q, and %q", e.mode, odolabels.ComponentDevMode, odolabels.ComponentDeployMode, odolabels.ComponentAnyMode)
}
