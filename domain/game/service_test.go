package game

import (
	"testing"
	"time"

	testhelper "soloterm/shared/testing"
)

func TestService_Save(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDBWithMigration(t, Migrate)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	service := NewService(repo)

	t.Run("save new game", func(t *testing.T) {

		// Create a test game (with zero ID to trigger insert)
		game, _ := NewGame("Test Game")

		// Test: Save the game (insert)
		game, err := service.Save(game)
		if err != nil {
			t.Fatalf("Save() insert failed: %v", err)
		}
	})

	t.Run("update existing game", func(t *testing.T) {
		// Create initial game
		game, _ := NewGame("Original Name")
		game, err := service.Save(game)
		if err != nil {
			t.Fatalf("Save() insert failed: %v", err)
		}

		// Wait a moment to ensure timestamps differ
		time.Sleep(10 * time.Millisecond)

		// Update the game
		game.Name = "Updated Name"
		game, err = service.Save(game)
		if err != nil {
			t.Fatalf("Save() update failed: %v", err)
		}

		// Make sure the updated_at column was updated
		if game.UpdatedAt.Equal(game.CreatedAt) {
			t.Errorf("updated_at was not modified on update: original=%v, updated=%v",
				game.UpdatedAt, game.CreatedAt)
		}
	})

	t.Run("validation errors prevent save", func(t *testing.T) {
		game, _ := NewGame("Original Name")
		desc := "a"
		game.Description = &desc
		game, err := service.Save(game)
		if err == nil {
			t.Fatalf("Save() should have failed with a validation error")
		}
	})

}

func TestService_GetByID(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDBWithMigration(t, Migrate)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	service := NewService(repo)

	// Create test game
	game, _ := NewGame("Test Game")
	game, err := service.Save(game)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	gameId := game.ID

	t.Run("get existing", func(t *testing.T) {
		// Test: Get existing game by ID
		retrieved, err := repo.GetByID(gameId)
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}

		// Verify: Retrieved game matches
		if retrieved.ID != gameId {
			t.Errorf("Expected ID %q, got %q", gameId, retrieved.ID)
		}
		if retrieved.Name != "Test Game" {
			t.Errorf("Expected name 'Test Game', got %q", retrieved.Name)
		}
	})
}

func TestService_GetAll(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDBWithMigration(t, Migrate)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	service := NewService(repo)

	// Create some games
	game, _ := NewGame("Test Game")
	service.Save(game)
	game, _ = NewGame("Another Game")
	service.Save(game)

	t.Run("get all", func(t *testing.T) {
		// Test: Get all games
		retrieved, err := service.GetAll()
		if err != nil {
			t.Fatalf("GetAll() failed: %v", err)
		}

		// Verify: Retrieved game count matches
		if len(retrieved) != 2 {
			t.Errorf("Expected to retrieve 2 games, got %d", len(retrieved))
		}
	})
}

func TestService_Delete(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDBWithMigration(t, Migrate)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	service := NewService(repo)

	// Create game to delete
	game, _ := NewGame("Test Game")
	game, err := service.Save(game)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	t.Run("delete existing", func(t *testing.T) {
		// Now delete it
		err := service.Delete(game.ID)
		if err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}

		// Verify it was deleted
		_, err = service.GetByID(game.ID)
		if err == nil {
			t.Error("Expected error for deleted entry, got nil")
		}
	})

}
