package session

import (
	"fmt"
	"sync"

	"k8s.io/klog/v2"

	"github.com/coding-hui/ai-terminal/internal/options"
)

var (
	historyStoreFactories = make(map[string]Factory)
	historyStores         = make(map[string]History)

	lock = sync.Mutex{}
)

type Factory interface {
	// Type unique type of the history store
	Type() string
	// Create relevant history store by type
	Create(options options.Config, chatEngineMode string) (History, error)
}

func GetHistoryStore(cfg options.Config, chatEngineMode string) (History, error) {
	dsType := cfg.DataStore.Type
	if store, ok := historyStores[chatEngineMode+dsType]; ok {
		return store, nil
	}
	if factory, ok := historyStoreFactories[dsType]; ok {
		if store, err := factory.Create(cfg, chatEngineMode); err != nil {
			klog.Errorf("failed to create chat history store %s: %v", dsType, err)
		} else {
			lock.TryLock()
			defer lock.Unlock()
			historyStores[chatEngineMode+dsType] = store
			return store, nil
		}
	}
	return nil, fmt.Errorf("chat history store %s is not supported", dsType)
}

func RegisterHistoryStore(factory Factory) {
	historyStoreFactories[factory.Type()] = factory
}
