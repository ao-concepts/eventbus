package eventbus_test

import (
	"math"
	"sync"
	"testing"

	"github.com/ao-concepts/eventbus"
	"github.com/ao-concepts/logging"
	"github.com/ao-concepts/storage"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)
	log := logging.New(logging.Debug, nil)
	c, err := storage.New(sqlite.Open(":memory:"), log)
	assert.Nil(err)

	assert.NotNil(eventbus.New(c, false))
	assert.NotNil(eventbus.New(nil, false))
	assert.NotNil(eventbus.New(c, true))
	assert.NotNil(eventbus.New(nil, true))
}

func TestBus_Subscribe(t *testing.T) {
	assert := assert.New(t)
	log := logging.New(logging.Debug, nil)
	c, err := storage.New(sqlite.Open(":memory:"), log)
	assert.Nil(err)
	bus := eventbus.New(c, false)

	assert.NotNil(bus.Subscribe("", nil))
	assert.NotNil(bus.Subscribe(":", nil))
	assert.NotNil(bus.Subscribe("test:create", nil))

	ch := make(chan eventbus.Event)
	assert.NotNil(bus.Subscribe("test:*:create", ch))
	assert.Nil(bus.Subscribe("test:create", ch))
	assert.Nil(bus.Subscribe("test:*", ch))
}

func TestBus_Publish_NonPersistent(t *testing.T) {
	assert := assert.New(t)
	// log := logging.New(logging.Debug, nil)
	// c, err := storage.New(sqlite.Open(":memory:"), log)
	// assert.Nil(err)
	bus := eventbus.New(nil, false)

	ch := make(chan eventbus.Event)
	assert.Nil(bus.Subscribe("test:create", ch))
	ch2 := make(chan eventbus.Event)
	assert.Nil(bus.Subscribe("test:create", ch2))
	ch3 := make(chan eventbus.Event)
	assert.Nil(bus.Subscribe("test:*", ch3))
	ch4 := make(chan eventbus.Event)
	assert.Nil(bus.Subscribe("test:delete", ch4))
	ch5 := make(chan eventbus.Event)
	assert.Nil(bus.Subscribe("other:create", ch5))

	wg := sync.WaitGroup{}
	chCounter := 0
	ch2Counter := 0
	ch3Counter := 0
	ch4Counter := 0
	ch5Counter := 0

	go func() {
		for {
			select {
			case d := <-ch:
				chCounter++
				assert.Equal("test-data", d.Data)
				wg.Done()
			case d := <-ch2:
				ch2Counter++
				assert.Equal("test-data", d.Data)
				wg.Done()
			case d := <-ch3:
				ch3Counter++
				assert.Equal("test-data", d.Data)
				wg.Done()
			case d := <-ch4:
				ch4Counter++
				assert.Equal("test-data", d.Data)
				wg.Done()
			case d := <-ch5:
				ch5Counter++
				assert.Equal("test-data", d.Data)
				wg.Done()
			}
		}
	}()

	// test:create
	wg.Add(3)
	assert.Nil(bus.Publish("test:create", "test-data"))
	wg.Wait()
	assert.Equal(1, chCounter)
	assert.Equal(1, ch2Counter)
	assert.Equal(1, ch3Counter)
	assert.Equal(0, ch4Counter)
	assert.Equal(0, ch5Counter)

	// test:delete
	wg.Add(2)
	assert.Nil(bus.Publish("test:delete", "test-data"))
	wg.Wait()
	assert.Equal(1, chCounter)
	assert.Equal(1, ch2Counter)
	assert.Equal(2, ch3Counter)
	assert.Equal(1, ch4Counter)
	assert.Equal(0, ch5Counter)

	// invalid
	assert.NotNil(bus.Publish("not:existing", "test-data"))
	assert.NotNil(bus.Publish("invalid:*", "test-data"))
}

func TestBus_Publish_Persistent_NotAll(t *testing.T) {
	assert := assert.New(t)
	log := logging.New(logging.Debug, nil)
	c, err := storage.New(sqlite.Open(":memory:"), log)
	assert.Nil(err)

	assert.Nil(c.UseTransaction(func(tx *storage.Transaction) (err error) {
		return tx.Gorm().AutoMigrate(&eventbus.Event{})
	}))

	bus := eventbus.New(c, false)

	ch := make(chan eventbus.Event)
	assert.Nil(bus.Subscribe("test:read", ch))
	ch2 := make(chan eventbus.Event)
	assert.Nil(bus.Subscribe("test:create", ch2))

	wg := sync.WaitGroup{}
	chCounter := 0
	ch2Counter := 0

	go func() {
		for {
			select {
			case d := <-ch:
				chCounter++
				assert.Equal("test-data", d.Data)
				wg.Done()
			case d := <-ch2:
				ch2Counter++
				assert.Equal("test-data", d.Data)
				wg.Done()
			}
		}
	}()

	// test:read
	wg.Add(1)
	assert.Nil(bus.Publish("test:read", "test-data"))
	wg.Wait()
	assert.Equal(1, chCounter)
	assert.Equal(0, ch2Counter)

	// test:create
	wg.Add(1)
	assert.Nil(bus.Publish("test:create", "test-data"))
	wg.Wait()
	assert.Equal(1, chCounter)
	assert.Equal(1, ch2Counter)

	assert.NotNil(bus.Publish("test:create", math.Inf(1)))

	assert.Nil(c.UseTransaction(func(tx *storage.Transaction) (err error) {
		r := eventbus.EventRepository{}
		events, err := r.GetAll(tx)
		assert.Nil(err)
		assert.Len(events, 1)
		return nil
	}))
}

func TestBus_Publish_Persistent_All(t *testing.T) {
	assert := assert.New(t)
	log := logging.New(logging.Debug, nil)
	c, err := storage.New(sqlite.Open(":memory:"), log)
	assert.Nil(err)

	assert.Nil(c.UseTransaction(func(tx *storage.Transaction) (err error) {
		return tx.Gorm().AutoMigrate(&eventbus.Event{})
	}))

	bus := eventbus.New(c, true)

	ch := make(chan eventbus.Event)
	assert.Nil(bus.Subscribe("test:read", ch))
	ch2 := make(chan eventbus.Event)
	assert.Nil(bus.Subscribe("test:create", ch2))

	wg := sync.WaitGroup{}
	chCounter := 0
	ch2Counter := 0

	go func() {
		for {
			select {
			case d := <-ch:
				chCounter++
				assert.Equal("test-data", d.Data)
				wg.Done()
			case d := <-ch2:
				ch2Counter++
				assert.Equal("test-data", d.Data)
				wg.Done()
			}
		}
	}()

	// test:read
	wg.Add(1)
	assert.Nil(bus.Publish("test:read", "test-data"))
	wg.Wait()
	assert.Equal(1, chCounter)
	assert.Equal(0, ch2Counter)

	// test:create
	wg.Add(1)
	assert.Nil(bus.Publish("test:create", "test-data"))
	wg.Wait()
	assert.Equal(1, chCounter)
	assert.Equal(1, ch2Counter)

	assert.NotNil(bus.Publish("test:create", math.Inf(1)))

	assert.Nil(c.UseTransaction(func(tx *storage.Transaction) (err error) {
		r := eventbus.EventRepository{}
		events, err := r.GetAll(tx)
		assert.Nil(err)
		assert.Len(events, 2)
		return nil
	}))
}
