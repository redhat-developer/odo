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

// junit_test.go contains unit tests for junit package

package junit

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

var emptySuites = `
<testsuites>
</testsuites>
`

var malSuitesString = `
<testsuites>
<testsuites>
`

var validSuiteString = `
<testsuite name="knative/test-infra">
	<properties>
		<property name="go.version" value="go1.6"/>
	</properties>
	<testcase name="TestBad" time="0.1">
		<failure>something bad</failure>
		<system-out>out: first line</system-out>
		<system-err>err: first line</system-err>
		<system-out>out: second line</system-out>
	</testcase>
	<testcase name="TestGood" time="0.1">
	</testcase>
	<testcase name="TestSkip" time="0.1">
		<skipped>do not test</skipped>
	</testcase>
</testsuite>
`

var validSuitesString = `
<testsuites>
	<testsuite name="knative/test-infra">
		<properties>
			<property name="go.version" value="go1.6"/>
		</properties>
		<testcase name="TestBad" time="0.1">
			<failure>something bad</failure>
			<system-out>out: first line</system-out>
			<system-err>err: first line</system-err>
			<system-out>out: second line</system-out>
		</testcase>
		<testcase name="TestGood" time="0.1">
		</testcase>
		<testcase name="TestSkip" time="0.1">
			<skipped>do not test</skipped>
		</testcase>
	</testsuite>
</testsuites>
`

func newTestCase(name string, status TestStatusEnum) *TestCase {
	testCase := TestCase{
		Name: name,
	}

	var tmp string // cast const to string
	switch {
	case status == Failed:
		tmp = string(Failed)
		testCase.Failure = &tmp
	case status == Skipped:
		tmp = string(Skipped)
		testCase.Skipped = &tmp
	}

	return &testCase
}

func TestUnmarshalEmptySuites(t *testing.T) {
	if _, err := UnMarshal([]byte(emptySuites)); err != nil {
		t.Errorf("Expected 'succeed', actual: 'failed parsing empty suites, '%s'", err)
	}
}

func TestUnmarshalMalFormed(t *testing.T) {
	if _, err := UnMarshal([]byte(malSuitesString)); err == nil {
		t.Errorf("Expected: failed, actual: succeeded parsing malformed xml, '%s'", err)
	}
}

func TestUnmarshalSuites(t *testing.T) {
	if _, err := UnMarshal([]byte(validSuitesString)); err != nil {
		t.Errorf("Expected: succeed, actual: failed parsing suites result, '%s'", err)
	}
}

func TestUnmarshalSuite(t *testing.T) {
	if _, err := UnMarshal([]byte(validSuiteString)); err != nil {
		t.Errorf("Expected: succeed, actual: failed parsing suite result, '%s'", err)
	}
}

func TestGetTestStatus(t *testing.T) {
	if status := newTestCase("TestGood", Passed).GetTestStatus(); Passed != status {
		t.Errorf("Expected '%s', actual '%s'", Passed, status)
	}
	if status := newTestCase("TestSkip", Skipped).GetTestStatus(); Skipped != status {
		t.Errorf("Expected '%s', actual '%s'", Skipped, status)
	}
	if status := newTestCase("TestBad", Failed).GetTestStatus(); Failed != status {
		t.Errorf("Expected '%s', actual '%s'", Failed, status)
	}
}

func TestAddTestSuite(t *testing.T) {
	testSuites := TestSuites{}
	testSuite0 := TestSuite{Name: "suite_0"}
	testSuite1 := TestSuite{Name: "suite_1"}

	if err := testSuites.AddTestSuite(&testSuite0); err != nil {
		t.Fatalf("Expected '', actual '%v'", err)
	}

	expectedErrString := "Test suite 'suite_0' already exists"
	if err := testSuites.AddTestSuite(&testSuite0); err == nil || err.Error() != expectedErrString {
		t.Fatalf("Expected: '%s', actual: '%v'", expectedErrString, err)
	}

	if err := testSuites.AddTestSuite(&testSuite1); err != nil {
		t.Fatalf("Expected '', actual '%v'", err)
	}

	if len(testSuites.Suites) != 2 {
		t.Fatalf("Expected 2, actual %d", len(testSuites.Suites))
	}
}

func TestCreateXMLErrorMsg(t *testing.T) {
	testDir := "test_output"
	os.RemoveAll(testDir) // clean up in case there were stale side-effects from previous runs
	if err := os.Mkdir(testDir, 0777); err != nil {
		t.Fatalf("cannot create directory %q", testDir)
	}
	defer os.RemoveAll(testDir) // clean up
	dest := path.Join(testDir, "TestCreateXMLErrorTestFile")
	CreateXMLErrorMsg("dummySuite", "dummyTest", "dummyError has occurred", dest)
	expected := `<testsuites><testsuite name="dummySuite" time="0" failures="1" tests="1"><testcase name="dummyTest" time="0" classname=""><failure>dummyError has occurred</failure><properties></properties></testcase><properties></properties></testsuite></testsuites>`

	got, err := ioutil.ReadFile(dest)
	if err != nil {
		t.Fatalf("cannot read %q, error %v", dest, err)
	}

	if string(got) != expected {
		t.Fatalf("expected:\n%q\n, got:\n%q", expected, got)
	}
}
