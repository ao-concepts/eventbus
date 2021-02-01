package eventbus

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/ao-concepts/storage"
)

// Bus that stores the events and dispatches them to the subscribers
type Bus struct {
	sync.RWMutex
	storage          *storage.Controller
	eventModel       *EventRepository
	persistAllEvents bool
	root             marker
}

// New eventbus constructor
func New(storage *storage.Controller, persistAllEvents bool) *Bus {
	return &Bus{
		storage:          storage,
		eventModel:       &EventRepository{},
		persistAllEvents: persistAllEvents,
		root: marker{
			tree: make(map[string]marker),
		},
	}
}

// Subscribe to a event on the bus
func (b *Bus) Subscribe(eventName string, listener EventChannel) error {
	if eventName == "" {
		return fmt.Errorf("A empty string is not a valid event name")
	}

	if eventName == ":" {
		return fmt.Errorf("A colon is not a valid event name")
	}

	if listener == nil {
		return fmt.Errorf("Cannot subscribe to a event by using a nil listener")
	}

	markerParts := strings.Split(eventName, ":")

	b.Lock()
	defer b.Unlock()

	return b.root.subscribe(markerParts, listener)
}

// Publish a event on all subscribing channels
func (b *Bus) Publish(eventName string, data interface{}) error {
	if strings.Contains(eventName, "*") {
		return fmt.Errorf("A published event cannot have a asterisk in its name")
	}

	markerParts := strings.Split(eventName, ":")

	b.RLock()
	defer b.RUnlock()

	if !b.root.exists(markerParts) {
		return fmt.Errorf("There is no listener registered for the event %s", eventName)
	}

	event := &Event{Data: data, Event: eventName}

	if err := b.persistEvent(event); err != nil {
		return err
	}

	b.root.publish(markerParts, event)

	return nil
}

// returns false if an error occured.
func (b *Bus) persistEvent(event *Event) error {
	if b.storage == nil {
		return nil
	}

	if !b.persistAllEvents && strings.HasSuffix(event.Event, ":read") {
		return nil
	}

	bytes, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}

	event.JSONData = string(bytes)

	return b.storage.UseTransaction(func(tx *storage.Transaction) (err error) {
		return b.eventModel.Insert(tx, event)
	})
}
