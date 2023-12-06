package loader

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tmwalaszek/hload/mock"
	"github.com/tmwalaszek/hload/model"

	"github.com/stretchr/testify/require"
)

var httpEngines = []string{"http", "fast_http"}

func TestHeaders(t *testing.T) {
	t.Parallel()

	var correctHeaders = []struct {
		Name            string
		HeadersStrings  []string
		ExpectedHeaders model.Headers
	}{
		{
			Name:           "Content-type json",
			HeadersStrings: []string{"Content-type: application/json"},
			ExpectedHeaders: model.Headers{
				"Content-type": []string{"application/json"},
			},
		},
		{
			Name:           "Set-Cookie multiple values",
			HeadersStrings: []string{"Cookie-set: language=pl", "Cookie-set: id=123"},
			ExpectedHeaders: model.Headers{
				"Cookie-set": []string{"language=pl", "id=123"},
			},
		},
		{
			Name:           "Content-type json with spaces",
			HeadersStrings: []string{" Content-type:application/json  "},
			ExpectedHeaders: model.Headers{
				"Content-type": []string{"application/json"},
			},
		},
	}

	for _, tc := range correctHeaders {
		t.Run(tc.Name, func(t *testing.T) {
			headers := make(model.Headers)
			for _, h := range tc.HeadersStrings {
				err := headers.Set(h)
				require.Nil(t, err)
			}

			require.Equal(t, tc.ExpectedHeaders, headers)
		})
	}

	var wrongHeaders = []struct {
		Name   string
		Header string
	}{
		{
			Name:   "Without :",
			Header: "Content-typeapplication/json",
		},
		{
			Name:   "With empty character in header name",
			Header: "Content- type: application/json",
		},
	}

	for _, tc := range wrongHeaders {
		t.Run(tc.Name, func(t *testing.T) {
			headers := make(model.Headers)
			err := headers.Set(tc.Header)
			require.NotNil(t, err)
		})
	}
}

func TestMixedRequests(t *testing.T) {
	t.Parallel()

	var tt = []struct {
		Name           string
		ReqCount       int
		Connections    int
		FailedRequests int
	}{
		{
			Name:           "20 requests with 5 of them failed - 1 worker",
			ReqCount:       20,
			Connections:    1,
			FailedRequests: 5,
		},
		{
			Name:           "20 requests with 10 of them failed - 2 workers",
			ReqCount:       20,
			Connections:    2,
			FailedRequests: 10,
		},
		{
			Name:           "1000 requests with 356 of them failed - 10 workers",
			ReqCount:       1000,
			Connections:    10,
			FailedRequests: 356,
		},
	}

	for _, engine := range httpEngines {
		for _, tc := range tt {
			t.Run(fmt.Sprintf("Testcase %s for engine %s", tc.Name, engine), func(t *testing.T) {
				handler, ts := mock.NewServer(tc.FailedRequests)
				defer ts.Close()

				u, err := url.JoinPath(ts.URL, "mixed")
				require.Nil(t, err)
				opts := &model.Loader{
					URL:        u,
					HTTPEngine: engine,
					Method:     "GET",
					LoaderReqDetails: model.LoaderReqDetails{
						ReqCount:    tc.ReqCount,
						Connections: tc.Connections,
					},
				}
				loader, err := NewLoader(opts)
				require.Nil(t, err)

				summary, err := loader.Do(context.Background())
				require.Nil(t, err)
				require.Equal(t, tc.ReqCount, summary.ReqCount)
				require.Equal(t, tc.ReqCount-tc.FailedRequests, summary.SuccessReq)
				require.Equal(t, tc.FailedRequests, summary.FailReq)
				require.Equal(t, tc.ReqCount, int(handler.Stats.RequestCount))
			})
		}
	}
}

