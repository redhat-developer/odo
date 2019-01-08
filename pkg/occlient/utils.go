package occlient

func isHidden(tags []string) bool {
	for _, tag := range tags {
		if tag == "hidden" {
			return true
		}
	}
	return false
}
