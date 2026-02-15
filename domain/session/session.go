package session

import (
	"soloterm/shared/validation"
	"time"
)

type Session struct {
	ID        int64     `db:"id"`
	GameID    int64     `db:"game_id"`
	Name      string    `db:"name"`
	Content   string    `db:"content"`
	GameName  string    `db:"game_name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewSession(gameID int64) (*Session, error) {
	session := &Session{
		ID:     0,
		GameID: gameID,
	}

	return session, nil
}

func (s *Session) Validate() *validation.Validator {
	v := validation.NewValidator()
	v.Check("name", len(s.Name) > 0, "cannot be blank")
	return v
}

func (s *Session) IsNew() bool {
	return s.ID == 0
}
