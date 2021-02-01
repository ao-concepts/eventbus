package eventbus_test

import (
	"testing"

	"github.com/ao-concepts/eventbus"
	"github.com/ao-concepts/logging"
	"github.com/ao-concepts/storage"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
)

func TestEventRepository(t *testing.T) {
	assert := assert.New(t)
	r := eventbus.EventRepository{}
	log := logging.New(logging.Debug, nil)
	c, err := storage.New(sqlite.Open(":memory:"), log)
	assert.Nil(err)

	assert.Nil(c.UseTransaction(func(tx *storage.Transaction) (err error) {
		assert.Nil(tx.Gorm().AutoMigrate(&eventbus.Event{}))
		events, err := r.GetAll(tx)
		assert.Nil(err)
		assert.Len(events, 0)

		assert.Nil(r.Insert(tx, &eventbus.Event{
			Event: "test",
			Data:  "some-data",
		}))

		events, err = r.GetAll(tx)
		assert.Nil(err)
		assert.Len(events, 1)
		return nil
	}))
}
