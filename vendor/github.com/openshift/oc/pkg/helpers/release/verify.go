package release

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/openpgp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	coreclientsetv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog"

	"github.com/openshift/library-go/pkg/manifest"
)

// Interface performs verification of the provided content. The default implementation
// in this package uses the container signature format defined at https://github.com/containers/image
// to authenticate that a given release image digest has been signed by a trusted party.
type Interface interface {
	// Verify should return nil if the provided release digest has suffient signatures to be considered
	// valid. It should return an error in all other cases.
	Verify(ctx context.Context, releaseDigest string) error

	Signatures() map[string][][]byte
}

// SignatureStore retrieves signatures for a digest or returns an error. It requires String()
// for describing the store.
type SignatureStore interface {
	// DigestSignatures returns zero or more signatures for the provided digest, or returns an
	// error if no signatures could be retrieved.
	DigestSignatures(ctx context.Context, digest string) ([][]byte, error)
	// String should return a description of where this store finds signatures in a short
	// clause intended for display in a description of the verifier.
	String() string
}

type rejectVerifier struct{}

func (rejectVerifier) Verify(ctx context.Context, releaseDigest string) error {
	return fmt.Errorf("verification is not possible")
}

func (rejectVerifier) Signatures() map[string][][]byte {
	return nil
}

// Reject fails always fails verification.
var Reject Interface = rejectVerifier{}

// ClientBuilder provides a method for generating an HTTP Client configured
// with cluster proxy settings, if they exist.
type ClientBuilder interface {
	// HTTPClient returns a client suitable for retrieving signatures. It is not
	// required to be unique per call, but may be called concurrently.
	HTTPClient() (*http.Client, error)
}

// DefaultClient uses the default http.Client for accessing signatures.
var DefaultClient = simpleClientBuilder{}

// simpleClientBuilder implements the ClientBuilder interface and may be used for testing.
type simpleClientBuilder struct{}

// HTTPClient from simpleClientBuilder creates an http.Client with no configuration.
func (s simpleClientBuilder) HTTPClient() (*http.Client, error) {
	return &http.Client{}, nil
}

// maxSignatureSearch prevents unbounded recursion on malicious signature stores (if
// an attacker was able to take ownership of the store to perform DoS on clusters).
const maxSignatureSearch = 10

// validReleaseDigest is a verification rule to filter clearly invalid digests.
var validReleaseDigest = regexp.MustCompile(`^[a-zA-Z0-9:]+$`)

// ReleaseVerifier implements a signature intersection operation on a provided release
// digest - all verifiers must have at least one valid signature attesting the release
// digest. If any failure occurs the caller should assume the content is unverified.
type ReleaseVerifier struct {
	verifiers map[string]openpgp.EntityList

	locations     []*url.URL
	clientBuilder ClientBuilder
	stores        []SignatureStore

	lock           sync.Mutex
	signatureCache map[string][][]byte
}

// NewReleaseVerifier creates a release verifier for the provided inputs.
func NewReleaseVerifier(verifiers map[string]openpgp.EntityList, locations []*url.URL, clientBuilder ClientBuilder) *ReleaseVerifier {
	return &ReleaseVerifier{
		verifiers:     verifiers,
		locations:     locations,
		clientBuilder: clientBuilder,

		signatureCache: make(map[string][][]byte),
	}
}

// WithStores copies the provided verifier and adds any provided stores to the list.
func (v *ReleaseVerifier) WithStores(stores ...SignatureStore) *ReleaseVerifier {
	return &ReleaseVerifier{
		verifiers:     v.verifiers,
		locations:     v.locations,
		clientBuilder: v.clientBuilder,

		stores:         append(append(make([]SignatureStore, 0, len(v.stores)+len(stores)), v.stores...), stores...),
		signatureCache: v.Signatures(),
	}
}

// Verifiers returns a copy of the verifiers in this payload.
func (v *ReleaseVerifier) Verifiers() map[string]openpgp.EntityList {
	out := make(map[string]openpgp.EntityList, len(v.verifiers))
	for k, v := range v.verifiers {
		out[k] = v
	}
	return out
}

