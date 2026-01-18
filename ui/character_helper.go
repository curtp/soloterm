package ui

import (
	"soloterm/domain/character"
)

// CharacterViewHelper coordinates character-related UI operations
type CharacterViewHelper struct {
	characterService *character.Service
}

// NewCharacterHandler creates a new character handler
func NewCharacterViewHelper(characterService *character.Service) *CharacterViewHelper {
	return &CharacterViewHelper{
		characterService: characterService,
	}
}

// LoadCharacters loads all characters
func (cv *CharacterViewHelper) LoadCharacters() (map[string][]*character.Character, error) {
	// Initialize the map
	charsBySystem := make(map[string][]*character.Character)

	// Load characters from database
	chars, err := cv.characterService.GetAll()
	if err != nil {
		return nil, err
	}

	// If there aren't any, return an empty map
	if len(chars) == 0 {
		return charsBySystem, nil
	}

	// Group them by system
	for _, c := range chars {
		charsBySystem[c.System] = append(charsBySystem[c.System], c)
	}

	return charsBySystem, nil
}
