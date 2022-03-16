package kclient

import (
	"context"
	"errors"
	"fmt"

	"github.com/devfile/library/pkg/devfile/generator"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// constants for volumes
const (
	PersistentVolumeClaimKind       = "PersistentVolumeClaim"
	PersistentVolumeClaimAPIVersion = "v1"
)

// CreatePVC creates a PVC resource in the cluster with the given name, size and labels
func (c *Client) CreatePVC(pvc corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	createdPvc, err := c.KubeClient.CoreV1().PersistentVolumeClaims(c.Namespace).Create(context.TODO(), &pvc, metav1.CreateOptions{FieldManager: FieldManager})
	if err != nil {
		return nil, fmt.Errorf("unable to create PVC: %w", err)
	}
	return createdPvc, nil
}

// DeletePVC deletes the required PVC resource from the cluster
func (c *Client) DeletePVC(pvcName string) error {
	return c.KubeClient.CoreV1().PersistentVolumeClaims(c.Namespace).Delete(context.TODO(), pvcName, metav1.DeleteOptions{})
}

// ListPVCs returns the PVCs based on the given selector
func (c *Client) ListPVCs(selector string) ([]corev1.PersistentVolumeClaim, error) {
	pvcList, err := c.KubeClient.CoreV1().PersistentVolumeClaims(c.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get PVCs for selector: %v: %w", selector, err)
	}

	return pvcList.Items, nil
}

// ListPVCNames returns the PVC names for the given selector
func (c *Client) ListPVCNames(selector string) ([]string, error) {
	pvcs, err := c.ListPVCs(selector)
	if err != nil {
		return nil, fmt.Errorf("unable to get PVCs from selector: %w", err)
	}

	var names []string
	for _, pvc := range pvcs {
		names = append(names, pvc.Name)
	}

	return names, nil
}

// GetPVCFromName returns the PVC of the given name
func (c *Client) GetPVCFromName(pvcName string) (*corev1.PersistentVolumeClaim, error) {
	return c.KubeClient.CoreV1().PersistentVolumeClaims(c.Namespace).Get(context.TODO(), pvcName, metav1.GetOptions{})
}

// UpdatePVCLabels updates the given PVC with the given labels
func (c *Client) UpdatePVCLabels(pvc *corev1.PersistentVolumeClaim, labels map[string]string) error {
	pvc.Labels = labels
	_, err := c.KubeClient.CoreV1().PersistentVolumeClaims(c.Namespace).Update(context.TODO(), pvc, metav1.UpdateOptions{FieldManager: FieldManager})
	if err != nil {
		return fmt.Errorf("unable to remove storage label from PVC: %w", err)
	}
	return nil
}

// GetAndUpdateStorageOwnerReference updates the given storage with the given owner references
func (c *Client) GetAndUpdateStorageOwnerReference(pvc *corev1.PersistentVolumeClaim, ownerReference ...metav1.OwnerReference) error {
	if len(ownerReference) <= 0 {
		return errors.New("owner references are empty")
	}
	// get the latest version of the PVC to avoid conflict errors
	latestPVC, err := c.KubeClient.CoreV1().PersistentVolumeClaims(c.Namespace).Get(context.TODO(), pvc.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	for _, owRf := range ownerReference {
		latestPVC.SetOwnerReferences(append(pvc.GetOwnerReferences(), owRf))
	}
	_, err = c.KubeClient.CoreV1().PersistentVolumeClaims(c.Namespace).Update(context.TODO(), latestPVC, metav1.UpdateOptions{FieldManager: FieldManager})
	if err != nil {
		return err
	}
	return nil
}

// UpdateStorageOwnerReference updates the given storage with the given owner references
func (c *Client) UpdateStorageOwnerReference(pvc *corev1.PersistentVolumeClaim, ownerReference ...metav1.OwnerReference) error {
	if len(ownerReference) <= 0 {
		return errors.New("owner references are empty")
	}

	updatedPVC := generator.GetPVC(generator.PVCParams{
		ObjectMeta: pvc.ObjectMeta,
		TypeMeta:   pvc.TypeMeta,
	})

	updatedPVC.OwnerReferences = ownerReference
	updatedPVC.Spec = pvc.Spec

	_, err := c.KubeClient.CoreV1().PersistentVolumeClaims(c.Namespace).Update(context.TODO(), updatedPVC, metav1.UpdateOptions{FieldManager: FieldManager})
	if err != nil {
		return err
	}
	return nil
}
