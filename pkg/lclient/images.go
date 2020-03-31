package lclient

import (
	"bytes"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
	"k8s.io/klog/glog"
)

// PullImage uses Docker to pull the specified image. If there are any issues pulling the image,
// it returns an error.
func (dc *Client) PullImage(image string) error {

	out, err := dc.Client.ImagePull(dc.Context, image, types.ImagePullOptions{})
	defer out.Close()

	if err != nil {
		return errors.Wrapf(err, "Unable to pull image")
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(out)

	if glog.V(4) {
		_, err := io.Copy(nil, out)
		if err != nil {
			return err
		}
	}

	newStr := buf.String()
	glog.V(4).Infof(newStr)
	return nil
}
