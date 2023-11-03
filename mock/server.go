package mock

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

type BenchmarkStats struct {
	RequestCount uint64
}

type headerCount struct {
	Value string
	Count int
}

type LoaderHandler struct {
	Stats               BenchmarkStats
	MixedFailedRequests int
	Args                []string
	// Only one body, loader can send only one Body in benchmark, might change later
	// If more than one body is in here, then we have an issue
	Body    [][]byte
	Headers map[string]*headerCount

	mx sync.Mutex
}

func (h *LoaderHandler) ResetStats() {
	h.Stats.RequestCount = 0
	h.Args = []string{}
	h.Body = [][]byte{}
	h.Headers = make(map[string]*headerCount)
}

func (h *LoaderHandler) appendIfNotExists(body []byte) {
	for _, b := range h.Body {
		if reflect.DeepEqual(b, body) {
			return
		}
	}

	h.Body = append(h.Body, body)
}

func (h *LoaderHandler) HandleHeaderRequests(w http.ResponseWriter, r *http.Request) {
	h.mx.Lock()
	defer h.mx.Unlock()

	contentType := r.Header.Get("Content-Type")
	cookieSet := r.Header.Get("Cookie-set")

	if len(contentType) != 0 {
		v, ok := h.Headers["Content-type"]
		if !ok {
			h.Headers["Content-type"] = &headerCount{
				Value: contentType,
				Count: 1,
			}
		} else {
			v.Count++
		}
	}

	if len(cookieSet) != 0 {
		v, ok := h.Headers["Cookie-set"]
		if !ok {
			h.Headers["Cookie-set"] = &headerCount{
				Value: cookieSet,
				Count: 1,
			}
		} else {
			v.Count++
		}
	}
}

func (h *LoaderHandler) HandleBodyRequests(w http.ResponseWriter, r *http.Request) {
	h.mx.Lock()
	defer h.mx.Unlock()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	copiedBody := make([]byte, len(body))
	copy(copiedBody, body)
	h.appendIfNotExists(copiedBody)
}

// Benchmark should contain only one request here, we want to check whether the args are set only
func (h *LoaderHandler) HandleArgsRequest(w http.ResponseWriter, r *http.Request) {
	h.mx.Lock()
	defer h.mx.Unlock()

	args := strings.Split(r.URL.RawQuery, "&")
	h.Args = args
}

func (h *LoaderHandler) HandleOKRequests(w http.ResponseWriter, r *http.Request) {
	atomic.AddUint64(&h.Stats.RequestCount, 1)
	fmt.Fprintf(w, "OK")
}

func (h *LoaderHandler) HandleLongRequests(w http.ResponseWriter, r *http.Request) {
	time.Sleep(10 * time.Second)
	w.WriteHeader(http.StatusOK)
}

func (h *LoaderHandler) HandleAbortRequests(w http.ResponseWriter, r *http.Request) {
	atomic.AddUint64(&h.Stats.RequestCount, 1)
	w.WriteHeader(http.StatusNotFound)
}

// HandleMixedRequests first n requests will failed with 404
func (h *LoaderHandler) HandleMixedRequests(w http.ResponseWriter, r *http.Request) {
	h.mx.Lock()
	defer h.mx.Unlock()

	if int(h.Stats.RequestCount) < h.MixedFailedRequests {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	h.Stats.RequestCount++
}

func NewServer(mixedFailedRequests int) (*LoaderHandler, *httptest.Server) {
	mux := http.NewServeMux()
	h := &LoaderHandler{
		MixedFailedRequests: mixedFailedRequests,
	}

	mux.HandleFunc("/ok", h.HandleOKRequests)
	mux.HandleFunc("/mixed", h.HandleMixedRequests)
	mux.HandleFunc("/args", h.HandleArgsRequest)
	mux.HandleFunc("/body", h.HandleBodyRequests)
	mux.HandleFunc("/header", h.HandleHeaderRequests)
	mux.HandleFunc("/abort", h.HandleAbortRequests)
	mux.HandleFunc("/long", h.HandleLongRequests)

	ts := httptest.NewServer(mux)

	return h, ts
}

// NewBenchmarkServer will start a docker container with nginx
func NewBenchmarkServer(ctx context.Context, wg *sync.WaitGroup, readyChan chan struct{}, errChan chan error) {
	go func() {
		defer wg.Done()
		pool, err := dockertest.NewPool("")
		if err != nil {
			errChan <- err
			return
		}

		resource, err := pool.RunWithOptions(&dockertest.RunOptions{
			Name:       "nginx",
			Repository: "nginx",
			Tag:        "latest",
			Env:        []string{"NGINX_PORT=8080"},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"80": {
					docker.PortBinding{
						HostIP:   "0.0.0.0",
						HostPort: "8080",
					},
				},
			},
			ExposedPorts: []string{"80"},
		})

		if errors.Is(err, docker.ErrContainerAlreadyExists) {
			readyChan <- struct{}{}
			return
		}

		if err != nil {
			errChan <- err
			return
		}

		time.Sleep(5 * time.Second)

		readyChan <- struct{}{}

		<-ctx.Done()
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purse resources")
		}
	}()
}
