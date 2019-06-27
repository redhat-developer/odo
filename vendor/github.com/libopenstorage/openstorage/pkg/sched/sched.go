package sched

import (
	"container/list"
	"fmt"
	"sync"
	"time"

	"github.com/libopenstorage/openstorage/pkg/dbg"
)

type TaskID uint64

const (
	TaskNone      = TaskID(0)
	numGoRoutines = 10
)

func ValidTaskID(t TaskID) bool { return t != TaskNone }

type ScheduleTask func(Interval)

type Scheduler interface {
	// Schedule given task at given interval.
	// Returns associated task id if scheduled successfully,
	// or a non-nil error in case of error.
	Schedule(task ScheduleTask, interval Interval,
		runAt time.Time, onlyOnce bool) (TaskID, error)

	// Cancel given task.
	Cancel(taskID TaskID) error

	// Restart scheduling.
	Start()

	// Stop scheduling.
	Stop()
}

var instance Scheduler

type taskInfo struct {
	// ID unique task identifier
	ID TaskID
	// task function to run
	task ScheduleTask
	// interval at which task is scheduled
	interval Interval
	// runtAt is next time at which task is going to be scheduled
	runAt time.Time
	// onlyOnce one time execution only
	onlyOnce bool
	// valid is true until task is not cancelled
	valid bool
	// lock for the enqueued member
	lock sync.Mutex
	// enqueued is true if task is scheduled to run
	enqueued bool
}

type manager struct {
	sync.Mutex
	// minimumInterval defines minumum task scheduling interval
	minimumInterval time.Duration
	// tasks is list of scheduled tasks
	tasks *list.List
	// currTaskID grows monotonically and gives next taskID
	currTaskID TaskID
	// ticker ticks every minimumInterval
	ticker *time.Ticker
	// started is true if schedular is not stopped
	started bool
	// enqueuedTasksLock protects enqueuedTasks
	enqueuedTasksLock sync.Mutex
	// cv signals when there are enqueuedTasks
	cv *sync.Cond
	// enqueuedTasks is list of tasks that must be run now
	enqueuedTasks *list.List
}

func (s *manager) Schedule(
	task ScheduleTask,
	interval Interval,
	runAt time.Time,
	onlyOnce bool,
) (TaskID, error) {
	s.Lock()
	defer s.Unlock()

	if task == nil {
		return TaskNone, fmt.Errorf("Invalid task specified")
	}
	now := time.Now()
	if interval.nextAfter(now).Sub(now) < time.Second {
		return TaskNone, fmt.Errorf("Minimum interval is a second")
	}

	s.currTaskID++
	t := &taskInfo{ID: s.currTaskID,
		task:     task,
		interval: interval,
		runAt:    interval.nextAfter(runAt),
		valid:    true,
		onlyOnce: onlyOnce,
		lock:     sync.Mutex{},
		enqueued: false}

	s.tasks.PushBack(t)
	return t.ID, nil
}

func (s *manager) Cancel(
	taskID TaskID,
) error {
	s.Lock()
	defer s.Unlock()

	for e := s.tasks.Front(); e != nil; e = e.Next() {
		t := e.Value.(*taskInfo)
		if t.ID == taskID {
			t.valid = false
			s.tasks.Remove(e)
			return nil
		}
	}
	return fmt.Errorf("Invalid task ID: %v", taskID)
}

func (s *manager) Stop() {
	s.Lock()
	s.ticker.Stop()
	s.started = false
	s.Unlock()

	// Stop running any scheduled tasks.
	s.enqueuedTasksLock.Lock()
	s.enqueuedTasks.Init()
	s.enqueuedTasksLock.Unlock()
}

func (s *manager) Start() {
	s.Lock()
	defer s.Unlock()

	if !s.started {
		s.ticker = time.NewTicker(s.minimumInterval)
	}
}

func (s *manager) scheduleTasks() {
	for {
		select {
		case <-s.ticker.C:
			now := time.Now()
			s.Lock()
			tasksReady := make([]*taskInfo, 0)
			for e := s.tasks.Front(); e != nil; e = e.Next() {
				t := e.Value.(*taskInfo)
				t.lock.Lock()
				if !t.enqueued &&
					(now.Equal(t.runAt) || now.After(t.runAt)) {
					tasksReady = append(tasksReady, t)
					t.enqueued = true
					if t.onlyOnce {
						s.tasks.Remove(e)
					}
				}
				t.lock.Unlock()
			}
			s.Unlock()
			s.enqueuedTasksLock.Lock()
			for _, t := range tasksReady {
				s.enqueuedTasks.PushBack(t)
			}
			s.cv.Broadcast()
			s.enqueuedTasksLock.Unlock()
		}
	}
}

func (s *manager) runTasks() {
	for {
		s.cv.L.Lock()
		if s.enqueuedTasks.Len() == 0 {
			s.cv.Wait()
		}
		var t *taskInfo
		if s.enqueuedTasks.Len() > 0 {
			t = s.enqueuedTasks.Front().Value.(*taskInfo)
			s.enqueuedTasks.Remove(s.enqueuedTasks.Front())
		}
		s.cv.L.Unlock()
		if t != nil && t.valid {
			t.task(t.interval)
			t.lock.Lock()
			t.runAt = t.interval.nextAfter(time.Now())
			t.enqueued = false
			t.lock.Unlock()
		}
	}
}

func New(minimumInterval time.Duration) Scheduler {
	m := &manager{
		tasks:             list.New(),
		currTaskID:        0,
		minimumInterval:   minimumInterval,
		ticker:            time.NewTicker(minimumInterval),
		enqueuedTasksLock: sync.Mutex{},
		enqueuedTasks:     list.New()}
	m.cv = sync.NewCond(&m.enqueuedTasksLock)
	for i := 0; i < numGoRoutines; i++ {
		go m.runTasks()
	}
	m.Start()
	go m.scheduleTasks()
	return m
}

func Init(minimumInterval time.Duration) {
	dbg.Assert(instance == nil, "Scheduler already initialized")
	instance = New(minimumInterval)
}

func Instance() Scheduler {
	return instance
}
