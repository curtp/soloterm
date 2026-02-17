package character

import (
	"testing"
	"time"

	testhelper "soloterm/shared/testing"
)

func TestCharacterService_Save(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	attrRepo := NewAttributeRepository(db)
	attrService := NewAttributeService(attrRepo)
	service := NewService(repo, attrService)

	t.Run("save new character", func(t *testing.T) {

		// Create a test character (with zero ID to trigger insert)
		character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")

		// Test: Save the character (insert)
		character, err := service.Save(character)
		if err != nil {
			t.Fatalf("Save() insert failed: %v", err)
		}
	})

	t.Run("update existing character", func(t *testing.T) {
		// Create initial character
		character, _ := NewCharacter("Original Name", "FlexD6", "Fighter", "Human")
		character, err := service.Save(character)
		if err != nil {
			t.Fatalf("Save() insert failed: %v", err)
		}

		// Wait a moment to ensure timestamps differ
		time.Sleep(10 * time.Millisecond)

		// Update the character
		character.Name = "Updated Name"
		character, err = service.Save(character)
		if err != nil {
			t.Fatalf("Save() update failed: %v", err)
		}

		// Make sure the updated_at column was updated
		if character.UpdatedAt.Equal(character.CreatedAt) {
			t.Errorf("updated_at was not modified on update: original=%v, updated=%v",
				character.UpdatedAt, character.CreatedAt)
		}
	})

	t.Run("validation errors prevent save", func(t *testing.T) {
		character, _ := NewCharacter("", "", "", "")
		character, err := service.Save(character)
		if err == nil {
			t.Fatalf("Save() should have failed with a validation error")
		}
	})

}

func TestCharacterService_GetByID(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	attrRepo := NewAttributeRepository(db)
	attrService := NewAttributeService(attrRepo)
	service := NewService(repo, attrService)

	// Create test character
	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")
	character, err := service.Save(character)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	characterId := character.ID

	t.Run("get existing", func(t *testing.T) {
		// Test: Get existing character by ID
		retrieved, err := repo.GetByID(characterId)
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}

		// Verify: Retrieved character matches
		if retrieved.ID != characterId {
			t.Errorf("Expected ID %q, got %q", characterId, retrieved.ID)
		}
		if retrieved.Name != "Test Character" {
			t.Errorf("Expected name 'Test Character', got %q", retrieved.Name)
		}
	})
}

func TestCharacterService_GetAll(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	attrRepo := NewAttributeRepository(db)
	attrService := NewAttributeService(attrRepo)
	service := NewService(repo, attrService)

	// Create some characters
	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")
	service.Save(character)
	character, _ = NewCharacter("Another Character", "FlexD6", "Fighter", "Human")
	service.Save(character)

	t.Run("get all", func(t *testing.T) {
		// Test: Get all characters
		retrieved, err := service.GetAll()
		if err != nil {
			t.Fatalf("GetAll() failed: %v", err)
		}

		// Verify: Retrieved character count matches
		if len(retrieved) != 2 {
			t.Errorf("Expected to retrieve 2 characters, got %d", len(retrieved))
		}
	})
}

func TestCharacterService_Delete(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	attrRepo := NewAttributeRepository(db)
	attrService := NewAttributeService(attrRepo)
	service := NewService(repo, attrService)

	// Create character to delete
	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")
	character, err := service.Save(character)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	t.Run("delete existing", func(t *testing.T) {
		// Now delete it
		err := service.Delete(character.ID)
		if err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}

		// Verify it was deleted
		_, err = service.GetByID(character.ID)
		if err == nil {
			t.Error("Expected error for deleted entry, got nil")
		}
	})

}

func TestCharacterService_Duplicate(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	attrRepo := NewAttributeRepository(db)
	attrService := NewAttributeService(attrRepo)
	service := NewService(repo, attrService)

	// Create character and attributes to duplicate
	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")
	character, err := service.Save(character)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	attr, _ := NewAttribute(character.ID, 0, 1, "Health: Max", "10")
	attr, _ = attrService.Save(attr)

	attr, _ = NewAttribute(character.ID, 0, 1, "Health: Current", "8")
	attr, _ = attrService.Save(attr)

	// Now duplicate it
	char, err := service.Duplicate(character.ID)
	if err != nil {
		t.Fatalf("Duplicate() failed: %v", err)
	}

	// Verify it was duplicated - should be 2 characters in the db and the new char should have attributes

	chars, err := service.GetAll()
	if err != nil {
		t.Error("Issue retrieving all chars")
	}
	if len(chars) != 2 {
		t.Errorf("Expected 2 characters, got %d", len(chars))
	}

	if char.Name != "Test Character (Copy)" {
		t.Error("Name did not change")
	}

	attrs, err := attrService.GetForCharacter(char.ID)
	if err != nil {
		t.Error("Issue retrieving attributes")
	}
	if len(attrs) != 2 {
		t.Errorf("Expected 2 attributes, got %d", len(attrs))
	}

	// Original is still in tact
	attrs, err = attrService.GetForCharacter(character.ID)
	if err != nil {
		t.Error("Issue retrieving attributes")
	}
	if len(attrs) != 2 {
		t.Errorf("Expected 2 attributes, got %d", len(attrs))
	}

}
