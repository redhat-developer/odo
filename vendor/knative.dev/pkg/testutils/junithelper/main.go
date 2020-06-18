/*
Copyright 2019 The Knative Authors

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
	"flag"

	"knative.dev/pkg/test/junit"
)

var (
	suite  string
	name   string
	errMsg string
	dest   string
)

func main() {
	flag.StringVar(&suite, "suite", "", "Name of suite")
	flag.StringVar(&name, "name", "", "Name of test")
	flag.StringVar(&errMsg, "err-msg", "", "Error message, empty means test passed, default empty")
	flag.StringVar(&dest, "dest", "junit_result.xml", "Where junit xml writes to")
	flag.Parse()

	junit.CreateXMLErrorMsg(suite, name, errMsg, dest)
}