func TestLoaderOKDurationLimiter(t *testing.T) {
	t.Parallel()

	handler, ts := mock.NewServer(0)
	defer ts.Close()

	u, err := url.JoinPath(ts.URL, "ok")
	require.Nil(t, err)

	opts := &model.Loader{
		URL:    u,
		Method: "GET",
	}

	var tt = []struct {
		Name        string
		Duration    time.Duration
		Connections int
		RateLimit   int
	}{
		{
			Name:        "5 seconds benchmark - 1 worker",
			Duration:    time.Second * 5,
			Connections: 1,
			RateLimit:   10,
		},
		{
			Name:        "10 seconds benchmark - 2 worker",
			Duration:    time.Second * 10,
			Connections: 2,
			RateLimit:   20,
		},
		{
			Name:        "10 seconds benchmark - 2 worker",
			Duration:    time.Second * 10,
			Connections: 4,
			RateLimit:   20,
		},
	}

	for _, engine := range httpEngines {
		for _, tc := range tt {
			t.Run(fmt.Sprintf("Testcase %s for engine %s", tc.Name, engine), func(t *testing.T) {
				handler.ResetStats()
				opts.Duration = tc.Duration
				opts.Connections = tc.Connections
				opts.RateLimit = tc.RateLimit
				opts.HTTPEngine = engine

				loader, err := NewLoader(opts)
				require.Nil(t, err)

				summary, err := loader.Do(context.Background())
				require.Nil(t, err)
				maxAllowed := tc.RateLimit*(int(tc.Duration)/1000000000) + (tc.RateLimit * 2)

				require.LessOrEqual(t, summary.ReqCount, maxAllowed)
			})
		}
	}
}

func TestLoaderOKDelay(t *testing.T) {
	t.Parallel()

	handler, ts := mock.NewServer(0)
	defer ts.Close()

	u, err := url.JoinPath(ts.URL, "ok")
	require.Nil(t, err)

	opts := &model.Loader{
		URL:    u,
		Method: "GET",
	}

	var tt = []struct {
		Name       string
		ReqCount   int
		ReqDelay   time.Duration
		Connection int
	}{
		{
			Name:       "10 requests delay 1 second 1 connection",
			ReqCount:   10,
			ReqDelay:   time.Second,
			Connection: 1,
		},
		{
			Name:       "20 requests delay 1 second 1 connection",
			ReqCount:   20,
			ReqDelay:   time.Second,
			Connection: 1,
		},
		{
			Name:       "10 requests delay 1 second 2 connection",
			ReqCount:   10,
			ReqDelay:   time.Second,
			Connection: 2,
		},
		{
			Name:       "20 requests delay 1 second 2 connection",
			ReqCount:   20,
			ReqDelay:   time.Second,
			Connection: 2,
		},
	}

	for _, engine := range httpEngines {
		for _, tc := range tt {
			t.Run(fmt.Sprintf("Testcase %s for engine %s", tc.Name, engine), func(t *testing.T) {
				handler.ResetStats()
				opts.ReqCount = tc.ReqCount
				opts.RequestDelay = tc.ReqDelay
				opts.Connections = tc.Connection
				opts.HTTPEngine = engine

				loader, err := NewLoader(opts)
				require.Nil(t, err)

				start := time.Now()
				_, err = loader.Do(context.Background())
				end := time.Now()
				elapsed := end.Sub(start)
				require.Nil(t, err)
				expected := (time.Duration(tc.ReqCount-1) * time.Second) / time.Duration(tc.Connection)
				require.Equal(t, expected.Truncate(time.Second), elapsed.Truncate(time.Second))
			})
		}
	}

}

func TestLoaderOKDuration(t *testing.T) {
	t.Parallel()

	handler, ts := mock.NewServer(0)
	defer ts.Close()

	u, err := url.JoinPath(ts.URL, "ok")
	require.Nil(t, err)

	opts := &model.Loader{
		URL:    u,
		Method: "GET",
	}

	var tt = []struct {
		Name        string
		Duration    time.Duration
		Connections int
	}{
		{
			Name:        "5 seconds benchmark - 1 worker",
			Duration:    time.Second * 5,
			Connections: 1,
		},
		{
			Name:        "10 seconds benchmark - 2 worker",
			Duration:    time.Second * 10,
			Connections: 2,
		},
	}

	for _, engine := range httpEngines {
		for _, tc := range tt {
			t.Run(fmt.Sprintf("Testcase %s for engine %s", tc.Name, engine), func(t *testing.T) {
				handler.ResetStats()
				opts.Duration = tc.Duration
				opts.Connections = tc.Connections
				opts.HTTPEngine = engine

				loader, err := NewLoader(opts)
				require.Nil(t, err)

				start := time.Now()
				_, err = loader.Do(context.Background())
				end := time.Now()
				elapsed := end.Sub(start)
				require.Nil(t, err)
				require.Equal(t, elapsed.Truncate(time.Second), tc.Duration)
			})
		}
	}
}

