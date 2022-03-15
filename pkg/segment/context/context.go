package context

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/redhat-developer/odo/pkg/kclient"

	dfutil "github.com/devfile/library/pkg/util"

	"k8s.io/klog"
)

const ComponentType = "componentType"
const ClusterType = "clusterType"
const TelemetryStatus = "isTelemetryEnabled"
const DevfileName = "devfileName"
const Language = "language"
const ProjectType = "projectType"

const NOTFOUND = "not-found"

type contextKey struct{}

var key = contextKey{}

// properties is a struct used to store data in a context and it comes with locking mechanism
type properties struct {
	lock    sync.Mutex
	storage map[string]interface{}
}

// NewContext returns a context more specifically to be used for telemetry data collection
func NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, key, &properties{storage: make(map[string]interface{})})
}

// GetContextProperties retrieves all the values set in a given context
func GetContextProperties(ctx context.Context) map[string]interface{} {
	cProperties := propertiesFromContext(ctx)
	if cProperties == nil {
		return make(map[string]interface{})
	}
	return cProperties.values()
}

// SetComponentType sets componentType property for telemetry data when a component is created/pushed
func SetComponentType(ctx context.Context, value string) {
	setContextProperty(ctx, ComponentType, dfutil.ExtractComponentType(value))
}

// SetClusterType sets clusterType property for telemetry data when a component is pushed or a project is created/set
func SetClusterType(ctx context.Context, client kclient.ClientInterface) {
	var value string
	if client == nil {
		value = NOTFOUND
	} else {
		// We are not checking ServerVersion to decide the cluster type because it does not always return the version,
		// it sometimes fails to retrieve the data if user is using minishift or plain oc cluster
		isOC, err := client.IsProjectSupported()
		if err != nil {
			klog.V(3).Info(fmt.Errorf("unable to detect project support: %w", err))
			value = NOTFOUND
		} else {
			if isOC {
				isOC4, err := client.IsCSVSupported()
				// TODO: Add a unit test for this case
				if err != nil {
					value = "openshift"
				} else {
					if isOC4 {
						value = "openshift4"
					} else {
						value = "openshift3"
					}
				}
			} else {
				value = "kubernetes"
			}
		}
	}
	setContextProperty(ctx, ClusterType, value)
}

// SetTelemetryStatus sets telemetry status before a command is run
func SetTelemetryStatus(ctx context.Context, isEnabled bool) {
	setContextProperty(ctx, TelemetryStatus, isEnabled)
}

func SetSignal(ctx context.Context, signal os.Signal) {
	setContextProperty(ctx, "receivedSignal", signal.String())
}

func SetDevfileName(ctx context.Context, devfileName string) {
	setContextProperty(ctx, DevfileName, devfileName)
}

func SetLanguage(ctx context.Context, language string) {
	setContextProperty(ctx, Language, language)
}

func SetProjectType(ctx context.Context, projectType string) {
	setContextProperty(ctx, ProjectType, projectType)
}

// GetTelemetryStatus gets the telemetry status that is set before a command is run
func GetTelemetryStatus(ctx context.Context) bool {
	isEnabled, ok := GetContextProperties(ctx)[TelemetryStatus]
	if ok {
		return isEnabled.(bool)
	}
	return false
}

// set safely sets value for a key in storage
func (p *properties) set(name string, value interface{}) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.storage[name] = value
}

// values safely retrieves a deep copy of the storage
func (p *properties) values() map[string]interface{} {
	p.lock.Lock()
	defer p.lock.Unlock()
	ret := make(map[string]interface{})
	for k, v := range p.storage {
		ret[k] = v
	}
	return ret
}

// propertiesFromContext retrieves the properties instance from the context
func propertiesFromContext(ctx context.Context) *properties {
	value := ctx.Value(key)
	if cast, ok := value.(*properties); ok {
		return cast
	}
	return nil
}

// setContextProperty sets the value of a key in given context for telemetry data
func setContextProperty(ctx context.Context, key string, value interface{}) {
	cProperties := propertiesFromContext(ctx)
	if cProperties != nil {
		cProperties.set(key, value)
	}
}
