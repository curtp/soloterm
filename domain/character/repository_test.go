package character

import (
	"testing"
	"time"

	testhelper "soloterm/shared/testing"
)

func TestRepository_Save(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)

	t.Run("save new character", func(t *testing.T) {

		// Create a test character (with zero ID to trigger insert)
		character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")

		// Test: Save the character (insert)
		err := repo.Save(character)
		if err != nil {
			t.Fatalf("Save() insert failed: %v", err)
		}

		// Verify: Database managed fields are set after insert
		if character.ID == 0 {
			t.Errorf("Expected ID to be set after insert, got %d", character.ID)
		}
		if character.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set after insert")
		}
		if character.UpdatedAt.IsZero() {
			t.Error("Expected UpdatedAt to be set after insert")
		}
	})

	t.Run("update existing character", func(t *testing.T) {
		// Create initial character
		character, _ := NewCharacter("Original Name", "FlexD6", "Fighter", "Human")
		err := repo.Save(character)
		if err != nil {
			t.Fatalf("Save() insert failed: %v", err)
		}

		// Verify the DB managed fields are set
		if character.CreatedAt.IsZero() || character.UpdatedAt.IsZero() || character.ID == 0 {
			t.Fatalf("Character was not saved with DB managed fields: %+v", character)
		}

		firstCreatedAt := character.CreatedAt
		firstUpdatedAt := character.UpdatedAt

		// Wait a moment to ensure timestamps differ
		time.Sleep(10 * time.Millisecond)

		// Update the character
		character.Name = "Updated Name"
		err = repo.Save(character)
		if err != nil {
			t.Fatalf("Save() update failed: %v", err)
		}

		// Make sure the updated_at column was updated
		if character.UpdatedAt.Equal(firstUpdatedAt) {
			t.Errorf("updated_at was not modified on update: original=%v, updated=%v",
				firstUpdatedAt, character.UpdatedAt)
		}

		// Verify created_at was not updated
		if !character.CreatedAt.Equal(firstCreatedAt) {
			t.Errorf("created_at was modified on update: original=%v, updated=%v",
				firstCreatedAt, character.CreatedAt)
		}

	})
}

func TestRepository_GetByID(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)

	// Create test character
	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")
	err := repo.Save(character)
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

	t.Run("get invalid id", func(t *testing.T) {
		// Test: Get non-existent character
		_, err = repo.GetByID(999999)
		if err == nil {
			t.Error("Expected error for non-existent ID, got nil")
		}

		// Test: Zero ID should return error
		_, err = repo.GetByID(0)
		if err == nil {
			t.Error("Expected error for zero ID, got nil")
		}
	})
}

func TestRepository_GetAll(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)

	// Create some characters
	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")
	repo.Save(character)
	character, _ = NewCharacter("Another Character", "FlexD6", "Fighter", "Human")
	repo.Save(character)

	t.Run("get all", func(t *testing.T) {
		// Test: Get all characters
		retrieved, err := repo.GetAll()
		if err != nil {
			t.Fatalf("GetAll() failed: %v", err)
		}

		// Verify: Retrieved character count matches
		if len(retrieved) != 2 {
			t.Errorf("Expected to retrieve 2 characters, got %d", len(retrieved))
		}
	})
}

func TestRepository_Delete(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)

	// Create character to delete
	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")
	err := repo.Save(character)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	t.Run("delete existing", func(t *testing.T) {
		// Now delete it
		count, err := repo.Delete(character.ID)
		if err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}

		if count != 1 {
			t.Errorf("Expected 1 row deleted, got %d", count)
		}

		// Verify it was deleted
		_, err = repo.GetByID(character.ID)
		if err == nil {
			t.Error("Expected error for deleted entry, got nil")
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		// This should return an error
		count, err := repo.Delete(9999)
		if err == nil {
			t.Error("Expected error for non-existent key, got nil")
		}

		if count != 0 {
			t.Errorf("Expected 0 rows deleted, got %d", count)
		}
	})

	t.Run("zero fails", func(t *testing.T) {
		// This should return an error
		count, err := repo.Delete(0)
		if err == nil {
			t.Error("Expected error for non-existent key, got nil")
		}

		if count != 0 {
			t.Errorf("Expected 0 rows deleted, got %d", count)
		}
	})

}