// String summarizes the verifier for human consumption
func (v *ReleaseVerifier) String() string {
	var keys []string
	for name := range v.verifiers {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	var builder strings.Builder
	builder.Grow(256)
	fmt.Fprintf(&builder, "All release image digests must have GPG signatures from")
	if len(keys) == 0 {
		fmt.Fprint(&builder, " <ERROR: no verifiers>")
	}
	for _, name := range keys {
		verifier := v.verifiers[name]
		fmt.Fprintf(&builder, " %s (", name)
		for i, entity := range verifier {
			if i != 0 {
				fmt.Fprint(&builder, ", ")
			}
			if entity.PrimaryKey != nil {
				fmt.Fprintf(&builder, strings.ToUpper(fmt.Sprintf("%x", entity.PrimaryKey.Fingerprint)))
				fmt.Fprint(&builder, ": ")
			}
			count := 0
			for identityName := range entity.Identities {
				if count != 0 {
					fmt.Fprint(&builder, ", ")
				}
				fmt.Fprintf(&builder, "%s", identityName)
				count++
			}
		}
		fmt.Fprint(&builder, ")")
	}

	hasLocations := len(v.locations) > 0
	hasStores := len(v.stores) > 0
	if hasLocations || hasStores {
		fmt.Fprintf(&builder, " - will check for signatures in containers/image format")
		if hasLocations {
			fmt.Fprintf(&builder, " at")
			for i, location := range v.locations {
				if i != 0 {
					fmt.Fprint(&builder, ",")
				}
				fmt.Fprintf(&builder, " %s", location.String())
			}
		}
		if hasStores {
			if hasLocations {
				fmt.Fprintf(&builder, " and from")
			} else {
				fmt.Fprintf(&builder, " from")
			}
			for i, store := range v.stores {
				if i != 0 {
					fmt.Fprint(&builder, ",")
				}
				fmt.Fprintf(&builder, " %s", store.String())
			}
		}
	} else {
		fmt.Fprintf(&builder, " - <ERROR: no locations or stores>")
	}
	return builder.String()
}

// Verify ensures that at least one valid signature exists for an image with digest
// matching release digest in any of the provided locations for all verifiers, or returns
// an error.
func (v *ReleaseVerifier) Verify(ctx context.Context, releaseDigest string) error {
	if len(v.verifiers) == 0 || (len(v.locations) == 0 && len(v.stores) == 0) {
		return fmt.Errorf("the release verifier is incorrectly configured, unable to verify digests")
	}
	if len(releaseDigest) == 0 {
		return fmt.Errorf("release images that are not accessed via digest cannot be verified")
	}
	if !validReleaseDigest.MatchString(releaseDigest) {
		return fmt.Errorf("the provided release image digest contains prohibited characters")
	}

	if v.hasVerified(releaseDigest) {
		return nil
	}

	parts := strings.SplitN(releaseDigest, ":", 3)
	if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) == 0 {
		return fmt.Errorf("the provided release image digest must be of the form ALGO:HASH")
	}
	algo, hash := parts[0], parts[1]
	name := fmt.Sprintf("%s=%s", algo, hash)

	remaining := make(map[string]openpgp.EntityList, len(v.verifiers))
	for k, v := range v.verifiers {
		remaining[k] = v
	}

	var signedWith [][]byte
	verifier := func(path string, signature []byte) (bool, error) {
		for k, keyring := range remaining {
			content, _, err := verifySignatureWithKeyring(bytes.NewReader(signature), keyring)
			if err != nil {
				klog.Infof("keyring %q could not verify signature: %v", k, err)
				klog.V(4).Infof("keyring %q could not verify signature: %v", k, err)
				continue
			}
			if err := verifyAtomicContainerSignature(content, releaseDigest); err != nil {
				klog.Infof("signature %q is not valid: %v", path, err)
				klog.V(4).Infof("signature %q is not valid: %v", path, err)
				continue
			}
			delete(remaining, k)
			signedWith = append(signedWith, signature)
		}
		return len(remaining) > 0, nil
	}

	// check the stores to see if they match any signatures
	for i, store := range v.stores {
		if len(remaining) == 0 {
			break
		}
		signatures, err := store.DigestSignatures(ctx, releaseDigest)
		if err != nil {
			klog.Infof("store %s could not load signatures: %v", store, err)
			klog.V(4).Infof("store %s could not load signatures: %v", store, err)
			continue
		}
		for _, signature := range signatures {
			hasUnsigned, err := verifier(fmt.Sprintf("store %d", i), signature)
			if err != nil {
				return err
			}
			if !hasUnsigned {
				break
			}
		}
	}

	var client *http.Client
	for _, location := range v.locations {
		if len(remaining) == 0 {
			break
		}
		switch location.Scheme {
		case "file":
			dir := filepath.Join(location.Path, name)
			if err := checkFileSignatures(ctx, dir, maxSignatureSearch, verifier); err != nil {
				return err
			}
		case "http", "https":
			if client == nil {
				var err error
				client, err = v.clientBuilder.HTTPClient()
				if err != nil {
					return err
				}
			}

			copied := *location
			copied.Path = path.Join(location.Path, name)
			if err := checkHTTPSignatures(ctx, client, copied, maxSignatureSearch, verifier); err != nil {
				return err
			}
		default:
			return fmt.Errorf("internal error: the store %s type is unrecognized, cannot verify signatures", location)
		}
	}

	if len(remaining) > 0 {
		if klog.V(4) {
			for k := range remaining {
				klog.Infof("Unable to verify %s against keyring %s", releaseDigest, k)
			}
		}
		return fmt.Errorf("unable to locate a valid signature for one or more sources")
	}

	v.cacheVerification(releaseDigest, signedWith)

	return nil
}

