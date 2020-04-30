package resources

type Resources map[string]interface{}

func Merge(from, to Resources) Resources {
	merged := Resources{}
	for k, v := range to {
		merged[k] = v
	}
	for k, v := range from {
		merged[k] = v
	}
	return merged
}
