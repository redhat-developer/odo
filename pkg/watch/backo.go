package watch

import (
	"time"

	"github.com/segmentio/backo-go"
	"k8s.io/klog/v2"
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

func (o *ExpBackoff) Delay() time.Duration {
	duration := o.backo.Duration(o.attempt)
	klog.V(4).Infof("wait for %v\n", duration)
	o.attempt++
	return duration
}

func (o *ExpBackoff) Reset() {
	o.attempt = 0
}
