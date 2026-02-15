// Package game provides core game domain logic, validation, and persistence.
// It defines the Game entity and its associated repository for database operations.
package character

import (
	"soloterm/shared/validation"
	"time"
)

const (
	MinNameLength           = 1
	MaxNameLength           = 50
	MinSystemLength         = 1
	MaxSystemLength         = 50
	MinAttributeNameLength  = 1
	MaxAttributeNameLength  = 50
	MinAttributeValueLength = 0
	MaxAttributeValueLength = 50
	MinRoleLength           = 1
	MaxRoleLength           = 50
	MinSpeciesLength        = 1
	MaxSpeciesLength        = 50
)

// Character represents a character in the system
type Character struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	System    string    `db:"system"`
	Role      string    `db:"role"`
	Species   string    `db:"species"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

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

func NewCharacter(name string, system string, role string, species string) (*Character, error) {
	character := &Character{
		ID:      0,
		Name:    name,
		System:  system,
		Role:    role,
		Species: species,
	}

	return character, nil
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

func (c *Character) Validate() *validation.Validator {
	v := validation.NewValidator()
	v.Check("name", c.Name != "", "is required")
	v.Check("name", len(c.Name) >= MinNameLength && len(c.Name) <= MaxNameLength, "must be between %d and %d characters", MinNameLength, MaxNameLength)
	v.Check("system", c.System != "", "is required")
	v.Check("system", len(c.System) >= MinSystemLength && len(c.System) <= MaxSystemLength, "must be between %d and %d characters", MinSystemLength, MaxSystemLength)
	v.Check("role", c.Role != "", "is required")
	v.Check("role", len(c.Role) >= MinRoleLength && len(c.Role) <= MaxRoleLength, "must be between %d and %d characters", MinRoleLength, MaxRoleLength)
	v.Check("species", c.Species != "", "is required")
	v.Check("species", len(c.Species) >= MinSpeciesLength && len(c.Species) <= MaxSpeciesLength, "must be between %d and %d characters", MinSpeciesLength, MaxSpeciesLength)

	return v
}

func (a *Attribute) Validate() *validation.Validator {
	v := validation.NewValidator()
	v.Check("name", a.Name != "", "is required")
	v.Check("name", len(a.Name) >= MinNameLength && len(a.Name) <= MaxNameLength, "must be between %d and %d characters", MinNameLength, MaxNameLength)
	v.Check("value", len(a.Value) >= MinAttributeValueLength && len(a.Value) <= MaxAttributeValueLength, "must be between %d and %d characters", MinAttributeValueLength, MaxAttributeValueLength)
	v.Check("character_id", a.CharacterID != 0, "is required")
	return v
}
