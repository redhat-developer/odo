package v2

import (
	"fmt"
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"strings"
)

// GetEvents returns the Events Object parsed from devfile
func (d *DevfileV2) GetEvents() v1.Events {
	if d.Events != nil {
		return *d.Events
	}
	return v1.Events{}
}

// AddEvents adds the Events Object to the devfile's events
// an event field is considered as invalid if it is already defined
// all event fields will be checked and processed, and returns a total error of all event fields
func (d *DevfileV2) AddEvents(events v1.Events) error {

	if d.Events == nil {
		d.Events = &v1.Events{}
	}
	var errorsList []string
	if len(events.PreStop) > 0 {
		if len(d.Events.PreStop) > 0 {
			errorsList = append(errorsList, (&common.FieldAlreadyExistError{Field: "event field", Name: "pre stop"}).Error())
		} else {
			d.Events.PreStop = events.PreStop
		}
	}

	if len(events.PreStart) > 0 {
		if len(d.Events.PreStart) > 0 {
			errorsList = append(errorsList, (&common.FieldAlreadyExistError{Field: "event field", Name: "pre start"}).Error())
		} else {
			d.Events.PreStart = events.PreStart
		}
	}

	if len(events.PostStop) > 0 {
		if len(d.Events.PostStop) > 0 {
			errorsList = append(errorsList, (&common.FieldAlreadyExistError{Field: "event field", Name: "post stop"}).Error())
		} else {
			d.Events.PostStop = events.PostStop
		}
	}

	if len(events.PostStart) > 0 {
		if len(d.Events.PostStart) > 0 {
			errorsList = append(errorsList, (&common.FieldAlreadyExistError{Field: "event field", Name: "post start"}).Error())
		} else {
			d.Events.PostStart = events.PostStart
		}
	}
	if len(errorsList) > 0 {
		return fmt.Errorf("errors while adding events:\n%s", strings.Join(errorsList, "\n"))
	}
	return nil
}

// UpdateEvents updates the devfile's events
// it only updates the events passed to it
func (d *DevfileV2) UpdateEvents(postStart, postStop, preStart, preStop []string) {

	if d.Events == nil {
		d.Events = &v1.Events{}
	}

	if postStart != nil {
		d.Events.PostStart = postStart
	}
	if postStop != nil {
		d.Events.PostStop = postStop
	}
	if preStart != nil {
		d.Events.PreStart = preStart
	}
	if preStop != nil {
		d.Events.PreStop = preStop
	}
}
