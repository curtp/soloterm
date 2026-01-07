package character

// Service handles character business logic
type Service struct {
	repo *Repository
}

// NewService creates a new character service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Save validates and saves a character (create or update)
func (s *Service) Save(c *Character) (*Character, error) {
	// Validate
	validator := c.Validate()
	if validator.HasErrors() {
		return nil, validator
	}

	// Save to database
	err := s.repo.Save(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Delete removes a character by ID
func (s *Service) Delete(id int64) error {
	_, err := s.repo.Delete(id)
	return err
}

// GetByID retrieves a character by ID
func (s *Service) GetByID(id int64) (*Character, error) {
	return s.repo.GetByID(id)
}

// GetAll retrieves all characters
func (s *Service) GetAll() ([]*Character, error) {
	return s.repo.GetAll()
}

// SaveAttribute validates and saves an attribute (create or update)
func (s *Service) SaveAttribute(a *Attribute) (*Attribute, error) {
	// Validate
	validator := a.Validate()
	if validator.HasErrors() {
		return nil, validator
	}

	// Save to database
	err := s.repo.SaveAttribute(a)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// DeleteAttribute removes an attribute by ID
func (s *Service) DeleteAttribute(id int64) error {
	_, err := s.repo.DeleteAttribute(id)
	return err
}

// GetAttributeByID retrieves an attribute by ID
func (s *Service) GetAttributeByID(id int64) (*Attribute, error) {
	return s.repo.GetAttributeByID(id)
}

// GetAll retrieves all character attributes
func (s *Service) GetAttributesForCharacter(character_id int64) ([]*Attribute, error) {
	return s.repo.GetAttributesForCharacter(character_id)
}
