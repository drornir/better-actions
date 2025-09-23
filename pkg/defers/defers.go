package defers

import "slices"

type Chain []func()

func (c *Chain) Add(f func()) {
	*c = append((*c), f)
}

func (c *Chain) Run() {
	for _, f := range slices.Backward(*c) {
		f()
	}
}

func (c *Chain) Noop() {}
