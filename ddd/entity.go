package ddd

type Entity struct {
	id ID
}

func (e Entity) ID() ID {
	return e.id
}

func NewEntity(id ID) Entity {
	return Entity{id}
}
