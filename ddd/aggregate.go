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
		return errors.New("não é possível adicionar evento com ID vazio")
	}
	if event.CreatedAt().IsZero() {
		return errors.New("não é possível adicionar evento sem data de criação")
	}
	agg.domainEvents[event.Id()] = event
	return nil
}

func (agg *Aggregate) RemoveEvent(event DomainEventer) {
	delete(agg.domainEvents, event.Id())
}
