/*
copyright 2019 the knative authors

licensed under the apache license, version 2.0 (the "license");
you may not use this file except in compliance with the license.
you may obtain a copy of the license at

    http://www.apache.org/licenses/license-2.0

unless required by applicable law or agreed to in writing, software
distributed under the license is distributed on an "as is" basis,
without warranties or conditions of any kind, either express or implied.
see the license for the specific language governing permissions and
limitations under the license.
*/

package kmeta

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestChildName(t *testing.T) {
	tests := []struct {
		parent string
		suffix string
		want   string
	}{{
		parent: "asdf",
		suffix: "-deployment",
		want:   "asdf-deployment",
	}, {
		parent: strings.Repeat("f", 63),
		suffix: "-deployment",
		want:   "ffffffffffffffffffff105d7597f637e83cc711605ac3ea4957-deployment",
	}, {
		parent: strings.Repeat("f", 63),
		suffix: "-deploy",
		want:   "ffffffffffffffffffffffff105d7597f637e83cc711605ac3ea4957-deploy",
	}, {
		parent: strings.Repeat("f", 63),
		suffix: strings.Repeat("f", 63),
		want:   "fffffffffffffffffffffffffffffff0502661254f13c89973cb3a83e0cbec0",
	}, {
		parent: "a",
		suffix: strings.Repeat("f", 63),
		want:   "ab5cfd486935decbc0d305799f4ce4414ffffffffffffffffffffffffffffff",
	}, {
		parent: strings.Repeat("b", 32),
		suffix: strings.Repeat("f", 32),
		want:   "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb329c7c81b9ab3ba71aa139066aa5625d",
	}, {
		parent: "aaaa",
		suffix: strings.Repeat("b---a", 20),
		want:   "aaaa7a3f7966594e3f0849720eced8212c18b---ab---ab---ab---ab---ab",
	}}

	for _, test := range tests {
		t.Run(test.parent+"-"+test.suffix, func(t *testing.T) {
			if got, want := ChildName(test.parent, test.suffix), test.want; got != want {
				t.Errorf("%s-%s: got: %63s want: %63s\ndiff:%s", test.parent, test.suffix, got, want, cmp.Diff(want, got))
			}
		})
	}
}
