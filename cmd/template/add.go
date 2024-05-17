package template

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-sqlite3"
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
		fmt.Fprintf(o.Err, "Error: %v", err)
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
			fmt.Fprintf(o.Err, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		r = o.In
	}

	b, err := io.ReadAll(r)
	if err != nil {
		fmt.Fprintf(o.Err, "Error: %v\n", err)
		os.Exit(1)
	}

	err = o.storage.InsertTemplate(o.TemplateName, string(b))
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) {
				fmt.Fprintf(o.Err, "Template %s already exists\n", o.TemplateName)
				os.Exit(1)
			}
		}
		fmt.Fprintf(o.Err, "Error: %v\n", err)
		os.Exit(1)
	}

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

	cmd.Flags().StringVarP(&opts.TemplateName, "name", "n", "", "Template name")
	cmd.Flags().StringVarP(&opts.FileName, "file", "f", "", "Load template from file")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}
