package motd

import (
	"fmt"
	"io"
	"strings"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog"
)

// Display the necessary MOTD if it's configured in the system.
// Note that the MOTD comes from the "motd" configMap, which
// should be in the "openshift" namespace. This needs to be configured
// by the deployer.
func DisplayMOTD(coreClient corev1client.CoreV1Interface, out io.Writer) error {
	motdcm, err := coreClient.ConfigMaps("openshift").Get("motd", metav1.GetOptions{})
	if kerrors.IsNotFound(err) || kerrors.IsForbidden(err) {
		// NOTE(jaosorior): If no motd is configured, it's fine. No need to
		// print anything. If we get a "Forbidden" error, this is because no
		// role binding has been configured yet. So we need to ignore the
		// error for now (until the role binding is included by default).
		return nil
	} else if err != nil {
		return err
	}

	motd, ok := motdcm.Data["message"]

	if !ok {
		klog.V(4).Infof("Unable to display MOTD. It exists but is misconfigured.")
		return nil
	}

	// Add newline if needed
	if !strings.HasSuffix(motd, "\n") {
		motd = motd + "\n"
	}
	fmt.Fprintf(out, "\n%s", motd)
	return nil
}
