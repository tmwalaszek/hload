package tags

import (
	"fmt"
	"os"
	"strings"

	"github.com/tmwalaszek/hload/cmd/cliio"
	"github.com/tmwalaszek/hload/cmd/common"
	"github.com/tmwalaszek/hload/model"
	"github.com/tmwalaszek/hload/storage"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type FindOptions struct {
	cliio.IO

	UUID string
	Name string

	Tags []*model.LoaderTag
}

func (o *FindOptions) Complete() {
	for _, tag := range viper.GetStringSlice("tag") {
		tags := strings.SplitN(tag, "=", 2)
		if len(tags) == 2 {
			o.Tags = append(o.Tags, &model.LoaderTag{
				Key:   tags[0],
				Value: tags[1],
			})
		}
	}
}

func (o *FindOptions) Run() {
	s, err := storage.NewStorage(viper.GetString("db"))
	if err != nil {
		fmt.Fprintf(o.Err, "Can't create storage handler: %v", err)
		os.Exit(1)
	}

	if o.UUID != "" {
		loaderTags, err := s.GetLoaderTags(o.UUID)
		if err != nil {
			fmt.Fprintf(o.Err, "Error while getting tags: %v", err)
			os.Exit(1)
		}

		output := common.WriteLoaderTags(o.UUID, loaderTags)
		fmt.Fprintf(o.Out, "%s\n", output)
		os.Exit(0)
	}

	if o.Name != "" {
		tags, err := s.GetLoaderTagsByKey(o.Name)
		if err != nil {
			fmt.Fprintf(o.Err, "Error while gettings tags: %v", err)
			os.Exit(1)
		}

		output := common.WriteLoaderTagsMap(tags)
		fmt.Fprintf(o.Out, "%s\n", output)
		os.Exit(0)
	}

	if len(o.Tags) > 0 {
		loaderConfs, err := s.GetLoaderByTags(o.Tags)
		if err != nil {
			fmt.Fprintf(o.Err, "Error while gettting loaders configuration: %v", err)
			os.Exit(1)
		}

		configs := make([]common.LoaderSummaries, 0)

		for _, loaderConf := range loaderConfs {
			l := common.LoaderSummaries{
				Loader: loaderConf,
			}

			configs = append(configs, l)
		}

		loaderConfiguration := common.Loaders{
			Loaders: configs,
			Short:   true,
		}

		fmt.Fprintf(o.Out, "%s\n", loaderConfiguration)
		os.Exit(0)
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
	// Tag pair key:value
	cmd.Flags().StringArrayP("tag", "t", []string{}, "Tag names pairs - key=value")

	return cmd
}
