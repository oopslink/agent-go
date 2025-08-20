package utils

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	errs "github.com/oopslink/agent-go/pkg/commons/errors"
)

func TestNewExponentialBackOff(t *testing.T) {
	backoff := NewExponentialBackOff()

	if backoff.InitialInterval != DefaultInitialInterval {
		t.Errorf("InitialInterval = %v, want %v", backoff.InitialInterval, DefaultInitialInterval)
	}

	if backoff.RandomizationFactor != DefaultRandomizationFactor {
		t.Errorf("RandomizationFactor = %v, want %v", backoff.RandomizationFactor, DefaultRandomizationFactor)
	}

	if backoff.Multiplier != DefaultMultiplier {
		t.Errorf("Multiplier = %v, want %v", backoff.Multiplier, DefaultMultiplier)
	}

	if backoff.MaxInterval != DefaultMaxInterval {
		t.Errorf("MaxInterval = %v, want %v", backoff.MaxInterval, DefaultMaxInterval)
	}
}

func TestExponentialBackOffNextBackOff(t *testing.T) {
	backoff := &ExponentialBackOff{
		InitialInterval:     100 * time.Millisecond,
		RandomizationFactor: 0.0, // No randomization for deterministic testing
		Multiplier:          2.0,
		MaxInterval:         1 * time.Second,
	}

	// First call should use InitialInterval
	first := backoff.NextBackOff()
	if first != 100*time.Millisecond {
		t.Errorf("First NextBackOff = %v, want %v", first, 100*time.Millisecond)
	}

	// Second call should be doubled
	second := backoff.NextBackOff()
	if second != 200*time.Millisecond {
		t.Errorf("Second NextBackOff = %v, want %v", second, 200*time.Millisecond)
	}

	// Third call should be doubled again
	third := backoff.NextBackOff()
	if third != 400*time.Millisecond {
		t.Errorf("Third NextBackOff = %v, want %v", third, 400*time.Millisecond)
	}

	// Fourth call should be doubled again
	fourth := backoff.NextBackOff()
	if fourth != 800*time.Millisecond {
		t.Errorf("Fourth NextBackOff = %v, want %v", fourth, 800*time.Millisecond)
	}

	// Fifth call should be capped at MaxInterval
	fifth := backoff.NextBackOff()
	if fifth != 1*time.Second {
		t.Errorf("Fifth NextBackOff = %v, want %v", fifth, 1*time.Second)
	}
}

func TestExponentialBackOffWithRandomization(t *testing.T) {
	backoff := &ExponentialBackOff{
		InitialInterval:     100 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          2.0,
		MaxInterval:         1 * time.Second,
	}

	// Test that randomization produces values within expected range
	for i := 0; i < 10; i++ {
		backoff.Reset()
		next := backoff.NextBackOff()

		// Should be within 50% of InitialInterval
		min := 50 * time.Millisecond
		max := 150 * time.Millisecond

		if next < min || next > max {
			t.Errorf("NextBackOff with randomization = %v, should be between %v and %v", next, min, max)
		}
	}
}

func TestZeroBackOff(t *testing.T) {
	backoff := &ZeroBackOff{}

	backoff.Reset() // Should not panic

	for i := 0; i < 10; i++ {
		next := backoff.NextBackOff()
		if next != 0 {
			t.Errorf("ZeroBackOff.NextBackOff() = %v, want 0", next)
		}
	}
}

func TestStopBackOff(t *testing.T) {
	backoff := &StopBackOff{}

	backoff.Reset() // Should not panic

	for i := 0; i < 10; i++ {
		next := backoff.NextBackOff()
		if next != Stop {
			t.Errorf("StopBackOff.NextBackOff() = %v, want %v", next, Stop)
		}
	}
}

func TestConstantBackOff(t *testing.T) {
	interval := 100 * time.Millisecond
	backoff := NewConstantBackOff(interval)

	backoff.Reset() // Should not panic

	for i := 0; i < 10; i++ {
		next := backoff.NextBackOff()
		if next != interval {
			t.Errorf("ConstantBackOff.NextBackOff() = %v, want %v", next, interval)
		}
	}
}

func TestRetrySuccess(t *testing.T) {
	ctx := context.Background()
	called := false

	operation := func() (string, error) {
		called = true
		return "success", nil
	}

	result, err := Retry(ctx, operation)

	if !called {
		t.Error("Operation should have been called")
	}

	if err != nil {
		t.Errorf("Retry() error = %v, want nil", err)
	}

	if result != "success" {
		t.Errorf("Retry() result = %v, want success", result)
	}
}

func TestRetryPermanentError(t *testing.T) {
	ctx := context.Background()
	called := false

	operation := func() (string, error) {
		called = true
		return "", errs.Permanent(errors.New("permanent error"))
	}

	result, err := Retry(ctx, operation)

	if !called {
		t.Error("Operation should have been called")
	}

	if err == nil {
		t.Error("Retry() should return error for permanent error")
	}

	if result != "" {
		t.Errorf("Retry() result = %v, want empty string", result)
	}
}

