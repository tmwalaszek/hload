package loader

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/tmwalaszek/hload/cmd/cliio"
	"github.com/tmwalaszek/hload/loader"
	"github.com/tmwalaszek/hload/model"
	"github.com/tmwalaszek/hload/storage"
	"github.com/tmwalaszek/hload/templates"

	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	DefaultConnections  = 10
	DefaultRequestCount = 1000
)

type RunOptions struct {
	Conf    *model.Loader
	Storage *storage.Storage

	Config             string
	SummaryDescription string
	LoaderConfigName   string

	Engine Engine

	Save                   bool
	SaveRequests           bool
	SaveAggregatedRequests bool

	ShowFullStats       bool
	ShowAggregatedStats bool

	// If set to true then it runs from start subcommand
	Start bool

	UUID string

	render *templates.RenderTemplate

	cliio.IO
}

type EngineType string

const (
	EngineHTTP     EngineType = "http"
	EngineFastHTTP EngineType = "fast_http"
)

type Engine struct {
	Value EngineType
}

func (e *Engine) String() string {
	if e.Value == "" {
		return string(EngineFastHTTP)
	}

	return string(e.Value)
}

func (e *Engine) Set(value string) error {
	switch v := value; v {
	case "http":
		e.Value = EngineHTTP
	case "fast_http":
		e.Value = EngineFastHTTP
	default:
		return fmt.Errorf("unknwon value: %s", v)
	}
	return nil
}

func (e *Engine) Type() string {
	return "engine"
}

func (o *RunOptions) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
		log.Print("Received signal and will stop benchmark")
		cancel()

	}()

	progressChan := make(chan struct{})
	var l *loader.Loader
	var err error

	if o.Conf.ReqCount != 0 {
		l, err = loader.NewLoaderProgress(o.Conf, progressChan)
		if err != nil {
			log.Fatalf("Could not create loader: %v", err)
		}
	} else {
		l, err = loader.NewLoader(o.Conf)
		if err != nil {
			log.Fatalf("Could not create loader: %v", err)
		}
	}

	var newLoaderUUID string
	if o.Save && o.UUID == "" {
		newLoaderUUID, err = o.Storage.InsertLoaderConfiguration(o.Conf)
		if err != nil {
			log.Fatalf("Could not save loader configuration: %v", err)
		}
	}

	printLoaderDescription(o)
	fmt.Printf("\n\n")
	var summary *model.Summary

	pw := progress.NewWriter()
	pw.SetAutoStop(true)

	trackerDuration := &progress.Tracker{
		Message: "Duration",
		Total:   100,
		Units:   progress.UnitsDefault,
	}

	if o.Conf.ReqCount != 0 {
		trackerReqCount := &progress.Tracker{
			Message: "Requests progress",
			Total:   int64(o.Conf.ReqCount),
			Units:   progress.UnitsDefault,
		}

		pw.AppendTracker(trackerReqCount)
		go tickReqCount(progressChan, trackerReqCount)
	} else {
		pw.AppendTracker(trackerDuration)
		go tickDuration(o.Conf.Duration, trackerDuration)
	}

	pwFinish := make(chan struct{})
	go func() {
		pw.Render()
		pwFinish <- struct{}{}
	}()

	summary, err = l.Do(ctx)
	if err != nil {
		log.Fatalf("Could not run loader: %v", err)
	}

	pw.Stop()
	<-pwFinish

	var loaderUUID string
	if o.Save {
		if o.UUID == "" {
			loaderUUID = newLoaderUUID
		} else {
			loaderUUID = o.UUID
		}

		_, err = o.Storage.InsertSummary(loaderUUID, summary, o.SaveRequests, o.SaveAggregatedRequests)
		if err != nil {
			log.Fatalf("Error saving summary: %v", err)
		}

	}

	fmt.Fprintf(o.Out, "\n")
	o.printSummary(summary)

	if o.Save && !o.Start {
		fmt.Fprintf(o.Out, "\n")
		fmt.Fprintf(o.Out, "New loader configuration saved: %s\n", loaderUUID)
	}

	if o.Save && o.Start {
		fmt.Fprintf(o.Out, "\n")
		fmt.Fprintf(o.Out, "New summary saved for %s loader\n", loaderUUID)
	}
}

