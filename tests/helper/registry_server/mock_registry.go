package registry_server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	_singleStackVersionName = "___SINGLE_VERSION__"
)

// MockRegistryServer is an implementation of a Devfile Registry Server,
// inspired by the own Devfile Registry tests at https://github.com/devfile/registry-support/blob/main/index/server/pkg/server/endpoint_test.go.
type MockRegistryServer struct {
	started bool
	server  *httptest.Server
}

// DevfileStack is the main struct for devfile stack
type DevfileStack struct {
	Name     string                `json:"name"`
	Versions []DevfileStackVersion `json:"versions,omitempty"`
}

type DevfileStackVersion struct {
	Version         string   `json:"version,omitempty"`
	IsDefault       bool     `json:"default"`
	SchemaVersion   string   `json:"schemaVersion,omitempty"`
	StarterProjects []string `json:"starterProjects"`
}

var manifests map[string]map[string]ocispec.Manifest

func init() {
	manifests = make(map[string]map[string]ocispec.Manifest)
	stackRoot := filepath.Join(getRegistryBasePath(), "stacks")
	_, err := os.Stat(stackRoot)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		log.Fatalf("file not found: %v - reason: %s. Did you run 'make generate-test-registry-build'?", stackRoot, err)
	}

	listFilesInDir := func(p string, excludeIf func(f os.FileInfo) bool) (res []string, err error) {
		file, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		files, err := file.Readdir(0)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			if excludeIf != nil && excludeIf(f) {
				continue
			}
			res = append(res, f.Name())
		}
		return res, nil
	}

	newManifest := func() ocispec.Manifest {
		return ocispec.Manifest{
			Versioned: specs.Versioned{SchemaVersion: 2},
			Config: ocispec.Descriptor{
				MediaType: "application/vnd.devfileio.devfile.config.v2+json",
			},
		}
	}

	buildLayerForFile := func(fpath string) (layer ocispec.Descriptor, err error) {
		stat, err := os.Stat(fpath)
		if err != nil {
			return ocispec.Descriptor{}, err
		}

		dgest, err := digestFile(fpath)
		if err != nil {
			return ocispec.Descriptor{}, err
		}
		layer.Digest = digest.Digest(dgest)

		f := filepath.Base(filepath.Clean(fpath))
		if f == "devfile.yaml" {
			layer.MediaType = "application/vnd.devfileio.devfile.layer.v1"
		} else if strings.HasSuffix(f, ".tar") {
			layer.MediaType = "application/x-tar"
		}

		layer.Size = stat.Size()
		layer.Annotations = map[string]string{
			"org.opencontainers.image.title": f,
		}
		return layer, nil
	}

	excludeIfDirFn := func(f os.FileInfo) bool {
		return f.IsDir()
	}
	excludeIfNotDirFn := func(f os.FileInfo) bool {
		return !f.IsDir()
	}

	dirsInStacksRoot, err := listFilesInDir(stackRoot, excludeIfNotDirFn)
	if err != nil {
		log.Fatalf(err.Error())
	}
	for _, f := range dirsInStacksRoot {
		manifests[f] = make(map[string]ocispec.Manifest)
		versionList, err := listFilesInDir(filepath.Join(stackRoot, f), excludeIfNotDirFn)
		if err != nil {
			log.Fatalf(err.Error())
		}
		if len(versionList) == 0 {
			// Possible stack with single unnamed version
			stackFiles, err := listFilesInDir(filepath.Join(stackRoot, f), excludeIfDirFn)
			if err != nil {
				log.Fatalf(err.Error())
			}
			manifest := newManifest()
			for _, vf := range stackFiles {
				layer, err := buildLayerForFile(filepath.Join(stackRoot, f, vf))
				if err != nil {
					log.Fatalf(err.Error())
				}
				manifest.Layers = append(manifest.Layers, layer)
			}
			manifests[f][_singleStackVersionName] = manifest
			continue
		}
		for _, v := range versionList {
			versionFiles, err := listFilesInDir(filepath.Join(stackRoot, f, v), excludeIfDirFn)
			if err != nil {
				log.Fatalf(err.Error())
			}
			manifest := newManifest()
			for _, vf := range versionFiles {
				layer, err := buildLayerForFile(filepath.Join(stackRoot, f, v, vf))
				if err != nil {
					log.Fatalf(err.Error())
				}
				manifest.Layers = append(manifest.Layers, layer)
			}
			manifests[f][v] = manifest
		}
	}
}

func NewMockRegistryServer() *MockRegistryServer {
	r := mux.NewRouter()
	m := MockRegistryServer{
		server: httptest.NewUnstartedServer(handlers.LoggingHandler(GinkgoWriter, r)),
	}

	m.setupRoutes(r)
	return &m
}

