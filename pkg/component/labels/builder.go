package labels

import (
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
)

type builder struct {
	m labels.Set
}

func Builder() builder {
	return builder{
		m: make(map[string]string),
	}
}

func (o builder) Labels() map[string]string {
	return o.m
}

func (o builder) Selector() string {
	return o.m.String()
}

func (o builder) SelectorFlag() string {
	return fmt.Sprintf("--selector=%s", o.m.String())
}

func (o builder) WithComponentName(name string) builder {
	o.m[KubernetesInstanceLabel] = name
	return o
}

func (o builder) WithAppName(name string) builder {
	o.m[KubernetesPartOfLabel] = name
	return o
}

func (o builder) WithApp(name string) builder {
	o.m[App] = name
	return o
}

func (o builder) WithManager(manager string) builder {
	o.m[KubernetesManagedByLabel] = manager
	return o
}

func (o builder) WithProjectType(typ string) builder {
	o.m[OdoProjectTypeAnnotation] = typ
	return o
}

func (o builder) WithMode(mode string) builder {
	o.m[OdoModeLabel] = mode
	return o
}

func (o builder) WithSourcePVC(s string) builder {
	o.m[SourcePVCLabel] = s
	return o
}

func (o builder) WithDevfileStorageName(name string) builder {
	o.m[DevfileStorageLabel] = name
	return o
}

func (o builder) WithStorageName(name string) builder {
	o.m[KubernetesStorageNameLabel] = name
	return o
}

func (o builder) WithComponent(name string) builder {
	o.m["component"] = name
	return o
}
