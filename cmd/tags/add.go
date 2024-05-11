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
		tags := strings.SplitN(tag, "=", 2)
		var key, value string
		if len(tags) == 1 {
			if tags[0] == "" {
				fmt.Fprintf(o.Err, "Empty tag name: %s", tag)
				os.Exit(1)
			}

			key = tags[0]
		} else {
			key = tags[0]
			value = tags[1]
		}

		loaderTags = append(loaderTags, &model.LoaderTag{
			Key:   key,
			Value: value,
		})
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

	//cmd.PreRun = func(cmd *cobra.Command, args []string) {
	//	err := viper.BindPFlags(cmd.Flags())
	//	if err != nil {
	//		log.Fatalf("Can't bind flags: %v", err)
	//	}
	//}

	cmd.Flags().StringVarP(&opts.UUID, "uuid", "u", "", "Loader configuration UUID")
	cmd.Flags().StringArrayVarP(&opts.TagNames, "tag", "t", []string{}, "Tag names pairs")

	_ = cmd.MarkFlagRequired("uuid")
	_ = cmd.MarkFlagRequired("tag")

	return cmd
}
