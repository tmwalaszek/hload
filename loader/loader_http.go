package loader

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/tmwalaszek/hload/model"

	"github.com/valyala/fasthttp"
)

type LoaderHTTP struct {
	client *http.Client
	opts   *model.Loader
}

func NewLoaderHTTP(opts *model.Loader) (*LoaderHTTP, error) {
	var tlsConfig tls.Config

	_, err := url.Parse(opts.URL)
	if err != nil {
		return nil, fmt.Errorf("bad url format: %w", err)
	}

	if opts.Connections == 0 {
		opts.Connections = DefaultConnection
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

	client := &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost: opts.Connections,
			IdleConnTimeout: opts.KeepAlive,
			TLSClientConfig: &tlsConfig,
		},
		Timeout: opts.Timeout,
	}

	return &LoaderHTTP{
		opts:   opts,
		client: client,
	}, nil
}

func (l *LoaderHTTP) Request() *model.RequestStat {
	var bodyReader *bytes.Reader
	var req *http.Request
	var err error

	if len(l.opts.Body) != 0 && (l.opts.Method == fasthttp.MethodPost || l.opts.Method == fasthttp.MethodPut) {
		bodyReader = bytes.NewReader(l.opts.Body)
		req, err = http.NewRequest(l.opts.Method, l.opts.URL, bodyReader)
	} else {
		req, err = http.NewRequest(l.opts.Method, l.opts.URL, nil)
	}

	if err != nil {
		return &model.RequestStat{
			Error: err.Error(),
		}
	}

	for key, value := range l.opts.Headers {
		if len(value) > 1 {
			for _, v := range value {
				req.Header.Set(key, v)
			}
		}

		if len(value) == 1 {
			req.Header.Set(key, value[0])
		}
	}

	if len(l.opts.Parameters) > 0 {
		r := rand.Intn(len(l.opts.Parameters))

		q := req.URL.Query()
		for key, value := range l.opts.Parameters[r] {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	start := time.Now()
	resp, err := l.client.Do(req)
	end := time.Now()
	duration := time.Since(start)

	if err != nil {
		return &model.RequestStat{
			Start:    start,
			End:      end,
			Duration: duration,
			Error:    err.Error(),
		}
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &model.RequestStat{
			Start:    start,
			End:      end,
			Duration: duration,
			RetCode:  resp.StatusCode,
			Error:    err.Error(),
		}
	}

	return &model.RequestStat{
		Start:    start,
		End:      end,
		Duration: duration,
		BodySize: len(body),
		RetCode:  resp.StatusCode,
	}
}
