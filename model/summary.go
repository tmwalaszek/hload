package model

import "time"

type Summary struct {
	UUID string `db:"uuid" json:"id"`
	URL  string `db:"url" json:"url"`

	Description string `db:"description" json:"description"`

	Start     time.Time     `db:"start" json:"start"`
	End       time.Time     `db:"end" json:"end"`
	TotalTime time.Duration `db:"total_time" json:"total_time"`

	ReqCount        int `db:"requests_count" json:"requests_count"`
	SuccessReq      int `db:"success_req" json:"success_req"` // Requests with return code 2x
	FailReq         int `db:"fail_req" json:"fail_req"`       // Requests with return code != 2x
	DataTransferred int `db:"data_transferred" json:"data_transferred"`

	ReqPerSec float64 `db:"req_per_sec" json:"req_per_sec"` // Request per second

	AvgReqTime time.Duration `db:"avg_req_time" json:"avg_req_time"` // Average request time
	MinReqTime time.Duration `db:"min_req_time" json:"min_req_time"` // Min request time
	MaxReqTime time.Duration `db:"max_req_time" json:"max_req_time"` // Max request time

	P50ReqTime time.Duration `db:"p50_req_time" json:"p_50_req_time"` // 50th percentile
	P75ReqTime time.Duration `db:"p75_req_time" json:"p_75_req_time"` // 75th percentile
	P90ReqTime time.Duration `db:"p90_req_time" json:"p_90_req_time"` // 90th percentile
	P99ReqTime time.Duration `db:"p99_req_time" json:"p_99_req_time"` // 99th percentile

	StdDeviation float64 `db:"std_deviation" json:"std_deviation"` // Standard deviation

	LoaderConf string `db:"loader_uuid" json:"-"`

	Errors    map[string]int `json:"errors,omitempty"`
	HTTPCodes map[int]int    `json:"http_codes,omitempty"`

	AggregatedStats []*AggregatedStat `json:"aggregated_stats,omitempty"`
	RequestStats    []*RequestStat    `json:"request_stats,omitempty"`
}
