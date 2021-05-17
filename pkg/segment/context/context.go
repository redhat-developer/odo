package context

import (
	"context"
	"sync"
)

type contextKey struct{}

var key = contextKey{}

type properties struct {
	lock    sync.Mutex
	storage map[string]interface{}
}

func (p *properties) set(name string, value interface{}) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.storage[name] = value
}

func (p *properties) values() map[string]interface{} {
	p.lock.Lock()
	defer p.lock.Unlock()
	ret := make(map[string]interface{})
	for k, v := range p.storage {
		ret[k] = v
	}
	return ret
}

// NewContext returns a context more specifically to be used for telemetry data collection
func NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, key, &properties{storage: make(map[string]interface{})})
}

func propertiesFromContext(ctx context.Context) *properties {
	value := ctx.Value(key)
	if cast, ok := value.(*properties); ok {
		return cast
	}
	return nil
}

// SetContextProperty sets the value of a key in given context for telemetry data
func SetContextProperty(ctx context.Context, key string, value interface{}) {
	properties := propertiesFromContext(ctx)
	if properties != nil {
		properties.set(key, value)
	}
}

// GetContextProperties retrieves all the values set in a given context
func GetContextProperties(ctx context.Context) map[string]interface{} {
	properties := propertiesFromContext(ctx)
	if properties == nil {
		return make(map[string]interface{})
	}
	return properties.values()
}
