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

func TestAttributeService_Reorder(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	attrRepo := NewAttributeRepository(db)
	attrService := NewAttributeService(attrRepo)

	// Each subtest gets its own character so GetForCharacter only sees that test's attrs.
	newChar := func(name string) *Character {
		c, _ := NewCharacter(name, "FlexD6", "Fighter", "Human")
		repo.Save(c)
		return c
	}
	save := func(charID int64, group, pos int, name string) *Attribute {
		a, _ := NewAttribute(charID, group, pos, name, "val")
		attrRepo.Save(a)
		return a
	}

	t.Run("individual move down within group swaps positions", func(t *testing.T) {
		char := newChar("C1")
		a := save(char.ID, 0, 0, "A")
		b := save(char.ID, 0, 1, "B")

		id, err := attrService.Reorder(char.ID, a.ID, 1, false)
		if err != nil {
			t.Fatalf("Reorder() failed: %v", err)
		}
		if id != a.ID {
			t.Errorf("expected returned ID=%d, got %d", a.ID, id)
		}

		aAfter, _ := attrRepo.GetByID(a.ID)
		bAfter, _ := attrRepo.GetByID(b.ID)
		if aAfter.Group != 0 || aAfter.PositionInGroup != 1 {
			t.Errorf("A: expected g0/p1, got g%d/p%d", aAfter.Group, aAfter.PositionInGroup)
		}
		if bAfter.Group != 0 || bAfter.PositionInGroup != 0 {
			t.Errorf("B: expected g0/p0, got g%d/p%d", bAfter.Group, bAfter.PositionInGroup)
		}
	})

	t.Run("individual move up within group swaps positions", func(t *testing.T) {
		char := newChar("C2")
		a := save(char.ID, 0, 0, "A")
		b := save(char.ID, 0, 1, "B")

		attrService.Reorder(char.ID, b.ID, -1, false)

		aAfter, _ := attrRepo.GetByID(a.ID)
		bAfter, _ := attrRepo.GetByID(b.ID)
		if bAfter.PositionInGroup != 0 {
			t.Errorf("B: expected p0, got p%d", bAfter.PositionInGroup)
		}
		if aAfter.PositionInGroup != 1 {
			t.Errorf("A: expected p1, got p%d", aAfter.PositionInGroup)
		}
	})

	t.Run("individual move down across group boundary swaps group and position", func(t *testing.T) {
		char := newChar("C3")
		a := save(char.ID, 0, 0, "A")
		b := save(char.ID, 1, 0, "B")

		attrService.Reorder(char.ID, a.ID, 1, false)

		aAfter, _ := attrRepo.GetByID(a.ID)
		bAfter, _ := attrRepo.GetByID(b.ID)
		if aAfter.Group != 1 || aAfter.PositionInGroup != 0 {
			t.Errorf("A: expected g1/p0, got g%d/p%d", aAfter.Group, aAfter.PositionInGroup)
		}
		if bAfter.Group != 0 || bAfter.PositionInGroup != 0 {
			t.Errorf("B: expected g0/p0, got g%d/p%d", bAfter.Group, bAfter.PositionInGroup)
		}
	})

	t.Run("individual move up at top is no-op", func(t *testing.T) {
		char := newChar("C4")
		a := save(char.ID, 0, 0, "A")
		save(char.ID, 0, 1, "B")

		id, err := attrService.Reorder(char.ID, a.ID, -1, false)
		if err != nil {
			t.Fatalf("Reorder() failed: %v", err)
		}
		if id != 0 {
			t.Errorf("expected no-op (id=0), got %d", id)
		}
		aAfter, _ := attrRepo.GetByID(a.ID)
		if aAfter.Group != 0 || aAfter.PositionInGroup != 0 {
			t.Errorf("A should be unchanged: expected g0/p0, got g%d/p%d", aAfter.Group, aAfter.PositionInGroup)
		}
	})

	t.Run("individual move down at bottom is no-op", func(t *testing.T) {
		char := newChar("C5")
		save(char.ID, 0, 0, "A")
		b := save(char.ID, 0, 1, "B")

		id, err := attrService.Reorder(char.ID, b.ID, 1, false)
		if err != nil {
			t.Fatalf("Reorder() failed: %v", err)
		}
		if id != 0 {
			t.Errorf("expected no-op (id=0), got %d", id)
		}
	})

	t.Run("group move down swaps group numbers", func(t *testing.T) {
		char := newChar("C6")
		a := save(char.ID, 0, 0, "A")
		b := save(char.ID, 1, 0, "B")

		id, err := attrService.Reorder(char.ID, a.ID, 1, true)
		if err != nil {
			t.Fatalf("Reorder() failed: %v", err)
		}
		if id != a.ID {
			t.Errorf("expected returned ID=%d, got %d", a.ID, id)
		}

		aAfter, _ := attrRepo.GetByID(a.ID)
		bAfter, _ := attrRepo.GetByID(b.ID)
		if aAfter.Group != 1 {
			t.Errorf("A: expected g1, got g%d", aAfter.Group)
		}
		if bAfter.Group != 0 {
			t.Errorf("B: expected g0, got g%d", bAfter.Group)
		}
	})

	t.Run("group move from child moves entire group", func(t *testing.T) {
		char := newChar("C7")
		header := save(char.ID, 0, 0, "Header")
		child := save(char.ID, 0, 1, "Child")
		other := save(char.ID, 1, 0, "Other")

		// Move using the child (not the header)
		attrService.Reorder(char.ID, child.ID, 1, true)

		headerAfter, _ := attrRepo.GetByID(header.ID)
		childAfter, _ := attrRepo.GetByID(child.ID)
		otherAfter, _ := attrRepo.GetByID(other.ID)
		if headerAfter.Group != 1 {
			t.Errorf("Header: expected g1, got g%d", headerAfter.Group)
		}
		if childAfter.Group != 1 {
			t.Errorf("Child: expected g1, got g%d", childAfter.Group)
		}
		if otherAfter.Group != 0 {
			t.Errorf("Other: expected g0, got g%d", otherAfter.Group)
		}
	})

	t.Run("group move up at top is no-op", func(t *testing.T) {
		char := newChar("C8")
		a := save(char.ID, 0, 0, "A")
		save(char.ID, 1, 0, "B")

		id, err := attrService.Reorder(char.ID, a.ID, -1, true)
		if err != nil {
			t.Fatalf("Reorder() failed: %v", err)
		}
		if id != 0 {
			t.Errorf("expected no-op (id=0), got %d", id)
		}
		aAfter, _ := attrRepo.GetByID(a.ID)
		if aAfter.Group != 0 {
			t.Errorf("A should be unchanged: expected g0, got g%d", aAfter.Group)
		}
	})

	t.Run("group move down at bottom is no-op", func(t *testing.T) {
		char := newChar("C9")
		save(char.ID, 0, 0, "A")
		b := save(char.ID, 1, 0, "B")

		id, err := attrService.Reorder(char.ID, b.ID, 1, true)
		if err != nil {
			t.Fatalf("Reorder() failed: %v", err)
		}
		if id != 0 {
			t.Errorf("expected no-op (id=0), got %d", id)
		}
	})

	t.Run("returns error for non-existent attribute", func(t *testing.T) {
		char := newChar("C10")
		_, err := attrService.Reorder(char.ID, 999999, 1, false)
		if err == nil {
			t.Error("expected error for non-existent attribute, got nil")
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