func TestLoaderWithArgs(t *testing.T) {
	t.Parallel()

	handler, ts := mock.NewServer(0)
	defer ts.Close()

	u, err := url.JoinPath(ts.URL, "args")
	require.Nil(t, err)

	opts := &model.Loader{
		URL:    u,
		Method: "GET",
	}

	p1 := make(model.Parameters, 0)
	err = p1.Set("key1=value1&key2=value2")
	require.Nil(t, err)

	var tt = []struct {
		Name        string
		ReqCount    int
		Connections int
		Parameters  model.Parameters
	}{
		{
			Name:        "10 requests 1 worker1",
			ReqCount:    1,
			Connections: 1,
			Parameters:  p1,
		},
	}

	for _, engine := range httpEngines {
		for _, tc := range tt {
			t.Run(fmt.Sprintf("Testcase %s for engine %s", tc.Name, engine), func(t *testing.T) {
				handler.ResetStats()
				opts.ReqCount = tc.ReqCount
				opts.Connections = tc.Connections
				opts.Parameters = tc.Parameters
				opts.HTTPEngine = engine

				loader, err := NewLoader(opts)
				require.Nil(t, err)

				_, err = loader.Do(context.Background())
				require.Nil(t, err)
				expectedArgs := strings.Split("key1=value1&key2=value2", "&")
				require.ElementsMatch(t, expectedArgs, handler.Args)
			})
		}
	}
}

func TestLoaderHeaderRequest(t *testing.T) {
	t.Parallel()

	handler, ts := mock.NewServer(0)
	defer ts.Close()

	u, err := url.JoinPath(ts.URL, "header")
	require.Nil(t, err)

	opts := &model.Loader{
		URL:    u,
		Method: "GET",
	}

	var tt = []struct {
		Name        string
		ReqCount    int
		Connections int
		Headers     []string
	}{
		{
			Name:        "json header",
			ReqCount:    2,
			Connections: 1,
			Headers:     []string{"Content-type: application/json"},
		},
		{
			Name:        "json+setcookie header",
			ReqCount:    2,
			Connections: 1,
			Headers:     []string{"Content-type: application/json", "Cookie-set: language=pl"},
		},
	}

	for _, engine := range httpEngines {
		for _, tc := range tt {
			t.Run(fmt.Sprintf("Testcase %s for engine %s", tc.Name, engine), func(t *testing.T) {
				handler.ResetStats()
				opts.ReqCount = tc.ReqCount
				opts.Connections = tc.Connections
				opts.HTTPEngine = engine

				header := make(model.Headers)
				for _, h := range tc.Headers {
					err := header.Set(h)
					require.Nil(t, err)
				}
				opts.Headers = header

				loader, err := NewLoader(opts)

				require.Nil(t, err)
				_, err = loader.Do(context.Background())
				require.Nil(t, err)

				for _, v := range tc.Headers {
					key := strings.Split(v, ":")[0]
					val := strings.Split(v, ":")[1]

					val = strings.TrimSpace(val)

					require.Contains(t, handler.Headers, key)
					require.Equal(t, handler.Headers[key].Value, val)
					require.Equal(t, handler.Headers[key].Count, tc.ReqCount)
				}
			})
		}
	}

}

func TestLoaderBenchmarkTimeout(t *testing.T) {
	t.Parallel()

	handler, ts := mock.NewServer(0)
	defer ts.Close()

	u, err := url.JoinPath(ts.URL, "long")
	require.Nil(t, err)

	opts := &model.Loader{
		URL:    u,
		Method: "GET",
		LoaderReqDetails: model.LoaderReqDetails{
			Connections: 1,
		},
	}

	var tt = []struct {
		Name             string
		ReqCount         int
		Connections      int
		BenchmarkTimeout time.Duration
		Pass             bool
	}{
		{
			Name:             "Benchmark should timeout - 1 connection",
			ReqCount:         100,
			Connections:      1,
			BenchmarkTimeout: 10 * time.Second,
			Pass:             false,
		},
		{
			Name:             "Benchmark should timeout - 2 connections",
			ReqCount:         100,
			BenchmarkTimeout: 10 * time.Second,
			Connections:      2,
			Pass:             false,
		},
		{
			Name:             "Benchmark should timeout - 20 connections",
			ReqCount:         100,
			BenchmarkTimeout: 10 * time.Second,
			Connections:      20,
			Pass:             false,
		},
	}

	for _, engine := range httpEngines {
		for _, tc := range tt {
			t.Run(fmt.Sprintf("Testcase %s for engine %s", tc.Name, engine), func(t *testing.T) {
				handler.ResetStats()
				opts.ReqCount = tc.ReqCount
				opts.Connections = tc.Connections
				opts.BenchmarkTimeout = tc.BenchmarkTimeout
				opts.HTTPEngine = engine

				loader, err := NewLoader(opts)
				require.Nil(t, err)

				s := time.Now()
				_, err = loader.Do(context.Background())
				d := time.Since(s)
				require.Nil(t, err)
				require.LessOrEqual(t, d/1e6, tc.BenchmarkTimeout)
			})
		}
	}
}

