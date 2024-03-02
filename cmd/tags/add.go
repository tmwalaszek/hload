package tags

import (
	"fmt"
	"os"
	"strings"

	"github.com/tmwalaszek/hload/cmd/cliio"
	"github.com/tmwalaszek/hload/model"
	"github.com/tmwalaszek/hload/storage"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type AddOptions struct {
	cliio.IO

	UUID string

	TagNames []string
}

func (o *AddOptions) Run() {
	s, err := storage.NewStorage(viper.GetString("db"))
	if err != nil {
		fmt.Fprintf(o.Err, "Can't create storage handler: %v", err)
		os.Exit(1)
	}

	loaderTags := make([]*model.LoaderTag, 0)
	for _, tag := range o.TagNames {
		tags := strings.Split(tag, ":")
		if len(tags) == 2 {
			loaderTags = append(loaderTags, &model.LoaderTag{
				Key:   tags[0],
				Value: tags[1],
			})
		}
	}

	err = s.InsertLoaderConfigurationTags(o.UUID, loaderTags)
	if err != nil {
		fmt.Fprintf(o.Err, "Error while inserting tags: %v", err)
		os.Exit(1)
	}

	fmt.Fprintf(o.Out, "Successfully added tags to loader %s", o.UUID)
}

func NewTagsAddCmd(cliIO cliio.IO) *cobra.Command {
	opts := AddOptions{
		IO: cliIO,
	}

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a tag to a loader configuration",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Run()
		},
	}

	cmd.Flags().StringVarP(&opts.UUID, "uuid", "u", "", "Loader configuration UUID")
	cmd.Flags().StringArrayVarP(&opts.TagNames, "tag", "t", []string{}, "Tag names pairs")

	_ = cmd.MarkFlagRequired("uuid")
	_ = cmd.MarkFlagRequired("tag")

	return cmd
}
