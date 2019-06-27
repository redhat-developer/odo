package keylock

import (
	"fmt"
	"sync"
)

// ErrKeyLockNotFound error type for lock object not found
type ErrKeyLockNotFound struct {
	// ID unique object identifier.
	ID string
}

func (e *ErrKeyLockNotFound) Error() string {
	return fmt.Sprintf("Lock with ID: %v not found", e.ID)
}

// ErrInvalidHandle error type for invalid lock handle.
type ErrInvalidHandle struct {
	// ID unique object identifier.
	ID string
}

func (e *ErrInvalidHandle) Error() string {
	return fmt.Sprintf("Invalid Handle with ID: %v", e.ID)
}

// KeyLock is a thread-safe interface for acquiring locks on arbitrary strings.
type KeyLock interface {
	// Acquire a lock associated with the specified ID.
	// Creates the lock if one doesn't already exist.
	Acquire(id string) LockHandle

	// Release the lock associated with the specified LockHandle
	// Returns an error if it is an invalid LockHandle.
	Release(h *LockHandle) error

	// Dump all locks.
	Dump() []string
}

// LockHandle is an opaque handle to an aquired lock.
type LockHandle struct {
	id     string
	genNum int64
	refcnt int64
	mutex  *sync.Mutex
}

type keyLock struct {
	sync.Mutex
	lockMap map[string]*LockHandle
}

var (
	klLock sync.Mutex
	klMap  = make(map[string]KeyLock)
)

// New returns a new instance of a KeyLock.
func New() KeyLock {
	return &keyLock{lockMap: make(map[string]*LockHandle)}
}

// ByName creates a new instance or returns an existing instance
// if found in the map.
func ByName(klName string) KeyLock {
	klLock.Lock()
	defer klLock.Unlock()

	kl, ok := klMap[klName]
	if !ok {
		kl = New()
		klMap[klName] = kl
	}
	return kl
}

func (kl *keyLock) Acquire(id string) LockHandle {
	h := kl.getOrCreateLock(id)
	h.mutex.Lock()
	h.genNum++
	return *h
}

func (kl *keyLock) Release(h *LockHandle) error {
	if len(h.id) == 0 {
		return &ErrInvalidHandle{}
	}

	kl.Lock()
	defer kl.Unlock()
	lockedH, exists := kl.lockMap[h.id]
	if !exists {
		return &ErrKeyLockNotFound{ID: h.id}
	}
	if h.genNum != lockedH.genNum {
		return &ErrInvalidHandle{ID: h.id}
	}
	lockedH.mutex.Unlock()
	lockedH.refcnt--
	if lockedH.refcnt == 0 {
		delete(kl.lockMap, h.id)
	}
	return nil
}

func (kl *keyLock) Dump() []string {
	kl.Lock()
	defer kl.Unlock()

	keys := make([]string, len(kl.lockMap))
	i := 0
	for k := range kl.lockMap {
		keys[i] = k
	}
	return keys
}

func (kl *keyLock) getOrCreateLock(id string) *LockHandle {
	kl.Lock()
	defer kl.Unlock()

	h, exists := kl.lockMap[id]
	if !exists {
		h = &LockHandle{
			mutex: &sync.Mutex{},
			id:    id,
		}
		kl.lockMap[id] = h
	}
	h.refcnt++
	return h
}
