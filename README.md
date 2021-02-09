# ao-concepts eventbus module

![CI](https://github.com/ao-concepts/eventbus/workflows/CI/badge.svg)
[![codecov](https://codecov.io/gh/ao-concepts/eventbus/branch/main/graph/badge.svg?token=AQVUZTRGQS)](https://codecov.io/gh/ao-concepts/eventbus)

This module provides a eventbus for application internal use with optional persistence. The eventbus uses a tree structure for efficient routing.

## Information

The ao-concepts ecosystem is still under active development and therefore the API of this module may have breaking changes until there is a first stable release.

If you are interested in contributing to this project, feel free to open a issue to discus a new feature, enhancement or improvement. If you found a bug or security vulnerability in this package, please start a issue, or open a PR against `master`.

## Installation

```shell
go get -u github.com/ao-concepts/eventbus
```

## Usage

A event name consists of a set of markers. All markers are separated by colons (e.g. `user:profile:update`).
These markers build a hierarchy by this hierarchy it is possible to listen to a set of events.
This means a subscriber to a `user:profile:*` event will listen to all events that start with `user:profile:`.

If the `Bus` is configured as persistent, it will automatically persist all published events that have subscribers and do not end with the `:read` suffix.
If you really want to persist all the events that have subscribers you can set `PersistAllEvents` to true. But be aware: this may result in a lot of irrelevant data.

Make sure you have migrated your database before using a persisted eventbus. An easy way o do this (not for production use) is to use gorms AutoMigrate function:

```go
log := logging.New(logging.Debug, nil)
c, err := storage.New(sqlite.Open(":memory:"), log)
if err != nil {
    log.ErrFatal(err)
}

if err := c.UseTransaction(func(tx *storage.Transaction) (err error) {
    return tx.Gorm().AutoMigrate(&eventbus.Event{})
}); err != nil {
    log.ErrFatal(err)
}
```

General usage:

```go
log := logging.New(logging.Debug, nil)
c, err := storage.New(sqlite.Open(":memory:"), log)
if err != nil {
    log.ErrFatal(err)
}

bus := eventbus.New(c, false)

ch := make(chan eventbus.Event)
err := bus.Subscribe("user:profile:update", ch)
if err != nil {
    log.ErrError(err)
}

err := bus.Publish("user:profile:update", &User{})
if err != nil {
    log.ErrError(err)
}
```

## Used packages 

This project uses some really great packages. Please make sure to check them out!

| Package                                                                  | Usage              |
| ------------------------------------------------------------------------ | ------------------ |
| [github.com/ao-concepts/storage](https://github.com/ao-concepts/storage) | Persistence helper |
| [github.com/stretchr/testify](https://github.com/stretchr/testify)       | Testing            |
| [gorm.io/gorm](gorm.io/gorm)                                             | Persisting events  |