func tickDuration(duration time.Duration, tracker *progress.Tracker) {
	onePercent := float64(duration) * 0.01
	t := time.Tick(time.Duration(onePercent))

	for !tracker.IsDone() {
		<-t
		tracker.Increment(1)
	}
}

func tickReqCount(progressChan chan struct{}, tracker *progress.Tracker) {
	for !tracker.IsDone() {
		<-progressChan
		tracker.Increment(1)
	}
}

// CompleteDB uses information saved in the database
func (o *RunOptions) CompleteDB() {
	var err error

	r, err := templates.NewRenderTemplate(viper.GetString("template"), viper.GetString("db"))
	if err != nil {
		fmt.Fprintf(o.Err, "Can't create render template: %v", err)
		os.Exit(1)
	}

	o.render = r

	o.Storage, err = storage.NewStorage(viper.GetString("db"))
	if err != nil {
		fmt.Fprintf(o.Err, "Can't create storage handler: %v", err)
		os.Exit(1)
	}

	o.UUID = viper.GetString("uuid")

	if o.UUID != "" {
		o.Conf, err = o.Storage.GetLoaderByID(o.UUID)
		if err != nil {
			fmt.Fprintf(o.Err, "Error fetching loader configuration from the database: %v", err)
			os.Exit(1)
		}
	}

	o.Save = viper.GetBool("save")
	o.SaveRequests = o.Conf.GatherFullRequestsStats
	o.SaveAggregatedRequests = o.Conf.GatherAggregateRequestsStats
}

