package loader

import (
	"fmt"
	"log"
	"os"

	"github.com/tmwalaszek/hload/cmd/cliio"
	"github.com/tmwalaszek/hload/storage"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type DeleteOptions struct {
	cliio.IO

	UUID string
}

func (o *DeleteOptions) Run() {
	s, err := storage.NewStorage(viper.GetString("db"))
	if err != nil {
		fmt.Fprintf(o.Err, "Error on creating new storage: %v", err)
		os.Exit(1)
	}

	err = s.DeleteLoader(o.UUID)
	if err != nil {
		fmt.Fprintf(o.Err, "Error deleting loader %s: %v", o.UUID, err)
		os.Exit(1)
	}

	fmt.Fprintf(o.Out, "Successfully deleted loadeer %s", o.UUID)
}

func NewLoaderDeleteCmd(cliIO cliio.IO) *cobra.Command {
	opts := DeleteOptions{
		IO: cliIO,
	}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete loader configurations with all data associated",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Run()
		},
	}

	cmd.Flags().StringVarP(&opts.UUID, "uuid", "u", "", "Loader configuration UUID")
	err := cmd.MarkFlagRequired("uuid")
	if err != nil {
		log.Fatal(err)
	}

	return cmd
}
