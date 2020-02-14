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

package pacers

import (
	"testing"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

func TestCombinedPacer(t *testing.T) {
	pacer1 := vegeta.Rate{Freq: 1, Per: time.Second}
	pacer2 := vegeta.Rate{Freq: 5, Per: time.Second}
	pacer3 := vegeta.Rate{Freq: 10, Per: time.Second}
	pacers := []vegeta.Pacer{pacer1, pacer2, pacer3}
	durations := []time.Duration{5 * time.Second, 5 * time.Second, 10 * time.Second}
	pacer, _ := NewCombined(pacers, durations)

	for _, tt := range []struct {
		name            string
		elapsedTime     time.Duration
		elapsedHits     uint64
		expectedNextHit time.Duration
		expectedStop    bool
	}{{
		name:            "test the first hit",
		elapsedTime:     0 * time.Second,
		elapsedHits:     0,
		expectedNextHit: 1 * time.Second,
	}, {
		name:            "test the switch pacer hit",
		elapsedTime:     5 * time.Second,
		elapsedHits:     5,
		expectedNextHit: 200 * time.Millisecond,
	}, {
		name:            "test the pacer middle hit",
		elapsedTime:     7 * time.Second,
		elapsedHits:     15,
		expectedNextHit: 200 * time.Millisecond,
	}, {
		name:            "test the last hit",
		elapsedTime:     20 * time.Second,
		elapsedHits:     130,
		expectedNextHit: 1 * time.Second,
	}, {
		name:            "test the loop back pacer hit",
		elapsedTime:     22 * time.Second,
		elapsedHits:     132,
		expectedNextHit: 1 * time.Second,
	}, {
		name:            "test the catch up hit",
		elapsedTime:     24 * time.Second,
		elapsedHits:     130,
		expectedNextHit: 0,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			nextHit, _ := pacer.Pace(tt.elapsedTime, tt.elapsedHits)
			if nextHit != tt.expectedNextHit {
				t.Errorf(
					"expected next hit for elapseTime %v and elapsedHits %d is %v, got %v",
					tt.elapsedTime, tt.elapsedHits,
					tt.expectedNextHit, nextHit,
				)
			}
		})
	}
}

func TestInvalidCombinedPacer(t *testing.T) {
	for _, tt := range []struct {
		name      string
		pacers    []vegeta.Pacer
		durations []time.Duration
	}{{
		name:      "pacers must not be empty",
		pacers:    make([]vegeta.Pacer, 0),
		durations: []time.Duration{10 * time.Second},
	}, {
		name:      "durations must not be empty",
		pacers:    []vegeta.Pacer{vegeta.Rate{Freq: 10, Per: 10 * time.Second}},
		durations: make([]time.Duration, 0),
	}, {
		name: "pacers and durations must have the same length",
		pacers: []vegeta.Pacer{
			vegeta.Rate{Freq: 10, Per: 10 * time.Second},
			vegeta.Rate{Freq: 10, Per: 5 * time.Second},
		},
		durations: []time.Duration{10 * time.Second},
	}, {
		name:      "pacers length must be more than 1",
		pacers:    []vegeta.Pacer{vegeta.Rate{Freq: 10, Per: 10 * time.Second}},
		durations: []time.Duration{10 * time.Second},
	}, {
		name: "duration for each pacer must be longer than 1 second",
		pacers: []vegeta.Pacer{
			vegeta.Rate{Freq: 10, Per: 10 * time.Second},
			vegeta.Rate{Freq: 10, Per: 5 * time.Second},
		},
		durations: []time.Duration{500 * time.Millisecond, 10 * time.Second},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCombined(tt.pacers, tt.durations)
			if err == nil {
				t.Errorf("the provided configuration should be invalid: %v, %v", tt.pacers, tt.durations)
			}
		})
	}
}
