package tags

import (
	"github.com/tmwalaszek/hload/cmd/cliio"

	"github.com/spf13/cobra"
)

func NewTagsCmd(cliIO cliio.IO) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "Manage tags",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Usage()
		},
	}

	cmd.AddCommand(NewTagsFindCmd(cliIO))
	cmd.AddCommand(NewTagsAddCmd(cliIO))
	cmd.AddCommand(NewTagsDelCmd(cliIO))
	return cmd
}