// Signatures returns a copy of any cached signatures that have been validated
// so far. It may return no signatures.
func (v *ReleaseVerifier) Signatures() map[string][][]byte {
	copied := make(map[string][][]byte)
	v.lock.Lock()
	defer v.lock.Unlock()
	for k, v := range v.signatureCache {
		copied[k] = v
	}
	return copied
}

// LoadConfigMapVerifierDataFromUpdate fetches the first config map in the payload with the correct annotation.
// It returns an error if the data is not valid, or no verifier if no config map is found. See the verify
// package for more details on the algorithm for verification. If the annotation is set, a verifier or error
// is always returned.
func LoadConfigMapVerifierDataFromUpdate(manifests []manifest.Manifest, clientBuilder ClientBuilder, configMapClient coreclientsetv1.ConfigMapsGetter) (Interface /**StorePersister,*/, error) {
	configMapGVK := corev1.SchemeGroupVersion.WithKind("ConfigMap")
	for _, manifest := range manifests {
		if manifest.GVK != configMapGVK {
			continue
		}
		if _, ok := manifest.Obj.GetAnnotations()[ReleaseAnnotationConfigMapVerifier]; !ok {
			continue
		}
		src := fmt.Sprintf("the config map %s/%s", manifest.Obj.GetNamespace(), manifest.Obj.GetName())
		data, _, err := unstructured.NestedStringMap(manifest.Obj.Object, "data")
		if err != nil {
			return nil, errors.Wrapf(err, "%s is not valid: %v", src, err)
		}
		verifier, err := NewFromConfigMapData(src, data, clientBuilder)
		if err != nil {
			return nil, err
		}
		return verifier, nil
	}
	return nil, nil
}

// hasVerified returns true if the digest has already been verified.
func (v *ReleaseVerifier) hasVerified(releaseDigest string) bool {
	v.lock.Lock()
	defer v.lock.Unlock()
	_, ok := v.signatureCache[releaseDigest]
	return ok
}

const maxSignatureCacheSize = 64

// cacheVerification caches the result of signature check for a digest for later retrieval.
func (v *ReleaseVerifier) cacheVerification(releaseDigest string, signedWith [][]byte) {
	v.lock.Lock()
	defer v.lock.Unlock()

	if len(signedWith) == 0 || len(releaseDigest) == 0 || v.signatureCache == nil {
		return
	}
	// remove the new entry
	delete(v.signatureCache, releaseDigest)
	// ensure the cache doesn't grow beyond our cap
	for k := range v.signatureCache {
		if len(v.signatureCache) < maxSignatureCacheSize {
			break
		}
		delete(v.signatureCache, k)
	}
	v.signatureCache[releaseDigest] = signedWith
}

// checkFileSignatures reads signatures as "signature-1", "signature-2", etc out of a directory until
// either the provided fn returns an error, false, or no such file exists. No more than maxSignaturesToCheck
// will be read.
func checkFileSignatures(ctx context.Context, dir string, maxSignaturesToCheck int, fn func(path string, signature []byte) (bool, error)) error {
	base := filepath.Join(dir, "signature-")
	for i := 1; i < maxSignatureSearch; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		path := base + strconv.Itoa(i)
		data, err := ioutil.ReadFile(path)
		if os.IsNotExist(err) {
			break
		}
		if err != nil {
			klog.V(4).Infof("unable to load signature: %v", err)
			continue
		}
		ok, err := fn(path, data)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
	}
	return nil
}

var errNotFound = fmt.Errorf("no more signatures to check")

