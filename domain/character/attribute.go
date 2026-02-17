package character

import (
	"soloterm/shared/validation"
	"time"
)

const (
	MinAttributeNameLength  = 1
	MaxAttributeNameLength  = 50
	MinAttributeValueLength = 0
	MaxAttributeValueLength = 50
)

type Attribute struct {
	ID                  int64     `db:"id"`
	CharacterID         int64     `db:"character_id"`
	Group               int       `db:"attribute_group"`
	PositionInGroup     int       `db:"position_in_group"`
	Name                string    `db:"name"`
	Value               string    `db:"value"`
	CreatedAt           time.Time `db:"created_at"`
	UpdatedAt           time.Time `db:"updated_at"`
	GroupCount          int       `db:"group_count"`
	GroupCountAfterZero int       `db:"group_count_after_zero"`
}

func NewAttribute(character_id int64, group int, position_in_group int, name string, value string) (*Attribute, error) {
	attribute := &Attribute{
		ID:              0,
		CharacterID:     character_id,
		Group:           group,
		PositionInGroup: position_in_group,
		Name:            name,
		Value:           value,
	}
	return attribute, nil
}

func (a *Attribute) Validate() *validation.Validator {
	v := validation.NewValidator()
	v.Check("name", a.Name != "", "is required")
	v.Check("name", len(a.Name) >= MinNameLength && len(a.Name) <= MaxNameLength, "must be between %d and %d characters", MinNameLength, MaxNameLength)
	v.Check("value", len(a.Value) >= MinAttributeValueLength && len(a.Value) <= MaxAttributeValueLength, "must be between %d and %d characters", MinAttributeValueLength, MaxAttributeValueLength)
	v.Check("character_id", a.CharacterID != 0, "is required")
	return v
}
