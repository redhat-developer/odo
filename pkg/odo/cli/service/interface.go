package service

import "github.com/spf13/cobra"

// ServiceProviderBackend is implemented by the backends supported by odo
// It is used in "odo service create" and "odo service delete"
type ServiceProviderBackend interface {
	CompleteServiceCreate(options *CreateOptions, cmd *cobra.Command, args []string) error
	ValidateServiceCreate(options *CreateOptions) error
	RunServiceCreate(options *CreateOptions) error

	ServiceDefined(options *DeleteOptions) (bool, error)
	ServiceExists(options *DeleteOptions) (bool, error)
	DeleteService(options *DeleteOptions, serviceName, app string) error
}
