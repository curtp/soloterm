package character

// AttributeService handles attribute business logic
type AttributeService struct {
	repo *AttributeRepository
}

// NewAttributeService creates a new attribute service
func NewAttributeService(repo *AttributeRepository) *AttributeService {
	return &AttributeService{repo: repo}
}

// Save validates and saves an attribute (create or update)
func (s *AttributeService) Save(a *Attribute) (*Attribute, error) {
	// Validate
	validator := a.Validate()
	if validator.HasErrors() {
		return nil, validator
	}

	// Save to database
	err := s.repo.Save(a)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// Delete removes an attribute by ID
func (s *AttributeService) Delete(id int64) error {
	_, err := s.repo.Delete(id)
	return err
}

// GetByID retrieves an attribute by ID
func (s *AttributeService) GetByID(id int64) (*Attribute, error) {
	return s.repo.GetByID(id)
}

// GetForCharacter retrieves all attributes for a character
func (s *AttributeService) GetForCharacter(character_id int64) ([]*Attribute, error) {
	return s.repo.GetForCharacter(character_id)
}
