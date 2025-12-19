package ui

import (
	"soloterm/domain/log"

	"github.com/jmoiron/sqlx"
)

// LogHandler contains handlers for log-related operations
type LogHandler struct {
	db      *sqlx.DB
	logRepo *log.Repository
}

// NewLogHandler creates a new LogHandler instance
func NewLogHandler(db *sqlx.DB) *LogHandler {
	return &LogHandler{
		db:      db,
		logRepo: log.NewRepository(db),
	}
}

// SaveLog saves the log and returns any validation errors
func (h *LogHandler) SaveLog(id *int64, gameID int64, logType log.LogType, description string, result string, narrative string) (*log.Log, error) {
	l := &log.Log{
		GameID:      gameID,
		LogType:     logType,
		Description: description,
		Result:      result,
		Narrative:   narrative,
	}

	// If an ID was provided, then set it on the log. That means it's an update
	if id != nil {
		l.ID = *id
	}

	// Validate the data
	validator := l.Validate()
	if validator.HasErrors() {
		// Return validation error (we'll handle this specially)
		return nil, &ValidationError{Validator: validator}
	}

	// Save to database
	err := h.logRepo.Save(l)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// Delete deletes a log entry by ID
func (h *LogHandler) Delete(id int64) error {
	_, err := h.logRepo.Delete(id)
	return err
}

func (h *LogHandler) GetByID(id int64) (*log.Log, error) {
	return h.logRepo.GetByID(id)
}

func (h *LogHandler) GetAllForGame(gameID int64) ([]*log.Log, error) {
	return h.logRepo.GetAllForGame(gameID)
}

func (h *LogHandler) GetLogsForSession(gameID int64, sessionDate string) ([]*log.Log, error) {
	return h.logRepo.GetLogsForSession(gameID, sessionDate)
}

func (h *LogHandler) GetSessionsForGame(gameID int64) ([]*log.Session, error) {
	return h.logRepo.GetSessionsForGame(gameID)
}
