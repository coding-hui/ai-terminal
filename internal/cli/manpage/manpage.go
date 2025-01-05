// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package manpage

import (
	"fmt"
	"os"

	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/spf13/cobra"
)

// NewCmdManPage creates the `manpage` command.
func NewCmdManPage(rootCmd *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:                   "manpage",
		Short:                 "Generates manpages",
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		Hidden:                true,
		Args:                  cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			manPage, err := mcobra.NewManPage(1, rootCmd)
			if err != nil {
				//nolint:wrapcheck
				return err
			}
			_, err = fmt.Fprint(os.Stdout, manPage.Build(roff.NewDocument()))
			//nolint:wrapcheck
			return err
		},
	}
}
