package watch

import (
	"fmt"

	"github.com/segmentio/backo-go"
)

type ExpBackoff struct {
	attempt int
	backo   *backo.Backo
}

func NewExpBackoff() *ExpBackoff {
	return &ExpBackoff{
		backo: backo.DefaultBacko(),
	}
}

func (o *ExpBackoff) Delay() {
	fmt.Printf("wait for %v\n", o.backo.Duration(o.attempt))
	o.backo.Sleep(o.attempt)
	o.attempt++
}

func (o *ExpBackoff) Reset() {
	o.attempt = 0
}
