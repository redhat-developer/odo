package variables

// checkForInvalidError checks for InvalidKeysError and stores the key in the map
func checkForInvalidError(invalidKeys map[string]bool, err error) {
	if verr, ok := err.(*InvalidKeysError); ok {
		for _, key := range verr.Keys {
			invalidKeys[key] = true
		}
	}
}
