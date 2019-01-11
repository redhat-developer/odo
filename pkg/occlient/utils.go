package occlient

func hasTag(tags []string, requiredTag string) bool {
	for _, tag := range tags {
		if tag == requiredTag {
			return true
		}
	}
	return false
}
