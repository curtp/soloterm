package snippet

import (
	"soloterm/shared/validation"
	"time"
)

const (
	MinNameLength    = 1
	MaxNameLength    = 50
	MinContentLength = 1
	MaxContentLength = 1000
)

// Snippet represents a snippet in the system
type Snippet struct {
	ID        int64     `db:"id"`
	GameID    *int64    `db:"game_id"`
	Name      string    `db:"name"`
	Content   string    `db:"content"`
	Position  int       `db:"position"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewSnippet(name string, content string, gameID *int64) (*Snippet, error) {
	snippet := &Snippet{
		ID:       0,
		Name:     name,
		Content:  content,
		GameID:   gameID,
		Position: 0,
	}
	return snippet, nil
}

func (sr *Snippet) Validate() *validation.Validator {
	v := validation.NewValidator()
	v.Check("name", sr.Name != "", "is required")
	v.Check("name", len(sr.Name) >= MinNameLength && len(sr.Name) <= MaxNameLength, "must be between %d and %d characters", MinNameLength, MaxNameLength)
	v.Check("content", sr.Content != "", "is required")
	v.Check("content", len(sr.Content) >= MinContentLength && len(sr.Content) <= MaxContentLength, "must be between %d and %d characters", MinContentLength, MaxContentLength)
	return v
}

func (sr *Snippet) IsNew() bool {
	return sr.ID == 0
}
