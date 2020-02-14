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

package gke

import "testing"

func TestGetClusterLocation(t *testing.T) {
	datas := []struct {
		region, zone string
		want         string
	}{
		{"a", "b", "a-b"},
		{"a", "", "a"},
	}
	for _, data := range datas {
		if got := GetClusterLocation(data.region, data.zone); got != data.want {
			t.Errorf("Cluster location with region %q and zone %q = %q, want: %q",
				data.region, data.zone, got, data.want)
		}
	}
}

func TestRegionZoneFromLoc(t *testing.T) {
	datas := []struct {
		loc        string
		wantRegion string
		wantZone   string
	}{
		{"a-b-c", "a-b", "c"},
		{"a-b", "a-b", ""},
		{"a", "a", ""},
		{"", "", ""},
	}
	for _, data := range datas {
		gotRegion, gotZone := RegionZoneFromLoc(data.loc)
		if gotRegion != data.wantRegion {
			t.Errorf("Cluster region from location %q = %q, want: %q",
				data.loc, gotRegion, data.wantRegion)
		}
		if gotZone != data.wantZone {
			t.Errorf("Cluster zone from location %q = %q, want: %q",
				data.loc, gotZone, data.wantZone)
		}
	}
}
