package event

import (
	"slices"

	"isp-config-service/entity"
)

type Compactor struct {
}

func NewCompactor() Compactor {
	return Compactor{}
}

func (c Compactor) Compact(events []entity.Event) []entity.Event {
	uniqueEvents := make([]entity.Event, 0, len(events))
	uniqueEventKeys := make(map[string]bool, len(events))
	for i := len(events) - 1; i >= 0; i-- {
		event := events[i]
		key := event.Key()
		if uniqueEventKeys[key] {
			continue
		}
		uniqueEvents = append(uniqueEvents, event)
		uniqueEventKeys[key] = true
	}
	slices.Reverse(uniqueEvents)
	return uniqueEvents
}
