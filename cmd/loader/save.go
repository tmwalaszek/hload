package loader

import (
	"fmt"
	"log"
	"os"

	"github.com/tmwalaszek/hload/cmd/cliio"
	"github.com/tmwalaszek/hload/model"
	"github.com/tmwalaszek/hload/storage"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type SaveOptions struct {
	cliio.IO
}

func (o *SaveOptions) Run() {
	s, err := storage.NewStorage(viper.GetString("db"))
	if err != nil {
		fmt.Fprintf(o.Err, "Error opening database file: %v", err)
		os.Exit(1)
	}

	viper.SetConfigFile(viper.GetString("file"))
	if err := viper.MergeInConfig(); err != nil {
		fmt.Fprintf(o.Err, "Error on loading benchmark configuration: %v\n", err)
		os.Exit(1)
	}

	host := viper.GetString("url")
	if host == "" {
		fmt.Fprint(o.Err, "The url option has to provided")
		os.Exit(1)
	}

	headers := make(model.Headers)
	params := make(model.Parameters, 0)

	for _, header := range viper.GetStringSlice("headers") {
		err := headers.Set(header)
		if err != nil {
			fmt.Fprintf(o.Err, "Error setting header: %v", err)
			os.Exit(1)
		}
	}

	for _, param := range viper.GetStringSlice("parameters") {
		err := params.Set(param)
		if err != nil {
			fmt.Fprintf(o.Err, "Error setting parameters: %v", err)
			os.Exit(1)
		}
	}

	opts := &model.Loader{
		URL:              host,
		Name:             viper.GetString("name"),
		Description:      viper.GetString("description"),
		Method:           viper.GetString("method"),
		SkipVerify:       viper.GetBool("insecure"),
		HTTPEngine:       viper.GetString("engine"),
		CA:               []byte(viper.GetString("ca")),
		Cert:             []byte(viper.GetString("cert")),
		Key:              []byte(viper.GetString("key")),
		Body:             []byte(viper.GetString("body")),
		BenchmarkTimeout: viper.GetDuration("benchmark_timeout"),
		LoaderReqDetails: model.LoaderReqDetails{
			ReqCount:     viper.GetInt("requests"),
			AbortAfter:   viper.GetInt("abort"),
			Connections:  viper.GetInt("connections"),
			Duration:     viper.GetDuration("duration"),
			KeepAlive:    viper.GetDuration("keep_alive"),
			RequestDelay: viper.GetDuration("request_delay"),
			ReadTimeout:  viper.GetDuration("read_timeout"),
			WriteTimeout: viper.GetDuration("write_timeout"),
			Timeout:      viper.GetDuration("timeout"),
			RateLimit:    viper.GetInt("rate_limit"),
		},
		Headers:    headers,
		Parameters: params,
	}

	id, err := s.InsertLoaderConfiguration(opts)
	if err != nil {
		fmt.Fprintf(o.Err, "Error on inserting loader opts: %v", err)
		os.Exit(1)
	}

	fmt.Fprintf(o.Out, "Loader configuration added successfuly with id: %s", id)
}

func NewLoaderSaveCmd(cliIO cliio.IO) *cobra.Command {
	opts := SaveOptions{
		IO: cliIO,
	}

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save loader configuration",
		Long:  "Save loader configuration. Loader configuration has to be provided in the yaml file",
		Run: func(cmd *cobra.Command, args []string) {
			err := viper.BindPFlags(cmd.Flags())
			if err != nil {
				log.Fatal(err)
			}

			opts.Run()
		},
	}

	cmd.Flags().StringP("file", "f", "", "Loader configuration yaml file")

	err := cmd.MarkFlagRequired("file")
	if err != nil {
		log.Fatal(err)
	}

	return cmd
}
