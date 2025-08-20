// NOTICES: this code is copy from https://github.com/cenkalti/backoff
//
// Package utils provides utility functions for the agent-go framework.
// This file contains retry utilities with exponential backoff strategies
// for handling transient failures in distributed systems.

package utils

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"

	errs "github.com/oopslink/agent-go/pkg/commons/errors"
)

// ExponentialBackOff implements an exponential backoff algorithm for retry operations.
// It increases the delay between retries exponentially while adding randomization
// to prevent thundering herd problems.
type ExponentialBackOff struct {
	InitialInterval     time.Duration // Initial delay between retries
	RandomizationFactor float64       // Factor for randomizing the delay (0.0 to 1.0)
	Multiplier          float64       // Factor to multiply the delay by on each retry
	MaxInterval         time.Duration // Maximum delay between retries

	currentInterval time.Duration // Current delay interval (internal state)
}

// Default values for ExponentialBackOff.
const (
	DefaultInitialInterval     = 500 * time.Millisecond // Default initial delay
	DefaultRandomizationFactor = 0.5                    // Default randomization factor
	DefaultMultiplier          = 1.5                    // Default multiplier for exponential growth
	DefaultMaxInterval         = 60 * time.Second       // Default maximum delay
)

// NewExponentialBackOff creates an instance of ExponentialBackOff using default values.
// The default configuration provides a good balance between retry frequency and backoff growth.
func NewExponentialBackOff() *ExponentialBackOff {
	return &ExponentialBackOff{
		InitialInterval:     DefaultInitialInterval,
		RandomizationFactor: DefaultRandomizationFactor,
		Multiplier:          DefaultMultiplier,
		MaxInterval:         DefaultMaxInterval,
	}
}

// Reset the interval back to the initial retry interval and restarts the timer.
// Reset must be called before using b.
func (b *ExponentialBackOff) Reset() {
	b.currentInterval = b.InitialInterval
}

// NextBackOff calculates the next backoff interval using the formula:
//
//	Randomized interval = RetryInterval * (1 ± RandomizationFactor)
//
// The interval increases exponentially up to MaxInterval, with randomization
// to prevent synchronized retries from multiple clients.
func (b *ExponentialBackOff) NextBackOff() time.Duration {
	if b.currentInterval == 0 {
		b.currentInterval = b.InitialInterval
	}

	next := getRandomValueFromInterval(b.RandomizationFactor, rand.Float64(), b.currentInterval)
	b.incrementCurrentInterval()
	return next
}

// Increments the current interval by multiplying it with the multiplier.
// Checks for overflow and caps the interval at MaxInterval.
func (b *ExponentialBackOff) incrementCurrentInterval() {
	// Check for overflow, if overflow is detected set the current interval to the max interval.
	if float64(b.currentInterval) >= float64(b.MaxInterval)/b.Multiplier {
		b.currentInterval = b.MaxInterval
	} else {
		b.currentInterval = time.Duration(float64(b.currentInterval) * b.Multiplier)
	}
}

// Returns a random value from the following interval:
//
//	[currentInterval - randomizationFactor * currentInterval, currentInterval + randomizationFactor * currentInterval].
//
// This randomization helps prevent thundering herd problems when multiple
// clients retry operations simultaneously.
func getRandomValueFromInterval(randomizationFactor, random float64, currentInterval time.Duration) time.Duration {
	if randomizationFactor == 0 {
		return currentInterval // make sure no randomness is used when randomizationFactor is 0.
	}
	var delta = randomizationFactor * float64(currentInterval)
	var minInterval = float64(currentInterval) - delta
	var maxInterval = float64(currentInterval) + delta

	// Get a random value from the range [minInterval, maxInterval].
	// The formula used below has a +1 because if the minInterval is 1 and the maxInterval is 3 then
	// we want a 33% chance for selecting either 1, 2 or 3.
	return time.Duration(minInterval + (random * (maxInterval - minInterval + 1)))
}

// timer interface abstracts timer functionality for testing and customization.
type timer interface {
	Start(duration time.Duration) // Start the timer with the given duration
	Stop()                        // Stop the timer and free resources
	C() <-chan time.Time          // Return the timer's channel
}

// defaultTimer implements Timer interface using time.Timer
type defaultTimer struct {
	timer *time.Timer // The underlying time.Timer
}

// C returns the timers channel which receives the current time when the timer fires.
func (t *defaultTimer) C() <-chan time.Time {
	return t.timer.C
}

// Start starts the timer to fire after the given duration
func (t *defaultTimer) Start(duration time.Duration) {
	if t.timer == nil {
		t.timer = time.NewTimer(duration)
	} else {
		t.timer.Reset(duration)
	}
}

// Stop is called when the timer is not used anymore and resources may be freed.
func (t *defaultTimer) Stop() {
	if t.timer != nil {
		t.timer.Stop()
	}
}

// BackOff is a backoff policy for retrying an operation.
type BackOff interface {
	// NextBackOff returns the duration to wait before retrying the operation,
	// backoff.Stop to indicate that no more retries should be made.
	//
	// Example usage:
	//
	//     duration := backoff.NextBackOff()
	//     if duration == backoff.Stop {
	//         // Do not retry operation.
	//     } else {
	//         // Sleep for duration and retry operation.
	//     }
	//
	NextBackOff() time.Duration

	// Reset to initial state.
	Reset()
}

// Stop indicates that no more retries should be made for use in NextBackOff().
const Stop time.Duration = -1

