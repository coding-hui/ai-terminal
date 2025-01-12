package conversation

import (
	"encoding/gob"
	"fmt"
	"io"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
)

func encode(w io.Writer, messages *[]llms.ChatMessageModel) error {
	if err := gob.NewEncoder(w).Encode(messages); err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	return nil
}

func decode(r io.Reader, messages *[]llms.ChatMessageModel) error {
	if err := gob.NewDecoder(r).Decode(messages); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}
