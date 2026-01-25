package concurrency

import (
	"context"
	"iter"
)

// ClosedOrDone creates an iterator that yields values from the given channel until it is closed or the context is done.
// If it returns before the channel is closed, it consumes the channel asynchronously until it is closed so it can be cleaned up.
func ClosedOrDone[T any](ch <-chan T, ctx context.Context) iter.Seq[T] {
	return func(yield func(T) bool) {
		drain := func() {
			for range ch {
			}
		}

		defer func() { go drain() }()

		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-ch:
				if !ok || !yield(v) {
					return
				}
			}
		}
	}
}
