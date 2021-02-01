package eventbus

import "fmt"

type marker struct {
	tree             map[string]marker
	specialListeners []EventChannel
	groupListeners   []EventChannel
}

// Subscribe a listenerto a event on a marker path
func (m *marker) subscribe(markerParts []string, listener EventChannel) error {
	if len(markerParts) == 0 {
		m.specialListeners = append(m.specialListeners, listener)
		return nil
	}

	first := markerParts[0]

	if first == "*" {
		if len(markerParts) > 1 {
			return fmt.Errorf("A asterisk is only allowed at the end of a marker path")
		}

		m.groupListeners = append(m.groupListeners, listener)
		return nil
	}

	if _, ok := m.tree[first]; !ok {
		m.tree[first] = marker{
			tree: make(map[string]marker),
		}
	}

	subMarker := m.tree[first]
	err := subMarker.subscribe(markerParts[1:], listener)
	m.tree[first] = subMarker
	return err
}

// Publish data to all matching subscribers
func (m *marker) publish(markerParts []string, event *Event) {
	publishToListeners(event, m.groupListeners)

	if len(markerParts) == 0 {
		publishToListeners(event, m.specialListeners)
		return
	}

	first := markerParts[0]

	if subMarker, ok := m.tree[first]; ok {
		subMarker.publish(markerParts[1:], event)
	}
}

// Exists checks if a marker path exists
func (m *marker) exists(markerParts []string) bool {
	if len(markerParts) == 0 {
		return true
	}

	first := markerParts[0]

	if subMarker, ok := m.tree[first]; ok {
		return subMarker.exists(markerParts[1:])
	}

	return false
}

func publishToListeners(event *Event, chans []EventChannel) {
	chansCopy := append([]EventChannel{}, chans...)

	go func(newChans []EventChannel) {
		for _, l := range newChans {
			l <- *event
		}
	}(chansCopy)
}