// checkHTTPSignatures reads signatures as "signature-1", "signature-2", etc as children of the provided URL
// over HTTP or HTTPS until either the provided fn returns an error, false, or the server returns 404. No
// more than maxSignaturesToCheck will be read. If the provided context is cancelled search will be terminated.
func checkHTTPSignatures(ctx context.Context, client *http.Client, u url.URL, maxSignaturesToCheck int, fn func(path string, signature []byte) (bool, error)) error {
	base := filepath.Join(u.Path, "signature-")
	sigURL := u
	for i := 1; i < maxSignatureSearch; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		sigURL.Path = base + strconv.Itoa(i)

		req, err := http.NewRequest("GET", sigURL.String(), nil)
		if err != nil {
			return fmt.Errorf("could not build request to check signature: %v", err)
		}
		req = req.WithContext(ctx)
		// load the body, being careful not to allow unbounded reads
		resp, err := client.Do(req)
		if err != nil {
			klog.V(4).Infof("unable to load signature: %v", err)
			continue
		}
		data, err := func() ([]byte, error) {
			body := resp.Body
			r := io.LimitReader(body, 50*1024)

			defer func() {
				// read the remaining body to avoid breaking the connection
				io.Copy(ioutil.Discard, r)
				body.Close()
			}()

			if resp.StatusCode == 404 {
				return nil, errNotFound
			}
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				if i == 1 {
					klog.V(4).Infof("Could not find signature at store location %v", sigURL)
				}
				return nil, fmt.Errorf("unable to retrieve signature from %v: %d", sigURL, resp.StatusCode)
			}

			return ioutil.ReadAll(resp.Body)
		}()
		if err == errNotFound {
			klog.Info("not found")
			break
		}
		if err != nil {
			klog.Info(err)
			klog.V(4).Info(err)
			continue
		}
		if len(data) == 0 {
			continue
		}

		ok, err := fn(sigURL.String(), data)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
	}
	return nil
}

// verifySignatureWithKeyring performs a containers/image verification of the provided signature
// message, checking for the integrity and authenticity of the provided message in r. It will return
// the identity of the signer if successful along with the message contents.
func verifySignatureWithKeyring(r io.Reader, keyring openpgp.EntityList) ([]byte, string, error) {
	md, err := openpgp.ReadMessage(r, keyring, nil, nil)
	if err != nil {
		return nil, "", fmt.Errorf("could not read the message: %v", err)
	}
	if !md.IsSigned {
		return nil, "", fmt.Errorf("not signed")
	}
	content, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		return nil, "", err
	}
	if md.SignatureError != nil {
		return nil, "", fmt.Errorf("signature error: %v", md.SignatureError)
	}
	if md.SignedBy == nil {
		return nil, "", fmt.Errorf("invalid signature")
	}
	if md.Signature != nil {
		if md.Signature.SigLifetimeSecs != nil {
			expiry := md.Signature.CreationTime.Add(time.Duration(*md.Signature.SigLifetimeSecs) * time.Second)
			if time.Now().After(expiry) {
				return nil, "", fmt.Errorf("signature expired on %s", expiry)
			}
		}
	} else if md.SignatureV3 == nil {
		return nil, "", fmt.Errorf("unexpected openpgp.MessageDetails: neither Signature nor SignatureV3 is set")
	}

	// follow conventions in containers/image
	return content, strings.ToUpper(fmt.Sprintf("%x", md.SignedBy.PublicKey.Fingerprint)), nil
}

// An atomic container signature has the following schema:
//
// {
// 	"critical": {
// 			"type": "atomic container signature",
// 			"image": {
// 					"docker-manifest-digest": "sha256:817a12c32a39bbe394944ba49de563e085f1d3c5266eb8e9723256bc4448680e"
// 			},
// 			"identity": {
// 					"docker-reference": "docker.io/library/busybox:latest"
// 			}
// 	},
// 	"optional": {
// 			"creator": "some software package v1.0.1-35",
// 			"timestamp": 1483228800,
// 	}
// }
type signature struct {
	Critical criticalSignature `json:"critical"`
	Optional optionalSignature `json:"optional"`
}

type criticalSignature struct {
	Type     string           `json:"type"`
	Image    criticalImage    `json:"image"`
	Identity criticalIdentity `json:"identity"`
}

type criticalImage struct {
	DockerManifestDigest string `json:"docker-manifest-digest"`
}

type criticalIdentity struct {
	DockerReference string `json:"docker-reference"`
}

type optionalSignature struct {
	Creator   string `json:"creator"`
	Timestamp int64  `json:"timestamp"`
}

// verifyAtomicContainerSignature verifiers that the provided data authenticates the
// specified release digest. If error is returned the provided data does NOT authenticate
// the release digest and the signature must be ignored.
func verifyAtomicContainerSignature(data []byte, releaseDigest string) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	var sig signature
	if err := d.Decode(&sig); err != nil {
		return fmt.Errorf("the signature is not valid JSON: %v", err)
	}
	if sig.Critical.Type != "atomic container signature" {
		return fmt.Errorf("signature is not the correct type")
	}
	if len(sig.Critical.Identity.DockerReference) == 0 {
		return fmt.Errorf("signature must have an identity")
	}
	if sig.Critical.Image.DockerManifestDigest != releaseDigest {
		return fmt.Errorf("signature digest does not match")
	}
	return nil
}
