package tags

import (
	"log"

	"github.com/spf13/viper"
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
		PreRun: func(cmd *cobra.Command, args []string) {
			err := viper.BindPFlags(cmd.Flags())
			if err != nil {
				log.Fatalf("Can't bind flags: %v", err)
			}
		},
	}

	cmd.AddCommand(NewTagsFindCmd(cliIO))
	cmd.AddCommand(NewTagsAddCmd(cliIO))
	cmd.AddCommand(NewTagsDelCmd(cliIO))
	cmd.AddCommand(NewTagsUpdateCommand(cliIO))
	return cmd
}
