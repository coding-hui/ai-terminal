package sqlite3

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coding-hui/ai-terminal/internal/convo"
)

func TestLoadContextStore(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	store := newLoadContextStore(db)

	t.Run("SaveContext and GetContext LoadContext", func(t *testing.T) {
		lc := &convo.LoadContext{
			ID:             uint64(1),
			Type:           "file",
			URL:            "",
			FilePath:       "/path/to/file",
			Content:        "test content",
			Name:           "test.txt",
			ConversationID: "conv1",
		}

		err := store.SaveContext(ctx, lc)
		require.NoError(t, err)

		retrieved, err := store.GetContext(ctx, lc.ID)
		require.NoError(t, err)
		assert.Equal(t, lc.ID, retrieved.ID)
		assert.Equal(t, lc.Type, retrieved.Type)
		assert.Equal(t, lc.FilePath, retrieved.FilePath)
		assert.Equal(t, lc.Content, retrieved.Content)
		assert.Equal(t, lc.Name, retrieved.Name)
		assert.Equal(t, lc.ConversationID, retrieved.ConversationID)
	})

	t.Run("GetContext non-existent LoadContext", func(t *testing.T) {
		_, err := store.GetContext(ctx, uint64(999))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "load context not found")
	})

	t.Run("ListContextsByteConvoID LoadContexts by convo", func(t *testing.T) {
		convID := "conv2"
		lc1 := &convo.LoadContext{
			ID:             uint64(2),
			Type:           "file",
			FilePath:       "/path/to/file1",
			Content:        "content1",
			Name:           "file1.txt",
			ConversationID: convID,
		}
		lc2 := &convo.LoadContext{
			ID:             uint64(3),
			Type:           "file",
			FilePath:       "/path/to/file2",
			Content:        "content2",
			Name:           "file2.txt",
			ConversationID: convID,
		}

		require.NoError(t, store.SaveContext(ctx, lc1))
		require.NoError(t, store.SaveContext(ctx, lc2))

		contexts, err := store.ListContextsByteConvoID(ctx, convID)
		require.NoError(t, err)
		assert.Len(t, contexts, 2)
	})

	t.Run("DeleteContexts LoadContext", func(t *testing.T) {
		lc := &convo.LoadContext{
			ID:             uint64(4),
			Type:           "file",
			FilePath:       "/path/to/file",
			Content:        "test content",
			Name:           "test.txt",
			ConversationID: "conv3",
		}

		require.NoError(t, store.SaveContext(ctx, lc))
		require.NoError(t, store.DeleteContexts(ctx, lc.ID))

		_, err := store.GetContext(ctx, lc.ID)
		require.Error(t, err)
	})

	t.Run("CleanContexts LoadContexts by convo", func(t *testing.T) {
		convID := "conv4"
		lc1 := &convo.LoadContext{
			ID:             uint64(5),
			Type:           "file",
			FilePath:       "/path/to/file1",
			Content:        "content1",
			Name:           "file1.txt",
			ConversationID: convID,
		}
		lc2 := &convo.LoadContext{
			ID:             uint64(6),
			Type:           "file",
			FilePath:       "/path/to/file2",
			Content:        "content2",
			Name:           "file2.txt",
			ConversationID: convID,
		}

		require.NoError(t, store.SaveContext(ctx, lc1))
		require.NoError(t, store.SaveContext(ctx, lc2))

		require.NoError(t, store.CleanContexts(ctx, convID))

		contexts, err := store.ListContextsByteConvoID(ctx, convID)
		require.NoError(t, err)
		assert.Empty(t, contexts)
	})

	t.Run("Update existing LoadContext", func(t *testing.T) {
		lc := &convo.LoadContext{
			ID:             uint64(7),
			Type:           "file",
			FilePath:       "/path/to/file",
			Content:        "initial content",
			Name:           "test.txt",
			ConversationID: "conv5",
		}

		require.NoError(t, store.SaveContext(ctx, lc))

		// Update content
		lc.Content = "updated content"
		require.NoError(t, store.SaveContext(ctx, lc))

		retrieved, err := store.GetContext(ctx, lc.ID)
		require.NoError(t, err)
		assert.Equal(t, "updated content", retrieved.Content)
	})
}

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(DefaultSchema)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	return db
}
