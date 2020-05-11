package serve

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/docker/distribution/manifest"
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"
)

func NewServeOptions(streams genericclioptions.IOStreams) *ServeOptions {
	return &ServeOptions{
		IOStreams:  streams,
		ListenAddr: ":5000",
	}
}

func New(parentName string, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewServeOptions(streams)
	cmd := &cobra.Command{
		Use:   "serve IMAGE",
		Short: "Serve a container registry from images mirrored to disk",
		Long: templates.LongDesc(`
			Serve a container registry

			This command will start an HTTP or HTTPS server that hosts a local directory of mirrored
			images. Use the 'oc image mirror --dir=DIR SRC=DST' command to populate that directory.
			The directory must have a 'v2' folder that contains repository sub directories.

			No authentication or authorization checks are performed and the source directory should
			only include content you wish network users to see.

			Experimental: This command is under active development and may change without notice.
		`),
		Run: func(cmd *cobra.Command, args []string) {
			kcmdutil.CheckErr(o.Complete(cmd, args))
			kcmdutil.CheckErr(o.Validate())
			kcmdutil.CheckErr(o.Run())
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&o.Dir, "dir", o.Dir, "The directory to serve images from.")
	flags.StringVar(&o.ListenAddr, "listen", o.ListenAddr, "A host:port to listen on. Defaults to *:5000")
	flags.StringVar(&o.TLSCertificatePath, "tls-crt", o.TLSCertificatePath, "Path to a TLS certificate to secure this server with.")
	flags.StringVar(&o.TLSKeyPath, "tls-key", o.TLSKeyPath, "Path to a TLS private key to secure this server with.")
	return cmd
}

type ServeOptions struct {
	genericclioptions.IOStreams

	Dir string

	ListenAddr         string
	TLSKeyPath         string
	TLSCertificatePath string
}

func (o *ServeOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *ServeOptions) Validate() error {
	return nil
}

func (o *ServeOptions) Run() error {
	if len(o.ListenAddr) == 0 {
		return fmt.Errorf("must specify an address to listen on")
	}
	if fi, err := os.Stat(o.Dir); err != nil || !fi.IsDir() {
		return fmt.Errorf("--dir must point to a directory: %v", err)
	}

	dir := http.Dir(o.Dir)
	fileHandler := http.FileServer(dir)
	http.DefaultServeMux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" && req.Method == "GET" {
			http.Redirect(w, req, "/v2/", http.StatusTemporaryRedirect)
			return
		}
		http.NotFound(w, req)
	})
	http.DefaultServeMux.HandleFunc("/v2/", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" && req.URL.Path == "/v2/" {
			w.Header().Set("Docker-Distribution-API-Version", "2.0")
		}
		if req.Method == "GET" {
			switch path.Base(path.Dir(req.URL.Path)) {
			case "blobs":
				w.Header().Set("Content-Type", "application/octet-stream")
			case "manifests":
				if f, err := dir.Open(req.URL.Path); err == nil {
					defer f.Close()
					if data, err := ioutil.ReadAll(f); err == nil {
						var versioned manifest.Versioned
						if err = json.Unmarshal(data, &versioned); err == nil {
							w.Header().Set("Content-Type", versioned.MediaType)
						}
					}
				}
			}
		}
		fileHandler.ServeHTTP(w, req)
	})
	if len(o.TLSKeyPath) > 0 || len(o.TLSCertificatePath) > 0 {
		klog.Infof("Serving TLS at %s ...", o.ListenAddr)
		return http.ListenAndServeTLS(o.ListenAddr, o.TLSCertificatePath, o.TLSKeyPath, nil)
	}
	klog.Infof("Serving at %s ...", o.ListenAddr)
	return http.ListenAndServe(o.ListenAddr, nil)
}
