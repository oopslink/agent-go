package eventbus

import (
	"context"
	"fmt"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"sync"
	"testing"
	"time"
)

func TestEventBus_BasicFunctionality(t *testing.T) {
	// Test event creation
	event := NewEvent("test.topic", "test data")
	if event.ID == "" {
		t.Error("Event ID should not be empty")
	}
	if event.Topic != "test.topic" {
		t.Errorf("Expected topic 'test.topic', got '%s'", event.Topic)
	}
	if event.Data != "test data" {
		t.Errorf("Expected data 'test data', got '%v'", event.Data)
	}
	if event.Timestamp.IsZero() {
		t.Error("Event timestamp should not be zero")
	}
}

func TestEventBus_SynchronousSubscription(t *testing.T) {
	eb := NewEventBus()

	var receivedEvents []*Event
	var mu sync.Mutex

	handler := func(ctx context.Context, event *Event) error {
		mu.Lock()
		defer mu.Unlock()
		receivedEvents = append(receivedEvents, event)
		return nil
	}

	// Subscribe synchronously (async=false, bufferSize doesn't matter for sync)
	subscriberID, err := eb.Subscribe("test.sync", handler, false, 0)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	if subscriberID == "" {
		t.Error("Subscriber ID should not be empty")
	}

	// Publish multiple events
	events := []*Event{
		NewEvent("test.sync", "message 1"),
		NewEvent("test.sync", "message 2"),
		NewEvent("test.sync", "message 3"),
	}

	for _, event := range events {
		err := eb.Publish(event)
		if err != nil {
			t.Errorf("Failed to publish event: %v", err)
		}
	}

	// For synchronous subscription, events should be processed immediately
	mu.Lock()
	defer mu.Unlock()

	if len(receivedEvents) != 3 {
		t.Errorf("Expected 3 events, got %d", len(receivedEvents))
	}

	// Check order preservation
	for i, receivedEvent := range receivedEvents {
		expectedData := fmt.Sprintf("message %d", i+1)
		if receivedEvent.Data != expectedData {
			t.Errorf("Event %d: expected data '%s', got '%v'", i, expectedData, receivedEvent.Data)
		}
		if receivedEvent.Topic != "test.sync" {
			t.Errorf("Event %d: expected topic 'test.sync', got '%s'", i, receivedEvent.Topic)
		}
	}
}

func TestEventBus_AsynchronousSubscription(t *testing.T) {
	eb := NewEventBus()

	var receivedEvents []*Event
	var mu sync.Mutex
	var wg sync.WaitGroup

	handler := func(ctx context.Context, event *Event) error {
		mu.Lock()
		defer mu.Unlock()
		receivedEvents = append(receivedEvents, event)
		wg.Done()
		return nil
	}

	// Subscribe asynchronously with buffer
	subscriberID, err := eb.Subscribe("test.async", handler, true, 10)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	if subscriberID == "" {
		t.Error("Subscriber ID should not be empty")
	}

	// Publish multiple events
	events := []*Event{
		NewEvent("test.async", "async message 1"),
		NewEvent("test.async", "async message 2"),
		NewEvent("test.async", "async message 3"),
	}

	wg.Add(len(events))

	for _, event := range events {
		err := eb.Publish(event)
		if err != nil {
			t.Errorf("Failed to publish event: %v", err)
		}
	}

	// Wait for async processing to complete
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	if len(receivedEvents) != 3 {
		t.Errorf("Expected 3 events, got %d", len(receivedEvents))
	}

	// Check order preservation for async subscription
	for i, receivedEvent := range receivedEvents {
		expectedData := fmt.Sprintf("async message %d", i+1)
		if receivedEvent.Data != expectedData {
			t.Errorf("Event %d: expected data '%s', got '%v'", i, expectedData, receivedEvent.Data)
		}
		if receivedEvent.Topic != "test.async" {
			t.Errorf("Event %d: expected topic 'test.async', got '%s'", i, receivedEvent.Topic)
		}
	}
}

