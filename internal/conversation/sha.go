package conversation

import (
	"crypto/rand"
	"crypto/sha1" //nolint: gosec
	"fmt"
	"regexp"
)

const (
	Sha1short         = 7
	Sha1minLen        = 4
	Sha1ReadBlockSize = 4096
)

var sha1reg = regexp.MustCompile(`\b[0-9a-f]{40}\b`)

func NewConversationID() string {
	b := make([]byte, Sha1ReadBlockSize)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", sha1.Sum(b)) //nolint: gosec
}
