package character

// Service handles character business logic
type Service struct {
	repo        *Repository
	attrService *AttributeService
}

// NewService creates a new character service
func NewService(repo *Repository, attrService *AttributeService) *Service {
	return &Service{repo: repo, attrService: attrService}
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

// Duplicate makes a copy of the character including all attributes and returns the character
func (s *Service) Duplicate(id int64) (*Character, error) {

	// Duplicate the character first
	char, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	char.ID = 0
	char.Name = char.Name + " (Copy)"
	char, err = s.Save(char)
	if err != nil {
		return nil, err
	}

	// Loop over the attributes and copy them over as-is
	attrs, err := s.attrService.GetForCharacter(id)
	if err != nil {
		return nil, err
	}

	// Copy each attribute using existing insert method
	for _, attr := range attrs {
		attr.ID = 0
		attr.CharacterID = char.ID
		_, err = s.attrService.Save(attr)
		if err != nil {
			return nil, err
		}
	}

	return char, nil

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
