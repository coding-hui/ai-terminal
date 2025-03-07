package convo

import (
	"context"
	"fmt"
	"sync"

	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/options"
)

var (
	conversationStoreFactories = make(map[string]Factory)
	conversationStores         = make(map[string]Store)

	lock = sync.Mutex{}
)

type Factory interface {
	// Type unique type of the convo store
	Type() string
	// Create relevant convo store by type
	Create(options *options.Config) (Store, error)
}

func GetConversationStore(cfg *options.Config) (Store, error) {
	dsType := cfg.DataStore.Type

	// Check if store already exists
	store, exists := conversationStores[dsType]
	if !exists {
		// Create new store if it doesn't exist
		factory, ok := conversationStoreFactories[dsType]
		if !ok {
			return nil, fmt.Errorf("chat convo store `%s` is not supported", dsType)
		}

		newStore, err := factory.Create(cfg)
		if err != nil {
			return nil, errbook.Wrap("Failed to create chat convo store: "+dsType, err)
		}

		lock.Lock()
		conversationStores[dsType] = newStore
		lock.Unlock()
		store = newStore
	}

	// Handle conversation ID if not provided
	if cfg.ConversationID == "" {
		// Try to get latest conversation
		latest, err := store.LatestConversation(context.Background())
		if err != nil {
			return nil, errbook.Wrap("Failed to get latest conversation", err)
		}

		if latest != nil && latest.ID != "" {
			cfg.ConversationID = latest.ID
		} else {
			// No conversations exist, generate new ID
			cfg.ConversationID = NewConversationID()
			//debug.Trace("conversation id not provided, generating new id `%s`", cfg.ConversationID)
		}
	}

	return store, nil
}

func RegisterConversationStore(factory Factory) {
	conversationStoreFactories[factory.Type()] = factory
}
