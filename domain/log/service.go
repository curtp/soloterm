package log

// Service handles log business logic
type Service struct {
	repo *Repository
}

// NewService creates a new log service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Save validates and saves a log entry (create or update)
func (s *Service) Save(l *Log) (*Log, error) {
	// Validate
	validator := l.Validate()
	if validator.HasErrors() {
		return nil, validator
	}

	// Save to database
	err := s.repo.Save(l)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// Delete removes a log entry by ID
func (s *Service) Delete(id int64) error {
	_, err := s.repo.Delete(id)
	return err
}

// GetByID retrieves a log entry by ID
func (s *Service) GetByID(id int64) (*Log, error) {
	return s.repo.GetByID(id)
}

// GetAllForGame retrieves all logs for a game
func (s *Service) GetAllForGame(gameID int64) ([]*Log, error) {
	return s.repo.GetAllForGame(gameID)
}

// GetSessionsForGame retrieves session groupings for a game
func (s *Service) GetSessionsForGame(gameID int64) ([]*Session, error) {
	return s.repo.GetSessionsForGame(gameID)
}

// GetLogsForSession retrieves logs for a specific session date
func (s *Service) GetLogsForSession(gameID int64, sessionDate string) ([]*Log, error) {
	return s.repo.GetLogsForSession(gameID, sessionDate)
}
