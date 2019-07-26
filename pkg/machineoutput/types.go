package machineoutput

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Error for machine readable output error messages
type Error struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Message           string `json:"message"`
}

// Success same as above, but copy-and-pasted just in case
// we change the output in the future
type Success struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Message           string `json:"message"`
}
