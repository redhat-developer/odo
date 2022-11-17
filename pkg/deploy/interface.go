package deploy

import (
	"context"
)

type Client interface {
	// Deploy resources from a devfile located in path, for the specified appName.
	// The filesystem specified is used to download and store the Dockerfiles needed to build the necessary container images,
	// in case such Dockerfiles are referenced as remote URLs in the Devfile.
	Deploy(ctx context.Context) error
}
