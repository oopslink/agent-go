package eventbus

import (
	"context"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/oopslink/agent-go/pkg/support/journal"
)

type EventHandler func(ctx context.Context, event *Event) error

func NewEvent(topic string, data any) *Event {
	return &Event{
		ID:        uuid.NewString(),
		Timestamp: time.Now(),
		Topic:     topic,
		Data:      data,
	}
}

type Event struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`

	Topic string `json:"topic"`
	Data  any    `json:"data"`
}

func newSubscriber(handler EventHandler, async bool, bufferSize int) *subscriber {
	return &subscriber{
		id: uuid.NewString(),

		eventChan: make(chan *Event, bufferSize),
		handler:   handler,
		async:     async,
	}
}

type subscriber struct {
	id string

	eventChan chan *Event
	handler   EventHandler
	async     bool

	mu     sync.Mutex
	closed bool
}

func (s *subscriber) start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return errors.New(ErrorCodeSubscriberAlreadyClosed)
	}

	if s.async {
		go s.asyncEventLoop()
	}

	return nil
}

func (s *subscriber) close() {
	s.mu.Lock()
	defer func() {
		s.closed = true
		s.mu.Unlock()
	}()

	close(s.eventChan)
}

func (s *subscriber) handleEvent(ctx context.Context, event *Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	if s.async {
		s.eventChan <- event
	} else {
		s.safeHandle(ctx, event)
	}
}

func (s *subscriber) asyncEventLoop() {
	for {
		select {
		case event, ok := <-s.eventChan:
			if !ok {
				return
			}
			s.safeHandle(context.Background(), event)
		}
	}
}

func (s *subscriber) safeHandle(ctx context.Context, event *Event) {
	defer func() {
		if r := recover(); r != nil {
			journal.Error("subscriber/handle", "subscriber/"+s.id, "recover from panic", "event", event, "error", r)
		}
	}()
	if err := s.handler(ctx, event); err != nil {
		journal.Error("subscriber/handle", "subscriber/"+s.id, "failed to handle event", "event", event, "error", err)
	}
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]*subscriber),
	}
}

type EventBus struct {
	mu          sync.RWMutex
	closed      bool
	subscribers map[string][]*subscriber
}

func (eb *EventBus) Subscribe(topic string, handler EventHandler, async bool, bufferSize int) (string, error) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		return "", errors.New(ErrorCodeEventBusAlreadyClosed)
	}

	sub := newSubscriber(handler, async, bufferSize)
	eb.subscribers[topic] = append(eb.subscribers[topic], sub)

	if async {
		if err := sub.start(); err != nil {
			return "", err
		}
	}

	return sub.id, nil
}

func (eb *EventBus) Unsubscribe(topic, subscriberId string) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		return errors.New(ErrorCodeEventBusAlreadyClosed)
	}

	subscribers, exists := eb.subscribers[topic]
	if !exists {
		return nil
	}

	for i, sub := range subscribers {
		if sub.id == subscriberId {
			sub.close()
			eb.subscribers[topic] = append(subscribers[:i], subscribers[i+1:]...)
			return nil
		}
	}

	return nil
}

func (eb *EventBus) Publish(event *Event) error {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if eb.closed {
		return errors.New(ErrorCodeEventBusAlreadyClosed)
	}

	for _, sub := range eb.subscribers[event.Topic] {
		sub.handleEvent(context.Background(), event)
	}
	return nil
}

func (eb *EventBus) Close() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.closed = true

	for _, subscribers := range eb.subscribers {
		for _, sub := range subscribers {
			sub.close()
		}
	}
}
