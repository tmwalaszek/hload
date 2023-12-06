package loader

import (
	"github.com/tmwalaszek/hload/cmd/cliio"

	"github.com/spf13/cobra"
)

func NewLoaderCmd(cliIO cliio.IO) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "loader",
		Short: "Run HTTP loader",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Usage()
		},
	}

	cmd.AddCommand(NewLoaderRunCmd(cliIO))
	cmd.AddCommand(NewLoaderSaveCmd(cliIO))
	cmd.AddCommand(NewLoaderStartCmd(cliIO))
	cmd.AddCommand(NewLoaderDeleteCmd(cliIO))
	cmd.AddCommand(NewLoaderFindCmd(cliIO))

	return cmd
}
