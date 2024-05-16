package tags

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tmwalaszek/hload/cmd/cliio"
	"github.com/tmwalaszek/hload/storage"
)

type UpdateOptions struct {
	cliio.IO

	s *storage.Storage

	UUID  string
	Key   string
	Value string
}

func (o *UpdateOptions) Complete() {
	s, err := storage.NewStorage(viper.GetString("db"))
	if err != nil {
		fmt.Fprintf(o.Err, "Can't create storage handler: %v", err)
		os.Exit(1)
	}

	o.s = s
}

func (o *UpdateOptions) Run() {
	err := o.s.UpdateLoaderTag(o.UUID, o.Key, o.Value)
	if err != nil {
		fmt.Fprintf(o.Err, "Can't update tag: %v", err)
		os.Exit(1)
	}

	fmt.Fprintf(o.Err, "Tag %s has been updated", o.Key)
	os.Exit(0)
}

func NewTagsUpdateCommand(cliIO cliio.IO) *cobra.Command {
	opts := UpdateOptions{
		IO: cliIO,
	}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a tag",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Complete()
			opts.Run()
		},
	}

	cmd.Flags().StringVarP(&opts.UUID, "uuid", "u", "", "tag UUID")
	cmd.Flags().StringVarP(&opts.Key, "key", "k", "", "tag key name")
	cmd.Flags().StringVarP(&opts.Value, "value", "v", "", "tag value")

	_ = cmd.MarkFlagRequired("uuid")
	_ = cmd.MarkFlagRequired("key")
	_ = cmd.MarkFlagRequired("value")

	return cmd
}
