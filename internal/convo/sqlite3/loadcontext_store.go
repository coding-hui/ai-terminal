package sqlite3

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/coding-hui/ai-terminal/internal/convo"
)

var (
	errLoadContextNotFound = errors.New("load context not found")
)

type sqliteLoadContextStore struct {
	db *sqlx.DB
}

func newLoadContextStore(db *sqlx.DB) *sqliteLoadContextStore {
	return &sqliteLoadContextStore{db: db}
}

func (s *sqliteLoadContextStore) SaveContext(ctx context.Context, lc *convo.LoadContext) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO load_contexts (
			id, type, url, file_path, content, name, conversation_id
		) VALUES (
			?, ?, ?, ?, ?, ?, ?
		) ON CONFLICT(id) DO UPDATE SET
			type = excluded.type,
			url = excluded.url,
			file_path = excluded.file_path,
			content = excluded.content,
			name = excluded.name,
			conversation_id = excluded.conversation_id
	`, lc.ID, lc.Type, lc.URL, lc.FilePath, lc.Content, lc.Name, lc.ConversationID)
	return err
}

func (s *sqliteLoadContextStore) GetContext(ctx context.Context, id uint64) (*convo.LoadContext, error) {
	var lc convo.LoadContext
	err := s.db.GetContext(ctx, &lc, `
		SELECT id, type, url, file_path, content, name, conversation_id, updated_at
		FROM load_contexts WHERE id = ?
	`, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errLoadContextNotFound, err)
	}
	return &lc, nil
}

func (s *sqliteLoadContextStore) ListContextsByteConvoID(ctx context.Context, conversationID string) ([]convo.LoadContext, error) {
	var contexts []convo.LoadContext
	err := s.db.SelectContext(ctx, &contexts, `
		SELECT id, type, url, file_path, content, name, conversation_id, updated_at
		FROM load_contexts WHERE conversation_id = ?
	`, conversationID)
	return contexts, err
}

func (s *sqliteLoadContextStore) DeleteContexts(ctx context.Context, id uint64) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM load_contexts WHERE id = ?
	`, id)
	return err
}

func (s *sqliteLoadContextStore) CleanContexts(ctx context.Context, conversationID string) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM load_contexts WHERE conversation_id = ?
	`, conversationID)
	return err
}
