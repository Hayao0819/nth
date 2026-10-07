package domain

import (
	"context"
	"errors"
	"sync"
)

var ErrRefreshCanceled = errors.New("browser session refresh canceled")

type RefreshRequest struct {
	profile string
	refresh func(context.Context) error
	done    chan struct{}
	once    sync.Once
	mu      sync.RWMutex
	result  error
}

func NewRefreshRequest(profile string, refresh func(context.Context) error) *RefreshRequest {
	return &RefreshRequest{
		profile: profile,
		refresh: refresh,
		done:    make(chan struct{}),
	}
}

func (r *RefreshRequest) Profile() string {
	return r.profile
}

func (r *RefreshRequest) Refresh(ctx context.Context) error {
	if r == nil || r.refresh == nil {
		return errors.New("browser session cannot be refreshed")
	}

	return r.refresh(ctx)
}

func (r *RefreshRequest) Complete(err error) {
	if r == nil {
		return
	}
	r.once.Do(func() {
		r.mu.Lock()
		r.result = err
		r.mu.Unlock()
		close(r.done)
	})
}

func (r *RefreshRequest) Wait(ctx context.Context) error {
	if r == nil {
		return errors.New("browser session refresh is unavailable")
	}
	select {
	case <-r.done:
		r.mu.RLock()
		defer r.mu.RUnlock()

		return r.result
	case <-ctx.Done():
		return context.Cause(ctx)
	}
}

func (r *RefreshRequest) Result() (error, bool) {
	if r == nil {
		return errors.New("browser session refresh is unavailable"), true
	}
	select {
	case <-r.done:
		r.mu.RLock()
		defer r.mu.RUnlock()

		return r.result, true
	default:
		return nil, false
	}
}
