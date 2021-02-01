package eventbus

import (
	"github.com/ao-concepts/storage"
	"gorm.io/gorm"
)

// Event that triggers a state change
type Event struct {
	gorm.Model
	Data     interface{} `gorm:"-"`
	JSONData string
	Event    string
}

// EventChannel channel that transports events
type EventChannel chan Event

// EventRepository event storage repository
type EventRepository struct {
	storage.Repository
}

// GetAll events
func (m *EventRepository) GetAll(tx *storage.Transaction) (entries []Event, err error) {
	return entries, tx.Gorm().Find(&entries).Error
}
