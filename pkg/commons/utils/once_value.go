// Package utils provides utility functions for the agent-go framework.
// This file contains a generic once-value utility for ensuring a value is only yielded once.
package utils

// OfOnceValue creates a new onceValue that will yield the given value only once.
// The value is yielded through a callback function, and subsequent calls to Get
// will not yield the value again. This is useful for ensuring a value is only
// processed or consumed once, even if the Get method is called multiple times.
func OfOnceValue[T any](v T) *onceValue[T] {
	return &onceValue[T]{
		v: v,
	}
}

// onceValue is a generic type that holds a value and ensures it's only yielded once.
// It uses a boolean flag to track whether the value has been yielded.
type onceValue[T any] struct {
	v   T    // The value to be yielded
	pop bool // Flag indicating whether the value has been yielded
}

// Get yields the value through the provided callback function, but only once.
// On the first call, it sets the pop flag to true and calls the yield function with the value.
// On subsequent calls, it returns immediately without calling the yield function.
// This ensures the value is only processed once, even if Get is called multiple times.
func (o *onceValue[T]) Get(yield func(v T)) {
	if o.pop {
		return
	}
	o.pop = true
	if yield != nil {
		yield(o.v)
	}
}
