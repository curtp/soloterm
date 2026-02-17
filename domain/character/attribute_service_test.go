package character

import (
	"testing"
	"time"

	testhelper "soloterm/shared/testing"
)

func TestAttributeService_Save(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	attrRepo := NewAttributeRepository(db)
	attrService := NewAttributeService(attrRepo)

	// Create a character to save the attribute to
	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")
	repo.Save(character)

	t.Run("save new attribute", func(t *testing.T) {

		attr, _ := NewAttribute(character.ID, 0, 1, "Health: Max", "10")
		attr, err := attrService.Save(attr)
		if err != nil {
			t.Fatalf("Save() insert failed: %v", err)
		}
	})

	t.Run("update existing attribute", func(t *testing.T) {
		// Create an attribute to update
		attr, _ := NewAttribute(character.ID, 0, 1, "Health: Max", "10")
		attr, err := attrService.Save(attr)

		// Wait a moment to ensure timestamps differ
		time.Sleep(10 * time.Millisecond)

		// Update the attribute
		attr.Value = "8"
		attr, err = attrService.Save(attr)
		if err != nil {
			t.Fatalf("Save() update failed: %v", err)
		}

		// Make sure the updated_at column was updated
		if attr.UpdatedAt.Equal(attr.CreatedAt) {
			t.Errorf("updated_at was not modified on update: original=%v, updated=%v",
				attr.UpdatedAt, attr.CreatedAt)
		}
	})

	t.Run("validation errors prevent save", func(t *testing.T) {
		attr, _ := NewAttribute(character.ID, 0, 1, "", "")
		attr, err := attrService.Save(attr)
		if err == nil {
			t.Fatalf("Save() should have failed with a validation error")
		}
	})

}

func TestAttributeService_GetByID(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	attrRepo := NewAttributeRepository(db)
	attrService := NewAttributeService(attrRepo)

	// Create character and attribute
	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")
	repo.Save(character)

	attr, _ := NewAttribute(character.ID, 0, 1, "Health: Max", "10")
	attr, _ = attrService.Save(attr)

	t.Run("get existing", func(t *testing.T) {
		retrieved, err := attrService.GetByID(attr.ID)
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}

		if retrieved.ID != attr.ID {
			t.Errorf("Expected ID %q, got %q", attr.ID, retrieved.ID)
		}
		if retrieved.Name != "Health: Max" {
			t.Errorf("Expected name 'Health: Max', got %q", retrieved.Name)
		}
		if retrieved.Value != "10" {
			t.Errorf("Expected value '10', got %q", retrieved.Value)
		}
	})

	t.Run("get non-existent", func(t *testing.T) {
		_, err := attrService.GetByID(9999)
		if err == nil {
			t.Fatalf("Expected error")
		}
	})
}

func TestAttributeService_GetForCharacter(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	attrRepo := NewAttributeRepository(db)
	attrService := NewAttributeService(attrRepo)

	// Create character and attributes
	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")
	repo.Save(character)

	attr, _ := NewAttribute(character.ID, 0, 1, "Health: Max", "10")
	attr, _ = attrService.Save(attr)

	attr, _ = NewAttribute(character.ID, 0, 1, "Health: Current", "8")
	attr, _ = attrService.Save(attr)

	attr, _ = NewAttribute(character.ID, 0, 1, "Skill: Swordmaster", "3")
	attr, _ = attrService.Save(attr)

	t.Run("get all attributes", func(t *testing.T) {
		retrieved, err := attrService.GetForCharacter(character.ID)
		if err != nil {
			t.Fatalf("GetForCharacter() failed: %v", err)
		}

		if len(retrieved) != 3 {
			t.Errorf("Exepected 3 attributes, retrieved %d", len(retrieved))
		}
	})
}

func TestAttributeService_Delete(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	attrRepo := NewAttributeRepository(db)
	attrService := NewAttributeService(attrRepo)

	// Create character and attribute
	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")
	repo.Save(character)

	attr, _ := NewAttribute(character.ID, 0, 1, "Health: Max", "10")
	attr, _ = attrService.Save(attr)

	attr, _ = NewAttribute(character.ID, 0, 1, "Health: Current", "8")
	attr, _ = attrService.Save(attr)

	t.Run("delete existing", func(t *testing.T) {
		// Now delete it
		err := attrService.Delete(attr.ID)
		if err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}

		// Verify it was deleted
		_, err = attrService.GetByID(attr.ID)
		if err == nil {
			t.Error("Expected error for deleted entry, got nil")
		}
	})

}
