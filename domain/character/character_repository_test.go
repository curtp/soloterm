package character

import (
	"testing"
	"time"

	testhelper "soloterm/shared/testing"
)

func TestCharacterRepository_Save(t *testing.T) {
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

func TestCharacterRepository_GetByID(t *testing.T) {
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

func TestCharacterRepository_GetAll(t *testing.T) {
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

func TestCharacterRepository_Delete(t *testing.T) {
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
