/*
Copyright 2020 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	"github.com/tektoncd/triggers/pkg/template"
	"k8s.io/apimachinery/pkg/runtime/serializer/streaming"
	"k8s.io/client-go/kubernetes/scheme"
)

var (
	rootCmd = &cobra.Command{
		Use:   "binding-eval",
		Short: "Tekton TriggerBinding evaluator",
		Run:   rootRun,
	}

	bindingPath string
	httpPath    string
)

func init() {
	rootCmd.Flags().StringVarP(&bindingPath, "binding", "b", "", "Path to trigger binding")
	rootCmd.Flags().StringVarP(&httpPath, "http_request", "r", "", "Path to HTTP request")
	if err := rootCmd.MarkFlagRequired("binding"); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func rootRun(cmd *cobra.Command, args []string) {
	if err := evalBinding(os.Stdout, bindingPath, httpPath); err != nil {
		log.Fatal(err)
	}
}

func evalBinding(w io.Writer, bindingPath, httpPath string) error {
	// Read HTTP request.
	r, err := readHTTP(httpPath)
	if err != nil {
		return fmt.Errorf("error reading HTTP file: %w", err)
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("error reading HTTP body: %w", err)
	}

	// Read bindings.
	bindings, err := readBindings(bindingPath)
	if err != nil {
		return fmt.Errorf("error reading bindings: %w", err)
	}

	t := template.ResolvedTrigger{
		TriggerBindings: bindings,
	}
	params, err := template.ResolveParams(t, body, r.Header)
	if err != nil {
		return fmt.Errorf("error resolving params: %w", err)
	}

	// Sort results for stable output.
	sort.SliceStable(params, func(i, j int) bool {
		return params[i].Name < params[j].Name
	})

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(params); err != nil {
		return fmt.Errorf("error encoding params: %w", err)
	}

	return nil
}

func readBindings(path string) ([]*v1alpha1.TriggerBinding, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error reading binding file: %w", err)
	}
	defer f.Close()

	var list []*v1alpha1.TriggerBinding
	decoder := streaming.NewDecoder(f, scheme.Codecs.UniversalDecoder())
	b := new(v1alpha1.TriggerBinding)
	for err == nil {
		_, _, err = decoder.Decode(nil, b)
		if err != nil {
			if err != io.EOF {
				return nil, fmt.Errorf("error decoding bindings: %w", err)
			}
			break
		}
		list = append(list, b)
	}
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error decoding bindings: %w", err)
	}

	return list, nil
}

func readHTTP(path string) (*http.Request, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()

	return http.ReadRequest(bufio.NewReader(f))
}

// Execute runs the command.
func Execute() error {
	return rootCmd.Execute()
}
