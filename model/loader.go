package model

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Loader provides general settings for the loader benchmark
// They are constant within all time of benchmark execution
type Loader struct {
	UUID        string `db:"uuid" json:"uuid,omitempty"`
	URL         string `db:"url" json:"url"`
	Name        string `db:"name" json:"name,omitempty"`
	Method      string `db:"method" json:"method,omitempty"`
	HTTPEngine  string `db:"http_engine" json:"http_engine,omitempty"`
	Description string `db:"description" json:"description,omitempty"`

	CreateDate time.Time `db:"create_date" json:"create_date,omitempty"`

	SkipVerify bool   `db:"skip_verify" json:"skip_verify,omitempty"`
	CA         []byte `db:"ca" json:"ca,omitempty"`
	Cert       []byte `db:"cert" json:"cert,omitempty"`
	Key        []byte `db:"key" json:"key,omitempty"`

	Body []byte `db:"body" json:"body,omitempty"`

	GatherFullRequestsStats      bool `json:"gather_full_requests_stats,omitempty" db:"gather_full_requests_stats"`
	GatherAggregateRequestsStats bool `json:"gather_aggregate_requests_stats,omitempty" db:"gather_aggregate_requests_stats"`

	AggregateWindow  time.Duration `db:"aggregate_window" json:"aggregate_window,omitempty"`
	BenchmarkTimeout time.Duration `db:"benchmark_timeout" json:"benchmark_timeout,omitempty"`

	Headers    Headers    `json:"headers,omitempty"`
	Parameters Parameters `json:"parameters,omitempty"`

	Tags []*LoaderTag `json:"tags,omitempty"`

	LoaderReqDetails
}

type LoaderReqDetails struct {
	ID          int64 `db:"id" json:"-"`
	ReqCount    int   `db:"request_count" json:"request_count,omitempty"`
	AbortAfter  int   `db:"abort_after" json:"abort_after,omitempty"`
	Connections int   `db:"connections" json:"connections,omitempty"`
	RateLimit   int   `db:"rate_limit" json:"rate_limit,omitempty"` // How many requests per second is allowed

	Duration     time.Duration `db:"duration" json:"duration,omitempty"`
	KeepAlive    time.Duration `db:"keep_alive" json:"keep_alive,omitempty"`
	RequestDelay time.Duration `db:"request_delay" json:"request_delay,omitempty"`
	ReadTimeout  time.Duration `db:"read_timeout" json:"read_timeout,omitempty"`
	WriteTimeout time.Duration `db:"write_timeout" json:"write_timeout,omitempty"`
	Timeout      time.Duration `db:"timeout" json:"timeout,omitempty"`

	LoaderConfigurationUUID string `db:"loader_uuid" json:"loader_uuid,omitempty"`
}

var ErrWrongHeaderFormat = errors.New("wrong header format")

type Headers map[string][]string

// Set header sets the header string into Headers map
func (h Headers) Set(header string) error {
	headerSplit := strings.SplitN(header, ":", 2)
	if len(headerSplit) != 2 {
		return ErrWrongHeaderFormat
	}

	headerName := strings.TrimSpace(headerSplit[0])
	headerValue := strings.TrimSpace(headerSplit[1])

	if strings.Contains(headerName, " ") {
		return ErrWrongHeaderFormat
	}

	_, ok := h[headerName]
	if !ok {
		h[headerName] = make([]string, 1)
		h[headerName][0] = headerValue

		return nil
	}

	h[headerName] = append(h[headerName], headerValue)

	return nil
}

type Parameters []map[string]string

// value needs to in format "key1=value2&key2=value2
func (p *Parameters) Set(value string) error {
	paramsMap := make(map[string]string)
	parameters := strings.Split(value, "&")

	for _, param := range parameters {
		keyValue := strings.Split(param, "=")
		if len(keyValue) != 2 {
			return fmt.Errorf("error parse parameter %s", value)
		}

		paramsMap[keyValue[0]] = keyValue[1]
	}

	*p = append(*p, paramsMap)
	return nil
}

type LoaderTag struct {
	Key        string    `db:"key" json:"key,omitempty"`
	Value      string    `db:"value" json:"value,omitempty"`
	CreateDate time.Time `db:"create_date" json:"create_date"`
	UpdateDate time.Time `db:"update_date" json:"update_date"`
}