func TestRetryWithMaxTries(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	operation := func() (string, error) {
		callCount++
		return "", errors.New("temporary error")
	}

	result, err := Retry(ctx, operation, WithMaxTries(3))

	if callCount != 3 {
		t.Errorf("Operation called %d times, want 3", callCount)
	}

	if err == nil {
		t.Error("Retry() should return error after max tries")
	}

	if result != "" {
		t.Errorf("Retry() result = %v, want empty string", result)
	}
}

func TestRetryWithContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	called := false

	operation := func() (string, error) {
		called = true
		cancel() // Cancel context after first call
		return "", errors.New("temporary error")
	}

	result, err := Retry(ctx, operation)

	if !called {
		t.Error("Operation should have been called")
	}

	if err == nil {
		t.Error("Retry() should return error when context is cancelled")
	}

	if result != "" {
		t.Errorf("Retry() result = %v, want empty string", result)
	}
}

func TestRetryWithRetryAfterError(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	operation := func() (string, error) {
		callCount++
		if callCount == 1 {
			return "", errs.RetryAfter(fmt.Errorf("retry err"), 1) // Retry after 1 second
		}
		return "success", nil
	}

	start := time.Now()
	result, err := Retry(ctx, operation)
	duration := time.Since(start)

	if callCount != 2 {
		t.Errorf("Operation called %d times, want 2", callCount)
	}

	if err != nil {
		t.Errorf("Retry() error = %v, want nil", err)
	}

	if result != "success" {
		t.Errorf("Retry() result = %v, want success", result)
	}

	// Should have waited at least 1 second
	if duration < time.Second {
		t.Errorf("Retry duration = %v, should be >= 1s", duration)
	}
}

func TestRetryWithNotify(t *testing.T) {
	ctx := context.Background()
	notifyCalled := false
	var notifyError error
	var notifyDuration time.Duration

	operation := func() (string, error) {
		return "", errors.New("temporary error")
	}

	notify := func(err error, duration time.Duration) {
		notifyCalled = true
		notifyError = err
		notifyDuration = duration
	}

	result, err := Retry(ctx, operation, WithMaxTries(2), WithNotify(notify))

	if !notifyCalled {
		t.Error("Notify function should have been called")
	}

	if notifyError == nil {
		t.Error("Notify should receive the error")
	}

	if notifyDuration == 0 {
		t.Error("Notify should receive a duration")
	}

	if err == nil {
		t.Error("Retry() should return error")
	}

	if result != "" {
		t.Errorf("Retry() result = %v, want empty string", result)
	}
}

func TestRetryWithCustomBackOff(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	operation := func() (string, error) {
		callCount++
		if callCount == 1 {
			return "", errors.New("temporary error")
		}
		return "success", nil
	}

	backoff := NewConstantBackOff(100 * time.Millisecond)
	result, err := Retry(ctx, operation, WithBackOff(backoff))

	if callCount != 2 {
		t.Errorf("Operation called %d times, want 2", callCount)
	}

	if err != nil {
		t.Errorf("Retry() error = %v, want nil", err)
	}

	if result != "success" {
		t.Errorf("Retry() result = %v, want success", result)
	}
}

func TestRetryWithMaxElapsedTime(t *testing.T) {
	ctx := context.Background()

	operation := func() (string, error) {
		return "", errors.New("temporary error")
	}

	start := time.Now()
	result, err := Retry(ctx, operation, WithMaxElapsedTime(100*time.Millisecond))
	duration := time.Since(start)

	if err == nil {
		t.Error("Retry() should return error after max elapsed time")
	}

	if result != "" {
		t.Errorf("Retry() result = %v, want empty string", result)
	}

	// Should have stopped within the time limit
	if duration > 200*time.Millisecond {
		t.Errorf("Retry duration = %v, should be <= 200ms", duration)
	}
}

func TestDefaultTimer(t *testing.T) {
	timer := &defaultTimer{}

	// Test Start
	timer.Start(100 * time.Millisecond)
	if timer.timer == nil {
		t.Error("Timer should be created after Start")
	}

	// Test C
	select {
	case <-timer.C():
		// Timer fired, which is expected
	case <-time.After(200 * time.Millisecond):
		t.Error("Timer should fire within 200ms")
	}

	// Test Stop
	timer.Stop()
}

func TestGetRandomValueFromInterval(t *testing.T) {
	currentInterval := 100 * time.Millisecond
	randomizationFactor := 0.5

	// Test with no randomization
	result := getRandomValueFromInterval(0.0, 0.5, currentInterval)
	if result != currentInterval {
		t.Errorf("getRandomValueFromInterval with 0.0 factor = %v, want %v", result, currentInterval)
	}

	// Test with randomization
	for i := 0; i < 10; i++ {
		result := getRandomValueFromInterval(randomizationFactor, 0.5, currentInterval)
		min := 50 * time.Millisecond
		max := 150 * time.Millisecond

		if result < min || result > max {
			t.Errorf("getRandomValueFromInterval = %v, should be between %v and %v", result, min, max)
		}
	}
}
