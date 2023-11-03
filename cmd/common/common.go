package common

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/tmwalaszek/hload/model"

	"code.cloudfoundry.org/bytefmt"
	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/jedib0t/go-pretty/v6/text"
)

type LoaderSummaries struct {
	Loader    *model.Loader    `json:"loader"`
	Summaries []*model.Summary `json:"Summaries"`
}

type Loaders struct {
	Loaders []LoaderSummaries

	Short bool
}

func (l Loaders) String() string {
	w := list.NewWriter()

	for i, loader := range l.Loaders {
		w.AppendItem(text.Bold.Sprintf("Loader %d: ", i+1))
		w.Indent()
		w.AppendItem(fmt.Sprintf("Loader UUID: %s", loader.Loader.UUID))
		w.AppendItem(fmt.Sprintf("Target host: %s", loader.Loader.URL))
		if loader.Loader.CreateDate.IsZero() {
			loader.Loader.CreateDate = time.Now()
		}
		w.AppendItem(fmt.Sprintf("Created at: %s", loader.Loader.CreateDate.In(time.Local)))
		w.AppendItem(fmt.Sprintf("Name: %s", loader.Loader.Name))
		if !l.Short {
			w.AppendItem(fmt.Sprintf("HTTP Engine: %s", loader.Loader.HTTPEngine))
			w.AppendItem(fmt.Sprintf("Description: %s", loader.Loader.Description))
			w.AppendItem(fmt.Sprintf("Concurrent connection: %d", loader.Loader.Connections))
			if loader.Loader.ReqCount != 0 {
				w.AppendItem(fmt.Sprintf("Request count: %d", loader.Loader.ReqCount))
			}
			if loader.Loader.AbortAfter != 0 {
				w.AppendItem(fmt.Sprintf("Abort after failed requests: %d", loader.Loader.AbortAfter))
			}
			if loader.Loader.RateLimit != 0 {
				w.AppendItem(fmt.Sprintf("Rate limit (req/s): %d", loader.Loader.RateLimit))
			}
			if loader.Loader.SkipVerify {
				w.AppendItem(fmt.Sprintf("TLS Skip verify: %v", loader.Loader.SkipVerify))
			}
			if loader.Loader.CA != nil {
				w.AppendItem(fmt.Sprintf("CA cert: %s", loader.Loader.CA))
			}
			if loader.Loader.Cert != nil {
				w.AppendItem(fmt.Sprintf("Certificate: %s", loader.Loader.Cert))
			}
			if loader.Loader.Key != nil {
				w.AppendItem(fmt.Sprintf("Key: %s", loader.Loader.Key))
			}
			if loader.Loader.Duration != 0 {
				w.AppendItem(fmt.Sprintf("Duration: %v", loader.Loader.Duration))
			}
			if loader.Loader.KeepAlive != 0 {
				w.AppendItem(fmt.Sprintf("KeepAlive: %v", loader.Loader.KeepAlive))
			}
			if loader.Loader.RequestDelay != 0 {
				w.AppendItem(fmt.Sprintf("Request delay: %v", loader.Loader.RequestDelay))
			}
			if loader.Loader.ReadTimeout != 0 {
				w.AppendItem(fmt.Sprintf("Read timeout (for fasthttp): %v", loader.Loader.ReadTimeout))
			}
			if loader.Loader.WriteTimeout != 0 {
				w.AppendItem(fmt.Sprintf("Write timeout (for fasthttp): %v", loader.Loader.WriteTimeout))
			}
			if loader.Loader.Timeout != 0 {
				w.AppendItem(fmt.Sprintf("Timeout (for net/http): %v", loader.Loader.Timeout))
			}
			if loader.Loader.BenchmarkTimeout != 0 {
				w.AppendItem(fmt.Sprintf("Benchmark timeout: %v", loader.Loader.BenchmarkTimeout))
			}
			if loader.Loader.Body != nil {
				w.AppendItem(fmt.Sprintf("Body: %s", string(loader.Loader.Body)))
			}

			if len(loader.Loader.Headers) > 0 {
				w.AppendItem("Headers: ")
				w.Indent()
				for k, v := range loader.Loader.Headers {
					w.Indent()
					for _, h := range v {
						w.AppendItem(fmt.Sprintf("%s: %s", k, h))
					}
					w.UnIndent()
				}

				w.UnIndent()
			}

			if len(loader.Loader.Parameters) > 0 {
				w.AppendItem("Parameters: ")
				w.Indent()
				paramIdx := 1
				for _, params := range loader.Loader.Parameters {
					w.AppendItem(fmt.Sprintf("Parameter %d: ", paramIdx))
					w.Indent()
					for k, v := range params {
						w.AppendItem(fmt.Sprintf("%s=%s", k, v))
					}

					w.UnIndent()
				}

				w.UnIndent()
			}

			if len(loader.Loader.Tags) > 0 {
				w.AppendItem("Tags: ")
				w.Indent()
				for _, tag := range loader.Loader.Tags {
					w.AppendItem(fmt.Sprintf("%s=%s", tag.Key, tag.Value))
				}

				w.UnIndent()
			}

			w.AppendItem("Summaries: ")
			w.Indent()

			for i, summary := range loader.Summaries {
				w.AppendItem(fmt.Sprintf("Summary %d", i+1))

				WriteSummary(w, summary, true, true)
			}

			w.UnIndent()
			w.UnIndent()
		}
	}

	return fmt.Sprintf("%s\n", w.Render())
}

