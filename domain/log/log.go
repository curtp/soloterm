// Package log provides log entry domain logic for tracking game events.
// It supports multiple log types (character actions, oracle questions, mechanics)
// and manages log persistence and session grouping.
package log

import (
	"soloterm/shared/validation"
	"time"
)

type LogType string

const (
	CHARACTER_ACTION LogType = "character_action"
	ORACLE_QUESTION  LogType = "oracle_question"
	MECHANICS        LogType = "mechanics"
	STORY            LogType = "story"
)

// LogTypeDisplayName returns the user-friendly display name for the log type
func (lt LogType) LogTypeDisplayName() string {
	switch lt {
	case CHARACTER_ACTION:
		return "Character Action"
	case ORACLE_QUESTION:
		return "Oracle Question"
	case MECHANICS:
		return "Mechanics"
	case STORY:
		return "Story"
	default:
		return string(lt)
	}
}

// Game represents a game in the system
type Log struct {
	ID          int64     `db:"id"`
	GameID      int64     `db:"game_id"`
	LogType     LogType   `db:"log_type"`
	Description string    `db:"description"`
	Narrative   string    `db:"narrative"`
	Result      string    `db:"result"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// Returns an array of LogTypes
func LogTypes() []LogType {
	return []LogType{
		CHARACTER_ACTION,
		MECHANICS,
		ORACLE_QUESTION,
		STORY,
	}
}

func NewLog(gameID int64) (*Log, error) {
	log := &Log{
		ID:     0,
		GameID: gameID,
	}

	return log, nil
}

func (l *Log) Validate() *validation.Validator {
	v := validation.NewValidator()
	v.Check("log_type", l.LogType == MECHANICS || l.LogType == CHARACTER_ACTION || l.LogType == ORACLE_QUESTION || l.LogType == STORY, "%s type is unknown", l.LogType)

	// Require description and result for everything other than story logs
	if l.LogType != STORY {
		v.Check("description", len(l.Description) > 0, "cannot be blank")
		v.Check("result", len(l.Result) > 0, "cannot be blank")
	}

	// Narrative is only required for character actions or story logs
	if l.LogType == CHARACTER_ACTION || l.LogType == STORY {
		v.Check("narrative", len(l.Narrative) > 0, "cannot be blank")
	}
	return v
}
