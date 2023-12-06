package loader

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/tmwalaszek/hload/model"

	"github.com/caio/go-tdigest/v4"
	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"
)

type Requester interface {
	Request() *model.RequestStat
}

const DefaultConnection = 10
const HTTPEngine = "http"
const FastHTTPEngine = "fast_http"

type Loader struct {
	opts *model.Loader

	reqChan   chan struct{}
	statsChan chan *model.RequestStat
	requester Requester

	progressChan chan struct{}
}

func NewLoader(opts *model.Loader) (*Loader, error) {
	return newLoader(opts)
}

func NewLoaderProgress(opts *model.Loader, progressChan chan struct{}) (*Loader, error) {
	l, err := newLoader(opts)
	if err != nil {
		return nil, err
	}

	l.progressChan = progressChan
	return l, nil
}

func newLoader(opts *model.Loader) (*Loader, error) {
	var tlsConfig tls.Config

	_, err := url.Parse(opts.URL)
	if err != nil {
		return nil, fmt.Errorf("bad url format: %w", err)
	}

	if opts.Connections < 0 {
		return nil, errors.New("number of connections has to be positive")
	}

	if opts.Connections == 0 {
		return nil, errors.New("number of connection has to be set")
	}

	if opts.Method == "" {
		return nil, errors.New("HTTP request method has to set")
	}

	if opts.AbortAfter < 0 {
		return nil, errors.New("number of abort after requests has to be positive")
	}

	if opts.ReqCount < 0 {
		return nil, errors.New("number of requests count has to be positive")
	}

	if opts.ReqCount == 0 && opts.Duration == 0 {
		return nil, errors.New("requests count or duration has to be set")
	}

	if opts.SkipVerify {
		tlsConfig.InsecureSkipVerify = opts.SkipVerify
	} else {
		if len(opts.CA) != 0 {
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(opts.CA)

			tlsConfig.RootCAs = caCertPool

			if len(opts.Cert) != 0 && len(opts.Key) != 0 {
				cert, err := tls.LoadX509KeyPair(string(opts.Cert), string(opts.Key))
				if err != nil {
					return nil, fmt.Errorf("could not load X509 key pair: %w", err)
				}

				tlsConfig.Certificates = []tls.Certificate{cert}
			}
		}
	}

	var reqChan chan struct{}
	if opts.Connections == 1 {
		reqChan = make(chan struct{})
	} else {
		reqChan = make(chan struct{}, opts.Connections)
	}

	statsChan := make(chan *model.RequestStat)
	var requester Requester

	switch v := opts.HTTPEngine; v {
	case HTTPEngine:
		requester, err = NewLoaderHTTP(opts)
	case FastHTTPEngine:
		requester, err = NewLoaderFastHTTP(opts)
	default:
		return nil, fmt.Errorf("wrong http engine %s", v)
	}

	if err != nil {
		return nil, err
	}

	return &Loader{
		opts:      opts,
		requester: requester,

		reqChan:   reqChan,
		statsChan: statsChan,
	}, nil
}

func (l *Loader) aggregateStat(stat *model.RequestStat, start time.Time, aggStats *[]*model.AggregatedStat) {
	diff := stat.Start.Sub(start)
	win := int(diff / l.opts.AggregateWindow)

	if len(*aggStats) <= win {
		for i := 0; i <= (win - len(*aggStats)); i++ {
			aggStat := &model.AggregatedStat{
				Start: start.Add(l.opts.AggregateWindow * time.Duration(win)),
				End:   start.Add(l.opts.AggregateWindow * time.Duration(win+1)),
			}

			*aggStats = append(*aggStats, aggStat)
		}
	}

	if (*aggStats)[win].MinRequestTime == 0 && (*aggStats)[win].MaxRequestTime == 0 {
		(*aggStats)[win].MinRequestTime = stat.Duration
		(*aggStats)[win].MaxRequestTime = stat.Duration
	} else {
		if stat.Duration > (*aggStats)[win].MaxRequestTime {
			(*aggStats)[win].MaxRequestTime = stat.Duration
		}

		if stat.Duration < (*aggStats)[win].MinRequestTime {
			(*aggStats)[win].MinRequestTime = stat.Duration
		}
	}
	(*aggStats)[win].AvgRequestTime += stat.Duration
	(*aggStats)[win].RequestCount++
}

