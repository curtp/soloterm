package ui

import (
	"soloterm/domain/session"
)

// SessionViewHelper coordinates session-related UI operations
type SessionViewHelper struct {
	sessionService *session.Service
}

// Create a new session helper which uses the session and log services provided
func NewSessionViewHelper(sessionService *session.Service) *SessionViewHelper {
	return &SessionViewHelper{
		sessionService: sessionService,
	}
}

// Loads a session for the session ID provided
func (sh *SessionViewHelper) LoadSession(sessionID int64) (*session.Session, error) {
	session, err := sh.sessionService.GetByID(sessionID)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// Loads all sessions for th game ID provided. This excludes the content
func (sh *SessionViewHelper) LoadSessions(gameID int64) ([]*session.Session, error) {
	sessions, err := sh.sessionService.GetAllForGame(gameID)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (sh *SessionViewHelper) Delete(sessionID int64) error {
	return sh.sessionService.Delete(sessionID)
}
