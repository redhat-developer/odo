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

func TestSteadyUpPacer(t *testing.T) {
	minRate := vegeta.Rate{Freq: 1, Per: time.Second}
	maxRate := vegeta.Rate{Freq: 5, Per: time.Second}
	pacer, _ := NewSteadyUp(minRate, maxRate, 10*time.Second)

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
		expectedNextHit: 853658536 * time.Nanosecond,
	}, {
		name:            "test the up hit",
		elapsedTime:     3 * time.Second,
		elapsedHits:     5,
		expectedNextHit: 520407468 * time.Nanosecond,
	}, {
		name:            "test the catch up hit",
		elapsedTime:     4 * time.Second,
		elapsedHits:     2,
		expectedNextHit: 0,
	}, {
		name:            "test the steady hit",
		elapsedTime:     7052432251 * time.Nanosecond,
		elapsedHits:     17,
		expectedNextHit: 258228895 * time.Nanosecond,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			nextHit, _ := pacer.Pace(tt.elapsedTime, tt.elapsedHits)
			if nextHit != tt.expectedNextHit {
				t.Errorf(
					"expected next hit for elapsedTime %v and elapsedHits %d is %v, got %v",
					tt.elapsedTime, tt.elapsedHits,
					tt.expectedNextHit, nextHit,
				)
			}
		})
	}
}

func TestInvalidSteadyUpPacer(t *testing.T) {
	for _, tt := range []struct {
		name       string
		min        vegeta.Rate
		max        vegeta.Rate
		upDuration time.Duration
	}{{
		name:       "up duration must be larger than 0",
		min:        vegeta.Rate{Freq: 10, Per: time.Second},
		max:        vegeta.Rate{Freq: 5, Per: time.Second},
		upDuration: 0,
	}, {
		name:       "min rate must be larger than 0",
		min:        vegeta.Rate{Freq: 0, Per: time.Second},
		max:        vegeta.Rate{Freq: 5, Per: time.Second},
		upDuration: 10 * time.Second,
	}, {
		name:       "max rate must be larger than 0",
		min:        vegeta.Rate{Freq: 10, Per: time.Second},
		max:        vegeta.Rate{Freq: 0, Per: time.Second},
		upDuration: 10 * time.Second,
	}, {
		name:       "min rate must be smaller than max rate",
		min:        vegeta.Rate{Freq: 10, Per: time.Second},
		max:        vegeta.Rate{Freq: 6, Per: time.Second},
		upDuration: 10 * time.Second,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSteadyUp(tt.min, tt.max, tt.upDuration)
			if err == nil {
				t.Errorf("the provided configuration should be invalid: %v, %v, %v", tt.min, tt.max, tt.upDuration)
			}
		})
	}
}
