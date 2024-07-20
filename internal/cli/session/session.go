package session

import (
	"context"
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"k8s.io/klog/v2"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"

	"github.com/coding-hui/ai-terminal/internal/cli/options"
	"github.com/coding-hui/ai-terminal/internal/util/mongodb"
)

const (
	defaultSession        = "default"
	currentKey            = "Current"
	currentSessionId      = "SessionId"
	mongoDefaultDBName    = "chat_history"
	messageCollectionName = "message_store"
	sessionCollectionName = "session"
)

type Manager struct {
	Datastore options.DataStoreOptions

	mongoClient *mongo.Client
	mongoDb     *mongo.Database
}

func NewSessionManager(ds options.DataStoreOptions) (*Manager, error) {
	var mongoClient *mongo.Client
	var mongoDb *mongo.Database
	var err error
	if ds.Type == "mongo" {
		mongoClient, err = mongodb.NewClient(context.Background(), ds.Url)
		if err != nil {
			return nil, err
		}
		mongoDb = mongoClient.Database(mongoDefaultDBName)
	}
	return &Manager{
		Datastore:   ds,
		mongoClient: mongoClient,
		mongoDb:     mongoDb,
	}, nil
}

type chatMessageModel struct {
	SessionID string `bson:"SessionId" json:"SessionId"`
	History   string `bson:"History"   json:"History"`
}

type currentSessionModel struct {
	SessionID string `bson:"SessionId" json:"SessionId"`
	Current   bool   `bson:"Current" json:"Current"`
}

func (m *Manager) CurrentSession() string {
	ctx := context.Background()

	filter := bson.M{currentKey: true}
	result := m.mongoDb.Collection(sessionCollectionName).FindOne(ctx, filter)
	if result.Err() != nil {
		klog.Warning("failed to find current session. use default.", result.Err())
		return defaultSession
	}

	var currentSession currentSessionModel
	if err := result.Decode(&currentSession); err != nil {
		klog.Warning("failed to decode current session. use default.", result.Err())
		return defaultSession
	}

	return currentSession.SessionID
}

func (m *Manager) SetCurrentSession(session string) error {
	ctx := context.Background()

	filter := bson.M{currentKey: true}
	replacement := bson.M{currentSessionId: session, currentKey: true}

	_ = m.mongoDb.Collection(sessionCollectionName).FindOneAndDelete(ctx, filter)
	_, err := m.mongoDb.Collection(sessionCollectionName).InsertOne(ctx, replacement)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) AllSession() (map[string][]llms.ChatMessage, error) {
	ctx := context.Background()
	session := make(map[string][]llms.ChatMessage)

	cursor, err := m.mongoDb.Collection(messageCollectionName).Find(ctx, bson.M{})
	if err != nil {
		return session, err
	}

	var _messages []chatMessageModel
	if err := cursor.All(ctx, &_messages); err != nil {
		return session, err
	}
	for _, msg := range _messages {
		m := llms.ChatMessageModel{}
		if err := json.Unmarshal([]byte(msg.History), &m); err != nil {
			return session, err
		}
		if _, ok := session[msg.SessionID]; !ok {
			session[msg.SessionID] = []llms.ChatMessage{}
		}
		session[msg.SessionID] = append(session[msg.SessionID], m.ToChatMessage())
	}

	return session, nil
}
