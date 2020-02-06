package pipelines

import "log"

// Bootstrap is the main driver for getting OpenShift pipelines for GitOps
// configured with a basic configuration.
func Bootstrap(quayUsername, baseRepo, prefix string) error {
	log.Printf("Bootstrapping %s, %s, %#v", quayUsername, baseRepo, prefix)
	return nil
}