// Complete method setting the configuration and merging the configuration file with command line options
func (o *RunOptions) Complete() {
	var err error

	if viper.GetString("name") == "" {
		o.LoaderConfigName = fmt.Sprintf("Configuration %v", time.Now().Format("Mon, 02 Jan 2006 15:04:05.000"))
	}

	if viper.GetString("loader_config") != "" {
		viper.SetConfigFile(viper.GetString("loader_config"))

		if err := viper.MergeInConfig(); err != nil {
			fmt.Fprintf(o.Err, "Error on loading benchmark configuration %s: %v\n", viper.GetString("loader_config"), err)
			os.Exit(1)
		}
	}

	// We need to have storage when id is not zero or save if true
	if viper.GetBool("save") {
		o.Storage, err = storage.NewStorage(viper.GetString("db"))
		if err != nil {
			fmt.Fprintf(o.Err, "Can't create storage handler: %v", err)
			os.Exit(1)
		}
	}

	o.Save = viper.GetBool("save")
	o.SaveRequests = viper.GetBool("save-requests-stats")
	o.SaveAggregatedRequests = viper.GetBool("save-aggregate-requests-stats")
	o.ShowFullStats = viper.GetBool("show-requests-stats")
	o.ShowAggregatedStats = viper.GetBool("show-aggregate-requests-stats")

	headers := make(model.Headers)
	params := make(model.Parameters, 0)

	for _, value := range viper.GetStringSlice("header") {
		err := headers.Set(value)
		if err != nil {
			fmt.Fprintf(o.Err, "Error setting header %s: %v", value, err)
			os.Exit(1)
		}
	}

	if viper.GetString("cookie") != "" {
		cookieHeader := fmt.Sprintf("Cookie: %s", viper.GetString("cookie"))
		err = headers.Set(cookieHeader)
		if err != nil {
			fmt.Fprintf(o.Err, "Error setting cookie header: %v", err)
			os.Exit(1)
		}
	}

	for _, value := range viper.GetStringSlice("parameter") {
		err := params.Set(value)
		if err != nil {
			fmt.Fprintf(o.Err, "Error setting parameter %s: %v", value, err)
			os.Exit(1)
		}
	}

	var body, caBody, certBody, keyBody []byte
	if viper.GetString("ca") != "" {
		caBody, err = os.ReadFile(viper.GetString("ca"))
		if err != nil {
			fmt.Fprintf(o.Err, "Could not read CA file: %v", err)
			os.Exit(1)
		}
	}

	if viper.GetString("cert") != "" {
		certBody, err = os.ReadFile(viper.GetString("cert"))
		if err != nil {
			fmt.Fprintf(o.Err, "Could not read Cert file: %v", err)
			os.Exit(1)
		}
	}

	if viper.GetString("key") != "" {
		keyBody, err = os.ReadFile("key")
		if err != nil {
			fmt.Fprintf(o.Err, "Could not read Key file: %v", err)
			os.Exit(1)
		}
	}

	if viper.GetString("body") != "" {
		body, err = os.ReadFile(viper.GetString("body"))
		if err != nil {
			fmt.Fprintf(o.Err, "Count not read body file: %v", err)
			os.Exit(1)
		}
	}

	requestCount := viper.GetInt("requests")
	duration := viper.GetDuration("duration")

	if requestCount == 0 && duration == 0 {
		requestCount = DefaultRequestCount
	}

	// If we have duration set to something, set the requestsCount to 0
	if duration != 0 {
		requestCount = 0
	}

	method := viper.GetString("method")
	if method == "" {
		method = http.MethodGet
	}

	connections := viper.GetInt("connections")
	if connections == 0 {
		connections = DefaultConnections
	}

	o.Conf = &model.Loader{
		URL:                          viper.GetString("host"),
		Name:                         o.LoaderConfigName,
		Description:                  viper.GetString("description"),
		Method:                       method,
		SkipVerify:                   viper.GetBool("insecure"),
		HTTPEngine:                   o.Engine.String(),
		CA:                           caBody,
		Cert:                         certBody,
		Key:                          keyBody,
		Body:                         body,
		BenchmarkTimeout:             viper.GetDuration("benchmark-timeout"),
		AggregateWindow:              viper.GetDuration("aggregate-window"),
		GatherFullRequestsStats:      o.SaveRequests || o.ShowFullStats,
		GatherAggregateRequestsStats: o.SaveAggregatedRequests || o.ShowAggregatedStats,
		LoaderReqDetails: model.LoaderReqDetails{
			ReqCount:     requestCount,
			AbortAfter:   viper.GetInt("abort"),
			Connections:  connections,
			Duration:     duration,
			KeepAlive:    viper.GetDuration("keep_alive"),
			RequestDelay: viper.GetDuration("request_delay"),
			ReadTimeout:  viper.GetDuration("read_timeout"),
			WriteTimeout: viper.GetDuration("write_timeout"),
			Timeout:      viper.GetDuration("timeout"),
			RateLimit:    viper.GetInt("rate-limit"),
		},
		Headers:    headers,
		Parameters: params,
	}

	if viper.GetString("save-loader") != "" {
		file, err := json.MarshalIndent(o.Conf, "", " ")
		if err != nil {
			log.Fatalf("Could not save loader configuration: %v", err)
		}

		err = os.WriteFile(viper.GetString("save-loader"), file, 0644)
		if err != nil {
			log.Fatalf("Could not save loader configuration: %v", err)
		}
	}
}

