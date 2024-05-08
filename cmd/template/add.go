package template

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tmwalaszek/hload/cmd/cliio"
	"github.com/tmwalaszek/hload/storage"
)

type AddOptions struct {
	cliio.IO
	storage *storage.Storage

	TemplateName string
	FileName     string
}

func (o *AddOptions) Complete() {
	s, err := storage.NewStorage(viper.GetString("db"))
	if err != nil {
		fmt.Fprintf(o.Err, "Can't create storage handler: %v", err)
		os.Exit(1)
	}

	o.storage = s
}

func (o *AddOptions) Run() {
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

	err = o.storage.InsertTemplate(o.TemplateName, string(b))
	if err != nil {
		fmt.Fprintf(o.Err, "Can't insert template: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "Added template %s\n", o.TemplateName)
	os.Exit(0)
}

func NewTemplateAddCmd(cliIO cliio.IO) *cobra.Command {
	opts := AddOptions{
		IO: cliIO,
	}

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a template",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Complete()
			opts.Run()
		},
	}

	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		if err != nil {
			log.Fatalf("Can't bind flags: %v", err)
		}
	}

	cmd.Flags().StringVarP(&opts.TemplateName, "name", "n", "", "Template name")
	cmd.Flags().StringVarP(&opts.FileName, "file", "f", "", "Load template from file")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}
