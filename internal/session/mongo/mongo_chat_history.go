package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/coding-hui/ai-terminal/internal/cli/options"
	"github.com/coding-hui/ai-terminal/internal/session"
	"github.com/coding-hui/ai-terminal/internal/util/mongodb"
)

const (
	mongoSessionIDKey      = "ChatID"
	mongoChatEngineModeKey = "ChatEngineMode"
)

func init() {
	session.RegisterHistoryStore(&mongoDBChatMessageHistoryFactory{})
}

type mongoDBChatMessageHistoryFactory struct{}

func (h *mongoDBChatMessageHistoryFactory) Type() string {
	return "mongo"
}

func (h *mongoDBChatMessageHistoryFactory) Create(cfg options.Config, chatEngineMode string) (session.History, error) {
	datastore := cfg.DataStore
	url := datastore.Url
	if url == "" {
		url = "mongodb://localhost:27017/"
	}
	chatHistory, err := NewMongoDBChatMessageSession(context.Background(),
		WithChatEngineMode(chatEngineMode),
		WithConnectionURL(url),
		WithDataBaseName(mongoDefaultDBName),
		WithCollectionName(mongoDefaultCollectionName),
	)
	if err != nil {
		return nil, err
	}
	return chatHistory, nil
}

type ChatMessageHistory struct {
	url            string
	chatEngineMode string
	databaseName   string
	collectionName string
	client         *mongo.Client
	collection     *mongo.Collection
}

type chatMessageModel struct {
	CreateTime     time.Time `bson:"CreateTime" json:"CreateTime"`
	ChatEngineMode string    `bson:"ChatEngineMode" json:"ChatEngineMode"`
	SessionID      string    `bson:"ChatID" json:"ChatID"`
	History        string    `bson:"History"   json:"History"`
}

// Statically assert that MongoDBChatMessageHistory implement the chat message history interface.
var _ session.History = &ChatMessageHistory{}

// NewMongoDBChatMessageSession creates a new MongoDBChatMessageHistory using chat message options.
func NewMongoDBChatMessageSession(ctx context.Context, options ...ChatMessageHistoryOption) (*ChatMessageHistory, error) {
	h, err := applyMongoDBChatOptions(options...)
	if err != nil {
		return nil, err
	}

	client, err := mongodb.NewClient(ctx, h.url)
	if err != nil {
		return nil, err
	}

	h.client = client

	h.collection = client.Database(h.databaseName).Collection(h.collectionName)
	// create session id index
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: mongoSessionIDKey, Value: 1}}},
		{Keys: bson.D{{Key: mongoChatEngineModeKey, Value: 1}}},
	}
	if _, err := h.collection.Indexes().CreateMany(ctx, indexes); err != nil {
		return nil, err
	}

	return h, nil
}

// AddAIMessage adds an AIMessage to the chat message history.
func (h *ChatMessageHistory) AddAIMessage(ctx context.Context, sessionID, message string) error {
	return h.AddMessage(ctx, sessionID, llms.AIChatMessage{Content: message})
}

// AddUserMessage adds a user to the chat message history.
func (h *ChatMessageHistory) AddUserMessage(ctx context.Context, sessionID, message string) error {
	return h.AddMessage(ctx, sessionID, llms.HumanChatMessage{Content: message})
}

// AddMessage adds a message to the store.
func (h *ChatMessageHistory) AddMessage(ctx context.Context, sessionID string, message llms.ChatMessage) error {
	_message, err := json.Marshal(llms.ConvertChatMessageToModel(message))
	if err != nil {
		return err
	}

	_, err = h.collection.InsertOne(ctx, chatMessageModel{
		CreateTime:     time.Now(),
		ChatEngineMode: h.chatEngineMode,
		SessionID:      sessionID,
		History:        string(_message),
	})

	return err
}

// SetMessages replaces existing messages in the store.
func (h *ChatMessageHistory) SetMessages(ctx context.Context, sessionID string, messages []llms.ChatMessage) error {
	_messages := []interface{}{}
	for _, message := range messages {
		_message, err := json.Marshal(llms.ConvertChatMessageToModel(message))
		if err != nil {
			return err
		}
		_messages = append(_messages, chatMessageModel{
			CreateTime:     time.Now(),
			ChatEngineMode: h.chatEngineMode,
			SessionID:      sessionID,
			History:        string(_message),
		})
	}

	if err := h.Clear(ctx, sessionID); err != nil {
		return err
	}

	_, err := h.collection.InsertMany(ctx, _messages)
	return err
}

// Messages returns all messages stored.
func (h *ChatMessageHistory) Messages(ctx context.Context, sessionID string) ([]llms.ChatMessage, error) {
	messages := []llms.ChatMessage{}
	filter := bson.M{mongoSessionIDKey: sessionID, mongoChatEngineModeKey: h.chatEngineMode}
	cursor, err := h.collection.Find(ctx, filter)
	if err != nil {
		return messages, err
	}

	_messages := []chatMessageModel{}
	if err := cursor.All(ctx, &_messages); err != nil {
		return messages, err
	}
	for _, message := range _messages {
		m := llms.ChatMessageModel{}
		if err := json.Unmarshal([]byte(message.History), &m); err != nil {
			return messages, err
		}
		messages = append(messages, m.ToChatMessage())
	}

	return messages, nil
}

// Sessions retrieves all sessions id from the MongoDB.
func (h *ChatMessageHistory) Sessions(ctx context.Context) ([]string, error) {
	sessions := []string{}
	filter := bson.M{mongoChatEngineModeKey: h.chatEngineMode}

	results, err := h.collection.Distinct(ctx, mongoSessionIDKey, filter)
	if err != nil {
		return sessions, err
	}

	for _, res := range results {
		sessions = append(sessions, fmt.Sprintf("%v", res))
	}

	return sessions, nil
}

// Clear clear session memory from MongoDB.
func (h *ChatMessageHistory) Clear(ctx context.Context, sessionID string) error {
	filter := bson.M{mongoSessionIDKey: sessionID}
	_, err := h.collection.DeleteMany(ctx, filter)
	return err
}

func (h *ChatMessageHistory) Exists(ctx context.Context, sessionID string) (bool, error) {
	filter := bson.M{mongoSessionIDKey: sessionID}
	count, err := h.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
