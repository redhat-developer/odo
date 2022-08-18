package genericclioptions

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

var _ error = NoDevfileError{}

type NoDevfileError struct {
	context string
}

func NewNoDevfileError(context string) NoDevfileError {
	return NoDevfileError{
		context: context,
	}
}

func (o NoDevfileError) Error() string {
	message := `The current directory does not represent an odo component. 
To get started:%s
  * Open this folder in your favorite IDE and start editing, your changes will be reflected directly on the cluster.
Visit https://odo.dev for more information.`

	if isEmpty, _ := location.DirIsEmpty(filesystem.DefaultFs{}, o.context); isEmpty {
		message = fmt.Sprintf(message, `
  * Use "odo init" to initialize an odo component in the folder.
  * Use "odo dev" to deploy it on cluster.`)
	} else {
		message = fmt.Sprintf(message, `
  * Use "odo dev" to initialize an odo component for this folder and deploy it on cluster.`)
	}
	return message
}

func IsNoDevfileError(err error) bool {
	_, ok := err.(NoDevfileError)
	return ok
}
