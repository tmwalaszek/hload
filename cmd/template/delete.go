package template

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tmwalaszek/hload/cmd/cliio"
	"github.com/tmwalaszek/hload/storage"
)

type DeleteOptions struct {
	cliio.IO
	storage *storage.Storage

	TemplateName string
}

func (o *DeleteOptions) Complete() {
	s, err := storage.NewStorage(viper.GetString("db"))
	if err != nil {
		fmt.Fprintf(o.Err, "Can't create storage handler: %v", err)
		os.Exit(1)
	}

	o.storage = s
}

func (o *DeleteOptions) Run() {
	err := o.storage.DeleteTemplate(o.TemplateName)
	if err != nil {
		fmt.Fprintf(o.Err, "Can't delete template %s: %v", o.TemplateName, err)
		os.Exit(1)
	}

	fmt.Fprintf(o.Err, "Template %s deleted", o.TemplateName)
	os.Exit(0)
}

func NewTemplateDeleteCmd(cliIO cliio.IO) *cobra.Command {
	opts := DeleteOptions{
		IO: cliIO,
	}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a template",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Complete()
			opts.Run()
		},
	}

	cmd.Flags().StringVarP(&opts.TemplateName, "name", "n", "", "Template name")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}
