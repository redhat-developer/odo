package servicebindingrequest

import (
	"fmt"
	"sort"
	"strings"

	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/annotations"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

// OLM represents the actions this operator needs to take upon Operator-Lifecycle-Manager resources,
// like ClusterServiceVersions (CSV) and CRDDescriptions.
type OLM struct {
	client dynamic.Interface // kubernetes dynamic client
	ns     string            // namespace
	logger *log.Log          // logger instance
}

const (
	csvResource = "clusterserviceversions"
)

var (
	olmLog = log.NewLog("olm")
)

// listCSVs simple list to all CSV objects in the cluster.
func (o *OLM) listCSVs() ([]unstructured.Unstructured, error) {
	log := o.logger
	gvr := olmv1alpha1.SchemeGroupVersion.WithResource(csvResource)
	resourceClient := o.client.Resource(gvr).Namespace(o.ns)
	csvs, err := resourceClient.List(metav1.ListOptions{})
	if err != nil && errors.IsNotFound(err) {
		log.Warning("ClusterServiceVersions CRD is not installed")
		return nil, nil
	} else if err != nil {
		log.Error(err, "during listing CSV objects from cluster")
		return nil, err
	}
	return csvs.Items, nil
}

// extractOwnedCRDs from a list of CSV objects.
func (o *OLM) extractOwnedCRDs(
	csvs []unstructured.Unstructured,
) ([]unstructured.Unstructured, error) {
	crds := []unstructured.Unstructured{}
	for _, csv := range csvs {
		ownedPath := []string{"spec", "customresourcedefinitions", "owned"}
		log := o.logger.WithValues("OwnedPath", ownedPath, "CSV.Name", csv.GetName())

		ownedCRDs, exists, err := unstructured.NestedSlice(csv.Object, ownedPath...)
		if err != nil {
			log.Error(err, "on extracting nested slice")
			return nil, err
		}
		if !exists {
			continue
		}

		for _, crd := range ownedCRDs {
			data := crd.(map[string]interface{})
			crds = append(crds, unstructured.Unstructured{Object: data})
		}
	}

	return crds, nil
}

// ListCSVOwnedCRDs return a unstructured list of CRD objects from "owned" section in CSVs.
func (o *OLM) ListCSVOwnedCRDs() ([]unstructured.Unstructured, error) {
	log := o.logger
	csvs, err := o.listCSVs()
	if err != nil {
		log.Error(err, "on listting CSVs")
		return nil, err
	}
	return o.extractOwnedCRDs(csvs)
}

// eachCRDDescriptionFn function to be called against each CRDDescription in a slice.
type eachCRDDescriptionFn func(crdDescription *olmv1alpha1.CRDDescription)

// loopCRDDescriptions takes a list of CRDDescriptions (extracted from "owned") and converts to a
// actual type instance, before calling out for informed function. This method can return error in
// case of issues to convert unstructured into CRDDescription.
func (o *OLM) loopCRDDescriptions(
	crdDescriptions []unstructured.Unstructured,
	fn eachCRDDescriptionFn,
) error {
	for _, u := range crdDescriptions {
		log := o.logger
		crdDescription := &olmv1alpha1.CRDDescription{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, crdDescription)
		if err != nil {
			log.Error(err, "on converting from unstructured to CRD")
			return err
		}
		log.Debug("Inspecting CRDDescription...", "CRDDescription", crdDescription)
		if crdDescription.Name == "" {
			log.Debug("Skipping empty CRDDescription!")
			continue
		}
		fn(crdDescription)
	}
	return nil
}

// SelectCRDByGVK return a single CRD based on a given GVK.
func (o *OLM) SelectCRDByGVK(gvk schema.GroupVersionKind, crd *unstructured.Unstructured) (*olmv1alpha1.CRDDescription, error) {
	log := o.logger.WithValues("Selector.GVK", gvk)
	ownedCRDs, err := o.ListCSVOwnedCRDs()
	if err != nil {
		log.Error(err, "on listing owned CRDs")
		return nil, err
	}

	var crdDescription *olmv1alpha1.CRDDescription

	// CRDDescription is both used by OLM to configure OLM descriptors in manifests existing in the
	// cluster but is also built from annotations present in the CRD
	if crd != nil {
		crdDescription, err = buildCRDDescriptionFromCRD(crd)
		if err != nil {
			return nil, err
		}
	}

	crdDescriptions := []*olmv1alpha1.CRDDescription{}

	err = o.loopCRDDescriptions(ownedCRDs, func(crdDescription *olmv1alpha1.CRDDescription) {
		log = o.logger.WithValues(
			"CRDDescription.Name", crdDescription.Name,
			"CRDDescription.Version", crdDescription.Version,
			"CRDDescription.Kind", crdDescription.Kind,
		)
		log.Debug("Inspecting CRDDescription object...")
		// checking for suffix since is expected to have object type as prefix
		if !strings.EqualFold(strings.ToLower(crdDescription.Kind), strings.ToLower(gvk.Kind)) {
			return
		}
		// matching resource version, unless when not informed
		if crdDescription.Version != "" &&
			!strings.EqualFold(strings.ToLower(gvk.Version), strings.ToLower(crdDescription.Version)) {
			return
		}
		log.Debug("CRDDescription object matches selector!")
		crdDescriptions = append(crdDescriptions, crdDescription)
	})
	if err != nil {
		return nil, err
	}

	if len(crdDescriptions) == 0 && crdDescription == nil {
		log.Debug("No CRD could be found for GVK.")
		return nil, fmt.Errorf("no crd could be found for gvk")
	} else if len(crdDescriptions) == 0 {
		// use the crdDescription built from CRD annotations as fallback
		return crdDescription, nil
	}

	return crdDescriptions[0], nil
}

// buildCRDDescriptionFromCR builds a CRDDescription from annotations present in the CR.
func buildCRDDescriptionFromCR(cr *unstructured.Unstructured, crdDescription *olmv1alpha1.CRDDescription) error {
	var (
		err error
	)

	gvk := schema.GroupVersionKind{
		Kind:    cr.GetKind(),
		Version: cr.GroupVersionKind().Version,
		Group:   cr.GroupVersionKind().Group,
	}
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)

	crdDescription.Name = gvr.Resource + "." + gvr.Group
	crdDescription.Kind = cr.GetKind()
	crdDescription.Version = cr.GroupVersionKind().Version

	specDescriptors, statusDescriptors, err := buildDescriptorsFromAnnotations(cr.GetAnnotations())
	if err != nil {
		return err
	}

	crdDescription.SpecDescriptors = append(crdDescription.SpecDescriptors, specDescriptors...)
	crdDescription.StatusDescriptors = append(crdDescription.StatusDescriptors, statusDescriptors...)

	return nil
}

