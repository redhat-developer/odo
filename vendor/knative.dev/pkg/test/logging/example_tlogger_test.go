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

package logging_test

import (
	"testing"

	"knative.dev/pkg/test/logging"
)

type testStruct struct {
	D string
	e float64
}

var (
	someStruct testStruct
	couldBeErr error
)

func init() { someStruct = testStruct{"hello", 42.0} }

// nolint:govet // godoc limitation; this would be named Test*(), of course
func Example(legacy *testing.T) {
	// Get our TLogger and ready the cleanup function
	t, cancel := logging.NewTLogger(legacy)
	defer cancel()

	// For the most part, you can pretend t is really a *testing.T
	// But you get better results by treating it as a leveled logger
	// with the same semantics as Infow() in Zap (keys & values alternating after the main argument).
	// Logging is leveled from 0-10; see https://github.com/go-logr/logr#how-do-i-choose-my-v-levels
	// In our tests currently, levels are sort-of used as follows:
	// 1: Describe broadly what the test is doing
	// 2: What specific action is occurring
	// 5: Trace level; print everything you could predict could be useful in diagnosing stubborn failures
	// 8: Just print anything leftover
	// Levels 5-9 also instruct the Kubernetes client library to print increasing amounts of data around control plane requests.
	t.V(1).Info("We're just presenting some log statements")
	t.V(2).Info("Don't forget about this struct", "SomeStruct", someStruct)
	t.V(5).Info("We just did a couple little steps, going to try something else")
	t.V(8).Info("What else is left?", "LifeTheUniverseAndEverything", 42, "Really anything", t)

	// Please use t.V(x).Info and avoid t.Logf to get structured logging
	// You get easier-to-read logs too!

	// When checking an error and you want to fail the test,
	// use one of the following:
	t.ErrorIfErr(couldBeErr, "Message about error", "key1", "value1", "key2", "value2", "keyYouGetTheIdea", 0)
	t.FatalIfErr(couldBeErr, "Message about why the test will now abort")

	// If failing the test, but you don't have an error object:
	t.Error("Message about failure", "keysAnd", "Values")
	t.Fatal("Message why failing now", "SupportKeysAndValues", true)
}
