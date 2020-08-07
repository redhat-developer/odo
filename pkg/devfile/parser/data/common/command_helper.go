package common

import "strings"

// GetID returns the ID of the command
func (dc DevfileCommand) GetID() string {
	if dc.Composite != nil {
		return dc.Composite.Id
	} else if dc.Exec != nil {
		return dc.Exec.Id
	}

	return ""
}

// SetIDToLower converts the command's id to lower case for more efficient processing and returns the new id
func (dc *DevfileCommand) SetIDToLower() string {
	var newId string
	if dc.Exec != nil {
		newId = strings.ToLower(dc.Exec.Id)
		dc.Exec.Id = newId
	} else if dc.Composite != nil {
		newId = strings.ToLower(dc.Composite.Id)
		dc.Composite.Id = newId
	}
	return newId
}

// GetGroup returns the group the command belongs to
func (dc DevfileCommand) GetGroup() *Group {
	if dc.Composite != nil {
		return dc.Composite.Group
	} else if dc.Exec != nil {
		return dc.Exec.Group
	}

	return nil
}
