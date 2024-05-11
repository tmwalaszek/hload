package template

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tmwalaszek/hload/cmd/cliio"
	"github.com/tmwalaszek/hload/storage"
)

type ChangeOptions struct {
	cliio.IO

	storage *storage.Storage

	TemplateName string
	FileName     string
}

func (o *ChangeOptions) Complete() {
	s, err := storage.NewStorage(viper.GetString("db"))
	if err != nil {
		fmt.Fprintf(o.Err, "Can't create storage handler: %v", err)
		os.Exit(1)
	}

	o.storage = s
}

func (o *ChangeOptions) Run() {
	var r io.Reader
	var err error

	if o.FileName != "" {
		r, err = os.Open(o.FileName)
		if err != nil {
			fmt.Fprintf(o.Err, "Can't open file %s: %v\n", o.FileName, err)
			os.Exit(1)
		}
	} else {
		r = o.In
	}

	b, err := io.ReadAll(r)
	if err != nil {
		fmt.Fprintf(o.Err, "Can't read file: %v\n", err)
		os.Exit(1)
	}

	err = o.storage.UpdateTemplate(o.TemplateName, string(b))
	if err != nil {
		fmt.Fprintf(o.Err, "Can't update template: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "Template %s updated\n", o.TemplateName)
	os.Exit(0)
}

func NewTemplateUpdateCmd(cliIO cliio.IO) *cobra.Command {
	opts := ChangeOptions{
		IO: cliIO,
	}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a template",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Complete()
			opts.Run()
		},
	}

	cmd.Flags().StringVarP(&opts.TemplateName, "name", "n", "", "Template name")
	cmd.Flags().StringVarP(&opts.FileName, "file", "f", "", "Load template from file")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}
