package oracle

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	// Blank import to register the games table migration (GetByID / GetAllForGame JOIN games)
	_ "soloterm/domain/game"
	testhelper "soloterm/shared/testing"
)

func TestRepository_Save(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)
	repo := NewRepository(db)

	t.Run("save new oracle", func(t *testing.T) {

		oracle, _ := NewOracle("fantasy", "foobar")
		oracle.Content = "oracle content"

		err := repo.Save(oracle)
		if err != nil {
			t.Fatalf("Creating a new oracle failed: %v", err)
		}

		// Verify: Database managed fields are set after insert
		assert.NotEqual(t, oracle.ID, 0, "Exepected oracle.ID not to be 0")
		assert.False(t, oracle.CreatedAt.IsZero(), "Expected created_at to be populated")
		assert.False(t, oracle.UpdatedAt.IsZero(), "Expected updated_at to be populated")

	})

	t.Run("update existing oracle", func(t *testing.T) {
		// Create initial oracle
		oracle, _ := NewOracle("sci-fi", "a name")
		oracle.Content = "oracle content"

		err := repo.Save(oracle)
		if err != nil {
			t.Fatalf("Save() insert failed: %v", err)
		}

		// Verify the DB managed fields are set
		assert.False(t, oracle.CreatedAt.IsZero(), "Expected created_at to be populated")
		assert.False(t, oracle.UpdatedAt.IsZero(), "Expected updated_at to be populated")

		firstCreatedAt := oracle.CreatedAt
		firstUpdatedAt := oracle.UpdatedAt

		// Wait a moment to ensure timestamps differ
		time.Sleep(10 * time.Millisecond)

		// Update the oracle
		oracle.Content = "Updated Result"
		err = repo.Save(oracle)
		if err != nil {
			t.Fatalf("Save() update failed: %v", err)
		}

		assert.NotEqual(t, oracle.UpdatedAt, firstUpdatedAt, "Expected updated_at to be updated")
		assert.Equal(t, oracle.CreatedAt, firstCreatedAt, "Exepected created_at to remain unchanged")

	})
}

func TestRepository_Get(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)

	// Create multiple oracles
	oracleID1 := testhelper.CreateTestOracle(t, db, "Oracle_1", "yes\r\nno")
	oracleID2 := testhelper.CreateTestOracle(t, db, "Oracle_2", "maybe\r\nprobably")

	t.Run("get existing", func(t *testing.T) {
		// Test: Get existing oracle by ID
		retrieved, err := repo.GetByID(oracleID1)
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}

		assert.Equalf(t, retrieved.ID, oracleID1, "Expected to retrieve id %d", oracleID1)

		retrieved, err = repo.GetByID(oracleID2)
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}

		assert.Equalf(t, retrieved.ID, oracleID2, "Expected to retrieve id %d", oracleID2)

	})

	t.Run("get invalid id", func(t *testing.T) {
		// Test: Get non-existent oracle
		_, err := repo.GetByID(999999)
		if err == nil {
			t.Error("Expected error for non-existent ID, got nil")
		}

		// Test: Zero ID should return error
		_, err = repo.GetByID(0)
		assert.NotNilf(t, err, "Expected an error for zero ID")
	})

	t.Run("get all", func(t *testing.T) {
		// Test: Get all oracles
		oracles, err := repo.GetAll()
		if err != nil {
			t.Fatalf("GetAll() failed: %v", err)
		}

		assert.Equalf(t, 2, len(oracles), "Expected 2 oracles, got %d", len(oracles))
	})

	t.Run("get by name", func(t *testing.T) {
		// Test: Get oracle by name
		oracle, err := repo.GetByName("Oracle_1")
		if err != nil {
			t.Fatalf("GetByName() failed: %v", err)
		}

		assert.NotNil(t, oracle)

		oracle, err = repo.GetByName("Oracle")
		if err == nil {
			t.Fatal("Expected an error")
		}

		assert.Nil(t, oracle)

	})

}

func TestRepository_Delete(t *testing.T) {
	// Setup
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)
	repo := NewRepository(db)
	oracleID := testhelper.CreateTestOracle(t, db, "TestOracle1", "yes\r\nno")

	t.Run("delete existing oracle", func(t *testing.T) {
		// Now delete it
		count, err := repo.Delete(oracleID)
		if err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}

		assert.Equal(t, int64(1), count)

		// Verify it was deleted
		_, err = repo.GetByID(oracleID)
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
