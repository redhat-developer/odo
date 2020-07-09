package common

// GetID returns the ID of the command
func (dc DevfileCommand) GetID() string {
	if dc.Composite != nil {
		return dc.Composite.Id
	}

	return dc.Exec.Id
}

// GetKind returns the kind of the command
func (dc DevfileCommand) GetKind() DevfileCommandGroupType {
	if dc.Composite != nil {
		return dc.Composite.Group.Kind
	}

	return dc.Exec.Group.Kind
}
