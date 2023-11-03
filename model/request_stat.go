package model

import "time"

// RequestStat describe HTTP request status
type RequestStat struct {
	Start    time.Time     `json:"start" db:"start"`
	End      time.Time     `json:"end" db:"end"`
	Duration time.Duration `json:"duration" db:"duration"`

	BodySize int `json:"body_size" db:"body_size"`

	RetCode int    `json:"ret_code" db:"ret_code"`
	Error   string `json:"error" db:"error"`
}

// AggregatedStat provides a average request time within a timeframe from start to end
type AggregatedStat struct {
	Start    time.Time     `json:"start" db:"start"`
	End      time.Time     `json:"end" db:"end"`
	Duration time.Duration `json:"duration" db:"duration"`

	AvgRequestTime time.Duration `json:"avg_request_time" db:"avg_request_time"`
	MaxRequestTime time.Duration `json:"max_request_time" db:"max_request_time"`
	MinRequestTime time.Duration `json:"min_request_time" db:"min_request_time"`
	RequestCount   int           `json:"request_count" db:"request_count"`
}
