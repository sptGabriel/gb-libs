package ddd

import "time"

type (
	DomainAction    string
	DomainOperation string
	DomainEventName string

	DomainDataModification struct {
		NewValue      string
		FieldName     string
		OriginalValue string
	}
)

type DomainEventer interface {
	Id() ID
	EntityId() ID
	Name() DomainEventName
	Operation() DomainOperation
	Modifications() []DomainDataModification
	CreatedAt() time.Time
}
