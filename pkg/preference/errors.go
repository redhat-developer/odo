package preference

type MinimumDurationValueError struct{}

func NewMinimumDurationValueError() MinimumDurationValueError {
	return MinimumDurationValueError{}
}

func (v MinimumDurationValueError) Error() string {
	return "value must be a positive Duration (e.g. 4s, 5m, 1h); minimum value: 1s"
}