func TestLoaderAbortRequest(t *testing.T) {
	t.Parallel()

	handler, ts := mock.NewServer(0)
	defer ts.Close()

	u, err := url.JoinPath(ts.URL, "abort")
	require.Nil(t, err)

	opts := &model.Loader{
		URL:    u,
		Method: "GET",
	}

	var tt = []struct {
		Name        string
		ReqCount    int
		Connections int
		Abort       int
	}{
		{
			Name:        "Test body 1",
			ReqCount:    100,
			Connections: 1,
			Abort:       2,
		},
	}

	for _, engine := range httpEngines {
		for _, tc := range tt {
			t.Run(fmt.Sprintf("Testcase %s for engine %s", tc.Name, engine), func(t *testing.T) {
				handler.ResetStats()
				opts.ReqCount = tc.ReqCount
				opts.Connections = tc.Connections
				opts.AbortAfter = tc.Abort
				opts.HTTPEngine = engine

				loader, err := NewLoader(opts)
				require.Nil(t, err)

				summary, err := loader.Do(context.Background())
				require.Nil(t, err)
				require.Equal(t, summary.ReqCount, int(handler.Stats.RequestCount))
				require.LessOrEqual(t, summary.ReqCount, tc.Abort+(tc.Connections*2)+1)
			})
		}
	}
}

func TestLoaderBodyRequest(t *testing.T) {
	t.Parallel()

	handler, ts := mock.NewServer(0)
	defer ts.Close()

	u, err := url.JoinPath(ts.URL, "body")
	require.Nil(t, err)

	opts := &model.Loader{
		URL:    u,
		Method: "POST",
	}

	var tt = []struct {
		Name        string
		ReqCount    int
		Connections int
		Body        []byte
	}{
		{
			Name:        "Test body 1",
			ReqCount:    1,
			Connections: 1,
			Body:        []byte("test body"),
		},
		{
			Name:        "Test body2",
			ReqCount:    10,
			Connections: 1,
			Body:        []byte("Test body with many requests"),
		},
		{
			Name:        "Test body2",
			ReqCount:    10,
			Connections: 4,
			Body:        []byte("Test body with many requests"),
		},
	}

	for _, engine := range httpEngines {
		for _, tc := range tt {
			t.Run(fmt.Sprintf("Testcase %s for engine %s", tc.Name, engine), func(t *testing.T) {
				handler.ResetStats()
				opts.ReqCount = tc.ReqCount
				opts.Connections = tc.Connections
				opts.Body = tc.Body
				opts.HTTPEngine = engine

				loader, err := NewLoader(opts)
				require.Nil(t, err)

				summary, err := loader.Do(context.Background())
				require.Nil(t, err)
				require.Equal(t, tc.ReqCount, summary.ReqCount)
				require.Len(t, handler.Body, 1)
				require.Equal(t, tc.Body, handler.Body[0])
			})
		}
	}

}
func TestLoaderOKRequests(t *testing.T) {
	t.Parallel()

	handler, ts := mock.NewServer(0)
	defer ts.Close()

	u, err := url.JoinPath(ts.URL, "ok")
	require.Nil(t, err)

	opts := &model.Loader{
		URL:    u,
		Method: "GET",
	}

	var tt = []struct {
		Name        string
		ReqCount    int
		Connections int
	}{
		{
			Name:        "10 requests 1 worker1",
			ReqCount:    10,
			Connections: 1,
		},
		{
			Name:        "10 requests 2 workers",
			ReqCount:    10,
			Connections: 2,
		},
		{
			Name:        "10 requests 10 workers",
			ReqCount:    10,
			Connections: 10,
		},
		{
			Name:        "10 requests 20 workers",
			ReqCount:    10,
			Connections: 20,
		},
		{
			Name:        "100 requests 3 workers",
			ReqCount:    100,
			Connections: 3,
		},
	}

	for _, engine := range httpEngines {
		for _, tc := range tt {
			t.Run(fmt.Sprintf("Testcase %s for engine %s", tc.Name, engine), func(t *testing.T) {
				handler.ResetStats()
				opts.ReqCount = tc.ReqCount
				opts.Connections = tc.Connections
				opts.HTTPEngine = engine

				loader, err := NewLoader(opts)
				require.Nil(t, err)

				summary, err := loader.Do(context.Background())
				require.Nil(t, err)
				require.Equal(t, tc.ReqCount, summary.ReqCount)
				require.Equal(t, tc.ReqCount, int(handler.Stats.RequestCount))
			})
		}
	}
}

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())

	v, _ := strconv.ParseBool(os.Getenv("DOCKER_BENCHMARK_NGINX"))
	var wg sync.WaitGroup
	// If DOCKER_BENCHMARK_NGINX env var is set to true
	if v {
		readyChan := make(chan struct{})
		errChan := make(chan error)

		wg.Add(1)
		mock.NewBenchmarkServer(ctx, &wg, readyChan, errChan)

		select {
		case <-readyChan:
		case err := <-errChan:
			log.Fatalf("Could not start benchmark server: %v", err)
		}
	}

	code := m.Run()

	if v {
		cancel()
		wg.Wait()
	}

	os.Exit(code)
}

