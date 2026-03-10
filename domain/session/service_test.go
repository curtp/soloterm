package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "soloterm/domain/game"
	testhelper "soloterm/shared/testing"
)

func TestService_Save(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)
	svc := NewService(NewRepository(db))
	gameID := testhelper.CreateTestGame(t, db, "Test Game")

	t.Run("rejects blank name", func(t *testing.T) {
		s, _ := NewSession(gameID)
		s.Name = ""
		_, err := svc.Save(s)
		assert.Error(t, err, "Expected validation error for blank name")
		assert.Equal(t, int64(0), s.ID, "Expected session to not be persisted")
	})

	t.Run("saves valid session", func(t *testing.T) {
		s, _ := NewSession(gameID)
		s.Name = "Session One"
		s.Content = "Some content"
		result, err := svc.Save(s)
		require.NoError(t, err)
		assert.NotEqual(t, int64(0), result.ID)
	})
}
