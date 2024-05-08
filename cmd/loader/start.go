package loader

import (
	"fmt"
	"os"

	"github.com/tmwalaszek/hload/cmd/cliio"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewLoaderStartCmd(cliIO cliio.IO) *cobra.Command {
	opts := RunOptions{
		IO: cliIO,
	}

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a benchmark already saved in the database",
		Run: func(cmd *cobra.Command, args []string) {
			err := viper.BindPFlags(cmd.Flags())
			if err != nil {
				fmt.Fprintf(cliIO.Err, "Could not bind flags: %v", err)
				os.Exit(1)
			}

			opts.CompleteDB()
			opts.Start = true
			opts.Run()
		},
	}

	cmd.Flags().BoolP("save", "s", true, "Save the summary")
	cmd.Flags().StringP("uuid", "u", "", "Loader configuration UUID from database")

	_ = cmd.MarkFlagRequired("uuid")

	return cmd
}
