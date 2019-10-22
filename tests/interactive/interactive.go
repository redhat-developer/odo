package interactive

import (
	"github.com/AlecAivazis/survey/core"
)

func init() {
	// disable color output for all prompts to simplify testing
	core.DisableColor = true
}