// buildCRDDescriptionFromCRD builds a CRDDescription from annotations present in the CRD.
func buildCRDDescriptionFromCRD(crd *unstructured.Unstructured) (*olmv1alpha1.CRDDescription, error) {
	var (
		ok  bool
		err error
	)

	crdDescription := &olmv1alpha1.CRDDescription{
		Name: crd.GetName(),
	}

	crdDescription.Kind, ok, err = unstructured.NestedString(crd.Object, "spec", "names", "kind")
	if err != nil || !ok {
		return nil, err
	}

	crdDescription.Version, ok, err = unstructured.NestedString(crd.Object, "spec", "version")
	if err != nil || !ok {
		return nil, err
	}

	specDescriptors, statusDescriptors, err := buildDescriptorsFromAnnotations(crd.GetAnnotations())
	if err != nil {
		return nil, err
	}
	crdDescription.SpecDescriptors = append(crdDescription.SpecDescriptors, specDescriptors...)
	crdDescription.StatusDescriptors = append(crdDescription.StatusDescriptors, statusDescriptors...)

	return crdDescription, nil
}

// buildDescriptorsFromAnnotations builds two descriptors collection, one for spec descriptors and
// another for status descriptors.
func buildDescriptorsFromAnnotations(in map[string]string) (
	[]olmv1alpha1.SpecDescriptor,
	[]olmv1alpha1.StatusDescriptor,
	error,
) {
	var specDescriptors []olmv1alpha1.SpecDescriptor
	var statusDescriptors []olmv1alpha1.StatusDescriptor

	acc := make(map[string][]string)

	for n, v := range in {
		// Iterate all annotations and compose related Spec and Status descriptors, where those
		// descriptors should be grouped by field path.	So, for example, the "status.dbCredentials"
		// field path should accumulate all related annotations, so the StatusDescriptor referring
		// "status.dbCredentials" have both "user" and "password" XDescriptors.

		bindingInfo, err := annotations.NewBindingInfo(n, v)
		if err != nil {
			// continue to the next annotation if an error is returned
			continue
		}

		descriptors, exists := acc[bindingInfo.ResourceReferencePath]
		if !exists {
			descriptors = make([]string, 0)
		}
		descriptors = append(descriptors, bindingInfo.Descriptor)
		acc[bindingInfo.ResourceReferencePath] = descriptors
	}

	// create the status and/or spec descriptors based on the
	for fieldPath, descriptors := range acc {
		sort.Strings(descriptors)
		path := strings.Split(fieldPath, ".")
		if path[0] == "status" {
			statusDescriptors = append(statusDescriptors, olmv1alpha1.StatusDescriptor{
				Path:         path[1],
				XDescriptors: descriptors,
			})
		} else if path[0] == "spec" {
			specDescriptors = append(specDescriptors, olmv1alpha1.SpecDescriptor{
				Path:         path[1],
				XDescriptors: descriptors,
			})
		}
	}

	return specDescriptors, statusDescriptors, nil
}