func TestEventBus_MixedSyncAsyncSubscription(t *testing.T) {
	eb := NewEventBus()

	var syncEvents []*Event
	var asyncEvents []*Event
	var syncMu, asyncMu sync.Mutex
	var wg sync.WaitGroup

	syncHandler := func(ctx context.Context, event *Event) error {
		syncMu.Lock()
		defer syncMu.Unlock()
		syncEvents = append(syncEvents, event)
		return nil
	}

	asyncHandler := func(ctx context.Context, event *Event) error {
		asyncMu.Lock()
		defer asyncMu.Unlock()
		asyncEvents = append(asyncEvents, event)
		wg.Done()
		return nil
	}

	// Subscribe both sync and async to same topic
	syncID, err := eb.Subscribe("test.mixed", syncHandler, false, 0)
	if err != nil {
		t.Fatalf("Failed to subscribe sync: %v", err)
	}

	asyncID, err := eb.Subscribe("test.mixed", asyncHandler, true, 5)
	if err != nil {
		t.Fatalf("Failed to subscribe async: %v", err)
	}

	if syncID == asyncID {
		t.Error("Sync and async subscribers should have different IDs")
	}

	// Publish events
	numEvents := 5
	wg.Add(numEvents) // Only async handler contributes to WaitGroup

	for i := 1; i <= numEvents; i++ {
		event := NewEvent("test.mixed", fmt.Sprintf("mixed message %d", i))
		err := eb.Publish(event)
		if err != nil {
			t.Errorf("Failed to publish event %d: %v", i, err)
		}
	}

	// Wait for async processing
	wg.Wait()

	// Check both handlers received all events
	syncMu.Lock()
	syncCount := len(syncEvents)
	syncMu.Unlock()

	asyncMu.Lock()
	asyncCount := len(asyncEvents)
	asyncMu.Unlock()

	if syncCount != numEvents {
		t.Errorf("Sync handler: expected %d events, got %d", numEvents, syncCount)
	}

	if asyncCount != numEvents {
		t.Errorf("Async handler: expected %d events, got %d", numEvents, asyncCount)
	}

	// Verify order preservation for both handlers
	syncMu.Lock()
	for i, event := range syncEvents {
		expectedData := fmt.Sprintf("mixed message %d", i+1)
		if event.Data != expectedData {
			t.Errorf("Sync event %d: expected '%s', got '%v'", i, expectedData, event.Data)
		}
	}
	syncMu.Unlock()

	asyncMu.Lock()
	for i, event := range asyncEvents {
		expectedData := fmt.Sprintf("mixed message %d", i+1)
		if event.Data != expectedData {
			t.Errorf("Async event %d: expected '%s', got '%v'", i, expectedData, event.Data)
		}
	}
	asyncMu.Unlock()
}

