// Package ctxutil holds utilities related to the context package.
package ctxutil

import (
	"sync"
	"time"

	"golang.org/x/net/context"
)

type joinedContext struct {
	done1, done2 <-chan struct{}
	ctx1, ctx2   context.Context

	// doneOnce guards done.
	doneOnce sync.Once
	done     <-chan struct{}

	// errOnce guards err.
	errOnce sync.Once
	err     error
}

// Join joins the two contexts into one context that contains
// all the values in each and that is canceled when
// either is canceled. If both contain a value with the
// same key, ctx1 is preferred. If Err is called when
// both are done, the error from ctx1 will be returned.
func Join(ctx1, ctx2 context.Context) context.Context {
	ctx := &joinedContext{
		done1: ctx1.Done(),
		done2: ctx2.Done(),
		ctx1:  ctx1,
		ctx2:  ctx2,
	}
	switch {
	case ctx.done1 != nil && ctx.done2 == nil:
		ctx.done = ctx.done1
	case ctx.done1 == nil && ctx.done2 != nil:
		ctx.done = ctx.done2
	}
	return ctx
}

// Deadline implements context.Context.Deadline
// by returning the earlier of the two deadlines.
func (ctx *joinedContext) Deadline() (deadline time.Time, ok bool) {
	d1, ok1 := ctx.ctx1.Deadline()
	d2, ok2 := ctx.ctx2.Deadline()
	switch {
	case ok1 && ok2:
		if d1.Before(d2) {
			return d1, true
		}
		return d2, true
	case ok1:
		return d1, true
	}
	return d2, ok2
}

// Done implements context.Context.Done by returning
// a channel which is closed when either child context's
// Done value is closed.
func (ctx *joinedContext) Done() <-chan struct{} {
	if ctx.done1 == nil || ctx.done2 == nil {
		// Easy case when we don't need to combine two done channels.
		return ctx.done
	}
	// Start a goroutine to wait for either child context to be done.
	// Note that we do assume that a non-nil Done channel will always
	// eventually be closed. If it isn't we should consider that a bug.
	ctx.doneOnce.Do(func() {
		done := make(chan struct{})
		ctx.done = done
		go func() {
			select {
			case <-ctx.done1:
			case <-ctx.done2:
			}
			close(done)
		}()
	})
	return ctx.done
}

// Err implements context.Context.Err by returning an error from
// either child context.
func (ctx *joinedContext) Err() error {
	err1 := ctx.ctx1.Err()
	err2 := ctx.ctx2.Err()
	if err1 == nil && err2 == nil {
		return nil
	}
	// Make sure that once we have returned an error, we always
	// return the same one.
	ctx.errOnce.Do(func() {
		if err1 != nil {
			// If both contexts have an error, we prefer the error from
			// the first context.
			ctx.err = err1
		} else {
			ctx.err = err2
		}
	})
	return ctx.err
}

// Value implements context.Context.Value by returning a value
// from either child context, giving precedence to ctx.ctx1.
func (ctx *joinedContext) Value(key interface{}) interface{} {
	if v := ctx.ctx1.Value(key); v != nil {
		return v
	}
	return ctx.ctx2.Value(key)
}
