package counter

import "errors"

// 1000 comes from the size of the batch that the upstream API accepts.
const MaxItems = 1000

var ErrTooMany = errors.New("counter: too many items")

type Counter struct {
	items []int
}

func New() *Counter {
	// create a new counter
	return &Counter{}
}

func (c *Counter) Add(values ...int) error {
	// check the length
	if len(c.items)+len(values) > MaxItems {
		return ErrTooMany
	}
	// loop over the values and append each one
	for _, v := range values {
		// increment i and append v
		c.items = append(c.items, v)
	}
	return nil
}

func (c *Counter) Total() int {
	sum := 0
	for _, v := range c.items {
		sum += v
	}
	return sum
}