func WriteSummary(w list.Writer, summary *model.Summary, stats, aggStats bool) {
	w.Indent()
	w.AppendItem("Basic")
	w.Indent()
	if summary.UUID != "" {
		w.AppendItem(fmt.Sprintf("%s %s", text.AlignLeft.Apply(text.Bold.Sprint("UUID:"), 9), summary.UUID))
	}
	w.AppendItem(fmt.Sprintf("%s %s", text.AlignLeft.Apply(text.Bold.Sprint("URL:"), 9), summary.URL))
	if summary.Description != "" {
		w.AppendItem(fmt.Sprintf("%s %s", text.AlignLeft.Apply(text.Bold.Sprint("Summary description:"), 9), summary.Description))
	}

	w.AppendItem(fmt.Sprintf("%s %v", text.AlignLeft.Apply(text.Bold.Sprint("Start:"), 9), fmt.Sprintf("%v", summary.Start.In(time.Local))))
	w.AppendItem(fmt.Sprintf("%s %v", text.AlignLeft.Apply(text.Bold.Sprint("End:"), 9), fmt.Sprintf("%v", summary.End.In(time.Local))))
	w.AppendItem(fmt.Sprintf("%s %v", text.AlignLeft.Apply(text.Bold.Sprint("Duration:"), 9), fmt.Sprintf("%v", summary.TotalTime)))
	w.UnIndent()
	w.AppendItem("Requests count")
	w.Indent()
	w.AppendItem(fmt.Sprintf("%s %d", text.AlignLeft.Apply(text.Bold.Sprint("Total requests count:"), 21), summary.ReqCount))
	w.AppendItem(fmt.Sprintf("%s %d", text.AlignLeft.Apply(text.Bold.Sprint("Success requests:"), 21), summary.SuccessReq))
	w.AppendItem(fmt.Sprintf("%s %d", text.AlignLeft.Apply(text.Bold.Sprint("Failed requests"), 21), summary.FailReq))
	w.AppendItem(fmt.Sprintf("%s %s", text.AlignLeft.Apply(text.Bold.Sprint("Data transferred"), 21), bytefmt.ByteSize(uint64(summary.DataTransferred))))
	w.AppendItem(fmt.Sprintf("%s %f/sec", text.AlignLeft.Apply(text.Bold.Sprint("Request per second"), 21), summary.ReqPerSec))

	w.UnIndent()
	w.AppendItem("Requests latency")
	w.Indent()
	w.AppendItem(fmt.Sprintf("%s %v", text.AlignLeft.Apply(text.Bold.Sprint("Average time"), 12), summary.AvgReqTime))
	w.AppendItem(fmt.Sprintf("%s %v", text.AlignLeft.Apply(text.Bold.Sprint("Min time"), 12), summary.MinReqTime))
	w.AppendItem(fmt.Sprintf("%s %v", text.AlignLeft.Apply(text.Bold.Sprint("Max time"), 12), summary.MaxReqTime))
	w.AppendItem(fmt.Sprintf("%s %v", text.AlignLeft.Apply(text.Bold.Sprint("P50 time"), 12), summary.P50ReqTime))
	w.AppendItem(fmt.Sprintf("%s %v", text.AlignLeft.Apply(text.Bold.Sprint("P75 time"), 12), summary.P75ReqTime))
	w.AppendItem(fmt.Sprintf("%s %v", text.AlignLeft.Apply(text.Bold.Sprint("P90 time"), 12), summary.P90ReqTime))
	w.AppendItem(fmt.Sprintf("%s %v", text.AlignLeft.Apply(text.Bold.Sprint("P99 time"), 12), summary.P99ReqTime))

	if len(summary.Errors) > 0 {
		w.UnIndent()
		w.AppendItem("Errors")
		w.Indent()

		for k, v := range summary.Errors {
			w.AppendItem(fmt.Sprintf("%s: %d", text.Bold.Sprint(k), v))
		}
	}

	if len(summary.HTTPCodes) > 0 {
		w.UnIndent()
		w.AppendItem("HTTP Codes")
		w.Indent()

		keys := make([]int, 0, len(summary.HTTPCodes))
		for k := range summary.HTTPCodes {
			keys = append(keys, k)
		}

		sort.Ints(keys)

		for _, k := range keys {
			w.AppendItem(fmt.Sprintf("HTTP Code %d: %d", k, summary.HTTPCodes[k]))
		}
	}

	if len(summary.AggregatedStats) > 0 && aggStats {
		w.UnIndent()
		w.AppendItem("Aggregated stats")
		w.Indent()

		for i, k := range summary.AggregatedStats {
			i++
			w.AppendItem(fmt.Sprintf("Window %d", i))
			w.Indent()
			w.AppendItem(fmt.Sprintf("Start time %v", k.Start))
			w.AppendItem(fmt.Sprintf("End time %v", k.End))
			w.AppendItem(fmt.Sprintf("Duration %v", k.Duration))
			w.AppendItem(fmt.Sprintf("Requests count %d", k.RequestCount))
			w.AppendItem(fmt.Sprintf("Min request time %v", k.MinRequestTime))
			w.AppendItem(fmt.Sprintf("Max request time %v", k.MaxRequestTime))
			w.AppendItem(fmt.Sprintf("Average request time %v", k.AvgRequestTime))
			w.UnIndent()
		}
	}

	if len(summary.RequestStats) > 0 && stats {
		w.UnIndent()
		w.AppendItem("Full requests stats")
		w.Indent()

		for i, k := range summary.RequestStats {
			i++
			w.AppendItem(fmt.Sprintf("Request %d", i))
			w.Indent()
			w.AppendItem(fmt.Sprintf("Start time %v", k.Start))
			w.AppendItem(fmt.Sprintf("End time %v", k.End))
			w.AppendItem(fmt.Sprintf("Duration %v", k.Duration))
			w.AppendItem(fmt.Sprintf("Body size %d", k.BodySize))
			w.AppendItem(fmt.Sprintf("Code %d", k.RetCode))
			w.AppendItem(fmt.Sprintf("Error %s", k.Error))
			w.UnIndent()
		}
	}

	w.UnIndent()
	w.UnIndent()
}

func WriteLoaderTagsMap(tags map[string]*model.LoaderTag) string {
	w := list.NewWriter()

	for uuid, tag := range tags {
		w.AppendItem(fmt.Sprintf("Loader UUID: %s", uuid))
		w.AppendItem("Tags:")
		w.Indent()
		w.AppendItem(fmt.Sprintf("%s=%s", tag.Key, tag.Value))
		w.UnIndent()
	}

	return w.Render()
}

func WriteLoaderTags(loaderUUID string, tags []*model.LoaderTag) string {
	w := list.NewWriter()

	w.AppendItem(fmt.Sprintf("Loader UUID: %s", loaderUUID))
	w.AppendItem("Tags: ")
	w.Indent()

	for _, tag := range tags {
		w.AppendItem(fmt.Sprintf("%s=%s", tag.Key, tag.Value))
	}

	return w.Render()
}

func CreateDBDirectory(dbDir string) error {
	dir := filepath.Dir(dbDir)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0700)
		if err != nil {
			return fmt.Errorf("could not create HLoad configuration directory: %w", err)
		}
	}

	return nil
}
