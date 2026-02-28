package game

// Service handles game business logic
type Service struct {
	repo *Repository
}

// NewService creates a new game service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Save validates and saves a game (create or update)
func (s *Service) Save(g *Game) (*Game, error) {
	// Validate
	validator := g.Validate()
	if validator.HasErrors() {
		return nil, validator
	}

	// Save to database
	err := s.repo.Save(g)
	if err != nil {
		return nil, err
	}

	return g, nil
}

// SaveNotes updates the notes on a game
func (s *Service) SaveNotes(gameID int64, notes string) error {
	return s.repo.SaveNotes(gameID, notes)
}

// Delete removes a game by ID
func (s *Service) Delete(id int64) error {
	_, err := s.repo.Delete(id)
	return err
}

// GetByID retrieves a game by ID
func (s *Service) GetByID(id int64) (*Game, error) {
	return s.repo.GetByID(id)
}

// GetAll retrieves all games
func (s *Service) GetAll() ([]*Game, error) {
	return s.repo.GetAll()
}