// ZeroBackOff is a fixed backoff policy whose backoff time is always zero,
// meaning that the operation is retried immediately without waiting, indefinitely.
type ZeroBackOff struct{}

func (b *ZeroBackOff) Reset() {}

func (b *ZeroBackOff) NextBackOff() time.Duration { return 0 }

// StopBackOff is a fixed backoff policy that always returns backoff.Stop for
// NextBackOff(), meaning that the operation should never be retried.
type StopBackOff struct{}

func (b *StopBackOff) Reset() {}

func (b *StopBackOff) NextBackOff() time.Duration { return Stop }

// ConstantBackOff is a backoff policy that always returns the same backoff delay.
// This is in contrast to an exponential backoff policy,
// which returns a delay that grows longer as you call NextBackOff() over and over again.
type ConstantBackOff struct {
	Interval time.Duration // Fixed delay between retries
}

func (b *ConstantBackOff) Reset()                     {}
func (b *ConstantBackOff) NextBackOff() time.Duration { return b.Interval }

// NewConstantBackOff creates a new ConstantBackOff with the specified interval.
func NewConstantBackOff(d time.Duration) *ConstantBackOff {
	return &ConstantBackOff{Interval: d}
}

// DefaultMaxElapsedTime sets a default limit for the total retry duration.
const DefaultMaxElapsedTime = 15 * time.Minute

// Operation is a function that attempts an operation and may be retried.
type Operation[T any] func() (T, error)

// Notify is a function called on operation error with the error and backoff duration.
type Notify func(error, time.Duration)

// retryOptions holds configuration settings for the retry mechanism.
type retryOptions struct {
	BackOff        BackOff       // Strategy for calculating backoff periods.
	Timer          timer         // Timer to manage retry delays.
	Notify         Notify        // Optional function to notify on each retry error.
	MaxTries       uint          // Maximum number of retry attempts.
	MaxElapsedTime time.Duration // Maximum total time for all retries.
}

// RetryOption is a function that configures retry behavior.
type RetryOption func(*retryOptions)

// WithBackOff configures a custom backoff strategy.
func WithBackOff(b BackOff) RetryOption {
	return func(args *retryOptions) {
		args.BackOff = b
	}
}

// withTimer sets a custom timer for managing delays between retries.
func withTimer(t timer) RetryOption {
	return func(args *retryOptions) {
		args.Timer = t
	}
}

// WithNotify sets a notification function to handle retry errors.
func WithNotify(n Notify) RetryOption {
	return func(args *retryOptions) {
		args.Notify = n
	}
}

// WithMaxTries limits the number of all attempts.
func WithMaxTries(n uint) RetryOption {
	return func(args *retryOptions) {
		args.MaxTries = n
	}
}

// WithMaxElapsedTime limits the total duration for retry attempts.
func WithMaxElapsedTime(d time.Duration) RetryOption {
	return func(args *retryOptions) {
		args.MaxElapsedTime = d
	}
}

// Retry attempts the operation until success, a permanent error, or backoff completion.
// It ensures the operation is executed at least once.
//
// The function supports various retry strategies and can be configured with options:
// - Exponential backoff (default)
// - Constant backoff
// - Zero backoff (immediate retries)
// - Stop backoff (no retries)
//
// Special error types are handled:
// - PermanentError: stops retrying immediately
// - RetryAfterError: uses the specified duration for the next retry
//
// Returns the operation result or error if retries are exhausted or context is cancelled.
func Retry[T any](ctx context.Context, operation Operation[T], opts ...RetryOption) (T, error) {
	// Initialize default retry options.
	args := &retryOptions{
		BackOff:        NewExponentialBackOff(),
		Timer:          &defaultTimer{},
		MaxElapsedTime: DefaultMaxElapsedTime,
	}

	// Apply user-provided options to the default settings.
	for _, opt := range opts {
		opt(args)
	}

	defer args.Timer.Stop()

	startedAt := time.Now()
	args.BackOff.Reset()
	for numTries := uint(1); ; numTries++ {
		// Execute the operation.
		res, err := operation()
		if err == nil {
			return res, nil
		}

		// Stop retrying if maximum tries exceeded.
		if args.MaxTries > 0 && numTries >= args.MaxTries {
			return res, err
		}

		// Handle permanent errors without retrying.
		var permanent *errs.PermanentError
		if errors.As(err, &permanent) {
			return res, err
		}

		// Stop retrying if context is cancelled.
		if cerr := context.Cause(ctx); cerr != nil {
			return res, cerr
		}

		// Calculate next backoff duration.
		next := args.BackOff.NextBackOff()
		if next == Stop {
			return res, err
		}

		// Reset backoff if RetryAfterError is encountered.
		var retryAfter *errs.RetryAfterError
		if errors.As(err, &retryAfter) {
			next = retryAfter.Duration()
			args.BackOff.Reset()
		}

		// Stop retrying if maximum elapsed time exceeded.
		if args.MaxElapsedTime > 0 && time.Since(startedAt)+next > args.MaxElapsedTime {
			return res, err
		}

		// Notify on error if a notifier function is provided.
		if args.Notify != nil {
			args.Notify(err, next)
		}

		// Wait for the next backoff period or context cancellation.
		args.Timer.Start(next)
		select {
		case <-args.Timer.C():
		case <-ctx.Done():
			return res, context.Cause(ctx)
		}
	}
}
