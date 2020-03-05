package storage

// HelperAdapter defines functions that kubernetes storage adapters may implement
type HelperAdapter interface {
	GetVolumeNameToPVCName() map[string]string
}
