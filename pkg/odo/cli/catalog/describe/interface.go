package describe

// CatalogProviderBackend is implemented by the catalog backends supported by odo
// It is used in "odo catalog describe service".
type CatalogProviderBackend interface {
	// the second argument can be a list of anything that needs to be sent to populate internal
	// structs
	CompleteDescribeService(*DescribeServiceOptions, []string) error
	ValidateDescribeService(*DescribeServiceOptions) error
	RunDescribeService(*DescribeServiceOptions) error
}