func (l *Loader) Do(ctx context.Context) (*model.Summary, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	for i := 0; i < l.opts.Connections; i++ {
		wg.Add(1)
		go l.worker(&wg)
	}

	done := make(chan struct{}, 1)
	go l.manageWorkers(ctx, done, &wg)

	requestsTimes := make([]*model.RequestStat, 0)
	errorsMap := make(map[string]int)
	httpCodes := make(map[int]int)

	var success, fail int
	var dataTransferred int

	start := time.Now().UTC().Truncate(time.Second)

	var aborted bool
	var minDuration, maxDuration, avgDuration time.Duration

	aggStats := make([]*model.AggregatedStat, 0, 1000)

	// We need to create the first aggregated window
	if l.opts.AggregateWindow != 0 && l.opts.GatherAggregateRequestsStats {
		aggStat := &model.AggregatedStat{
			Start: start,
			End:   start.Add(l.opts.AggregateWindow),
		}
		aggStats = append(aggStats, aggStat)
	}

	t, err := tdigest.New()
	if err != nil {
		return nil, fmt.Errorf("tdigest error: %w", err)
	}
MAIN:
	for {
		select {
		case stat, ok := <-l.statsChan:
			if !ok {
				break MAIN
			}

			if l.progressChan != nil {
				l.progressChan <- struct{}{}
			}

			if minDuration == 0 && maxDuration == 0 {
				maxDuration = stat.Duration
				minDuration = stat.Duration
			} else {
				if stat.Duration > maxDuration {
					maxDuration = stat.Duration
				}

				if stat.Duration < minDuration {
					minDuration = stat.Duration
				}
			}

			avgDuration += stat.Duration

			// TODO(tmwalaszek) this should not really happen so we fatal here at the moment
			err = t.Add(float64(stat.Duration))
			if err != nil {
				log.Fatalf("error in request duration stat: %v", err)
			}

			// calculate window
			if l.opts.AggregateWindow != 0 && l.opts.GatherAggregateRequestsStats {
				l.aggregateStat(stat, start, &aggStats)
			}

			if l.opts.GatherFullRequestsStats {
				r := &model.RequestStat{
					Start:    stat.Start,
					End:      stat.End,
					Duration: stat.Duration,
					RetCode:  stat.RetCode,
					BodySize: stat.BodySize,
					Error:    stat.Error,
				}

				requestsTimes = append(requestsTimes, r)
			}

			if stat.RetCode >= 200 && stat.RetCode < 300 && stat.Error == "" {
				success++
				dataTransferred += stat.BodySize
			} else {
				fail++
				var errString string
				if stat.Error != "" {
					errString = stat.Error
				} else {
					errString = fasthttp.StatusMessage(stat.RetCode)
				}

				if _, ok := errorsMap[errString]; !ok {
					errorsMap[errString] = 1
				} else {
					errorsMap[errString]++
				}
			}

			if stat.Error == "" {
				if _, ok := httpCodes[stat.RetCode]; !ok {
					httpCodes[stat.RetCode] = 1
				} else {
					httpCodes[stat.RetCode]++
				}
			}

			if fail >= l.opts.AbortAfter && l.opts.AbortAfter != 0 {
				if !aborted {
					done <- struct{}{}
					aborted = true
				}
			}
		case <-ctx.Done():
			break MAIN
		}
	}

	end := time.Now().UTC().Truncate(time.Second)
	totalTime := time.Since(start)

	for i := range aggStats {
		if i < len(aggStats)-1 {
			aggStats[i].Duration = aggStats[i].End.Sub(aggStats[i].Start)
		} else {
			aggStats[i].Duration = end.Sub(aggStats[i].Start)
		}
	}

	p50 := time.Duration(t.Quantile(0.5))
	p75 := time.Duration(t.Quantile(0.75))
	p90 := time.Duration(t.Quantile(0.9))
	p99 := time.Duration(t.Quantile(0.99))

	if success != 0 {
		avgDuration = time.Duration(int64(avgDuration) / int64(success))
	} else {
		avgDuration = 0
	}

	reqCount := success + fail
	var reqPerSecond float64

	if totalTime > time.Second {
		reqPerSecond = float64(success) / (float64(totalTime) / float64(time.Second))
	} else {
		reqPerSecond = float64(success)
	}

	summary := &model.Summary{
		URL:             l.opts.URL,
		Start:           start,
		End:             end,
		TotalTime:       totalTime,
		DataTransferred: dataTransferred,
		ReqPerSec:       reqPerSecond,
		ReqCount:        reqCount,
		SuccessReq:      success,
		FailReq:         fail,
		AvgReqTime:      avgDuration,
		MinReqTime:      minDuration,
		MaxReqTime:      maxDuration,
		P50ReqTime:      p50,
		P75ReqTime:      p75,
		P90ReqTime:      p90,
		P99ReqTime:      p99,
		Errors:          errorsMap,
		HTTPCodes:       httpCodes,
		AggregatedStats: aggStats,
		RequestStats:    requestsTimes,
	}

	return summary, nil
}

