package template

import (
	"github.com/spf13/cobra"
	"github.com/tmwalaszek/hload/cmd/cliio"
)

func NewTemplateCmd(cliIO cliio.IO) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage templates",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Usage()
		},
	}

	cmd.AddCommand(NewTemplateAddCmd(cliIO))
	cmd.AddCommand(NewTemplateFindCmd(cliIO))
	cmd.AddCommand(NewTemplateUpdateCmd(cliIO))
	cmd.AddCommand(NewTemplateDeleteCmd(cliIO))
	return cmd
}
