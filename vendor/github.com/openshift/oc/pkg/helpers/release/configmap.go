package release

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/crypto/openpgp"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// ReleaseAnnotationConfigMapVerifier is an annotation set on a config map in the
// release payload to indicate that this config map controls signing for the payload.
// Only the first config map within the payload should be used, regardless of whether
// it has data. See NewFromConfigMapData for more.
const ReleaseAnnotationConfigMapVerifier = "release.openshift.io/verification-config-map"

// NamespaceLabelConfigMap is the Namespace label applied to a configmap
// containing signatures.
const NamespaceLabelConfigMap = "openshift-config-managed"

// ReleaseLabelConfigMap is a label applied to a configmap inside the
// NamespaceLabelConfigMap namespace that indicates it contains signatures
// for release image digests. Any binaryData key that starts with the digest
// is added to the list of signatures checked.
const ReleaseLabelConfigMap = "release.openshift.io/verification-signatures"

// digestToKeyPrefix changes digest to use '-' in place of ':',
// {algo}-{hash} instead of {algo}:{hash}, because colons are not
// allowed in ConfigMap keys.
func digestToKeyPrefix(digest string) (string, error) {
	parts := strings.SplitN(digest, ":", 3)
	if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) == 0 {
		return "", fmt.Errorf("the provided digest must be of the form ALGO:HASH")
	}
	algo, hash := parts[0], parts[1]
	return fmt.Sprintf("%s-%s", algo, hash), nil
}

// GetSignaturesAsConfigmap returns the given signatures in a configmap. Uses
// digestToKeyPrefix to replace colon with dash when saving digest to configmap.
func GetSignaturesAsConfigmap(digest string, signatures [][]byte) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: NamespaceLabelConfigMap,
			Labels: map[string]string{
				ReleaseLabelConfigMap: "",
			},
		},
		BinaryData: make(map[string][]byte),
	}
	prefix, err := digestToKeyPrefix(digest)
	if err != nil {
		return nil, err
	}
	cm.Name = prefix
	for i := 0; i < len(signatures); i++ {
		cm.BinaryData[fmt.Sprintf("%s-%d", prefix, i+1)] = signatures[i]
	}
	return cm, nil
}

// NewFromConfigMapData expects to receive the data field of the first config map in the release
// image payload with the annotation "release.openshift.io/verification-config-map". Only the
// first payload item in lexographic order will be considered - all others are ignored. The
// verifier returned by this method
//
// The presence of one or more config maps instructs the CVO to verify updates before they are
// downloaded.
//
// The keys within the config map in the data field define how verification is performed:
//
// verifier-public-key-*: One or more GPG public keys in ASCII form that must have signed the
//                        release image by digest.
//
// store-*: A URL (scheme file://, http://, or https://) location that contains signatures. These
//          signatures are in the atomic container signature format. The URL will have the digest
//          of the image appended to it as "<STORE>/<ALGO>=<DIGEST>/signature-<NUMBER>" as described
//          in the container image signing format. The docker-image-manifest section of the
//          signature must match the release image digest. Signatures are searched starting at
//          NUMBER 1 and incrementing if the signature exists but is not valid. The signature is a
//          GPG signed and encrypted JSON message. The file store is provided for testing only at
//          the current time, although future versions of the CVO might allow host mounting of
//          signatures.
//
// See https://github.com/containers/image/blob/ab49b0a48428c623a8f03b41b9083d48966b34a9/docs/signature-protocols.md
// for a description of the signature store
//
// The returned verifier will require that any new release image will only be considered verified
// if each provided public key has signed the release image digest. The signature may be in any
// store and the lookup order is internally defined.
func NewFromConfigMapData(src string, data map[string]string, clientBuilder ClientBuilder) (*ReleaseVerifier, error) {
	verifiers := make(map[string]openpgp.EntityList)
	var stores []*url.URL
	for k, v := range data {
		switch {
		case strings.HasPrefix(k, "verifier-public-key-"):
			keyring, err := loadArmoredOrUnarmoredGPGKeyRing([]byte(v))
			if err != nil {
				return nil, errors.Wrapf(err, "%s has an invalid key %q that must be a GPG public key: %v", src, k, err)
			}
			verifiers[k] = keyring
		case strings.HasPrefix(k, "store-"):
			v = strings.TrimSpace(v)
			u, err := url.Parse(v)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "file") {
				return nil, fmt.Errorf("%s has an invalid key %q: must be a valid URL with scheme file://, http://, or https://", src, k)
			}
			stores = append(stores, u)
		default:
			klog.Warningf("An unexpected key was found in %s and will be ignored (expected store-* or verifier-public-key-*): %s", src, k)
		}
	}
	if len(stores) == 0 {
		return nil, fmt.Errorf("%s did not provide any signature stores to read from and cannot be used", src)
	}
	if len(verifiers) == 0 {
		return nil, fmt.Errorf("%s did not provide any GPG public keys to verify signatures from and cannot be used", src)
	}

	return NewReleaseVerifier(verifiers, stores, clientBuilder), nil
}

func loadArmoredOrUnarmoredGPGKeyRing(data []byte) (openpgp.EntityList, error) {
	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(data))
	if err == nil {
		return keyring, nil
	}
	return openpgp.ReadKeyRing(bytes.NewReader(data))
}
