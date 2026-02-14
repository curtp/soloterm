package session

// Service handles session business sessionic
type Service struct {
	repo *Repository
}

// NewService creates a new session service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Save validates and saves a session entry (create or update)
func (s *Service) Save(l *Session) (*Session, error) {
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

// Delete removes a session entry by ID
func (s *Service) Delete(id int64) error {
	_, err := s.repo.Delete(id)
	return err
}

// GetByID retrieves a session entry by ID
func (s *Service) GetByID(id int64) (*Session, error) {
	return s.repo.GetByID(id)
}

// GetAllForGame retrieves all sessions for a game
func (s *Service) GetAllForGame(gameID int64) ([]*Session, error) {
	return s.repo.GetAllForGame(gameID)
}
