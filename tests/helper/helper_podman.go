package helper

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/pkg/podman"
)

func getBooleanValueFromEnvVar(envvar string, defaultValue bool) bool {
	strVal := os.Getenv("PODMAN_USE_NAMESPACES")
	boolValue, err := strconv.ParseBool(strVal)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

func GenerateAndSetContainersConf(dir string) {
	useNamespaces := getBooleanValueFromEnvVar("PODMAN_USE_NAMESPACES", true)
	if !useNamespaces {
		return
	}
	ns := GenerateProjectName()
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
	// ⚠  Apply Kubernetes/Openshift components are not supported on Podman. Skipping: k8s-deploybydefault-true-and-referenced.
	re := regexp.MustCompile(`(?:Kubernetes|OpenShift|Kubernetes/Openshift) components are not supported on Podman\.\s*Skipping:\s*([^\n]+)\.`)
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

// GenerateDelayedPodman returns a podman cmd that sleeps for delaySecond before responding;
// this function is usually used in combination with PODMAN_CMD_INIT_TIMEOUT odo preference
func GenerateDelayedPodman(commonVarContext string, delaySecond int) string {
	delayer := filepath.Join(commonVarContext, "podman-cmd-delayer")
	fileContent := fmt.Sprintf(`#!/bin/bash

echo Delaying command execution... >&2
sleep %d
echo "$@"
`, delaySecond)
	err := CreateFileWithContentAndPerm(delayer, fileContent, 0755)
	Expect(err).ToNot(HaveOccurred())
	return delayer
}
