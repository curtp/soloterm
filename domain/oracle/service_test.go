package oracle

import (
	"testing"

	"github.com/stretchr/testify/assert"

	_ "soloterm/domain/game"
	testhelper "soloterm/shared/testing"
)

func TestService_Lookup(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	repo := NewRepository(db)
	svc := *NewService(repo)

	// Insert two tables with the same name in different categories
	insertOracle(t, svc, "Fantasy", "descriptors", "ancient\nmystical\ncursed", 0, 0)
	insertOracle(t, svc, "Cyberpunk", "descriptors", "neon\nchrome\nhacked", 1, 0)
	insertOracle(t, svc, "Fantasy", "moods", "dark\nforeboding", 0, 1)

	t.Run("unqualified merges all matching tables", func(t *testing.T) {
		entries, ok := svc.Lookup("descriptors")
		assert.True(t, ok)
		// Both Fantasy and Cyberpunk descriptors combined = 6 entries
		assert.Equal(t, 6, len(entries))
		assert.Contains(t, entries, "ancient")
		assert.Contains(t, entries, "neon")
	})

	t.Run("unqualified is case-insensitive", func(t *testing.T) {
		entries, ok := svc.Lookup("Descriptors")
		assert.True(t, ok)
		assert.Equal(t, 6, len(entries))
	})

	t.Run("qualified returns only the specified category's table", func(t *testing.T) {
		entries, ok := svc.Lookup("fantasy/descriptors")
		assert.True(t, ok)
		assert.Equal(t, 3, len(entries))
		assert.Contains(t, entries, "ancient")
		assert.NotContains(t, entries, "neon")
	})

	t.Run("qualified is case-insensitive for both parts", func(t *testing.T) {
		entries, ok := svc.Lookup("CYBERPUNK/DESCRIPTORS")
		assert.True(t, ok)
		assert.Equal(t, 3, len(entries))
		assert.Contains(t, entries, "neon")
	})

	t.Run("unqualified single match returns that table's entries", func(t *testing.T) {
		entries, ok := svc.Lookup("moods")
		assert.True(t, ok)
		assert.Equal(t, 2, len(entries))
	})

	t.Run("unknown table returns false", func(t *testing.T) {
		_, ok := svc.Lookup("nonexistent")
		assert.False(t, ok)
	})

	t.Run("qualified unknown category returns false", func(t *testing.T) {
		_, ok := svc.Lookup("scifi/descriptors")
		assert.False(t, ok)
	})
}

// insertOracle is a test helper that inserts a fully-specified oracle row.
func insertOracle(t *testing.T, svc Service, category, name, content string, catPos, posInCat int) {
	t.Helper()
	oracle, err := NewOracle(category, name)
	assert.Nil(t, err)
	oracle.Content = content
	oracle.CategoryPosition = catPos
	oracle.PositionInCategory = posInCat
	_, err = svc.Save(oracle)
	assert.Nil(t, err)
}
