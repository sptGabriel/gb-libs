package ddd

import (
	"encoding/json"

	"github.com/google/uuid"
)

type ID string

func (id ID) String() string {
	return string(id)
}

func (id ID) IsEmpty() bool {
	return id == ``
}

func (id ID) ToUUID() (uuid.UUID, error) {
	return uuid.Parse(id.String())
}

func (id ID) MarshalJSON() ([]byte, error) {
	uid, err := id.ToUUID()
	if err != nil {
		return nil, err
	}

	return json.Marshal(uid)
}

func (id *ID) UnmarshalJSON(data []byte) error {
	b := uuid.UUID{}
	err := json.Unmarshal(data, &b)
	if err != nil {
		return err
	}

	*id = NewFromUUID(b)

	return nil
}

func NewFromUUID(uid uuid.UUID) ID {
	return ID(uid.String())
}

func New() ID {
	return ID(uuid.New().String())
}
