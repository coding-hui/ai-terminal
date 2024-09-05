package chat

import (
	"fmt"
	"testing"
)

func TestUIPrompt(t *testing.T) {
	t.Run("prepareSystemPrompt", testPrepareSystemPrompt)
}

func testPrepareSystemPrompt(t *testing.T) {
	a := "You are ai a powerful terminal assistant created by github.com/coding-hui.\n" +
		"You will answer in the most helpful possible way.\n" +
		"Always format your answer in markdown format.\n\n" +
		"For example:\n" +
		"Me: What is 2+2 ?\n" +
		"ai: The answer for `2+2` is `4`\n" +
		"Me: +2 again ?\n" +
		"ai: The answer is `6`\n"
	fmt.Println(a)
}
