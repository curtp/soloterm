package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	testhelper "soloterm/shared/testing"
)

func TestRepository_Save(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)
	repo := NewRepository(db)

	t.Run("save new session", func(t *testing.T) {

		session, _ := NewSession(1)
		session.Name = "a name"
		session.Content = "session content"

		err := repo.Save(session)
		if err != nil {
			t.Fatalf("Creating a new session failed: %v", err)
		}

		// Verify: Database managed fields are set after insert
		assert.NotEqual(t, session.ID, 0, "Exepected session.ID not to be 0")
		assert.False(t, session.CreatedAt.IsZero(), "Expected created_at to be populated")
		assert.False(t, session.UpdatedAt.IsZero(), "Expected updated_at to be populated")

	})

	t.Run("update existing session", func(t *testing.T) {
		// Create initial session
		session, _ := NewSession(1)
		session.Name = "a name"
		session.Content = "session content"

		err := repo.Save(session)
		if err != nil {
			t.Fatalf("Save() insert failed: %v", err)
		}

		// Verify the DB managed fields are set
		assert.False(t, session.CreatedAt.IsZero(), "Expected created_at to be populated")
		assert.False(t, session.UpdatedAt.IsZero(), "Expected updated_at to be populated")

		firstCreatedAt := session.CreatedAt
		firstUpdatedAt := session.UpdatedAt

		// Wait a moment to ensure timestamps differ
		time.Sleep(10 * time.Millisecond)

		// Update the session
		session.Content = "Updated Result"
		err = repo.Save(session)
		if err != nil {
			t.Fatalf("Save() update failed: %v", err)
		}

		assert.NotEqual(t, session.UpdatedAt, firstUpdatedAt, "Expected updated_at to be updated")
		assert.Equal(t, session.CreatedAt, firstCreatedAt, "Exepected created_at to remain unchanged")

	})
}

func TestRepository_Get(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)

	// Create multiple sessions to interact with
	session1, _ := NewSession(1)
	session1.Name = "a name"
	session1.Content = "session content"

	err := repo.Save(session1)
	if err != nil {
		t.Fatalf("Save() insert failed: %v", err)
	}

	session1ID := session1.ID

	session2, _ := NewSession(1)
	session2.Name = "another name"
	session2.Content = "session content"

	err = repo.Save(session2)
	if err != nil {
		t.Fatalf("Save() insert failed: %v", err)
	}

	session3, _ := NewSession(2)
	session3.Name = "a name"
	session3.Content = "session content"

	err = repo.Save(session3)
	if err != nil {
		t.Fatalf("Save() insert failed: %v", err)
	}

	t.Run("get existing", func(t *testing.T) {
		// Test: Get existing session by ID
		retrieved, err := repo.GetByID(session1ID)
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}

		assert.Equalf(t, retrieved.ID, session1ID, "Expected to retrieve id %d", session1ID)
	})

	t.Run("get invalid id", func(t *testing.T) {
		// Test: Get non-existent session
		_, err = repo.GetByID(999999)
		if err == nil {
			t.Error("Expected error for non-existent ID, got nil")
		}

		// Test: Zero ID should return error
		_, err = repo.GetByID(0)
		assert.NotNilf(t, err, "Expected an error for zero ID")
	})

	t.Run("get all for game", func(t *testing.T) {
		// Test: Get all sessions for a specific session
		sessions, err := repo.GetAllForGame(1)
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}

		assert.Equalf(t, 2, len(sessions), "Expected 2 sessions, got %d", len(sessions))
		assert.Equal(t, "", sessions[0].Content)
	})
}

func TestRepository_Delete(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)
	repo := NewRepository(db)

	session, _ := NewSession(1)
	session.Name = "a name"
	session.Content = "some content"

	err := repo.Save(session)
	if err != nil {
		t.Fatalf("Save() update failed: %v", err)
	}

	session2, _ := NewSession(1)
	session2.Name = "another name"
	session2.Content = "some content"

	err = repo.Save(session2)
	if err != nil {
		t.Fatalf("Save() update failed: %v", err)
	}

	session3, _ := NewSession(1)
	session3.Name = "yet another name"
	session3.Content = "some content"

	err = repo.Save(session3)
	if err != nil {
		t.Fatalf("Save() update failed: %v", err)
	}

	t.Run("delete existing session", func(t *testing.T) {
		// Now delete it
		count, err := repo.Delete(session.ID)
		if err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}

		assert.Equal(t, int64(1), count)

		// Verify it was deleted
		_, err = repo.GetByID(session.ID)
		assert.NotNil(t, err)
	})

	t.Run("invalid id", func(t *testing.T) {
		// This should return an error
		count, err := repo.Delete(9999)
		if err == nil {
			t.Error("Expected error for non-existent key, got nil")
		}

		assert.Equal(t, int64(0), count)
	})

	t.Run("0 id throws an error", func(t *testing.T) {
		// This should return an error
		count, err := repo.Delete(0)
		if err == nil {
			t.Error("Expected error for non-existent key, got nil")
		}

		assert.Equal(t, int64(0), count)
	})
}

func TestRepository_DeleteAllForGame(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)
	repo := NewRepository(db)

	session, _ := NewSession(1)
	session.Name = "a name"
	session.Content = "some content"

	err := repo.Save(session)
	if err != nil {
		t.Fatalf("Save() update failed: %v", err)
	}

	session2, _ := NewSession(1)
	session2.Name = "another name"
	session2.Content = "some content"

	err = repo.Save(session2)
	if err != nil {
		t.Fatalf("Save() update failed: %v", err)
	}

	session3, _ := NewSession(1)
	session3.Name = "yet another name"
	session3.Content = "some content"

	err = repo.Save(session3)
	if err != nil {
		t.Fatalf("Save() update failed: %v", err)
	}

	t.Run("invalid session id does nothing", func(t *testing.T) {
		// Now delete it
		count, err := repo.DeleteAllForGame(9999)
		assert.NotNil(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("0 throws an error", func(t *testing.T) {
		// This should return an error
		count, err := repo.DeleteAllForGame(0)
		assert.NotNil(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("delete all for session", func(t *testing.T) {
		// Delete them
		count, err := repo.DeleteAllForGame(1)
		if err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}
		assert.Equal(t, int64(3), count)

		// Verify it was deleted
		_, err = repo.GetByID(session.ID)
		assert.NotNil(t, err)
	})

}
