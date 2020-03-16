package lclient

import (
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
)

// PullImage uses Docker to pull the specified image. If there are any issues pulling the image,
// it returns an error.
func (dc *Client) PullImage(image string) error {

	codewindOut, err := dc.Client.ImagePull(dc.Context, image, types.ImagePullOptions{})

	if err != nil {
		return errors.Wrapf(err, "Unable to pull docker image")
	}

	defer codewindOut.Close()
	io.Copy(os.Stdout, codewindOut)
	return nil
}
