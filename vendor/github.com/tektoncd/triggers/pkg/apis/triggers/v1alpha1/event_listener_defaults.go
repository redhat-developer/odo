package v1alpha1

import (
	"context"
)

// SetDefaults sets the defaults on the object.
func (el *EventListener) SetDefaults(ctx context.Context) {
	if IsUpgradeViaDefaulting(ctx) {
		// Most likely the EventListener passed here is already running
		for i := range el.Spec.Triggers {
			t := &el.Spec.Triggers[i]
			upgradeBinding(t)
			upgradeInterceptor(t)
			removeParams(t)
		}
	}
}

func upgradeBinding(t *EventListenerTrigger) {
	if t.DeprecatedBinding != nil {
		if len(t.Bindings) > 0 {
			// Do nothing since it will be a Validation Error.
		} else {
			// Set the binding to bindings
			t.Bindings = append(t.Bindings, &EventListenerBinding{
				Name: t.DeprecatedBinding.Name,
			})
			t.DeprecatedBinding = nil
		}
	}
}

func upgradeInterceptor(t *EventListenerTrigger) {
	if t.DeprecatedInterceptor != nil {
		if len(t.Interceptors) > 0 {
			// Do nothing since it will be a Validation Error.
			return
		}

		t.Interceptors = []*EventInterceptor{t.DeprecatedInterceptor}
		t.DeprecatedInterceptor = nil
	}
}

func removeParams(t *EventListenerTrigger) {
	if t.DeprecatedParams != nil {
		t.DeprecatedParams = nil
	}
}
