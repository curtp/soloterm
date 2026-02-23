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

func TestAttributeRepository_UpdatePosition(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	attrRepo := NewAttributeRepository(db)

	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")
	repo.Save(character)

	t.Run("updates group and position", func(t *testing.T) {
		attr, _ := NewAttribute(character.ID, 0, 0, "Strength", "10")
		attrRepo.Save(attr)

		err := attrRepo.UpdatePosition(attr.ID, 2, 3)
		if err != nil {
			t.Fatalf("UpdatePosition() failed: %v", err)
		}

		retrieved, err := attrRepo.GetByID(attr.ID)
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}
		if retrieved.Group != 2 {
			t.Errorf("Expected Group=2, got %d", retrieved.Group)
		}
		if retrieved.PositionInGroup != 3 {
			t.Errorf("Expected PositionInGroup=3, got %d", retrieved.PositionInGroup)
		}
	})

	t.Run("no error for non-existent id", func(t *testing.T) {
		err := attrRepo.UpdatePosition(999999, 0, 0)
		if err != nil {
			t.Errorf("UpdatePosition() expected no error for non-existent id, got: %v", err)
		}
	})
}

func TestAttributeRepository_SwapGroups(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	attrRepo := NewAttributeRepository(db)

	character, _ := NewCharacter("Test Character", "FlexD6", "Fighter", "Human")
	repo.Save(character)

	t.Run("swaps two single-item groups", func(t *testing.T) {
		a, _ := NewAttribute(character.ID, 10, 0, "Alpha", "1")
		attrRepo.Save(a)
		b, _ := NewAttribute(character.ID, 20, 0, "Beta", "2")
		attrRepo.Save(b)

		err := attrRepo.SwapGroups(character.ID, 10, 20)
		if err != nil {
			t.Fatalf("SwapGroups() failed: %v", err)
		}

		aAfter, _ := attrRepo.GetByID(a.ID)
		bAfter, _ := attrRepo.GetByID(b.ID)

		if aAfter.Group != 20 {
			t.Errorf("Alpha: expected Group=20, got %d", aAfter.Group)
		}
		if bAfter.Group != 10 {
			t.Errorf("Beta: expected Group=10, got %d", bAfter.Group)
		}
	})

	t.Run("swaps all items in multi-item groups", func(t *testing.T) {
		header, _ := NewAttribute(character.ID, 30, 0, "Stats", "")
		attrRepo.Save(header)
		child1, _ := NewAttribute(character.ID, 30, 1, "HP", "10")
		attrRepo.Save(child1)
		child2, _ := NewAttribute(character.ID, 30, 2, "XP", "0")
		attrRepo.Save(child2)

		standalone, _ := NewAttribute(character.ID, 40, 0, "Gear", "")
		attrRepo.Save(standalone)

		err := attrRepo.SwapGroups(character.ID, 30, 40)
		if err != nil {
			t.Fatalf("SwapGroups() failed: %v", err)
		}

		for _, id := range []int64{header.ID, child1.ID, child2.ID} {
			a, _ := attrRepo.GetByID(id)
			if a.Group != 40 {
				t.Errorf("id=%d: expected Group=40, got %d", id, a.Group)
			}
		}
		gearAfter, _ := attrRepo.GetByID(standalone.ID)
		if gearAfter.Group != 30 {
			t.Errorf("Gear: expected Group=30, got %d", gearAfter.Group)
		}
	})

	t.Run("does not affect other characters", func(t *testing.T) {
		other, _ := NewCharacter("Other Character", "FlexD6", "Rogue", "Elf")
		repo.Save(other)

		bystander, _ := NewAttribute(other.ID, 50, 0, "Bystander", "x")
		attrRepo.Save(bystander)
		target, _ := NewAttribute(character.ID, 50, 0, "Target", "y")
		attrRepo.Save(target)
		target2, _ := NewAttribute(character.ID, 60, 0, "Target2", "z")
		attrRepo.Save(target2)

		err := attrRepo.SwapGroups(character.ID, 50, 60)
		if err != nil {
			t.Fatalf("SwapGroups() failed: %v", err)
		}

		bystanderAfter, _ := attrRepo.GetByID(bystander.ID)
		if bystanderAfter.Group != 50 {
			t.Errorf("Bystander on other character should not be affected: expected Group=50, got %d", bystanderAfter.Group)
		}
	})

	t.Run("does not affect unrelated groups", func(t *testing.T) {
		unrelated, _ := NewAttribute(character.ID, 70, 0, "Unrelated", "q")
		attrRepo.Save(unrelated)
		g80, _ := NewAttribute(character.ID, 80, 0, "G80", "r")
		attrRepo.Save(g80)
		g90, _ := NewAttribute(character.ID, 90, 0, "G90", "s")
		attrRepo.Save(g90)

		err := attrRepo.SwapGroups(character.ID, 80, 90)
		if err != nil {
			t.Fatalf("SwapGroups() failed: %v", err)
		}

		unrelatedAfter, _ := attrRepo.GetByID(unrelated.ID)
		if unrelatedAfter.Group != 70 {
			t.Errorf("Unrelated group should not be affected: expected Group=70, got %d", unrelatedAfter.Group)
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
