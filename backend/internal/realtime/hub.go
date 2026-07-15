package realtime

import (
	"errors"
	"sync"
)

const subscriberQueueSize = 16

var (
	ErrHubClosed = errors.New("realtime hub is closed")
	ErrHubFull   = errors.New("realtime screening subscriber limit reached")
)

type Hub struct {
	mutex           sync.Mutex
	subscribers     map[string]map[chan SeatEvent]struct{}
	maxPerScreening int
	closed          bool
}

func NewHub(maxPerScreening int) *Hub {
	return &Hub{
		subscribers:     make(map[string]map[chan SeatEvent]struct{}),
		maxPerScreening: maxPerScreening,
	}
}

func (hub *Hub) Subscribe(screeningID string) (<-chan SeatEvent, func(), error) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	if hub.closed {
		return nil, nil, ErrHubClosed
	}

	screeningSubscribers := hub.subscribers[screeningID]
	if screeningSubscribers == nil {
		screeningSubscribers = make(map[chan SeatEvent]struct{})
		hub.subscribers[screeningID] = screeningSubscribers
	}
	if hub.maxPerScreening > 0 && len(screeningSubscribers) >= hub.maxPerScreening {
		return nil, nil, ErrHubFull
	}

	events := make(chan SeatEvent, subscriberQueueSize)
	screeningSubscribers[events] = struct{}{}

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			hub.remove(screeningID, events)
		})
	}

	return events, unsubscribe, nil
}

func (hub *Hub) Publish(event SeatEvent) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	for events := range hub.subscribers[event.ScreeningID] {
		select {
		case events <- event:
		default:
			close(events)
			delete(hub.subscribers[event.ScreeningID], events)
		}
	}

	if len(hub.subscribers[event.ScreeningID]) == 0 {
		delete(hub.subscribers, event.ScreeningID)
	}
}

func (hub *Hub) Close() {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	if hub.closed {
		return
	}
	hub.closed = true

	for _, screeningSubscribers := range hub.subscribers {
		for events := range screeningSubscribers {
			close(events)
		}
	}
	hub.subscribers = make(map[string]map[chan SeatEvent]struct{})
}

func (hub *Hub) remove(screeningID string, events chan SeatEvent) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	screeningSubscribers := hub.subscribers[screeningID]
	if _, exists := screeningSubscribers[events]; !exists {
		return
	}

	close(events)
	delete(screeningSubscribers, events)
	if len(screeningSubscribers) == 0 {
		delete(hub.subscribers, screeningID)
	}
}
