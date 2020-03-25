package lclient

import (
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
	"k8s.io/klog/glog"
)

// PullImage uses Docker to pull the specified image. If there are any issues pulling the image,
// it returns an error.
func (dc *Client) PullImage(image string) error {

	out, err := dc.Client.ImagePull(dc.Context, image, types.ImagePullOptions{})

	if err != nil {
		return errors.Wrapf(err, "Unable to pull image")
	}

	defer out.Close()
	if glog.V(4) {
		_, err := io.Copy(os.Stdout, out)
		if err != nil {
			return err
		}
	}

	return nil
}
