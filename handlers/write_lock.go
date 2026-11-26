package handlers

import (
	"context"
	"fmt"
	"sync"

	ak "github.com/alkiranet/client-go/tenant"
)

// objectLocks runs the writes to one object one at a time within a process.
// An update reads the stored object, merges the patch and sends the whole
// body, and client-go sends no If-Match, so two overlapping updates of one
// object would both merge from the same read and the second PUT would drop
// the first change. Writes to different objects run concurrently.
type objectLocks struct {
	mu sync.Mutex
	m  map[string]*objectLock
}

type objectLock struct {
	slot chan struct{}
	refs int
}

// lock waits for the slot of key until ctx ends, and returns its release.
func (l *objectLocks) lock(ctx context.Context, key string) (func(), error) {
	l.mu.Lock()
	if l.m == nil {
		l.m = map[string]*objectLock{}
	}
	o := l.m[key]
	if o == nil {
		o = &objectLock{slot: make(chan struct{}, 1)}
		l.m[key] = o
	}
	o.refs++
	l.mu.Unlock()
	select {
	case o.slot <- struct{}{}:
		return func() {
			<-o.slot
			l.release(key, o)
		}, nil
	case <-ctx.Done():
		l.release(key, o)
		return nil, ctx.Err()
	}
}

func (l *objectLocks) release(key string, o *objectLock) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if o.refs--; o.refs == 0 {
		delete(l.m, key)
	}
}

// lockObject takes the per-object write slot of the client's tenant, this
// family and object; a call whose deadline ends while it waits sends nothing.
func (w writeCall[T]) lockObject(ctx context.Context, client *ak.AlkiraClient, object string) (func(), error) {
	unlock, err := w.p.locks.lock(ctx, tenantOf(client)+"\x00"+w.f.Name+"\x00"+object)
	if err != nil {
		return nil, fmt.Errorf("%w: another write to the same %s was still running: %w", ak.ErrNotSent, w.f.Noun, err)
	}
	return unlock, nil
}

// lockedWrite waits up to the write timeout for the per-object slot, then
// starts the write timeout, so a write queued behind a slow one still gets
// the full budget for its reads and its write. done releases both.
func (w writeCall[T]) lockedWrite(ctx context.Context, client *ak.AlkiraClient, object string) (context.Context, func(), error) {
	wait, stop := context.WithTimeout(ctx, w.p.timeout)
	unlock, err := w.lockObject(wait, client, object)
	stop()
	if err != nil {
		return nil, nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, w.p.timeout)
	return ctx, func() { cancel(); unlock() }, nil
}
