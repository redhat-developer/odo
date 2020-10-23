package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// convertEnvs converts environment variables from the devfile structure to kubernetes structure
func convertEnvs(vars []common.Env) []corev1.EnvVar {
	kVars := []corev1.EnvVar{}
	for _, env := range vars {
		kVars = append(kVars, corev1.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		})
	}
	return kVars
}

// convertPorts converts endpoint variables from the devfile structure to kubernetes ContainerPort
func convertPorts(endpoints []common.Endpoint) ([]corev1.ContainerPort, error) {
	containerPorts := []corev1.ContainerPort{}
	for _, endpoint := range endpoints {
		name := strings.TrimSpace(util.GetDNS1123Name(strings.ToLower(endpoint.Name)))
		name = util.TruncateString(name, 15)
		for _, c := range containerPorts {
			if c.Name == endpoint.Name {
				// the name has to be unique within a single container since it is considered as the URL name
				return nil, fmt.Errorf("devfile contains multiple endpoint entries with same name: %v", endpoint.Name)
			}
		}
		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name:          name,
			ContainerPort: endpoint.TargetPort,
		})
	}
	return containerPorts, nil
}

// getResourceReqs creates a kubernetes ResourceRequirements object based on resource requirements set in the devfile
func getResourceReqs(comp common.DevfileComponent) corev1.ResourceRequirements {
	reqs := corev1.ResourceRequirements{}
	limits := make(corev1.ResourceList)
	if comp.Container.MemoryLimit != "" {
		memoryLimit, err := resource.ParseQuantity(comp.Container.MemoryLimit)
		if err == nil {
			limits[corev1.ResourceMemory] = memoryLimit
		}
		reqs.Limits = limits
	}
	return reqs
}

// GetSyncFolder gets the sync folder path for source code.
// sourceVolumePath: mount path of the empty dir volume the odo uses to sync source code
// projects: list of projects from devfile
func GetSyncFolder(sourceVolumePath string, projects []common.DevfileProject) (string, error) {
	// if there are no projects in the devfile, source would be synced to $PROJECTS_ROOT
	if len(projects) == 0 {
		return sourceVolumePath, nil
	}

	// if there is one or more projects in the devfile, get the first project and check its clonepath
	project := projects[0]

	if project.ClonePath != "" {
		if strings.HasPrefix(project.ClonePath, "/") {
			return "", fmt.Errorf("the clonePath %s in the devfile project %s must be a relative path", project.ClonePath, project.Name)
		}
		if strings.Contains(project.ClonePath, "..") {
			return "", fmt.Errorf("the clonePath %s in the devfile project %s cannot escape the value defined by $PROJECTS_ROOT. Please avoid using \"..\" in clonePath", project.ClonePath, project.Name)
		}
		// If clonepath exist source would be synced to $PROJECTS_ROOT/clonePath
		return filepath.ToSlash(filepath.Join(sourceVolumePath, project.ClonePath)), nil
	}
	// If clonepath does not exist source would be synced to $PROJECTS_ROOT/projectName
	return filepath.ToSlash(filepath.Join(sourceVolumePath, project.Name)), nil

}
