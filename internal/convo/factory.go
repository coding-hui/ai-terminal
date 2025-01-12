package convo

import (
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
	// Type unique type of the history store
	Type() string
	// Create relevant history store by type
	Create(options options.Config) (Store, error)
}

func GetConversationStore(cfg options.Config) (Store, error) {
	dsType := cfg.DataStore.Type
	if store, ok := conversationStores[dsType]; ok {
		return store, nil
	}
	if factory, ok := conversationStoreFactories[dsType]; ok {
		if store, err := factory.Create(cfg); err != nil {
			return nil, errbook.Wrap("Failed to create chat history store: "+dsType, err)
		} else {
			lock.TryLock()
			defer lock.Unlock()
			conversationStores[dsType] = store
			return store, nil
		}
	}
	return nil, fmt.Errorf("chat history store %s is not supported", dsType)
}

func RegisterConversationStore(factory Factory) {
	conversationStoreFactories[factory.Type()] = factory
}