func NewLoaderRunCmd(cliIO cliio.IO) *cobra.Command {
	opts := RunOptions{
		IO: cliIO,
	}

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run HTTP loader",
		Run: func(cmd *cobra.Command, args []string) {
			err := viper.BindPFlags(cmd.Flags())
			if err != nil {
				fmt.Fprintf(cliIO.Err, "Could not bind flags: %v", err)
				os.Exit(1)
			}

			opts.Complete()
			opts.CompleteDB()
			opts.Run()
		},
	}

	cmd.Flags().StringVarP(&opts.Config, "loader_config", "f", "", "Loader configuration file")
	cmd.Flags().String("name", "", "Loader configuration name")
	cmd.Flags().String("description", "Default loader description", "Loader description will be saved in the database")
	cmd.Flags().String("summary-description", "", "Custom summary description that will be saved in the database")
	cmd.Flags().String("host", "", "Host")
	cmd.Flags().StringP("method", "m", http.MethodGet, "HTTP Method")
	cmd.Flags().String("ca", "", "CA path")
	cmd.Flags().String("cert", "", "Cert path")
	cmd.Flags().String("key", "", "Key path")
	cmd.Flags().String("body", "", "Path to the body file")
	cmd.Flags().String("save-loader", "", "Save the loader configuration to a file (json)")
	cmd.Flags().StringP("cookie", "b", "", "Send the data in the HTTP Cookie header")
	cmd.Flags().Var(&opts.Engine, "engine", "HTTP library used: fast_http or net/http")

	cmd.Flags().BoolP("insecure", "i", false, "TLS Skip verify")
	cmd.Flags().BoolP("save", "s", false, "Save loader configuration and result")
	cmd.Flags().Bool("save-requests-stats", false, "Save all requests stats - HIGH MEMORY USAGE")
	cmd.Flags().Bool("save-aggregate-requests-stats", false, "Save aggregated requests stats")
	cmd.Flags().Bool("show-requests-stats", false, "Show all the gather requests stats")
	cmd.Flags().Bool("show-aggregate-requests-stats", false, "Show all the aggregated requests stats")

	cmd.Flags().DurationP("duration", "d", 0, "Loader duration")
	cmd.Flags().Duration("keep-alive", 0, "HTTP Keep Alive")
	cmd.Flags().DurationP("request-delay", "D", 0, "Request delay")
	cmd.Flags().Duration("read-timeout", 0, "Read Timeout")
	cmd.Flags().Duration("write-timeout", 0, "Write Timeout")
	cmd.Flags().Duration("timeout", 0, "Timeout used for net/http error")
	cmd.Flags().Duration("benchmark-timeout", 0, "Benchmark timeout when Requests count option is used")
	cmd.Flags().DurationP("aggregate-window", "A", 10*time.Second, "Aggregate results into window buckets")

	cmd.Flags().IntP("rate-limit", "L", 0, "Rate limit requests per second")
	cmd.Flags().IntP("requests", "r", DefaultRequestCount, "Requests count")
	cmd.Flags().IntP("connections", "c", DefaultConnections, "Concurrent connections")
	cmd.Flags().IntP("abort", "a", 0, "Number of connections after which benchmark will be aborted")

	cmd.Flags().StringSliceP("header", "H", nil, "Header, can be used multiple times")
	cmd.Flags().StringSliceP("parameter", "P", nil, "HTTP parameters, can be used multiple times")

	err := cmd.MarkFlagRequired("host")
	if err != nil {
		log.Fatal(err)
	}

	return cmd
}

func printLoaderDescription(opts *RunOptions) {
	l := list.NewWriter()
	l.AppendItem("Running loader...")
	l.AppendItem("Loader basic parameters:")
	l.Indent()
	l.AppendItem(fmt.Sprintf("Created at: %s", time.Now()))
	l.AppendItem(fmt.Sprintf("Target host: %s", opts.Conf.URL))
	l.AppendItem(fmt.Sprintf("Concurrent connections: %d", opts.Conf.Connections))
	if opts.Conf.ReqCount != 0 {
		l.AppendItem(fmt.Sprintf("Requests count: %d", opts.Conf.ReqCount))
	}
	if opts.Conf.Duration != 0 {
		l.AppendItem(fmt.Sprintf("Duration: %v", opts.Conf.Duration))
	}

	fmt.Fprintf(opts.Out, "%s\n", l.Render())
}

func (o *RunOptions) printSummary(summary *model.Summary) {
	b, err := o.render.RenderSummary(summary, o.ShowFullStats, o.ShowAggregatedStats)
	if err != nil {
		log.Fatalf("Failed to render summary template: %v", err)
	}

	fmt.Fprintf(o.Out, "%s\n", string(b))
}
