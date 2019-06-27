package keylock

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLock(t *testing.T) {
	kl := New()

	cb := make(chan *LockHandle)
	go lock(t, kl, "foo", cb)
	_, err := wait(t, kl, cb)
	require.NoError(t, err, "wait")

	endState(t, kl, 1)
}

func TestLockRelease(t *testing.T) {
	kl := ByName("test")

	cb := make(chan *LockHandle)
	go lock(t, kl, "foo", cb)
	h, err := wait(t, kl, cb)
	require.NoError(t, err, "wait")
	require.NoError(t, kl.Release(h), "unlock")

	endState(t, kl, 0)
}

func TestDoubleLock(t *testing.T) {
	kl := ByName("test")

	cb1 := make(chan *LockHandle)
	go lock(t, kl, "foo", cb1)
	h1, err := wait(t, kl, cb1)
	require.NoError(t, err, "wait")

	cb2 := make(chan *LockHandle)
	go lock(t, kl, "foo", cb2)
	h2, err := wait(t, kl, cb2)
	require.Error(t, err, "wait")

	require.NoError(t, kl.Release(h1), "unlock")

	h2, err = wait(t, kl, cb2)
	require.NoError(t, err, "wait")
	require.NoError(t, kl.Release(h2), "unlock")

	endState(t, kl, 0)
}

func TestBadRelease(t *testing.T) {
	kl := ByName("test")

	cb1 := make(chan *LockHandle)
	go lock(t, kl, "foo", cb1)
	h1, err := wait(t, kl, cb1)
	require.NoError(t, err, "wait")

	cb2 := make(chan *LockHandle)
	go lock(t, kl, "foo", cb2)
	h2, err := wait(t, kl, cb2)
	require.Error(t, err, "wait")
	require.NoError(t, kl.Release(h1), "unlock")
	h2, err = wait(t, kl, cb2)
	require.NoError(t, err, "wait")

	require.Error(t, kl.Release(h1), "unlock")
	require.NoError(t, kl.Release(h2), "unlock")

	endState(t, kl, 0)
}

func TestLottaLocks(t *testing.T) {
	kl := ByName("test")

	cb := make(chan int)
	for i := 0; i < 100; i++ {
		go lockAndSleep(t, kl, "foo", cb)
	}
	for i := 0; i != 100; {
		i += <-cb
		fmt.Printf("+")
	}
	fmt.Println("+")
	endState(t, kl, 0)
}

func lockAndSleep(t *testing.T, kl KeyLock, key string, doneCb chan<- int) {
	cb := make(chan *LockHandle)
	go lock(t, kl, key, cb)

	fmt.Printf("-")
	h, err := wait(t, kl, cb)
	for err != nil {
		h, err = wait(t, kl, cb)
	}
	time.Sleep(2)
	require.NoError(t, kl.Release(h), "unlock")
	doneCb <- 1
}

func lock(t *testing.T, kl KeyLock, key string, cb chan<- *LockHandle) {
	h := kl.Acquire(key)
	cb <- &h
}

func wait(t *testing.T,
	kl KeyLock,
	cb chan *LockHandle,
) (*LockHandle, error) {
	select {
	case h := <-cb:
		return h, nil
	case <-time.After(2 * time.Second):
		return nil, fmt.Errorf("Timeout")
	}
}

func endState(t *testing.T, kl KeyLock, count int) {
	require.Equal(t, len(kl.Dump()), count, "Total lock count")
}
