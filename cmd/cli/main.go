// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/coding-hui/ai-terminal/internal/cli"
)

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	command := cli.NewDefaultAICommand()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
