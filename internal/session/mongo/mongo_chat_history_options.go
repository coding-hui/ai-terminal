package mongo

import (
	"errors"
)

const (
	mongoDefaultDBName         = "chat_history"
	mongoDefaultCollectionName = "message_store"
)

var (
	errMongoInvalidURL            = errors.New("invalid mongo url option")
	errMongoInvalidChatEngineMode = errors.New("invalid mongo chat engine mode option")
)

type ChatMessageHistoryOption func(m *ChatMessageHistory)

func applyMongoDBChatOptions(options ...ChatMessageHistoryOption) (*ChatMessageHistory, error) {
	h := &ChatMessageHistory{
		databaseName:   mongoDefaultDBName,
		collectionName: mongoDefaultCollectionName,
	}

	for _, option := range options {
		option(h)
	}

	if h.url == "" {
		return nil, errMongoInvalidURL
	}
	if h.chatEngineMode == "" {
		return nil, errMongoInvalidChatEngineMode
	}

	return h, nil
}

// WithConnectionURL is an option for specifying the MongoDB connection URL. Must be set.
func WithConnectionURL(connectionURL string) ChatMessageHistoryOption {
	return func(p *ChatMessageHistory) {
		p.url = connectionURL
	}
}

// WithChatEngineMode is an arbitrary key that is used to store the messages of a single chat session,
// like exec,chat etc. Must be set.
func WithChatEngineMode(chatEngineMode string) ChatMessageHistoryOption {
	return func(p *ChatMessageHistory) {
		p.chatEngineMode = chatEngineMode
	}
}

// WithCollectionName is an option for specifying the collection name.
func WithCollectionName(name string) ChatMessageHistoryOption {
	return func(p *ChatMessageHistory) {
		p.collectionName = name
	}
}

// WithDataBaseName is an option for specifying the database name.
func WithDataBaseName(name string) ChatMessageHistoryOption {
	return func(p *ChatMessageHistory) {
		p.databaseName = name
	}
}
