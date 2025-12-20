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
)

// DisplayName returns the user-friendly display name for the log type
func (lt LogType) DisplayName() string {
	switch lt {
	case CHARACTER_ACTION:
		return "Character Action"
	case ORACLE_QUESTION:
		return "Oracle Question"
	case MECHANICS:
		return "Mechanics"
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

func LogTypeFor(val string) LogType {
	var logType LogType

	switch val {
	case string(CHARACTER_ACTION):
		logType = CHARACTER_ACTION
	case string(ORACLE_QUESTION):
		logType = ORACLE_QUESTION
	case string(MECHANICS):
		logType = MECHANICS
	}
	return logType
}

func LogTypeLabelFor(val string) *string {
	var label string

	switch val {
	case string(CHARACTER_ACTION):
		label = "Character Action"
	case string(ORACLE_QUESTION):
		label = "Oracle Question"
	case string(MECHANICS):
		label = "Mechanics"
	}
	return &label
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
	v.Check("log_type", l.LogType == MECHANICS || l.LogType == CHARACTER_ACTION || l.LogType == ORACLE_QUESTION, "%s type is unknown", l.LogType)
	v.Check("description", len(l.Description) > 0, "cannot be blank")
	v.Check("result", len(l.Result) > 0, "cannot be blank")
	if l.LogType == CHARACTER_ACTION {
		v.Check("narrative", len(l.Narrative) > 0, "cannot be blank")
	}
	return v
}
