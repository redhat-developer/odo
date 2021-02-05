package validate

import "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"

func validatePreStart(preStart []string) (err error) {
	// This is odo specific validation. There is still discussion about how PreStart should be implemented.
	// https://github.com/devfile/api/issues/204
	// https://github.com/openshift/odo/issues/4187
	// This is here to prevent anyone from using PreStart event until we have a proper implementation
	if len(preStart) != 0 {
		return &UnsupportedFieldError{fieldName: "preStart"}
	}
	return nil
}

func validateEvents(events v1alpha2.Events) (err error) {

	if err := validatePreStart(events.PreStart); err != nil {
		return err
	}

	return nil
}
