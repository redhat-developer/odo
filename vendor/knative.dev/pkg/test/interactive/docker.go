/*
Copyright 2020 The Knative Authors

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

// Helper functions for running interactive CLI sessions from Go
package interactive

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

var defaultDockerCommands []string

func init() {
	defaultDockerCommands = []string{"docker", "run", "-it", "--rm", "--entrypoint", "bash"}
}

// Env represents a collection of environment variables and their values
type Env map[string]string

// Docker is mostly an Command preloaded with arguments which setup Docker for running an image interactively.
type Docker struct {
	Command
}

// PromoteFromEnv pulls the named environment variables from the environment and puts them in the Env.
// It does not stop on error and returns an error listing all the failed values
func (e Env) PromoteFromEnv(envVars ...string) error {
	var err error
	for _, env := range envVars {
		v := os.Getenv(env)
		if v == "" {
			err = fmt.Errorf("environment variable %q is not set; %v", env, err)
		} else {
			e[env] = v
		}
	}
	return err
}

// Creates an Docker with default Docker command arguments for running interactively
func NewDocker() Docker {
	return Docker{NewCommand(defaultDockerCommands...)}
}

// AddEnv adds arguments so all the environment variables present in e become part of the docker run's environment
func (d *Docker) AddEnv(e Env) {
	for k, v := range e {
		d.AddArgs("-e", fmt.Sprintf("%s=%s", k, v))
	}
}

// AddMount add arguments for the --mount command
func (d *Docker) AddMount(typeStr, source, target string, optAdditionalArgs ...string) {
	addl := ""
	if len(optAdditionalArgs) != 0 {
		addl = "," + strings.Join(optAdditionalArgs, ",")
	}
	d.AddArgs("--mount", fmt.Sprintf("type=%s,source=%s,target=%s%s", typeStr, source, target, addl))
}

// AddRWOverlay mounts a directory into the image at the desired location, but with an overlay
//  so internal changes do not modify the external directory.
// externalDirectory probably needs to be an absolute path
// Returns a function to clean up the mount (but does not delete the directory).
// Uses sudo and probably only works on Linux
func (d *Docker) AddRWOverlay(externalDirectory, internalDirectory string) func() {
	tmpDir, err := ioutil.TempDir("", "overlay")
	if err != nil {
		log.Fatal(err)
	}
	subDirs := []string{"upper", "work", "overlay"}
	for _, d := range subDirs {
		err = os.Mkdir(path.Join(tmpDir, d), 0777)
		if err != nil {
			log.Fatal(err)
		}
	}
	overlayDir := path.Join(tmpDir, "overlay")
	// The options for overlay mount are confusing
	// You need empty directories for upper and work (and overlay, though you can mount over directories that have files in them if you *want* to...)
	mount := NewCommand("sudo", "mount", "-t", "overlay", "-o",
		fmt.Sprintf("lowerdir=%s,upperdir=%s/upper,workdir=%s/work", externalDirectory, tmpDir, tmpDir),
		"none", overlayDir)
	// Print command to run so user knows why it is asking for sudo password (if it does)
	log.Println(mount)
	if err = mount.Run(); err != nil {
		log.Fatalf("Unable to create overlay mount, so giving up: %v", err)
	}
	d.AddMount("bind", overlayDir, internalDirectory)
	return func() {
		// Print command to run so user knows why it is asking for sudo password (if it does)
		umount := NewCommand("sudo", "umount", overlayDir)
		log.Println(umount)
		umount.Run()
	}
}
