package sqlite3

import (
	"context"
	"database/sql"
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
	res, err := s.db.ExecContext(ctx, s.db.Rebind(`
		UPDATE load_contexts
		SET
			type = ?,
			url = ?,
			file_path = ?,
			content = ?,
			name = ?,
			conversation_id = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE
			id = ?
	`), lc.Type, lc.URL, lc.FilePath, lc.Content, lc.Name, lc.ConversationID, lc.ID)
	if err != nil {
		return fmt.Errorf("SaveContext: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("SaveContext: %w", err)
	}

	if rows > 0 {
		return nil
	}

	resp, err := s.db.ExecContext(ctx, s.db.Rebind(`
		INSERT INTO load_contexts (
			type, url, file_path, content, name, conversation_id
		) VALUES (
			?, ?, ?, ?, ?, ?
		)
	`), lc.Type, lc.URL, lc.FilePath, lc.Content, lc.Name, lc.ConversationID)
	if err != nil {
		return fmt.Errorf("SaveContext: %w", err)
	}

	lastInsertId, err := resp.LastInsertId()
	if err != nil {
		return fmt.Errorf("SaveContext: %w", err)
	}
	lc.ID = uint64(lastInsertId)

	return nil
}

func (s *sqliteLoadContextStore) GetContext(ctx context.Context, id uint64) (*convo.LoadContext, error) {
	var lc convo.LoadContext
	err := s.db.GetContext(ctx, &lc, s.db.Rebind(`
		SELECT id, type, url, file_path, content, name, conversation_id, updated_at
		FROM load_contexts WHERE id = ?
	`), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %v", errLoadContextNotFound, err)
		}
		return nil, fmt.Errorf("GetContext: %w", err)
	}
	return &lc, nil
}

func (s *sqliteLoadContextStore) ListContextsByteConvoID(ctx context.Context, conversationID string) ([]convo.LoadContext, error) {
	var contexts []convo.LoadContext
	if err := s.db.SelectContext(ctx, &contexts, s.db.Rebind(`
		SELECT id, type, url, file_path, content, name, conversation_id, updated_at
		FROM load_contexts WHERE conversation_id = ?
	`), conversationID); err != nil {
		return nil, fmt.Errorf("ListContextsByteConvoID: %w", err)
	}
	return contexts, nil
}

func (s *sqliteLoadContextStore) DeleteContexts(ctx context.Context, id uint64) error {
	res, err := s.db.ExecContext(ctx, s.db.Rebind(`
		DELETE FROM load_contexts WHERE id = ?
	`), id)
	if err != nil {
		return fmt.Errorf("DeleteContexts: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteContexts: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("DeleteContexts: no rows affected")
	}

	return nil
}

func (s *sqliteLoadContextStore) CleanContexts(ctx context.Context, conversationID string) (int64, error) {
	res, err := s.db.ExecContext(ctx, s.db.Rebind(`
		DELETE FROM load_contexts WHERE conversation_id = ?
	`), conversationID)
	if err != nil {
		return 0, fmt.Errorf("CleanContexts: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("CleanContexts: %w", err)
	}

	if rows == 0 {
		return 0, fmt.Errorf("CleanContexts: no rows affected")
	}

	return rows, nil
}
