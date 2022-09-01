package watch

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/redhat-developer/odo/pkg/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodPhases map[metav1.Time]corev1.PodPhase

func NewPodPhases() PodPhases {
	return map[metav1.Time]corev1.PodPhase{}
}

func (o *PodPhases) Add(out io.Writer, k metav1.Time, pod *corev1.Pod) {
	v := pod.Status.Phase
	if pod.GetDeletionTimestamp() != nil {
		v = "Terminating"
	}
	display := false
	if (*o)[k] != v {
		display = true
	}
	(*o)[k] = v
	if display {
		o.Display(out)
	}
}

func (o *PodPhases) Delete(out io.Writer, pod *corev1.Pod) {
	k := pod.GetCreationTimestamp()
	if _, ok := (*o)[k]; ok {
		delete(*o, k)
		o.Display(out)
	}
}

func (o PodPhases) Display(out io.Writer) {

	if len(o) == 0 {
		log.Fwarning(out, "No pod exists [Background task]")
		return
	}

	keys := make([]metav1.Time, 0, len(o))
	for k := range o {
		keys = append(keys, k)
	}

	if len(keys) == 1 {
		phase := o[keys[0]]
		if phase == corev1.PodRunning {
			log.Fsuccess(out, fmt.Sprintf("Pod is %s [Background task]", phase))
			return
		}
		log.Fwarning(out, fmt.Sprintf("Pod is %s [Background task]", phase))
		return
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Before(&keys[j])
	})

	values := make([]string, 0, len(o))
	for _, k := range keys {
		values = append(values, string(o[k]))
	}
	log.Fwarning(out, fmt.Sprintf("Pods are %s [Background task]", strings.Join(values, ", ")))
}
