package sqlite3

import (
	"context"
	"os"

	"github.com/jmoiron/sqlx"

	"github.com/coding-hui/ai-terminal/internal/conversation"
	"github.com/coding-hui/ai-terminal/internal/errbook"
)

// DefaultLimit sets a default limit for select queries.
const DefaultLimit = 1000

// DefaultTableName sets a default table name.
const DefaultTableName = "ai_terminal_conversations"

// DefaultSchema sets a default schema to be run after connecting.
const DefaultSchema = `CREATE TABLE
		  IF NOT EXISTS conversations (
		    id string NOT NULL PRIMARY KEY,
		    title string NOT NULL,
		    model string NOT NULL,
		    updated_at datetime NOT NULL DEFAULT (strftime ('%Y-%m-%d %H:%M:%f', 'now')),
		    CHECK (id <> ''),
		    CHECK (title <> '')
		  );
CREATE INDEX IF NOT EXISTS idx_conv_id ON conversations (id);
CREATE INDEX IF NOT EXISTS idx_conv_title ON conversations (title)
`

// SqliteChatMessageHistoryOption is a function for creating new
// chat message history with other than the default values.
type SqliteChatMessageHistoryOption func(m *SqliteStore)

// WithDataPath is an option for NewSqliteChatMessageHistory for
// setting a path to the data directory.
func WithDataPath(path string) SqliteChatMessageHistoryOption {
	return func(m *SqliteStore) {
		m.DataPath = path
	}
}

// WithDB is an option for NewSqliteChatMessageHistory for adding
// a database connection.
func WithDB(db *sqlx.DB) SqliteChatMessageHistoryOption {
	return func(m *SqliteStore) {
		m.DB = db
	}
}

// WithContext is an option for NewSqliteChatMessageHistory
// to use a context internally when running Schema.
func WithContext(ctx context.Context) SqliteChatMessageHistoryOption {
	return func(m *SqliteStore) {
		m.Ctx = ctx //nolint:fatcontext
	}
}

// WithDBAddress is an option for NewSqliteChatMessageHistory for
// specifying an address or file path for when connecting the db.
func WithDBAddress(addr string) SqliteChatMessageHistoryOption {
	return func(m *SqliteStore) {
		m.DBAddress = addr
	}
}

// WithConversation is an option for NewSqliteChatMessageHistory for
// setting a session name or ConvoID for the history.
func WithConversation(convoID string) SqliteChatMessageHistoryOption {
	return func(m *SqliteStore) {
		m.ConvoID = convoID
	}
}

func applyChatOptions(options ...SqliteChatMessageHistoryOption) *SqliteStore {
	h := &SqliteStore{}

	for _, option := range options {
		option(h)
	}

	if h.Ctx == nil {
		h.Ctx = context.Background()
	}

	if h.DBAddress == "" {
		h.DBAddress = ":memory:"
	}

	if h.ConvoID == "" {
		h.ConvoID = "default"
	}

	if h.DataPath == "" {
		h.DataPath = "./data"
	}

	h.SimpleChatHistoryStore = conversation.NewSimpleChatHistoryStore(h.DataPath)

	if h.DB == nil {
		db, err := sqlx.Open("sqlite", h.DBAddress)
		if err != nil {
			errbook.HandleError(errbook.Wrap("Could not open database.", err))
			os.Exit(1)
		}
		h.DB = db
	}

	if err := h.DB.Ping(); err != nil {
		errbook.HandleError(errbook.Wrap("Could not connect to database.", err))
		os.Exit(1)
	}

	if _, err := h.DB.ExecContext(h.Ctx, DefaultSchema); err != nil {
		errbook.HandleError(errbook.Wrap("Could not create conversation db table.", err))
		os.Exit(1)
	}

	return h
}
