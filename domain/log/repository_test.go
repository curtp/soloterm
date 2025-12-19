package log

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

	t.Run("save new log", func(t *testing.T) {

		log, _ := NewLog(1)
		log.LogType = CHARACTER_ACTION
		log.Description = "some description"
		log.Result = "some result"
		log.Narrative = "some narrative"

		err := repo.Save(log)
		if err != nil {
			t.Fatalf("Creating a new log failed: %v", err)
		}

		// Verify: Database managed fields are set after insert
		if log.ID == 0 {
			t.Errorf("Expected ID to be set after insert, got %d", log.ID)
		}
		if log.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set after insert")
		}
		if log.UpdatedAt.IsZero() {
			t.Error("Expected UpdatedAt to be set after insert")
		}
	})

	t.Run("update existing log", func(t *testing.T) {
		// Create initial game
		log, _ := NewLog(1)
		log.LogType = CHARACTER_ACTION
		log.Description = "some description"
		log.Result = "some result"
		log.Narrative = "some narrative"

		err := repo.Save(log)
		if err != nil {
			t.Fatalf("Save() insert failed: %v", err)
		}

		// Verify the DB managed fields are set
		if log.CreatedAt.IsZero() || log.UpdatedAt.IsZero() {
			t.Fatalf("Game was not saved with DB managed fields: %+v", log)
		}

		firstCreatedAt := log.CreatedAt
		firstUpdatedAt := log.UpdatedAt

		// Wait a moment to ensure timestamps differ
		time.Sleep(10 * time.Millisecond)

		// Update the log
		log.Result = "Updated Result"
		err = repo.Save(log)
		if err != nil {
			t.Fatalf("Save() update failed: %v", err)
		}

		// Make sure the updated_at column was updated
		if log.UpdatedAt.Equal(firstUpdatedAt) {
			t.Errorf("updated_at was not modified on update: original=%v, updated=%v",
				firstUpdatedAt, log.UpdatedAt)
		}

		// Verify created_at was not updated
		if !log.CreatedAt.Equal(firstCreatedAt) {
			t.Errorf("created_at was modified on update: original=%v, updated=%v",
				firstCreatedAt, log.CreatedAt)
		}

	})
}

func TestRepository_Get(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDBWithMigration(t, Migrate)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)

	// Create multiple games to interact with
	log1, _ := NewLog(1)
	log1.LogType = CHARACTER_ACTION
	log1.Description = "some description"
	log1.Result = "some result"
	log1.Narrative = "some narrative"

	err := repo.Save(log1)
	if err != nil {
		t.Fatalf("Save() insert failed: %v", err)
	}

	log1ID := log1.ID

	log2, _ := NewLog(1)
	log2.LogType = CHARACTER_ACTION
	log2.Description = "some description"
	log2.Result = "some result"
	log2.Narrative = "some narrative"

	err = repo.Save(log2)
	if err != nil {
		t.Fatalf("Save() insert failed: %v", err)
	}

	log3, _ := NewLog(2)
	log3.LogType = CHARACTER_ACTION
	log3.Description = "some description"
	log3.Result = "some result"
	log3.Narrative = "some narrative"

	err = repo.Save(log3)
	if err != nil {
		t.Fatalf("Save() insert failed: %v", err)
	}

	t.Run("get existing", func(t *testing.T) {
		// Test: Get existing log by ID
		retrieved, err := repo.GetByID(log1ID)
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}

		// Verify: Retrieved log matches
		if retrieved.ID != log1ID {
			t.Errorf("Expected ID %q, got %q", log1ID, retrieved.ID)
		}
	})

	t.Run("get invalid id", func(t *testing.T) {
		// Test: Get non-existent log
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

	t.Run("get all for game", func(t *testing.T) {
		// Test: Get all logs for a specific game
		logs, err := repo.GetAllForGame(1)
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}

		// Verify the number of logs retrieved match the game
		if len(logs) != 2 {
			t.Errorf("Expected 2 logs, got %d", len(logs))
		}
	})
}

func TestRepository_Delete(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDBWithMigration(t, Migrate)
	defer testhelper.TeardownTestDB(t, db)
	repo := NewRepository(db)

	log, _ := NewLog(1)
	log.LogType = CHARACTER_ACTION
	log.Description = "some description"
	log.Result = "some result"
	log.Narrative = "some narrative"

	err := repo.Save(log)
	if err != nil {
		t.Fatalf("Save() update failed: %v", err)
	}

	log2, _ := NewLog(1)
	log2.LogType = CHARACTER_ACTION
	log2.Description = "some description"
	log2.Result = "some result"
	log2.Narrative = "some narrative"

	err = repo.Save(log2)
	if err != nil {
		t.Fatalf("Save() update failed: %v", err)
	}

	log3, _ := NewLog(1)
	log3.LogType = CHARACTER_ACTION
	log3.Description = "some description"
	log3.Result = "some result"
	log3.Narrative = "some narrative"

	err = repo.Save(log3)
	if err != nil {
		t.Fatalf("Save() update failed: %v", err)
	}

	t.Run("delete existing log", func(t *testing.T) {
		// Now delete it
		count, err := repo.Delete(log.ID)
		if err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}

		if count != 1 {
			t.Errorf("Expected 1 row deleted, got %d", count)
		}

		// Verify it was deleted
		_, err = repo.GetByID(log.ID)
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

	t.Run("delete all for game", func(t *testing.T) {
		// Delete them
		count, err := repo.DeleteAllForGame(1)
		if err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}

		if count != 2 {
			t.Errorf("Expected 2 row deleted, got %d", count)
		}

		// Verify it was deleted
		_, err = repo.GetByID(log.ID)
		if err == nil {
			t.Error("Expected error for deleted entry, got nil")
		}
	})

}
