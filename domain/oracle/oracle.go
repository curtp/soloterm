package oracle

import (
	"regexp"
	"soloterm/shared/validation"
	"time"
)

var validNameRegex = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

const (
	MinNameLength     = 1
	MaxNameLength     = 50
	MinCategoryLength = 1
	MaxCategoryLength = 50
)

// Oracle represents an oracle in the system
type Oracle struct {
	ID                 int64     `db:"id"`
	Category           string    `db:"category"`
	Name               string    `db:"name"`
	Content            string    `db:"content"`
	CategoryPosition   int       `db:"category_position"`
	PositionInCategory int       `db:"position_in_category"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

func NewOracle(category string, name string) (*Oracle, error) {
	oracle := &Oracle{
		ID:       0,
		Category: category,
		Name:     name,
	}
	return oracle, nil
}

func (o *Oracle) Validate() *validation.Validator {
	v := validation.NewValidator()
	v.Check("name", o.Name != "", "is required")
	v.Check("name", len(o.Name) >= MinNameLength && len(o.Name) <= MaxNameLength, "must be between %d and %d characters", MinNameLength, MaxNameLength)
	v.Check("name", o.Name == "" || validNameRegex.MatchString(o.Name), "may only contain letters, numbers, _ and -")
	v.Check("category", o.Category != "", "is required")
	v.Check("category", len(o.Category) >= MinCategoryLength && len(o.Category) <= MaxCategoryLength, "must be between %d and %d characters", MinCategoryLength, MaxCategoryLength)
	return v
}

func (o *Oracle) IsNew() bool {
	return o.ID == 0
}

// CategoryInfo summarises a single category's sort position data.
// Used by the form to compute positions for new/edited oracles.
type CategoryInfo struct {
	Name                  string `db:"name"`
	CategoryPosition      int    `db:"category_position"`
	MaxPositionInCategory int    `db:"max_position_in_category"`
}
