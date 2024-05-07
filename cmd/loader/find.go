package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/tmwalaszek/hload/cmd/cliio"
	"github.com/tmwalaszek/hload/model"
	"github.com/tmwalaszek/hload/storage"
	"github.com/tmwalaszek/hload/templates"
	"github.com/tmwalaszek/hload/time_formats"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type FindOptions struct {
	cliio.IO

	UUID string

	LoaderLimit  int
	SummaryLimit int
	FromEpoch    int64
	ToEpoch      int64

	Summary           bool
	ShowRequestsStats bool
	RangeFind         bool

	URL                string
	LoaderName         string
	LoaderDescription  string
	SummaryDescription string
	Output             string
	From               string
	To                 string

	db     *storage.Storage
	render *templates.RenderTemplate
}

func (o *FindOptions) Complete() {
	var err error

	s, err := storage.NewStorage(viper.GetString("db"))
	if err != nil {
		fmt.Fprintf(o.Err, "Can't create storage handler: %v", err)
		os.Exit(1)
	}

	o.db = s
	r, err := templates.NewRenderTemplate(viper.GetString("template"), viper.GetString("db"))
	if err != nil {
		fmt.Fprintf(o.Err, "Can't create render template: %v", err)
		os.Exit(1)
	}

	o.render = r

	// We don't support empty From and provided To atm
	if o.From != "" {
		o.FromEpoch, err = time_formats.TimeToEpoch(o.From)
		if err != nil {
			fmt.Fprintf(o.Err, "Error while parsing From date: %v", err)
			os.Exit(1)
		}

		if o.To != "" {
			o.ToEpoch, err = time_formats.TimeToEpoch(o.To)
			if err != nil {
				fmt.Fprintf(o.Err, "Error while parsing To date: %v", err)
				os.Exit(1)
			}
		} else {
			o.ToEpoch = time.Now().UTC().Unix()
		}

		o.RangeFind = true
	}
}

func (o *FindOptions) getSummaries(loaderUUID string) []*model.Summary {
	var summaries []*model.Summary
	var err error

	if o.Summary {
		if o.SummaryLimit != 0 {
			if o.ShowRequestsStats {
				summaries, err = o.db.GetSummaries(loaderUUID, storage.WithLimit(o.SummaryLimit), storage.WithRequests(), storage.WithFrom(o.FromEpoch), storage.WithTo(o.ToEpoch))
			} else {
				summaries, err = o.db.GetSummaries(loaderUUID, storage.WithLimit(o.SummaryLimit), storage.WithFrom(o.FromEpoch), storage.WithTo(o.ToEpoch))
			}
		}

		if err != nil {
			fmt.Fprintf(o.Err, "Error while getting summaries for loader configuration: %v", err)
			os.Exit(1)
		}
	}

	return summaries
}

func (o *FindOptions) Run() {
	loaders := make([]*model.Loader, 0)
	loaderSummary := make([]templates.LoaderSummaries, 0)
	loaderConfigurations := templates.Loaders{
		Short: false,
	}

	var err error

	if o.UUID != "" {
		loaderConf, err := o.db.GetLoaderByID(o.UUID)
		if err != nil {
			fmt.Fprintf(o.Err, "Error while gettting loaders configuration: %v", err)
			os.Exit(1)
		}

		if loaderConf == nil {
			fmt.Fprintf(o.Err, "Loader configuration %s not found\n", o.UUID)
			os.Exit(1)
		}

		summaries := o.getSummaries(loaderConf.UUID)
		loaderOpts := templates.LoaderSummaries{
			Loader:    loaderConf,
			Summaries: summaries,
		}
		loaderSummary = append(loaderSummary, loaderOpts)
	} else if o.LoaderDescription != "" {
		loaders, err = o.db.GetLoaderByDescription(o.LoaderDescription)
		if err != nil {
			fmt.Fprintf(o.Err, "Error getting loader configuration from the database: %v", err)
			os.Exit(1)
		}
	} else if o.RangeFind {
		loaders, err = o.db.GetLoadersByRange(o.FromEpoch, o.ToEpoch, o.LoaderLimit)
		if err != nil {
			fmt.Fprintf(o.Err, "Error getting loader configuration from the database: %v", err)
			os.Exit(1)
		}
	} else {
		loaders, err = o.db.GetLoaders(o.LoaderLimit)
		if err != nil {
			fmt.Fprintf(o.Err, "Error getting loader configuration from the database: %v", err)
			os.Exit(1)
		}
	}
	for _, loaderConf := range loaders {
		summaries := o.getSummaries(loaderConf.UUID)
		loaderOpts := templates.LoaderSummaries{
			Loader:    loaderConf,
			Summaries: summaries,
		}

		loaderSummary = append(loaderSummary, loaderOpts)
	}

	loaderConfigurations.Loaders = loaderSummary

	switch o.Output {
	case "json":
		output, err := json.MarshalIndent(loaderSummary, "", " ")
		if err != nil {
			fmt.Fprintf(o.Err, "Error while JSON marshaling output: %v", err)
			os.Exit(1)
		}

		fmt.Fprintf(o.Out, "%s\n", string(output))
	default:
		b, err := o.render.RenderOutput(&loaderConfigurations)
		if err != nil {
			fmt.Fprintf(o.Err, "Error while rendering output: %v", err)
		}
		fmt.Fprintf(o.Out, "%s", string(b))
	}
}

func NewLoaderFindCmd(cliIO cliio.IO) *cobra.Command {
	opts := FindOptions{
		IO: cliIO,
	}

	cmd := &cobra.Command{
		Use:   "find",
		Short: "Find loader configurations and results",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Complete()
			opts.Run()
		},
	}

	cmd.Flags().StringVarP(&opts.UUID, "uuid", "u", "", "Loader configuration UUID")
	cmd.Flags().StringVarP(&opts.URL, "url", "U", "", "Loader target URL")
	cmd.Flags().StringVarP(&opts.LoaderDescription, "description", "d", "", "Loader description")
	cmd.Flags().StringVarP(&opts.LoaderName, "name", "n", "", "Loader name")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "list", "Output")
	cmd.Flags().StringVarP(&opts.From, "from", "f", "", "From date")
	cmd.Flags().StringVarP(&opts.To, "to", "t", "", "To date")
	cmd.Flags().BoolVarP(&opts.Summary, "show-summary", "s", false, "Show the summaries of the loadedr configurations")
	cmd.Flags().IntVarP(&opts.LoaderLimit, "loader-limit", "l", 10, "Limit the number of loaders that matches the query")
	cmd.Flags().IntVarP(&opts.SummaryLimit, "summary-limit", "L", 5, "Limit the number of returned summaries")
	cmd.Flags().BoolVar(&opts.ShowRequestsStats, "show-request-stats", false, "Show requests stats - both full or aggregated")

	return cmd
}
