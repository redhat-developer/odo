package pipeline

import (
	"fmt"
	"github.com/redhat-developer/service-binding-operator/pkg/binding"
	"strings"
)

func (crdDescription *CRDDescription) BindingAnnotations() map[string]string {
	anns := make(map[string]string)
	for _, sd := range crdDescription.StatusDescriptors {
		objectType := getObjectType(sd.XDescriptors)
		for _, xd := range sd.XDescriptors {
			loadDescriptor(anns, sd.Path, xd, "status", objectType)
		}
	}

	for _, sd := range crdDescription.SpecDescriptors {
		objectType := getObjectType(sd.XDescriptors)
		for _, xd := range sd.XDescriptors {
			loadDescriptor(anns, sd.Path, xd, "spec", objectType)
		}
	}
	return anns
}

func getObjectType(descriptors []string) string {
	typeAnno := "urn:alm:descriptor:io.kubernetes:"
	for _, desc := range descriptors {
		if strings.HasPrefix(desc, typeAnno) {
			return strings.TrimPrefix(desc, typeAnno)
		}
	}
	return ""
}

func loadDescriptor(anns map[string]string, path string, descriptor string, root string, objectType string) {
	if !strings.HasPrefix(descriptor, binding.AnnotationPrefix) {
		return
	}

	keys := strings.Split(descriptor, ":")
	key := binding.AnnotationPrefix
	value := ""

	if len(keys) > 1 {
		key += "/" + keys[1]
	} else {
		key += "/" + path
	}

	p := []string{fmt.Sprintf("path={.%s.%s}", root, path)}
	if len(keys) > 1 {
		p = append(p, keys[2:]...)
	}
	if objectType != "" {
		p = append(p, []string{fmt.Sprintf("objectType=%s", objectType)}...)
	}

	value += strings.Join(p, ",")
	anns[key] = value
}
