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

type DeleteOptions struct {
	cliio.IO

	UUID string

	TagNames []string
}

func (o *DeleteOptions) Run() {
	s, err := storage.NewStorage(viper.GetString("db"))
	if err != nil {
		fmt.Fprintf(o.Err, "Can't create storage handler: %v", err)
		os.Exit(1)
	}

	loaderTags := make([]*model.LoaderTag, 0)
	for _, tag := range o.TagNames {
		tags := strings.SplitN(tag, "=", 2)
		if len(tags) == 2 {
			loaderTags = append(loaderTags, &model.LoaderTag{
				Key:   tags[0],
				Value: tags[1],
			})
		}
	}

	err = s.DeleteLoaderTag(o.UUID, loaderTags)
	if err != nil {
		fmt.Fprintf(o.Err, "Error while inserting tags: %v", err)
		os.Exit(1)
	}

	fmt.Fprintf(o.Out, "Successfully added tags to loader %s", o.UUID)
}

func NewTagsDelCmd(cliIO cliio.IO) *cobra.Command {
	opts := AddOptions{
		IO: cliIO,
	}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a tag from a loader configuration",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Run()
		},
	}

	cmd.Flags().StringVarP(&opts.UUID, "uuid", "u", "", "Loader configuration UUID")
	cmd.Flags().StringArrayVarP(&opts.TagNames, "tag", "t", []string{}, "Tag names pairs - key=value")

	cmd.MarkFlagRequired("uuid")
	cmd.MarkFlagRequired("tag")

	return cmd
}
