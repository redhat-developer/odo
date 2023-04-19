package helper

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/pkg/podman"
)

func GenerateAndSetContainersConf(dir string) {
	ns := GetProjectName()
	containersConfPath := filepath.Join(dir, "containers.conf")
	err := CreateFileWithContent(containersConfPath, fmt.Sprintf(`
[engine]
namespace=%q
`, ns))
	Expect(err).ShouldNot(HaveOccurred())
	os.Setenv("CONTAINERS_CONF", containersConfPath)
}

// ExtractK8sAndOcComponentsFromOutputOnPodman extracts the list of Kubernetes and OpenShift components from the "odo" output on Podman.
func ExtractK8sAndOcComponentsFromOutputOnPodman(out string) []string {
	lines, err := ExtractLines(out)
	Expect(err).ShouldNot(HaveOccurred())

	var handled []string
	// Example lines to match:
	// ⚠ Kubernetes components are not supported on Podman. Skipping: k8s-deploybydefault-true-and-referenced, k8s-deploybydefault-true-and-not-referenced.
	// ⚠ OpenShift components are not supported on Podman. Skipping: ocp-deploybydefault-true-and-referenced.
	// ⚠  Apply OpenShift components are not supported on Podman. Skipping: k8s-deploybydefault-true-and-referenced.
	// ⚠  Apply OpenShift components are not supported on Podman. Skipping: k8s-deploybydefault-true-and-referenced.
	re := regexp.MustCompile(`(?:Kubernetes|OpenShift) components are not supported on Podman\.\s*Skipping:\s*([^\n]+)\.`)
	for _, l := range lines {
		matches := re.FindStringSubmatch(l)
		if len(matches) > 1 {
			handled = append(handled, strings.Split(matches[1], ", ")...)
		}
	}

	return handled
}

// Returns version of installed podman
func GetPodmanVersion() string {
	cmd := exec.Command("podman", "version", "--format", "json")
	out, err := cmd.Output()
	Expect(err).ToNot(HaveOccurred(), func() string {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		return err.Error()
	})
	var result podman.SystemVersionReport
	err = json.Unmarshal(out, &result)
	Expect(err).ToNot(HaveOccurred())
	return result.Client.Version
}
