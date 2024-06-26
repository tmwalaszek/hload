package tags

import (
	"fmt"
	"os"
	"strings"

	"github.com/tmwalaszek/hload/cmd/cliio"
	"github.com/tmwalaszek/hload/model"
	"github.com/tmwalaszek/hload/storage"
	"github.com/tmwalaszek/hload/templates"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type FindOptions struct {
	cliio.IO

	UUID string
	Name string

	Tags []*model.LoaderTag

	render *templates.RenderTemplate
}

func (o *FindOptions) Complete() {
	for _, tag := range viper.GetStringSlice("tag") {
		tags := strings.SplitN(tag, "=", 2)
		var key, value string
		if len(tags) == 2 {
			value = tags[1]
		}

		key = tags[0]
		o.Tags = append(o.Tags, &model.LoaderTag{
			Key:   key,
			Value: value,
		})
	}

	r, err := templates.NewRenderTemplate(viper.GetString("template"), viper.GetString("db"))
	if err != nil {
		fmt.Fprintf(o.Err, "Error: %v", err)
		os.Exit(1)
	}

	o.render = r
}

func (o *FindOptions) Run() {
	s, err := storage.NewStorage(viper.GetString("db"))
	if err != nil {
		fmt.Fprintf(o.Err, "Error: %v", err)
		os.Exit(1)
	}

	if o.UUID != "" {
		loaderTags, err := s.GetLoaderTags(o.UUID)
		if err != nil {
			fmt.Fprintf(o.Err, "Error: %v", err)
			os.Exit(1)
		}

		if len(loaderTags) > 0 {
			output, err := o.render.RenderTags(o.UUID, loaderTags)
			if err != nil {
				fmt.Fprintf(o.Err, "Error: %v", err)
				os.Exit(1)
			}

			fmt.Fprintf(o.Out, "%s\n", string(output))
			os.Exit(0)
		}
	}

	if o.Name != "" {
		tags, err := s.GetLoaderTagsByKey(o.Name)
		if err != nil {
			fmt.Fprintf(o.Err, "Error: %v", err)
			os.Exit(1)
		}

		output, err := o.render.RenderTagsMap(tags)
		if err != nil {
			fmt.Fprintf(o.Err, "Error: %v", err)
			os.Exit(1)
		}

		if len(output) > 0 {
			fmt.Fprintf(o.Out, "%s\n", output)
			os.Exit(0)
		}

		os.Exit(0)
	}

	if len(o.Tags) > 0 {
		loaderConfs, err := s.GetLoaderByTags(o.Tags)
		if err != nil {
			fmt.Fprintf(o.Err, "Error: %v", err)
			os.Exit(1)
		}

		configs := make([]templates.LoaderSummaries, 0)

		for _, loaderConf := range loaderConfs {
			l := templates.LoaderSummaries{
				Loader: loaderConf,
			}

			configs = append(configs, l)
		}

		loaderConfiguration := &templates.Loaders{
			Loaders: configs,
			Short:   true,
		}

		b, err := o.render.RenderOutput(loaderConfiguration)
		if err != nil {
			fmt.Fprintf(o.Err, "Error: %v", err)
			os.Exit(1)
		}

		if len(b) > 0 {
			fmt.Fprintf(o.Out, "%s\n", string(b))
			os.Exit(0)
		}

		os.Exit(1)
	}
}

func NewTagsFindCmd(cliIO cliio.IO) *cobra.Command {
	opts := FindOptions{
		IO: cliIO,
	}

	cmd := &cobra.Command{
		Use:   "find",
		Short: "Find tags",
		Run: func(cmd *cobra.Command, args []string) {
			err := viper.BindPFlags(cmd.Flags())
			if err != nil {
				fmt.Fprintf(cliIO.Err, "Could not bind flags: %v", err)
				os.Exit(1)
			}

			opts.Complete()
			opts.Run()
		},
	}

	cmd.Flags().StringVarP(&opts.UUID, "uuid", "u", "", "Loader configuration UUID from database")
	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "Tag name")
	cmd.Flags().StringArrayP("tag", "t", []string{}, "Tag names pairs - key=value")

	return cmd
}
