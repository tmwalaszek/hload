package loader

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/tmwalaszek/hload/model"

	"github.com/valyala/fasthttp"
)

type LoaderFastHTTP struct {
	client *fasthttp.Client
	opts   *model.Loader
}

func NewLoaderFastHTTP(opts *model.Loader) (*LoaderFastHTTP, error) {
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

	client := &fasthttp.Client{
		Name:                "hload",
		MaxConnsPerHost:     opts.Connections,
		ReadTimeout:         opts.ReadTimeout,
		WriteTimeout:        opts.WriteTimeout,
		MaxIdleConnDuration: opts.KeepAlive,
		TLSConfig:           &tlsConfig,
	}

	return &LoaderFastHTTP{
		opts:   opts,
		client: client,
	}, nil
}

func (l *LoaderFastHTTP) Request() *model.RequestStat {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	args := fasthttp.AcquireArgs()

	req.SetRequestURI(l.opts.URL)
	req.Header.SetMethod(l.opts.Method)

	// Set all Headers into Request
	for key, value := range l.opts.Headers {
		if len(value) > 1 {
			for _, v := range value {
				req.Header.Add(key, v)
			}
		}
		if len(value) == 1 {
			req.Header.Set(key, value[0])
		}
	}

	if len(l.opts.Parameters) > 0 {
		r := rand.Intn(len(l.opts.Parameters))

		// Set args if any
		for key, value := range l.opts.Parameters[r] {
			args.Add(key, value)
		}

		if l.opts.Method == fasthttp.MethodGet {
			reqArgs := req.URI().QueryArgs()
			args.CopyTo(reqArgs)
		} else if l.opts.Method == fasthttp.MethodPost || l.opts.Method == fasthttp.MethodPut {
			reqArgs := req.PostArgs()
			args.CopyTo(reqArgs)
		}
	}

	if len(l.opts.Body) != 0 && (l.opts.Method == fasthttp.MethodPost || l.opts.Method == fasthttp.MethodPut) {
		req.SetBody(l.opts.Body)
	}

	start := time.Now()
	err := l.client.Do(req, resp)

	bodySize := len(resp.Body())

	end := time.Now()
	duration := time.Since(start)

	statusCode := resp.StatusCode()

	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(resp)
	fasthttp.ReleaseArgs(args)

	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	}

	return &model.RequestStat{
		Start:    start,
		End:      end,
		Duration: duration,
		BodySize: bodySize,
		RetCode:  statusCode,
		Error:    errorMsg,
	}
}
