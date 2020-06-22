package v1

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetStatusCondition(t *testing.T) {
	testCases := []struct {
		name               string
		testCondition      Condition
		startConditions    *[]Condition
		expectedConditions *[]Condition
	}{
		{
			name: "add when empty",
			testCondition: Condition{
				Type:    ConditionAvailable,
				Status:  "True",
				Reason:  "Testing",
				Message: "Basic message",
			},
			startConditions: &[]Condition{},
			expectedConditions: &[]Condition{
				{
					Type:    ConditionAvailable,
					Status:  "True",
					Reason:  "Testing",
					Message: "Basic message",
				},
			},
		},
		{
			name: "add to conditions",
			testCondition: Condition{
				Type:    ConditionAvailable,
				Status:  "True",
				Reason:  "TestingAvailableTrue",
				Message: "Available condition true",
			},
			startConditions: &[]Condition{
				{
					Type:              ConditionDegraded,
					Status:            "False",
					Reason:            "TestingDegradedFalse",
					Message:           "Degraded condition false",
					LastHeartbeatTime: metav1.NewTime(time.Now()),
				},
			},
			expectedConditions: &[]Condition{
				{
					Type:    ConditionAvailable,
					Status:  "True",
					Reason:  "TestingAvailableTrue",
					Message: "Available condition true",
				},
				{
					Type:    ConditionDegraded,
					Status:  "False",
					Reason:  "TestingDegradedFalse",
					Message: "Degraded condition false",
				},
			},
		},
		{
			name: "replace condition",
			testCondition: Condition{
				Type:    ConditionDegraded,
				Status:  "True",
				Reason:  "TestingDegradedTrue",
				Message: "Degraded condition true",
			},
			startConditions: &[]Condition{
				{
					Type:    ConditionDegraded,
					Status:  "False",
					Reason:  "TestingDegradedFalse",
					Message: "Degraded condition false",
				},
			},
			expectedConditions: &[]Condition{
				{
					Type:    ConditionDegraded,
					Status:  "True",
					Reason:  "TestingDegradedTrue",
					Message: "Degraded condition true",
				},
			},
		},
		{
			name: "last heartbeat",
			testCondition: Condition{
				Type:    ConditionDegraded,
				Status:  "True",
				Reason:  "TestingDegradedTrue",
				Message: "Degraded condition true",
			},
			startConditions: &[]Condition{
				{
					Type:    ConditionDegraded,
					Status:  "True",
					Reason:  "TestingDegradedFalse",
					Message: "Degraded condition false",
				},
			},
			expectedConditions: &[]Condition{
				{
					Type:    ConditionDegraded,
					Status:  "True",
					Reason:  "TestingDegradedTrue",
					Message: "Degraded condition true",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			SetStatusCondition(tc.startConditions, tc.testCondition)
			compareConditions(t, tc.startConditions, tc.expectedConditions)
		})
	}

	return
}

func TestRemoveStatusCondition(t *testing.T) {
	testCases := []struct {
		name               string
		testConditionType  ConditionType
		startConditions    *[]Condition
		expectedConditions *[]Condition
	}{
		{
			name:               "remove when empty",
			testConditionType:  ConditionAvailable,
			startConditions:    &[]Condition{},
			expectedConditions: &[]Condition{},
		},
		{
			name:              "basic remove",
			testConditionType: ConditionAvailable,
			startConditions: &[]Condition{
				{
					Type:              ConditionAvailable,
					Status:            "True",
					Reason:            "TestingAvailableTrue",
					Message:           "Available condition true",
					LastHeartbeatTime: metav1.NewTime(time.Now()),
				},
				{
					Type:              ConditionDegraded,
					Status:            "False",
					Reason:            "TestingDegradedFalse",
					Message:           "Degraded condition false",
					LastHeartbeatTime: metav1.NewTime(time.Now()),
				},
			},
			expectedConditions: &[]Condition{
				{
					Type:    ConditionDegraded,
					Status:  "False",
					Reason:  "TestingDegradedFalse",
					Message: "Degraded condition false",
				},
			},
		},
		{
			name:              "remove last condition",
			testConditionType: ConditionAvailable,
			startConditions: &[]Condition{
				{
					Type:    ConditionAvailable,
					Status:  "True",
					Reason:  "TestingAvailableTrue",
					Message: "Available condition true",
				},
			},
			expectedConditions: &[]Condition{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			RemoveStatusCondition(tc.startConditions, tc.testConditionType)
			compareConditions(t, tc.startConditions, tc.expectedConditions)
		})
	}

	return
}

func compareConditions(t *testing.T, gotConditions *[]Condition, expectedConditions *[]Condition) {
	for _, expectedCondition := range *expectedConditions {
		testCondition := FindStatusCondition(*gotConditions, expectedCondition.Type)
		if testCondition == nil {
			t.Errorf("Condition type '%v' not found in '%v'", expectedCondition.Type, *gotConditions)
		}
		if testCondition.Status != expectedCondition.Status {
			t.Errorf("Unexpected status '%v', expected '%v'", testCondition.Status, expectedCondition.Status)
		}
		if testCondition.Message != expectedCondition.Message {
			t.Errorf("Unexpected message '%v', expected '%v'", testCondition.Message, expectedCondition.Message)
		}
		// Test for lastHeartbeatTime
		if testCondition.LastHeartbeatTime.IsZero() {
			t.Error("lastHeartbeatTime should never be zero")
		}
		timeNow := metav1.NewTime(time.Now())
		if timeNow.Before(&testCondition.LastHeartbeatTime) {
			t.Errorf("Unexpected lastHeartbeatTime '%v', should be before '%v'", testCondition.LastHeartbeatTime, timeNow)
		}
	}
}