func TestRepository_SaveAttribute(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	// Create a test character (with zero ID to trigger insert)
	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")

	// Test: Save the character (insert)
	err := repo.Save(character)
	if err != nil {
		t.Fatalf("Save() insert failed: %v", err)
	}

	t.Run("save new attribute", func(t *testing.T) {

		// Create an attribute for the character
		attribute, err := NewAttribute(character.ID, 0, 1, "Health", "10")
		if err != nil {
			t.Fatalf("NewAttribute() failed: %v", err)
		}

		err = repo.SaveAttribute(attribute)
		if err != nil {
			t.Fatalf("SaveAttribute() insert failed: %v", err)
		}

		// Verify the DB managed fields are set
		if attribute.CreatedAt.IsZero() || attribute.UpdatedAt.IsZero() || attribute.ID == 0 {
			t.Fatalf("Attribute was not saved with DB managed fields: %+v", attribute)
		}
	})

	t.Run("update existing attribute", func(t *testing.T) {
		// Create initial attribute
		attribute, _ := NewAttribute(character.ID, 0, 1, "Health", "10")
		err := repo.SaveAttribute(attribute)
		if err != nil {
			t.Fatalf("Save() insert failed: %v", err)
		}

		firstCreatedAt := attribute.CreatedAt
		firstUpdatedAt := attribute.UpdatedAt

		// Wait a moment to ensure timestamps differ
		time.Sleep(10 * time.Millisecond)

		// Update the attribute
		attribute.Name = "Max Health"
		attribute.Value = "15"
		err = repo.SaveAttribute(attribute)
		if err != nil {
			t.Fatalf("Save() update failed: %v", err)
		}

		// Make sure the updated_at column was updated
		if attribute.UpdatedAt.Equal(firstUpdatedAt) {
			t.Errorf("updated_at was not modified on update: original=%v, updated=%v",
				firstUpdatedAt, attribute.UpdatedAt)
		}

		// Verify created_at was not updated
		if !attribute.CreatedAt.Equal(firstCreatedAt) {
			t.Errorf("created_at was modified on update: original=%v, updated=%v",
				firstCreatedAt, attribute.CreatedAt)
		}
	})
}

func TestRepository_GetAttributesForCharacter(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)

	// Create a character and some attributes
	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")
	repo.Save(character)

	attr, _ := NewAttribute(character.ID, 0, 1, "Health: Max", "10")
	repo.SaveAttribute(attr)
	attr, _ = NewAttribute(character.ID, 0, 1, "Health: Current", "8")
	repo.SaveAttribute(attr)
	attr, _ = NewAttribute(character.ID, 1, 1, "Skill: Swordmaster", "2")
	repo.SaveAttribute(attr)
	attr, _ = NewAttribute(character.ID, 1, 1, "Skill: Intimidating", "1")
	repo.SaveAttribute(attr)

	t.Run("get all attributes", func(t *testing.T) {
		retrieved, err := repo.GetAttributesForCharacter(character.ID)
		if err != nil {
			t.Fatalf("GetAttributesForCharacter() failed: %v", err)
		}

		// Verify: Retrieved character count matches
		if len(retrieved) != 4 {
			t.Errorf("Expected to retrieve 4 characters, got %d", len(retrieved))
		}
	})
}

func TestRepository_GetAttributeByID(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)

	// Create test character
	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")
	err := repo.Save(character)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}
	attr, _ := NewAttribute(character.ID, 0, 1, "Health: Max", "10")
	repo.SaveAttribute(attr)

	t.Run("get existing", func(t *testing.T) {
		retrieved, err := repo.GetAttributeByID(attr.ID)
		if err != nil {
			t.Fatalf("GetAttributeByID() failed: %v", err)
		}

		// Verify: Retrieved character matches
		if retrieved.ID != attr.ID {
			t.Errorf("Expected ID %q, got %q", attr.ID, retrieved.ID)
		}
		if retrieved.Name != "Health: Max" {
			t.Errorf("Expected name 'Health: Max', got %q", retrieved.Name)
		}
	})

	t.Run("get invalid id", func(t *testing.T) {
		// Test: Get non-existent attribute
		_, err = repo.GetAttributeByID(999999)
		if err == nil {
			t.Error("Expected error for non-existent ID, got nil")
		}

		// Test: Zero ID should return error
		_, err = repo.GetAttributeByID(0)
		if err == nil {
			t.Error("Expected error for zero ID, got nil")
		}
	})
}

func TestRepository_DeleteAttribute(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)

	// Create character and an attribute to delete
	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")
	err := repo.Save(character)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	deleteMe, _ := NewAttribute(character.ID, 0, 1, "Health: Max", "10")
	repo.SaveAttribute(deleteMe)
	attr, _ := NewAttribute(character.ID, 0, 1, "Skill: Insta-kill", "1")
	repo.SaveAttribute(attr)

	t.Run("delete existing", func(t *testing.T) {
		// Now delete it
		count, err := repo.DeleteAttribute(deleteMe.ID)
		if err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}

		if count != 1 {
			t.Errorf("Expected 1 row deleted, got %d", count)
		}

		// Verify it was deleted
		_, err = repo.GetAttributeByID(deleteMe.ID)
		if err == nil {
			t.Error("Expected error for deleted entry, got nil")
		}

		// Total attributes for characters should be 1
		attrs, err := repo.GetAttributesForCharacter(character.ID)
		if err != nil || len(attrs) != 1 {
			t.Errorf("Expected there to be 1 remaining attribute, but there were %d. Error: %v", len(attrs), err)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		// This should return an error
		count, err := repo.DeleteAttribute(9999)
		if err == nil {
			t.Error("Expected error for non-existent key, got nil")
		}

		if count != 0 {
			t.Errorf("Expected 0 rows deleted, got %d", count)
		}
	})

	t.Run("zero fails", func(t *testing.T) {
		// This should return an error
		count, err := repo.DeleteAttribute(0)
		if err == nil {
			t.Error("Expected error for non-existent key, got nil")
		}

		if count != 0 {
			t.Errorf("Expected 0 rows deleted, got %d", count)
		}
	})

}
