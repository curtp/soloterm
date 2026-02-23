package character

import "fmt"

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

// Reorder moves an attribute up or down in the display order.
// direction: -1 = up, +1 = down.
// groupMove: false = swap the selected item with its adjacent neighbor;
//            true  = move the selected item's entire group past the adjacent group.
// Returns the moved attribute's ID on success, 0 on a boundary no-op, or an error.
func (s *AttributeService) Reorder(characterID, attributeID int64, direction int, groupMove bool) (int64, error) {
	attrs, err := s.repo.GetForCharacter(characterID)
	if err != nil {
		return 0, err
	}

	// Find the index of the target attribute in the sorted list.
	idx := -1
	for i, a := range attrs {
		if a.ID == attributeID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return 0, fmt.Errorf("attribute %d not found for character %d", attributeID, characterID)
	}

	curr := attrs[idx]

	if groupMove {
		// Walk in direction until we find an item in a different group.
		neighborGroup := -1
		for i := idx + direction; i >= 0 && i < len(attrs); i += direction {
			if attrs[i].Group != curr.Group {
				neighborGroup = attrs[i].Group
				break
			}
		}
		if neighborGroup == -1 {
			return 0, nil // Already at boundary.
		}
		if err = s.repo.SwapGroups(characterID, curr.Group, neighborGroup); err != nil {
			return 0, err
		}
	} else {
		neighborIdx := idx + direction
		if neighborIdx < 0 || neighborIdx >= len(attrs) {
			return 0, nil // Already at boundary.
		}
		neighbor := attrs[neighborIdx]
		if err = s.repo.UpdatePosition(curr.ID, neighbor.Group, neighbor.PositionInGroup); err != nil {
			return 0, err
		}
		if err = s.repo.UpdatePosition(neighbor.ID, curr.Group, curr.PositionInGroup); err != nil {
			return 0, err
		}
	}

	return curr.ID, nil
}