func benchmarkLoader(opts *model.Loader, b *testing.B) {
	target := os.Getenv("BENCHMARK_SERVER")
	if target == "" {
		target = "http://127.0.0.1:8080"
	}

	opts.URL = target

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		loader, err := NewLoader(opts)
		if err != nil {
			b.Fatalf("Loader error: %v", err)
		}
		_, err = loader.Do(context.Background())
		if err != nil {
			b.Fatalf("Loader Do error: %v", err)
		}
	}
}

// Fasthttp benchmarks 1000 requests 1, 10, 100, 250, 500 connections
func BenchmarkLoaderFastHTTP1(b *testing.B) {
	opts := &model.Loader{
		LoaderReqDetails: model.LoaderReqDetails{
			ReqCount:    1000,
			Connections: 1,
		},
		HTTPEngine: FastHTTPEngine,
	}

	benchmarkLoader(opts, b)
}

func BenchmarkLoaderFastHTTP10(b *testing.B) {
	opts := &model.Loader{
		LoaderReqDetails: model.LoaderReqDetails{
			ReqCount:    1000,
			Connections: 10,
		},
		HTTPEngine: FastHTTPEngine,
	}

	benchmarkLoader(opts, b)
}

func BenchmarkLoaderFastHTTP100(b *testing.B) {
	opts := &model.Loader{
		LoaderReqDetails: model.LoaderReqDetails{
			ReqCount:    1000,
			Connections: 100,
		},
		HTTPEngine: FastHTTPEngine,
	}

	benchmarkLoader(opts, b)
}

// Net/http benchmarks 1000 requests 1, 10, 100 connections
func BenchmarkLoaderHTTP1(b *testing.B) {
	opts := &model.Loader{
		LoaderReqDetails: model.LoaderReqDetails{
			ReqCount:    1000,
			Connections: 1,
		},
		HTTPEngine: HTTPEngine,
	}

	benchmarkLoader(opts, b)
}

func BenchmarkLoaderHTTP10(b *testing.B) {
	opts := &model.Loader{
		LoaderReqDetails: model.LoaderReqDetails{
			ReqCount:    1000,
			Connections: 10,
		},
		HTTPEngine: HTTPEngine,
	}

	benchmarkLoader(opts, b)
}

func BenchmarkLoaderHTTP100(b *testing.B) {
	opts := &model.Loader{
		LoaderReqDetails: model.LoaderReqDetails{
			ReqCount:    1000,
			Connections: 100,
		},
		HTTPEngine: HTTPEngine,
	}

	benchmarkLoader(opts, b)
}