func (m *MockRegistryServer) Start() (url string, err error) {
	m.server.Start()
	m.started = true
	fmt.Fprintln(GinkgoWriter, "Mock Devfile Registry server started and available at", m.server.URL)
	return m.server.URL, nil
}

func (m *MockRegistryServer) Stop() error {
	m.server.Close()
	m.started = false
	return nil
}

func (m *MockRegistryServer) GetUrl() string {
	return m.server.URL
}

func (m *MockRegistryServer) IsStarted() bool {
	return m.started
}

func notFoundManifest(res http.ResponseWriter, req *http.Request, tag string) {
	var data string
	if req.Method == http.MethodGet {
		data = fmt.Sprintf(`
{
	"code": "MANIFEST_UNKNOWN",
	"message": "manifest unknown",
	"detail": {
		"tag": %s
	}
}
`, tag)
	}
	res.WriteHeader(http.StatusNotFound)
	_, err := res.Write([]byte(data))
	if err != nil {
		fmt.Fprintln(GinkgoWriter, "[warn] failed to write response; cause:", err)
	}
}

// notFound custom handler for anything not found
func notFound(res http.ResponseWriter, req *http.Request, data string) {
	res.WriteHeader(http.StatusNotFound)
	_, err := res.Write([]byte(data))
	if err != nil {
		fmt.Fprintln(GinkgoWriter, "[warn] failed to write response; cause:", err)
	}
}

func internalServerError(res http.ResponseWriter, req *http.Request, data string) {
	res.WriteHeader(http.StatusInternalServerError)
	_, err := res.Write([]byte(fmt.Sprintf(`{"detail": %q}`, data)))
	if err != nil {
		fmt.Fprintln(GinkgoWriter, "[warn] failed to write response; cause:", err)
	}
}

// setupRoutes setups the routing, based on the OpenAPI Schema defined at:
// https://github.com/devfile/registry-support/blob/main/index/server/openapi.yaml
func (m *MockRegistryServer) setupRoutes(r *mux.Router) {
	r.HandleFunc("/v2index", serveV2Index).Methods(http.MethodGet)
	r.HandleFunc("/v2/devfile-catalog/{stack}/manifests/{ref}", serveManifests).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/v2/devfile-catalog/{stack}/blobs/{digest}", serveBlobs).Methods(http.MethodGet)
	r.HandleFunc("/devfiles/{stack}", m.serveDevfileDefaultVersion).Methods(http.MethodGet)
	r.HandleFunc("/devfiles/{stack}/{version}", m.serveDevfileAtVersion).Methods(http.MethodGet)
}

func getRegistryBasePath() string {
	_, filename, _, _ := runtime.Caller(1)
	return filepath.Join(path.Dir(filename), "testdata", "registry-build")
}

func serveV2Index(res http.ResponseWriter, req *http.Request) {
	index := filepath.Join(getRegistryBasePath(), "index.json")
	d, err := os.ReadFile(index)
	if err != nil {
		internalServerError(res, req, err.Error())
		return
	}
	_, err = res.Write(d)
	if err != nil {
		fmt.Fprintln(GinkgoWriter, "[warn] failed to write response; cause:", err)
	}
}

func serveManifests(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	stack := vars["stack"]
	ref := vars["ref"]
	var (
		stackManifest ocispec.Manifest
		found         bool
		bytes         []byte
		err           error
	)

	if strings.HasPrefix(ref, "sha256:") {
		var stackManifests map[string]ocispec.Manifest
		stackManifests, found = manifests[stack]
		if !found {
			notFoundManifest(res, req, ref)
			return
		}
		found = false
		var dgst string
		for _, manifest := range stackManifests {
			dgst, err = digestEntity(manifest)
			if err != nil {
				internalServerError(res, req, "")
				return
			}
			if reflect.DeepEqual(ref, dgst) {
				stackManifest = manifest
				found = true
				break
			}
		}
		if !found {
			notFoundManifest(res, req, ref)
			return
		}
	} else {
		stackManifest, found = manifests[stack][ref]
		if !found {
			// Possible single unnamed version
			stackManifest, found = manifests[stack][_singleStackVersionName]
			if !found {
				notFoundManifest(res, req, ref)
				return
			}
		}
	}

	var j []byte
	if j, err = json.MarshalIndent(stackManifest, " ", " "); err != nil {
		fmt.Fprintln(GinkgoWriter, "[debug] stackManifest:", stackManifest)
	} else {
		fmt.Fprintln(GinkgoWriter, "[debug] stackManifest:", string(j))
	}

	if req.Method == http.MethodGet {
		bytes, err = json.Marshal(stackManifest)
		if err != nil {
			internalServerError(res, req, err.Error())
			return
		}
	}

	res.Header().Set("Content-Type", ocispec.MediaTypeImageManifest)
	res.WriteHeader(http.StatusOK)
	_, err = res.Write(bytes)
	if err != nil {
		fmt.Fprintln(GinkgoWriter, "[warn] failed to write response; cause:", err)
	}
}

