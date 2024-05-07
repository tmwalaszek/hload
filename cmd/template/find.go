package template

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tmwalaszek/hload/cmd/cliio"
	"github.com/tmwalaszek/hload/storage"
	"github.com/tmwalaszek/hload/templates"
)

type FindOptions struct {
	cliio.IO
	storage *storage.Storage

	Name  string
	Limit int
	Full  bool
	Base  bool
}

func (o *FindOptions) Complete() {
	s, err := storage.NewStorage(viper.GetString("db"))
	if err != nil {
		fmt.Fprintf(o.Err, "Can't create storage handler: %v", err)
		os.Exit(1)
	}

	o.storage = s
}

func (o *FindOptions) Run() {
	loc, err := time.LoadLocation("Local")
	if err != nil {
		fmt.Fprintf(o.Err, "Can't load time location: %v", err)
		os.Exit(1)
	}

	if o.Base {
		_, err := io.WriteString(o.Out, templates.ListTemplate)
		if err != nil {
			fmt.Fprintf(o.Err, "Can't write template: %v", err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	if o.Name != "" {
		temp, err := o.storage.GetTemplateByName(o.Name)
		if err != nil {
			fmt.Fprintf(o.Err, "Can't find template %s: %v", o.Name, err)
			os.Exit(1)
		}

		if o.Full {
			_, err = io.WriteString(o.Out, temp.Content)
			if err != nil {
				fmt.Fprintf(o.Err, "Can't write template: %v", err)
				os.Exit(1)
			}

			os.Exit(0)
		}

		fmt.Fprintf(o.Out, "Found template %s (last update time %v)\n", o.Name, temp.UpdateDate.In(loc))
		os.Exit(0)
	}

	templs, err := o.storage.GetTemplates(o.Limit)
	if err != nil {
		fmt.Fprintf(o.Err, "Can't find templates: %v", err)
		os.Exit(1)
	}

	if len(templs) > 0 {
		fmt.Fprintf(o.Out, "Found %d template(s)\n", len(templs))

		for _, t := range templs {
			fmt.Fprintf(o.Out, "  - %s (%v)\n", t.Name, t.UpdateDate.In(loc))
		}
		os.Exit(0)
	}

	fmt.Fprintf(o.Out, "No templates found\n")
}

func NewTemplateFindCmd(cliIO cliio.IO) *cobra.Command {
	opts := FindOptions{
		IO: cliIO,
	}

	cmd := &cobra.Command{
		Use:   "find",
		Short: "Find a template",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Complete()
			opts.Run()
		},
	}

	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "Template name")
	cmd.Flags().IntVarP(&opts.Limit, "limit", "l", 10, "Limit the number of results")
	cmd.Flags().BoolVarP(&opts.Full, "full", "f", false, "Show full template - with content")
	cmd.Flags().BoolVarP(&opts.Base, "base", "b", false, "Show the base template only. It will return the content only")

	return cmd
}