// extractGVKs loop owned objects and extract the GVK information from them.
func (o *OLM) extractGVKs(
	crdDescriptions []unstructured.Unstructured,
) ([]schema.GroupVersionKind, error) {
	log := o.logger
	gvks := []schema.GroupVersionKind{}
	err := o.loopCRDDescriptions(crdDescriptions, func(crdDescription *olmv1alpha1.CRDDescription) {
		log.Debug("Extracting GVK from CRDDescription", "CRDDescription.Name", crdDescription.Name)
		_, gv := schema.ParseResourceArg(crdDescription.Name)
		gvks = append(gvks, schema.GroupVersionKind{
			Group:   gv.Group,
			Version: crdDescription.Version,
			Kind:    crdDescription.Kind,
		})
	})
	if err != nil {
		return []schema.GroupVersionKind{}, err
	}
	return gvks, nil
}

// ListCSVOwnedCRDsAsGVKs return the list of owned CRDs from all CSV objects as a list of GVKs.
func (o *OLM) ListCSVOwnedCRDsAsGVKs() ([]schema.GroupVersionKind, error) {
	log := o.logger
	ownedCRDs, err := o.ListCSVOwnedCRDs()
	if err != nil {
		log.Error(err, "on listting CSVs")
		return nil, err
	}
	return o.extractGVKs(ownedCRDs)
}

// ListGVKsFromCSVNamespacedName return the list of owned GVKs for a given CSV namespaced named.
func (o *OLM) ListGVKsFromCSVNamespacedName(
	namespacedName types.NamespacedName,
) ([]schema.GroupVersionKind, error) {
	log := o.logger.WithValues("CSV.NamespacedName", namespacedName)
	log.Debug("Reading CSV to extract GVKs...")
	gvr := olmv1alpha1.SchemeGroupVersion.WithResource(csvResource)
	resourceClient := o.client.Resource(gvr).Namespace(namespacedName.Namespace)
	u, err := resourceClient.Get(namespacedName.Name, metav1.GetOptions{})
	if err != nil {
		// the CSV might have disappeared between discovery and check, so not found is not an error
		if errors.IsNotFound(err) {
			return []schema.GroupVersionKind{}, nil
		}
		log.Error(err, "on reading CSV object")
		return []schema.GroupVersionKind{}, err
	}

	unstructuredCSV := *u
	csvs := []unstructured.Unstructured{unstructuredCSV}

	ownedCRDs, err := o.extractOwnedCRDs(csvs)
	if err != nil {
		log.Error(err, "on extracting owned CRDs")
		return []schema.GroupVersionKind{}, err
	}

	return o.extractGVKs(ownedCRDs)
}

// NewOLM instantiate a new OLM.
func NewOLM(client dynamic.Interface, ns string) *OLM {
	return &OLM{
		client: client,
		ns:     ns,
		logger: olmLog,
	}
}
