/*
 Copyright 2019 The Tekton Authors

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

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/go-github/github"
)

const (
	// Environment variable containing GitHub secret token
	envSecret = "GITHUB_SECRET_TOKEN"
)

func main() {
	secretToken := os.Getenv(envSecret)
	if secretToken == "" {
		log.Fatalf("No secret token given")
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		//TODO: We should probably send over the EL eventID as a X-Tekton-Event-Id header as well
		payload, err := github.ValidatePayload(request, []byte(secretToken))
		id := github.DeliveryID(request)
		if err != nil {
			log.Printf("Error handling Github Event with delivery ID %s : %q", id, err)
			http.Error(writer, fmt.Sprint(err), http.StatusBadRequest)
		}
		log.Printf("Handling Github Event with delivery ID: %s; Payload: %s", id, payload)
		n, err := writer.Write(payload)
		if err != nil {
			log.Printf("Failed to write response for Github event ID: %s. Bytes writted: %d. Error: %q", id, n, err)
		}
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", 8080), nil))
}
