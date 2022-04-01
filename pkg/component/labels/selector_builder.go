package labels

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

type selectorBuilder struct {
	selector labels.Selector
}

func SelectorBuilder() selectorBuilder {
	return selectorBuilder{
		selector: labels.NewSelector(),
	}
}

func (o selectorBuilder) WithComponent(name string) selectorBuilder {
	req, err := labels.NewRequirement("component", selection.Equals, []string{name})
	if err != nil {
		panic(err)
	}
	o.selector = o.selector.Add(*req)
	return o
}

func (o selectorBuilder) WithoutSourcePVC(s string) selectorBuilder {
	req, err := labels.NewRequirement(sourcePVCLabel, selection.NotEquals, []string{s})
	if err != nil {
		panic(err)
	}
	o.selector = o.selector.Add(*req)
	return o
}

func (o selectorBuilder) Selector() string {
	return o.selector.String()
}