func TestEventBus_OrderPreservationUnderLoad(t *testing.T) {
	eb := NewEventBus()

	var receivedEvents []*Event
	var mu sync.Mutex
	var wg sync.WaitGroup

	handler := func(ctx context.Context, event *Event) error {
		mu.Lock()
		defer mu.Unlock()
		receivedEvents = append(receivedEvents, event)
		wg.Done()
		return nil
	}

	// Test with async subscription and larger buffer
	_, err := eb.Subscribe("test.order", handler, true, 100)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	numEvents := 50
	wg.Add(numEvents)

	// Publish events rapidly
	for i := 1; i <= numEvents; i++ {
		event := NewEvent("test.order", i) // Use integer for easier comparison
		err := eb.Publish(event)
		if err != nil {
			t.Errorf("Failed to publish event %d: %v", i, err)
		}
	}

	// Wait for all events to be processed
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	if len(receivedEvents) != numEvents {
		t.Errorf("Expected %d events, got %d", numEvents, len(receivedEvents))
	}

	// Verify strict order preservation
	for i, event := range receivedEvents {
		expectedValue := i + 1
		if event.Data != expectedValue {
			t.Errorf("Event %d: expected value %d, got %v", i, expectedValue, event.Data)
		}
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	eb := NewEventBus()

	var receivedCount int
	var mu sync.Mutex

	handler := func(ctx context.Context, event *Event) error {
		mu.Lock()
		defer mu.Unlock()
		receivedCount++
		return nil
	}

	// Subscribe
	subscriberID, err := eb.Subscribe("test.unsub", handler, false, 0)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish first event
	err = eb.Publish(NewEvent("test.unsub", "before unsubscribe"))
	if err != nil {
		t.Errorf("Failed to publish first event: %v", err)
	}

	// Unsubscribe
	err = eb.Unsubscribe("test.unsub", subscriberID)
	if err != nil {
		t.Errorf("Failed to unsubscribe: %v", err)
	}

	// Publish second event (should not be received)
	err = eb.Publish(NewEvent("test.unsub", "after unsubscribe"))
	if err != nil {
		t.Errorf("Failed to publish second event: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if receivedCount != 1 {
		t.Errorf("Expected 1 received event, got %d", receivedCount)
	}
}

func TestEventBus_MultipleTopics(t *testing.T) {
	eb := NewEventBus()

	var topic1Events, topic2Events []*Event
	var mu1, mu2 sync.Mutex

	handler1 := func(ctx context.Context, event *Event) error {
		mu1.Lock()
		defer mu1.Unlock()
		topic1Events = append(topic1Events, event)
		return nil
	}

	handler2 := func(ctx context.Context, event *Event) error {
		mu2.Lock()
		defer mu2.Unlock()
		topic2Events = append(topic2Events, event)
		return nil
	}

	// Subscribe to different topics
	_, err := eb.Subscribe("topic1", handler1, false, 0)
	if err != nil {
		t.Fatalf("Failed to subscribe to topic1: %v", err)
	}

	_, err = eb.Subscribe("topic2", handler2, false, 0)
	if err != nil {
		t.Fatalf("Failed to subscribe to topic2: %v", err)
	}

	// Publish to different topics
	err = eb.Publish(NewEvent("topic1", "data1"))
	if err != nil {
		t.Errorf("Failed to publish to topic1: %v", err)
	}

	err = eb.Publish(NewEvent("topic2", "data2"))
	if err != nil {
		t.Errorf("Failed to publish to topic2: %v", err)
	}

	err = eb.Publish(NewEvent("topic1", "data3"))
	if err != nil {
		t.Errorf("Failed to publish to topic1 again: %v", err)
	}

	// Check isolation
	mu1.Lock()
	if len(topic1Events) != 2 {
		t.Errorf("Topic1 handler: expected 2 events, got %d", len(topic1Events))
	}
	if topic1Events[0].Data != "data1" || topic1Events[1].Data != "data3" {
		t.Error("Topic1 handler received wrong events")
	}
	mu1.Unlock()

	mu2.Lock()
	if len(topic2Events) != 1 {
		t.Errorf("Topic2 handler: expected 1 event, got %d", len(topic2Events))
	}
	if topic2Events[0].Data != "data2" {
		t.Error("Topic2 handler received wrong event")
	}
	mu2.Unlock()
}

func TestEventBus_HandlerError(t *testing.T) {
	eb := NewEventBus()

	var callCount int
	var mu sync.Mutex

	errorHandler := func(ctx context.Context, event *Event) error {
		mu.Lock()
		defer mu.Unlock()
		callCount++
		return errors.Errorf(errors.InternalError, "handler error")
	}

	_, err := eb.Subscribe("test.error", errorHandler, false, 0)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish event - should not fail even if handler returns error
	err = eb.Publish(NewEvent("test.error", "test data"))
	if err != nil {
		t.Errorf("Publish should not fail when handler returns error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if callCount != 1 {
		t.Errorf("Expected handler to be called once, got %d calls", callCount)
	}
}

func TestEventBus_HandlerPanic(t *testing.T) {
	eb := NewEventBus()

	var callCount int
	var mu sync.Mutex
	var wg sync.WaitGroup

	panicHandler := func(ctx context.Context, event *Event) error {
		mu.Lock()
		defer mu.Unlock()
		callCount++
		wg.Done()
		panic("handler panic")
	}

	// Test with async handler to ensure panic recovery works in goroutine
	_, err := eb.Subscribe("test.panic", panicHandler, true, 1)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	wg.Add(1)

	// Publish event - should not crash the program
	err = eb.Publish(NewEvent("test.panic", "test data"))
	if err != nil {
		t.Errorf("Publish should not fail when handler panics: %v", err)
	}

	// Wait for async handler to complete
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	if callCount != 1 {
		t.Errorf("Expected handler to be called once, got %d calls", callCount)
	}
}

func TestEventBus_ConcurrentPublish(t *testing.T) {
	eb := NewEventBus()

	var receivedEvents []*Event
	var mu sync.Mutex
	var wg sync.WaitGroup

	handler := func(ctx context.Context, event *Event) error {
		mu.Lock()
		defer mu.Unlock()
		receivedEvents = append(receivedEvents, event)
		wg.Done()
		return nil
	}

	_, err := eb.Subscribe("test.concurrent", handler, true, 100)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	numGoroutines := 10
	eventsPerGoroutine := 10
	totalEvents := numGoroutines * eventsPerGoroutine

	wg.Add(totalEvents)

	// Start multiple goroutines publishing events concurrently
	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			for i := 0; i < eventsPerGoroutine; i++ {
				event := NewEvent("test.concurrent", fmt.Sprintf("g%d-e%d", goroutineID, i))
				if err := eb.Publish(event); err != nil {
					t.Errorf("Failed to publish event: %v", err)
				}
			}
		}(g)
	}

	// Wait for all events to be processed
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	if len(receivedEvents) != totalEvents {
		t.Errorf("Expected %d events, got %d", totalEvents, len(receivedEvents))
	}

	// Verify all events have correct topic
	for i, event := range receivedEvents {
		if event.Topic != "test.concurrent" {
			t.Errorf("Event %d: expected topic 'test.concurrent', got '%s'", i, event.Topic)
		}
	}
}

func TestEventBus_BufferOverflow(t *testing.T) {
	eb := NewEventBus()

	var processedCount int
	var mu sync.Mutex

	// Slow handler to cause buffer pressure
	slowHandler := func(ctx context.Context, event *Event) error {
		time.Sleep(10 * time.Millisecond) // Simulate slow processing
		mu.Lock()
		defer mu.Unlock()
		processedCount++
		return nil
	}

	// Subscribe with small buffer
	_, err := eb.Subscribe("test.buffer", slowHandler, true, 2)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish more events than buffer can hold
	numEvents := 5
	for i := 0; i < numEvents; i++ {
		err := eb.Publish(NewEvent("test.buffer", fmt.Sprintf("event %d", i)))
		if err != nil {
			t.Errorf("Failed to publish event %d: %v", i, err)
		}
	}

	// Give some time for processing
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	processed := processedCount
	mu.Unlock()

	// Should process at least some events (buffer + currently processing)
	if processed < 2 {
		t.Errorf("Expected at least 2 processed events, got %d", processed)
	}

	// Note: Due to the small buffer and slow handler, some events might be dropped
	// This is expected behavior - the test ensures the system doesn't crash
}

func TestEventBus_Seq(t *testing.T) {
	eb := NewEventBus()
	eb.Subscribe("test", func(ctx context.Context, event *Event) error {
		fmt.Println(fmt.Sprintf("[x] %s", event.Data))
		return nil
	}, true, 10)

	for i := 0; i < 100; i++ {
		msg := fmt.Sprintf("[%d] xxx", i)
		eb.Publish(NewEvent("test", msg))
	}

	time.Sleep(1 * time.Second)
}
