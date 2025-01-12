package sqlite3

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	_ "modernc.org/sqlite"

	"github.com/coding-hui/ai-terminal/internal/conversation"
	"github.com/coding-hui/ai-terminal/internal/options"
)

func init() {
	conversation.RegisterConversationStore(&sqliteStoreFactor{})
}

var (
	errNoMatches   = errors.New("no conversations found")
	errManyMatches = errors.New("multiple conversations matched the input")
)

type sqliteStoreFactor struct{}

func (s *sqliteStoreFactor) Type() string {
	return "db"
}

func (s *sqliteStoreFactor) Create(options options.Config) (conversation.Store, error) {
	return NewSqliteStore(
		WithDataPath(options.DataStore.CachePath),
		WithConversation(options.ChatID),
	), nil
}

type SqliteStore struct {
	// DB is the database connection.
	DB *sqlx.DB
	// Ctx is a context that can be used for the schema exec.
	Ctx context.Context
	// DBAddress is the address or file path for connecting the db.
	DBAddress string
	// ConvoID defines a session name or ConvoID for a conversation.
	ConvoID string
	// DataPath is the path to the data directory.
	DataPath string

	*conversation.SimpleChatHistoryStore
}

// Statically assert that SqliteStore implement the chat message history interface.
var _ conversation.Store = &SqliteStore{}

func NewSqliteStore(options ...SqliteChatMessageHistoryOption) *SqliteStore {
	return applyChatOptions(options...)
}

// Latest returns the last message in the chat history.
func (h *SqliteStore) Latest(ctx context.Context) (*conversation.Conversation, error) {
	var convo conversation.Conversation
	if err := h.DB.Get(&convo, `
		SELECT
		  *
		FROM
		  conversations
		ORDER BY
		  updated_at DESC
		LIMIT
		  1
	`); err != nil {
		return nil, fmt.Errorf("FindHead: %w", err)
	}
	return &convo, nil
}

// Get retrieves a conversation from the store
func (h *SqliteStore) Get(ctx context.Context, id string) (*conversation.Conversation, error) {
	var conversations []conversation.Conversation
	var err error

	if len(id) < conversation.Sha1minLen {
		err = h.findByExactTitle(ctx, &conversations, id)
	} else {
		err = h.findByIDOrTitle(ctx, &conversations, id)
	}
	if err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}

	if len(conversations) > 1 {
		return nil, errManyMatches
	}
	if len(conversations) == 1 {
		return &conversations[0], nil
	}
	return nil, errNoMatches
}

func (h *SqliteStore) findByExactTitle(ctx context.Context, result *[]conversation.Conversation, in string) error {
	if err := h.DB.SelectContext(ctx, result, h.DB.Rebind(`
		SELECT
		  *
		FROM
		  conversations
		WHERE
		  title = ?
	`), in); err != nil {
		return fmt.Errorf("findByExactTitle: %w", err)
	}
	return nil
}

func (h *SqliteStore) findByIDOrTitle(ctx context.Context, result *[]conversation.Conversation, in string) error {
	if err := h.DB.SelectContext(ctx, result, h.DB.Rebind(`
		SELECT
		  *
		FROM
		  conversations
		WHERE
		  id glob ?
		  OR title = ?
	`), in+"*", in); err != nil {
		return fmt.Errorf("findByIDOrTitle: %w", err)
	}
	return nil
}

// List retrieves all conversation ConvoID from the store
func (h *SqliteStore) List(ctx context.Context) ([]conversation.Conversation, error) {
	var convos []conversation.Conversation
	if err := h.DB.SelectContext(ctx, &convos, `
		SELECT
		  *
		FROM
		  conversations
		ORDER BY
		  updated_at DESC
	`); err != nil {
		return convos, fmt.Errorf("List: %w", err)
	}
	return convos, nil
}

// ListOlderThan retrieves all conversation ConvoID from the store that are older than the given time.
func (h *SqliteStore) ListOlderThan(ctx context.Context, t time.Duration) ([]conversation.Conversation, error) {
	var convos []conversation.Conversation
	if err := h.DB.SelectContext(ctx, &convos, h.DB.Rebind(`
		SELECT
		  *
		FROM
		  conversations
		WHERE
		  updated_at < ?
		`), time.Now().Add(-t)); err != nil {
		return nil, fmt.Errorf("ListOlderThan: %w", err)
	}
	return convos, nil
}

func (h *SqliteStore) Save(ctx context.Context, id, title, model string) error {
	res, err := h.DB.ExecContext(ctx, h.DB.Rebind(`
		UPDATE conversations
		SET
		  title = ?,
		  model = ?,
		  updated_at = CURRENT_TIMESTAMP
		WHERE
		  id = ?
	`), title, model, id)
	if err != nil {
		return fmt.Errorf("Save: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("Save: %w", err)
	}

	if rows > 0 {
		return nil
	}

	if _, err := h.DB.ExecContext(ctx, h.DB.Rebind(`
		INSERT INTO
		  conversations (id, title, model)
		VALUES
		  (?, ?, ?)
	`), id, title, model); err != nil {
		return fmt.Errorf("Save: %w", err)
	}

	return nil
}

func (h *SqliteStore) Delete(ctx context.Context, id string) error {
	if _, err := h.DB.ExecContext(ctx, h.DB.Rebind(`
		DELETE FROM conversations
		WHERE
		  id = ?
	`), id); err != nil {
		return fmt.Errorf("Delete: %w", err)
	}
	return nil
}

// Clear resets messages.
func (h *SqliteStore) Clear(ctx context.Context) error {
	if _, err := h.DB.ExecContext(ctx, `DELETE FROM conversations`); err != nil {
		return fmt.Errorf("Clear: %w", err)
	}
	return nil
}

// Exists checks if the given chat conversation exists.
func (h *SqliteStore) Exists(ctx context.Context, id string) (bool, error) {
	var count int
	if err := h.DB.GetContext(ctx, &count, h.DB.Rebind(`
		SELECT
		  COUNT(*)
		FROM
		  conversations
		WHERE
		  id =?
	`), id); err != nil {
		return false, fmt.Errorf("Exists: %w", err)
	}
	return count > 0, nil
}

func (h *SqliteStore) Close() error {
	return h.DB.Close() //nolint: wrapcheck
}
