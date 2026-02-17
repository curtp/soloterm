package character

import (
	"testing"
	"time"

	testhelper "soloterm/shared/testing"
)

func TestAttributeRepository_Save(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	attrRepo := NewAttributeRepository(db)

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

		err = attrRepo.Save(attribute)
		if err != nil {
			t.Fatalf("Save() insert failed: %v", err)
		}

		// Verify the DB managed fields are set
		if attribute.CreatedAt.IsZero() || attribute.UpdatedAt.IsZero() || attribute.ID == 0 {
			t.Fatalf("Attribute was not saved with DB managed fields: %+v", attribute)
		}
	})

	t.Run("update existing attribute", func(t *testing.T) {
		// Create initial attribute
		attribute, _ := NewAttribute(character.ID, 0, 1, "Health", "10")
		err := attrRepo.Save(attribute)
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
		err = attrRepo.Save(attribute)
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

func TestAttributeRepository_GetForCharacter(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	attrRepo := NewAttributeRepository(db)

	// Create a character and some attributes
	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")
	repo.Save(character)

	attr, _ := NewAttribute(character.ID, 0, 1, "Health: Max", "10")
	attrRepo.Save(attr)
	attr, _ = NewAttribute(character.ID, 0, 1, "Health: Current", "8")
	attrRepo.Save(attr)
	attr, _ = NewAttribute(character.ID, 1, 1, "Skill: Swordmaster", "2")
	attrRepo.Save(attr)
	attr, _ = NewAttribute(character.ID, 1, 1, "Skill: Intimidating", "1")
	attrRepo.Save(attr)

	t.Run("get all attributes", func(t *testing.T) {
		retrieved, err := attrRepo.GetForCharacter(character.ID)
		if err != nil {
			t.Fatalf("GetForCharacter() failed: %v", err)
		}

		// Verify: Retrieved attribute count matches
		if len(retrieved) != 4 {
			t.Errorf("Expected to retrieve 4 attributes, got %d", len(retrieved))
		}
	})
}

func TestAttributeRepository_GetByID(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	attrRepo := NewAttributeRepository(db)

	// Create test character
	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")
	err := repo.Save(character)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}
	attr, _ := NewAttribute(character.ID, 0, 1, "Health: Max", "10")
	attrRepo.Save(attr)

	t.Run("get existing", func(t *testing.T) {
		retrieved, err := attrRepo.GetByID(attr.ID)
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}

		// Verify: Retrieved attribute matches
		if retrieved.ID != attr.ID {
			t.Errorf("Expected ID %q, got %q", attr.ID, retrieved.ID)
		}
		if retrieved.Name != "Health: Max" {
			t.Errorf("Expected name 'Health: Max', got %q", retrieved.Name)
		}
	})

	t.Run("get invalid id", func(t *testing.T) {
		// Test: Get non-existent attribute
		_, err = attrRepo.GetByID(999999)
		if err == nil {
			t.Error("Expected error for non-existent ID, got nil")
		}

		// Test: Zero ID should return error
		_, err = attrRepo.GetByID(0)
		if err == nil {
			t.Error("Expected error for zero ID, got nil")
		}
	})
}

func TestAttributeRepository_Delete(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	attrRepo := NewAttributeRepository(db)

	// Create character and an attribute to delete
	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")
	err := repo.Save(character)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	deleteMe, _ := NewAttribute(character.ID, 0, 1, "Health: Max", "10")
	attrRepo.Save(deleteMe)
	attr, _ := NewAttribute(character.ID, 0, 1, "Skill: Insta-kill", "1")
	attrRepo.Save(attr)

	t.Run("delete existing", func(t *testing.T) {
		// Now delete it
		count, err := attrRepo.Delete(deleteMe.ID)
		if err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}

		if count != 1 {
			t.Errorf("Expected 1 row deleted, got %d", count)
		}

		// Verify it was deleted
		_, err = attrRepo.GetByID(deleteMe.ID)
		if err == nil {
			t.Error("Expected error for deleted entry, got nil")
		}

		// Total attributes for characters should be 1
		attrs, err := attrRepo.GetForCharacter(character.ID)
		if err != nil || len(attrs) != 1 {
			t.Errorf("Expected there to be 1 remaining attribute, but there were %d. Error: %v", len(attrs), err)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		// This should return an error
		count, err := attrRepo.Delete(9999)
		if err == nil {
			t.Error("Expected error for non-existent key, got nil")
		}

		if count != 0 {
			t.Errorf("Expected 0 rows deleted, got %d", count)
		}
	})

	t.Run("zero fails", func(t *testing.T) {
		// This should return an error
		count, err := attrRepo.Delete(0)
		if err == nil {
			t.Error("Expected error for non-existent key, got nil")
		}

		if count != 0 {
			t.Errorf("Expected 0 rows deleted, got %d", count)
		}
	})

}
