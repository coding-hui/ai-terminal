package coders

import (
	"testing"
)

func TestChooseExistingFence(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedOpen  string
		expectedClose string
	}{
		{
			name:          "backtick fence with SEARCH/REPLACE",
			input:         "```go\n<<<<<<< SEARCH\nsome code\n=======\nnew code\n>>>>>>> REPLACE\n```",
			expectedOpen:  "```",
			expectedClose: "```",
		},
		{
			name:          "code tag fence with SEARCH/REPLACE",
			input:         "<code>\n<<<<<<< SEARCH\nsome code\n=======\nnew code\n>>>>>>> REPLACE\n</code>",
			expectedOpen:  "<code>",
			expectedClose: "</code>",
		},
		{
			name:          "pre tag fence with SEARCH/REPLACE",
			input:         "<pre>\n<<<<<<< SEARCH\nsome code\n=======\nnew code\n>>>>>>> REPLACE\n</pre>",
			expectedOpen:  "<pre>",
			expectedClose: "</pre>",
		},
		{
			name:          "no matching fence",
			input:         "just some text\n<<<<<<< SEARCH\nsome code\n=======\nnew code\n>>>>>>> REPLACE\n",
			expectedOpen:  "```",
			expectedClose: "```",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, c := chooseExistingFence(tt.input)
			if o != tt.expectedOpen || c != tt.expectedClose {
				t.Errorf("chooseExistingFence() = (%q, %q), want (%q, %q)",
					o, c, tt.expectedOpen, tt.expectedClose)
			}
		})
	}
}
