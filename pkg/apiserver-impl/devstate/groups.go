package devstate

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"

	"k8s.io/utils/pointer"
)

func GetGroup(command v1alpha2.Command) string {
	if command.Exec != nil && command.Exec.Group != nil {
		return string(command.Exec.Group.Kind)
	}
	if command.Apply != nil && command.Apply.Group != nil {
		return string(command.Apply.Group.Kind)
	}
	if command.Composite != nil && command.Composite.Group != nil {
		return string(command.Composite.Group.Kind)
	}
	return ""
}

func GetDefault(command v1alpha2.Command) bool {
	if command.Exec != nil && command.Exec.Group != nil {
		return pointer.BoolDeref(command.Exec.Group.IsDefault, false)
	}
	if command.Apply != nil && command.Apply.Group != nil {
		return pointer.BoolDeref(command.Apply.Group.IsDefault, false)
	}
	if command.Composite != nil && command.Composite.Group != nil {
		return pointer.BoolDeref(command.Composite.Group.IsDefault, false)
	}
	return false
}

func SetGroup(command *v1alpha2.Command, group string) {
	if command.Exec != nil {
		if group == "" {
			command.Exec.Group = nil
			return
		}
		if command.Exec.Group == nil {
			command.Exec.Group = &v1alpha2.CommandGroup{}
		}
		command.Exec.Group.Kind = v1alpha2.CommandGroupKind(group)
		return
	}
	if command.Apply != nil {
		if group == "" {
			command.Apply.Group = nil
			return
		}
		if command.Apply.Group == nil {
			command.Apply.Group = &v1alpha2.CommandGroup{}
		}
		command.Apply.Group.Kind = v1alpha2.CommandGroupKind(group)
		return
	}
	if command.Composite != nil {
		if group == "" {
			command.Composite.Group = nil
			return
		}
		if command.Composite.Group == nil {
			command.Composite.Group = &v1alpha2.CommandGroup{}
		}
		command.Composite.Group.Kind = v1alpha2.CommandGroupKind(group)
		return
	}
}

func SetDefault(command *v1alpha2.Command, def bool) {
	if command.Exec != nil {
		if command.Exec.Group == nil {
			command.Exec.Group = &v1alpha2.CommandGroup{}
		}
		command.Exec.Group.IsDefault = pointer.Bool(def)
		return
	}
	if command.Apply != nil {
		if command.Apply.Group == nil {
			command.Apply.Group = &v1alpha2.CommandGroup{}
		}
		command.Apply.Group.IsDefault = pointer.Bool(def)
		return
	}
	if command.Composite != nil {
		if command.Composite.Group == nil {
			command.Composite.Group = &v1alpha2.CommandGroup{}
		}
		command.Composite.Group.IsDefault = pointer.Bool(def)
		return
	}
}
