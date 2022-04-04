package kclient

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

// TryWithBlockOwnerDeletion will execute `exec` a first time with `BlockOwnerDeletion` set to true in `ownerReference`
// If a Forbidden errors occurs, it will call `exec` again with the original `ownerReference`
func (c *Client) TryWithBlockOwnerDeletion(ownerReference metav1.OwnerReference, exec func(ownerReference metav1.OwnerReference) error) error {
	blockOwnerRef := ownerReference
	blockOwnerRef.BlockOwnerDeletion = pointer.BoolPtr(true)
	err := exec(blockOwnerRef)
	if err == nil {
		return nil
	}
	if apierrors.IsForbidden(err) {
		return exec(ownerReference)
	}
	return err
}
