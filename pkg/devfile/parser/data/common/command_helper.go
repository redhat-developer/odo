package common

// GetID returns the ID of the command
func (dc DevfileCommand) GetID() string {
	if dc.Composite != nil {
		return dc.Composite.Id
	} else if dc.Exec != nil {
		return dc.Exec.Id
	}

	return ""
}

// GetKind returns the kind of the command
func (dc DevfileCommand) GetKind() DevfileCommandGroupType {
	if dc.Composite != nil {
		return dc.Composite.Group.Kind
	}

	return dc.Exec.Group.Kind
}

// GetGroup returns the group the command belongs to
func (dc DevfileCommand) GetGroup() *Group {
	if dc.Composite != nil {
		return dc.Composite.Group
	}

	return dc.Exec.Group
}
