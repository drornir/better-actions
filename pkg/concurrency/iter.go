package concurrency

import (
	"context"
	"iter"
)

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
