package ddd

import (
	"errors"
	"slices"
)

type Aggregate struct {
	Entity
	domainEvents map[ID]DomainEventer
}

func NewAggregate(id ID) Aggregate {
	return Aggregate{
		Entity:       Entity{id: id},
		domainEvents: map[ID]DomainEventer{},
	}
}

func (agg *Aggregate) Events() []DomainEventer {
	events := []DomainEventer{}
	for _, ev := range agg.domainEvents {
		events = append(events, ev)
	}

	sortByDate := func(a, b DomainEventer) int {
		return a.CreatedAt().Compare(b.CreatedAt())
	}
	slices.SortFunc(events, sortByDate)

	return events
}

func (agg *Aggregate) ClearEvents() {
	agg.domainEvents = map[ID]DomainEventer{}
}

func (agg *Aggregate) AddEvent(event DomainEventer) error {
	if event.Id().IsEmpty() {
		return errors.New("cannot add event with empty ID")
	}
	if event.CreatedAt().IsZero() {
		return errors.New("cannot add event without creation date")
	}
	agg.domainEvents[event.Id()] = event
	return nil
}

func (agg *Aggregate) RemoveEvent(event DomainEventer) {
	delete(agg.domainEvents, event.Id())
}

// EventsOf returns all domain events that can be asserted to type E,
// sorted by creation date.
func EventsOf[E DomainEventer](agg *Aggregate) []E {
	var result []E
	for _, ev := range agg.domainEvents {
		if typed, ok := any(ev).(E); ok {
			result = append(result, typed)
		}
	}
	slices.SortFunc(result, func(a, b E) int {
		return a.CreatedAt().Compare(b.CreatedAt())
	})
	return result
}