func serveBlobs(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	stack := vars["stack"]
	sDigest := vars["digest"]
	stackRoot := filepath.Join(getRegistryBasePath(), "stacks", stack)
	var (
		blobPath string
		found    bool
		err      error
	)

	found = false
	err = filepath.WalkDir(stackRoot, func(path string, d fs.DirEntry, err error) error {
		var fdgst string

		if err != nil {
			return err
		}

		if found || d.IsDir() {
			return nil
		}

		fdgst, err = digestFile(path)
		if err != nil {
			return err
		}
		if reflect.DeepEqual(sDigest, fdgst) {
			blobPath = path
			found = true
		}

		return nil
	})
	if err != nil || !found {
		notFound(res, req, "")
		return
	}

	file, err := os.Open(blobPath)
	Expect(err).ShouldNot(HaveOccurred())
	defer file.Close()

	bytes, err := io.ReadAll(file)
	Expect(err).ShouldNot(HaveOccurred())

	res.WriteHeader(http.StatusOK)
	res.Header().Set("Content-Type", http.DetectContentType(bytes))
	_, err = res.Write(bytes)
	if err != nil {
		fmt.Fprintln(GinkgoWriter, "[warn] failed to write response; cause:", err)
	}
}

func (m *MockRegistryServer) serveDevfileDefaultVersion(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	stack := vars["stack"]

	defaultVersion, internalErr, err := findStackDefaultVersion(stack)
	if err != nil {
		if internalErr {
			internalServerError(res, req, "")
		} else {
			notFound(res, req, "")
		}
		return
	}

	http.Redirect(res, req, fmt.Sprintf("%s/devfiles/%s/%s", m.GetUrl(), stack, defaultVersion), http.StatusSeeOther)
}

func findStackDefaultVersion(stack string) (string, bool, error) {
	index, err := parseIndex()
	if index == nil {
		return "", true, err
	}
	for _, d := range index {
		if d.Name != stack {
			continue
		}
		for _, v := range d.Versions {
			if v.IsDefault {
				return v.Version, false, nil
			}
		}
	}
	return "", false, fmt.Errorf("default version not found for %q", stack)
}

func parseIndex() ([]DevfileStack, error) {
	// find the default version
	index := filepath.Join(getRegistryBasePath(), "index.json")
	d, err := os.ReadFile(index)
	if err != nil {
		return nil, err
	}

	var objmap []DevfileStack
	err = json.Unmarshal(d, &objmap)
	if err != nil {
		return nil, err
	}
	return objmap, nil
}

func (m *MockRegistryServer) serveDevfileAtVersion(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	stack := vars["stack"]
	version := vars["version"]

	// find layer for this version and redirect to the blob download URL
	manifestByVersionMap, ok := manifests[stack]
	if !ok {
		notFound(res, req, "")
		return
	}
	manifest, ok := manifestByVersionMap[version]
	if ok {
		// find blob with devfile
		for _, layer := range manifest.Layers {
			if layer.Annotations["org.opencontainers.image.title"] == "devfile.yaml" {
				http.Redirect(res, req, fmt.Sprintf("%s/v2/devfile-catalog/%s/blobs/%s", m.GetUrl(), stack, layer.Digest), http.StatusSeeOther)
				return
			}
		}
		notFound(res, req, "devfile.yaml not found")
		return
	}

	// find if devfile has a single version that matches the default version in index
	defaultVersion, internalErr, err := findStackDefaultVersion(stack)
	if err != nil {
		if internalErr {
			internalServerError(res, req, "")
		} else {
			notFound(res, req, "")
		}
		return
	}
	if defaultVersion != version {
		notFound(res, req, "default version for this stack is:"+defaultVersion)
		return
	}
	manifest, ok = manifestByVersionMap[defaultVersion]
	if ok {
		// find blob with devfile
		for _, layer := range manifest.Layers {
			if layer.Annotations["org.opencontainers.image.title"] == "devfile.yaml" {
				http.Redirect(res, req, fmt.Sprintf("%s/v2/devfile-catalog/%s/blobs/%s", m.GetUrl(), stack, layer.Digest), http.StatusSeeOther)
				return
			}
		}
		notFound(res, req, "devfile.yaml not found")
		return
	}
}

// digestEntity generates sha256 digest of any entity type
func digestEntity(e interface{}) (string, error) {
	bytes, err := json.Marshal(e)
	if err != nil {
		return "", err
	}

	return digest.FromBytes(bytes).String(), nil
}

// digestFile generates sha256 digest from file contents
func digestFile(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	dgst, err := digest.FromReader(file)
	if err != nil {
		return "", err
	}

	return dgst.String(), nil
}
