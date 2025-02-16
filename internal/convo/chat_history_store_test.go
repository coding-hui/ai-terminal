package convo

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	convoID := NewConversationID()

	t.Run("read non-existent", func(t *testing.T) {
		store := NewSimpleChatHistoryStore(t.TempDir())
		err := store.load("super-fake")
		require.ErrorIs(t, err, os.ErrNotExist)
		defer func() {
			_ = store.InvalidateMessages(context.Background(), convoID)
		}()
	})

	t.Run("set messages", func(t *testing.T) {
		ctx := context.Background()
		store := NewSimpleChatHistoryStore(t.TempDir())
		_ = store.InvalidateMessages(context.Background(), convoID)
		require.NoError(t, store.AddUserMessage(ctx, convoID, "hello"))
		require.NoError(t, store.AddAIMessage(ctx, convoID, "hi"))
		require.NoError(t, store.AddAIMessage(ctx, convoID, "bye"))

		// Messages should be in memory but not persisted
		messages, err := store.Messages(ctx, convoID)
		require.NoError(t, err)
		require.Equal(t, 3, len(messages))

		// After persist, messages should be saved
		require.NoError(t, store.PersistentMessages(ctx, convoID))
		persistedMessages, err := store.Messages(ctx, convoID)
		require.NoError(t, err)
		require.Equal(t, messages, persistedMessages)

		defer func() {
			_ = store.InvalidateMessages(context.Background(), convoID)
		}()
	})

	t.Run("delete", func(t *testing.T) {
		store := NewSimpleChatHistoryStore(t.TempDir())
		require.NoError(t, store.PersistentMessages(context.Background(), convoID))
		require.NoError(t, store.InvalidateMessages(context.Background(), convoID))
		require.ErrorIs(t, store.load(convoID), os.ErrNotExist)
		defer func() {
			_ = store.InvalidateMessages(context.Background(), convoID)
		}()
	})

	t.Run("concurrent access", func(t *testing.T) {
		ctx := context.Background()
		store := NewSimpleChatHistoryStore(t.TempDir())
		convoID := NewConversationID()

		var wg sync.WaitGroup
		const numWorkers = 100
		const numMessages = 100

		// Concurrent writers
		wg.Add(numWorkers)
		for i := 0; i < numWorkers; i++ {
			go func(i int) {
				defer wg.Done()
				for j := 0; j < numMessages; j++ {
					if i%2 == 0 {
						require.NoError(t, store.AddUserMessage(ctx, convoID, fmt.Sprintf("user %d-%d", i, j)))
					} else {
						require.NoError(t, store.AddAIMessage(ctx, convoID, fmt.Sprintf("ai %d-%d", i, j)))
					}
					time.Sleep(time.Millisecond) // Add some delay to increase contention
				}
			}(i)
		}

		// Concurrent readers
		wg.Add(numWorkers)
		for i := 0; i < numWorkers; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < numMessages; j++ {
					msgs, err := store.Messages(ctx, convoID)
					require.NoError(t, err)
					require.True(t, len(msgs) <= numWorkers*numMessages)
					time.Sleep(time.Millisecond)
				}
			}()
		}

		wg.Wait()

		// Verify final message count
		msgs, err := store.Messages(ctx, convoID)
		require.NoError(t, err)
		require.Equal(t, numWorkers*numMessages, len(msgs))

		// Verify persistence
		require.NoError(t, store.PersistentMessages(ctx, convoID))
		persistedMsgs, err := store.Messages(ctx, convoID)
		require.NoError(t, err)
		require.Equal(t, msgs, persistedMsgs)

		defer func() {
			_ = store.InvalidateMessages(context.Background(), convoID)
		}()
	})
}
