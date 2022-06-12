package task

import (
	"errors"
	"testing"
	"time"
)

func TestRetryable_RetryWithSchedule(t *testing.T) {
	var empty struct{}
	for _, tt := range []struct {
		name            string
		runner          func(nbInvocations int) (exitCondition bool, result interface{}, err error)
		errorIfTimeout  bool
		schedule        []time.Duration
		wantErr         bool
		wantInvocations int
	}{
		{
			name: "no schedule with runner returning no error",
			runner: func(_ int) (exitCondition bool, result interface{}, err error) {
				return false, empty, nil
			},
			wantInvocations: 1,
		},
		{
			name: "no schedule with runner returning an error",
			runner: func(_ int) (exitCondition bool, result interface{}, err error) {
				return false, empty, errors.New("some error")
			},
			wantErr:         true,
			wantInvocations: 1,
		},
		{
			name: "schedule with runner returning no error and exit condition never matched",
			runner: func(_ int) (exitCondition bool, result interface{}, err error) {
				return false, empty, nil
			},
			schedule: []time.Duration{
				10 * time.Millisecond,
				30 * time.Millisecond,
				50 * time.Millisecond,
			},
			wantInvocations: 3,
		},
		{
			name: "schedule with runner returning an error and exit condition never matched",
			runner: func(_ int) (exitCondition bool, result interface{}, err error) {
				return false, empty, errors.New("some error")
			},
			schedule: []time.Duration{
				10 * time.Millisecond,
				30 * time.Millisecond,
				50 * time.Millisecond,
			},
			wantInvocations: 3,
			wantErr:         true,
		},
		{
			name: "schedule with runner returning no error and exit condition never matched and error if timeout set to true",
			runner: func(_ int) (exitCondition bool, result interface{}, err error) {
				return false, empty, nil
			},
			schedule: []time.Duration{
				10 * time.Millisecond,
				30 * time.Millisecond,
				50 * time.Millisecond,
			},
			wantInvocations: 3,
			errorIfTimeout:  true,
			wantErr:         true,
		},
		{
			name: "schedule with runner returning an error and exit condition never matched and error if timeout set to true",
			runner: func(_ int) (exitCondition bool, result interface{}, err error) {
				return false, empty, errors.New("some error")
			},
			schedule: []time.Duration{
				10 * time.Millisecond,
				30 * time.Millisecond,
				50 * time.Millisecond,
			},
			wantInvocations: 3,
			errorIfTimeout:  true,
			wantErr:         true,
		},
		{
			name: "schedule with runner return no error and matching exit condition after 2nd invocation",
			runner: func(n int) (exitCondition bool, result interface{}, err error) {
				if n == 2 {
					return true, empty, nil
				}
				return false, empty, nil
			},
			schedule: []time.Duration{
				10 * time.Millisecond,
				30 * time.Millisecond,
				50 * time.Millisecond,
				100 * time.Millisecond,
			},
			wantInvocations: 2,
		},
		{
			name: "schedule with runner return an error and matching exit condition after 2nd invocation",
			runner: func(n int) (exitCondition bool, result interface{}, err error) {
				err = errors.New("some error")
				if n == 2 {
					return true, empty, err
				}
				return false, empty, err
			},
			schedule: []time.Duration{
				10 * time.Millisecond,
				30 * time.Millisecond,
				50 * time.Millisecond,
				100 * time.Millisecond,
			},
			wantInvocations: 2,
			wantErr:         true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var nbRunnerInvocations int
			_, err := NewRetryable(tt.name, func() (exitCondition bool, result interface{}, err error) {
				nbRunnerInvocations++
				return tt.runner(nbRunnerInvocations)
			}, tt.errorIfTimeout).RetryWithSchedule(tt.schedule...)

			if tt.wantErr != (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantInvocations != nbRunnerInvocations {
				t.Errorf("expected %d, got %d", tt.wantInvocations, nbRunnerInvocations)
			}
		})
	}
}
