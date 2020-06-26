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

package interactive

import (
	"os"
	"os/exec"
	"testing"
)

func TestEnv(t *testing.T) {
	e := make(Env)
	err := e.PromoteFromEnv("PWD")
	if err != nil {
		t.Error(err)
		t.Logf("The Environ is:\n%+v\n", os.Environ())
	}
	epwd, exists := e["PWD"]
	if !exists || epwd != os.Getenv("PWD") {
		t.Errorf(`$PWD promotion did not occur correctly: Env='%v'; os.Getenv("PWD")=%s`, e, os.Getenv("PWD"))
	}
	badName := "GEEZ_I_REALLY_H0PE_TH1S_DOES_NOT_EXIST"
	// and just in case:
	for {
		_, exists := os.LookupEnv(badName)
		if !exists {
			break
		}
		badName += "z"
	}
	err = e.PromoteFromEnv(badName)
	if err == nil {
		t.Error("Should have received error promoting non-existent variable")
	}
}

func TestAddEnv(t *testing.T) {
	d := NewDocker()
	e := make(Env)
	l := len(defaultDockerCommands)
	e["env"] = "var"
	d.AddEnv(e)
	d.run = func(c *exec.Cmd) error {
		if c.Args[l] != "-e" || c.Args[l+1] != "env=var" {
			t.Error("Env var wasn't added correctly")
			t.Log(mySpew.Sdump(c))
		}
		t.Log("Test Over")
		return nil
	}
	d.Run()
}

func TestAddMount(t *testing.T) {
	d := NewDocker()
	l := len(defaultDockerCommands)
	d.AddMount("bind", "mysource", "mytarget", "other1=banana1", "other2=banana2")
	d.run = func(c *exec.Cmd) error {
		if c.Args[l] != "--mount" || c.Args[l+1] != "type=bind,source=mysource,target=mytarget,other1=banana1,other2=banana2" {
			t.Error("Mount wasn't added correctly")
			t.Log(mySpew.Sdump(c))
		}
		t.Log("Test Over")
		return nil
	}
	d.Run()
}
