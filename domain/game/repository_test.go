package game

import (
	"testing"
	"time"

	testhelper "soloterm/shared/testing"
)

func TestRepository_Save(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDBWithMigration(t, Migrate)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)

	t.Run("save new game", func(t *testing.T) {

		// Create a test game (with zero ID to trigger insert)
		game, _ := NewGame("Test Game")

		// Test: Save the game (insert)
		err := repo.Save(game)
		if err != nil {
			t.Fatalf("Save() insert failed: %v", err)
		}

		// Verify: Database managed fields are set after insert
		if game.Id == 0 {
			t.Errorf("Expected ID to be set after insert, got %d", game.Id)
		}
		if game.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set after insert")
		}
		if game.UpdatedAt.IsZero() {
			t.Error("Expected UpdatedAt to be set after insert")
		}
	})

	t.Run("update existing game", func(t *testing.T) {
		// Create initial game
		game, _ := NewGame("Original Name")
		err := repo.Save(game)
		if err != nil {
			t.Fatalf("Save() insert failed: %v", err)
		}

		// Verify the DB managed fields are set
		if game.CreatedAt.IsZero() || game.UpdatedAt.IsZero() || game.Id == 0 {
			t.Fatalf("Game was not saved with DB managed fields: %+v", game)
		}

		firstCreatedAt := game.CreatedAt
		firstUpdatedAt := game.UpdatedAt

		// Wait a moment to ensure timestamps differ
		time.Sleep(10 * time.Millisecond)

		// Update the game
		game.Name = "Updated Name"
		err = repo.Save(game)
		if err != nil {
			t.Fatalf("Save() update failed: %v", err)
		}

		// Make sure the updated_at column was updated
		if game.UpdatedAt.Equal(firstUpdatedAt) {
			t.Errorf("updated_at was not modified on update: original=%v, updated=%v",
				firstUpdatedAt, game.UpdatedAt)
		}

		// Verify created_at was not updated
		if !game.CreatedAt.Equal(firstCreatedAt) {
			t.Errorf("created_at was modified on update: original=%v, updated=%v",
				firstCreatedAt, game.CreatedAt)
		}

	})
}

func TestRepository_GetByID(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDBWithMigration(t, Migrate)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)

	// Create test game
	game, _ := NewGame("Test Game")
	err := repo.Save(game)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	gameId := game.Id

	t.Run("get existing", func(t *testing.T) {
		// Test: Get existing game by ID
		retrieved, err := repo.GetByID(gameId)
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}

		// Verify: Retrieved game matches
		if retrieved.Id != gameId {
			t.Errorf("Expected ID %q, got %q", gameId, retrieved.Id)
		}
		if retrieved.Name != "Test Game" {
			t.Errorf("Expected name 'Test Game', got %q", retrieved.Name)
		}
	})

	t.Run("get invalid id", func(t *testing.T) {
		// Test: Get non-existent game
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

func TestRepository_Delete(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDBWithMigration(t, Migrate)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)

	// Create game to delete
	game, _ := NewGame("Test Game")
	err := repo.Save(game)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	t.Run("delete existing", func(t *testing.T) {
		// Now delete it
		count, err := repo.Delete(game.Id)
		if err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}

		if count != 1 {
			t.Errorf("Expected 1 row deleted, got %d", count)
		}

		// Verify it was deleted
		_, err = repo.GetByID(game.Id)
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
}
