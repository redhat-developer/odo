package sched

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"sync"
	"testing"
	"time"
)

type testCounter struct {
	sync.Mutex
	count int
}

func (tc *testCounter) incr() {
	tc.Lock()
	tc.count++
	tc.Unlock()
}

func TestStartStop(t *testing.T) {
	s := New(time.Second)
	s.Stop()
	require.True(t, true, "OK")
}

func TestSingle(t *testing.T) {
	s := New(time.Second)
	tc := testCounter{count: 0}
	task := func(Interval) {
		tc.incr()
	}

	s.Schedule(task, Periodic(2*time.Second), time.Now(), true)
	time.Sleep(6 * time.Second)
	require.Equal(t, tc.count, 1, "running only once")

	taskID, err := s.Schedule(task, Periodic(2*time.Second), time.Now(), false)
	require.Equal(t, err, nil, "schedule periodic")
	time.Sleep(12 * time.Second)
	require.True(t, tc.count <= 6, "periodic task")
	s.Stop()
	time.Sleep(3 * time.Second)
	stopCount := tc.count
	time.Sleep(4 * time.Second)
	require.True(t, tc.count == stopCount, "task scheduled when stopped")

	err = s.Cancel(taskID)
	require.Equal(t, err, nil, "cancel valid task")
	time.Sleep(6 * time.Second)
	require.True(t, tc.count < 7, "periodic task after stop")

	s.Stop()
}

func TestCancel(t *testing.T) {
	s := New(time.Second)
	tc := testCounter{count: 0}
	task := func(Interval) {
		tc.incr()
	}

	taskID, err := s.Schedule(task, Periodic(10*time.Second), time.Now(), true)
	require.Equal(t, err, nil, "scheduling task")
	time.Sleep(8 * time.Second)

	require.Equal(t, tc.count, 0, "task should not be scheduled")
	err = s.Cancel(taskID)
	require.Equal(t, err, nil, "cancel future task")

	time.Sleep(2 * time.Second)
	require.Equal(t, tc.count, 0, "cancelled task")
}

func TestMulti(t *testing.T) {
	s := New(time.Second)

	// Create a bunch of task counters and tasks.
	tcs := make([]*testCounter, 0)
	tasks := make([]func(Interval), 0)

	for i := 0; i < 10; i++ {
		tc := &testCounter{count: 0}
		task := func(Interval) {
			tc.incr()
		}
		tcs = append(tcs, tc)
		tasks = append(tasks, task)
	}

	// Schedule some tasks - few as runOnce and others as periodic.
	taskIDs := make([]TaskID, 0)
	for i, _ := range tasks {
		taskID, err := s.Schedule(tasks[i], Periodic(2*time.Second), time.Now(),
			i%3 == 0)
		require.Equal(t, err, nil, "schedule multi")
		taskIDs = append(taskIDs, taskID)
	}
	time.Sleep(20 * time.Second)

	// Check counters for runOnce and periodic, cancel the runOnce tasks.
	for i, tc := range tcs {
		if i%3 == 0 {
			require.True(t, tc.count == 1, "count for runOnce task")
			err := s.Cancel(taskIDs[i])
			require.NotEqual(t, err, nil, "cancelling runOnce task")
		} else {
			require.True(t, tc.count > 4, "periodic multi-task")
		}
	}

	// Stop the schedular and see that tasks are not scheduled anymore.
	s.Stop()
	time.Sleep(2 * time.Second)

	counters := make([]int, len(tcs))
	for i, tc := range tcs {
		counters[i] = tc.count
	}
	time.Sleep(6 * time.Second)
	for i, tc := range tcs {
		require.Equal(t, counters[i], tc.count,
			"counters increased after stopping schedular")
	}

	// Start the schedular and see tasks get scheduled
	s.Start()
	time.Sleep(4 * time.Second)

	for i, tc := range tcs {
		runOnce := i % 3
		if runOnce != 0 {
			require.True(t, tc.count == counters[i],
				"counters increased after stopping schedular")
			err := s.Cancel(taskIDs[i])
			require.Equal(t, err, nil, "cancelling multi-task")
		}
	}

	for _, taskID := range taskIDs {
		err := s.Cancel(taskID)
		require.NotEqual(t, err, nil, "cancelling invalid multi-task")
	}

	s.Stop()
}

func TestScheduleStrings(t *testing.T) {
	invalidSchedules := make(map[string][]string)
	invalidSchedules[DailyType] = []string{"10.15", "10.15,a", "10.15,", ",a",
		",5", ","}
	invalidSchedules[WeeklyType] = []string{"x", ",5", ",a5", "a,5", "Mon,5",
		"monday@1.4,5", "monday@1,5.5"}
	invalidSchedules[MonthlyType] = []string{"x", ",5", ",a5", "a,5",
		"10,", "10.4@", "10@1.2", "10@1,4.5", "10,1.1", "40@", "50", "@11.,55",
		"@11.,5.5", "10.1@11.,5.5", "10@11.,5", "40@11,5"}
	invalidSchedules[PeriodicType] = []string{"", ",", "x", ",x", "x,", "10.4",
		"10.4,a", "10,a", "10,4.4", "1.2,5"}

	ParseCLI[PeriodicType] = ParsePeriodic

	for schedType, intvs := range invalidSchedules {
		parser, _ := ParseCLI[schedType]
		for _, intv := range intvs {
			_, err := parser(intv)
			require.NotEqual(t, err, nil, "%s parsed as valid %s",
				intv, schedType)
		}
	}

	dockSched := []string{
		DailyType + nonYamlTypeSeparator + "10:43,5",
		MonthlyType + nonYamlTypeSeparator + "10@13:30,5",
		WeeklyType + nonYamlTypeSeparator + "monday@10:52,5",
		PeriodicType + nonYamlTypeSeparator + "10,5",
		policyTag + nonYamlTypeSeparator + "x1,x2,x3",
	}
	origLen := len(dockSched)
	for i := 0; i < origLen; i++ {
		for j := i + 1; j < origLen; j++ {
			dockSched = append(dockSched, dockSched[i]+scheduleSeparator+
				dockSched[j])
		}
	}
	for _, sched := range dockSched {
		_, _, err := ParseScheduleAndPolicies(sched)
		require.Equal(t, err, nil, "Parsing policy %s, err: %v", sched, err)
	}
}

func TestScheduleUpgrade(t *testing.T) {
	d := Daily(1, 2)
	m := Monthly(10, 4, 54)
	w := Weekly(time.Friday, 14, 30)
	p := Periodic(120 * time.Minute)
	oldSpecs := []IntervalSpec{
		d.Spec(), m.Spec(), w.Spec(), p.Spec(),
	}
	oldSpecString, err := yaml.Marshal(oldSpecs)
	str := string(oldSpecString)
	_, _, err = ParseScheduleAndPolicies(str)
	require.Equal(t, err, nil, "Parsing old intervals %v", err)
}