func outputFn[T any](wg *sync.WaitGroup, out chan struct{}, c <-chan T) {
	for range c {
		out <- struct{}{}
	}

	wg.Done()
}

func merge(c1, c2 <-chan time.Time, done ...<-chan struct{}) <-chan struct{} {
	var wg sync.WaitGroup
	out := make(chan struct{})

	wg.Add(len(done) + 2)

	go outputFn[time.Time](&wg, out, c1)
	go outputFn[time.Time](&wg, out, c2)

	for _, c := range done {
		go outputFn[struct{}](&wg, out, c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func (l *Loader) manageWorkers(ctx context.Context, done chan struct{}, wg *sync.WaitGroup) {
	breakAfter := make(<-chan time.Time)
	if l.opts.Duration != 0 {
		breakAfter = time.After(l.opts.Duration)
	}

	benchmarkTimeout := make(<-chan time.Time)
	if l.opts.BenchmarkTimeout != 0 {
		benchmarkTimeout = time.After(l.opts.BenchmarkTimeout)
	}

	var limiter *rate.Limiter
	if l.opts.RateLimit != 0 {
		limiter = rate.NewLimiter(rate.Limit(l.opts.RateLimit), l.opts.RateLimit)
	}

	mergedChan := merge(breakAfter, benchmarkTimeout, ctx.Done(), done)

	runRequestLoop := func() {
		ticker := time.NewTicker(time.Millisecond * 50)
		defer ticker.Stop()

		var i int
		for {
			select {
			case <-mergedChan:
				return
			default:
			}

			select {
			case l.reqChan <- struct{}{}:
				if limiter != nil {
					ok := limiter.Allow()
					if !ok {
						err := limiter.Wait(ctx)
						if err != nil {
							break
						}
					}
				}
				i++
			case <-ticker.C:
			}

			if i >= l.opts.ReqCount && l.opts.ReqCount > 0 {
				break
			}
		}
	}

	runRequestLoop()

	close(l.reqChan)

	wg.Wait()
	close(l.statsChan)
}

func (l *Loader) worker(wg *sync.WaitGroup) {
	savedReqTime := time.Time{}

	for {
		_, ok := <-l.reqChan
		if !ok {
			break
		}

		reqTime := time.Now()
		if l.opts.RequestDelay != 0 && !savedReqTime.IsZero() {
			s := savedReqTime.Add(l.opts.RequestDelay).Sub(reqTime)
			if s > 0 {
				time.Sleep(s)
			}
		}

		stat := l.requester.Request()
		savedReqTime = time.Now()

		l.statsChan <- stat
	}

	wg.Done()
}
