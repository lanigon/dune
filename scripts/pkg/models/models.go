package models

import (
	"strconv"
	"strings"
	"time"
)

// ExecuteResponse is returned by execute query/SQL endpoints.
type ExecuteResponse struct {
	ExecutionID string `json:"execution_id"`
	State       string `json:"state"`
}

// ExecutionStatus is returned by the status endpoint.
type ExecutionStatus struct {
	ExecutionID         string     `json:"execution_id"`
	QueryID             int        `json:"query_id"`
	State               string     `json:"state"`
	IsExecutionFinished bool       `json:"is_execution_finished"`
	SubmittedAt         time.Time  `json:"submitted_at"`
	ExpiresAt           time.Time  `json:"expires_at"`
	ExecutionStartedAt  *time.Time `json:"execution_started_at,omitempty"`
	ExecutionEndedAt    *time.Time `json:"execution_ended_at,omitempty"`
	ExecutionCostCredits float64   `json:"execution_cost_credits"`
	Error               *ExecError `json:"error,omitempty"`
}

type ExecError struct {
	Type     string          `json:"type"`
	Message  string          `json:"message"`
	Metadata *ErrorMetadata  `json:"metadata,omitempty"`
}

type ErrorMetadata struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// QueryResult is the top-level response for result endpoints.
type QueryResult struct {
	ExecutionID         string          `json:"execution_id"`
	QueryID             int             `json:"query_id"`
	State               string          `json:"state"`
	IsExecutionFinished bool            `json:"is_execution_finished"`
	SubmittedAt         time.Time       `json:"submitted_at"`
	ExpiresAt           time.Time       `json:"expires_at"`
	NextOffset          *int            `json:"next_offset,omitempty"`
	NextURI             *string         `json:"next_uri,omitempty"`
	Result              *ResultData     `json:"result,omitempty"`
	Error               *ExecError      `json:"error,omitempty"`
}

type ResultData struct {
	Rows     []map[string]interface{} `json:"rows"`
	Metadata ResultMetadata           `json:"metadata"`
}

type ResultMetadata struct {
	ColumnNames         []string `json:"column_names"`
	ColumnTypes         []string `json:"column_types"`
	RowCount            int      `json:"row_count"`
	TotalRowCount       int      `json:"total_row_count"`
	ResultSetBytes      int      `json:"result_set_bytes"`
	TotalResultSetBytes int      `json:"total_result_set_bytes"`
	DatapointCount      int      `json:"datapoint_count"`
	ExecutionTimeMillis int      `json:"execution_time_millis"`
	PendingTimeMillis   int      `json:"pending_time_millis"`
}

// Execution states.
const (
	StatePending          = "QUERY_STATE_PENDING"
	StateExecuting        = "QUERY_STATE_EXECUTING"
	StateCompleted        = "QUERY_STATE_COMPLETED"
	StateFailed           = "QUERY_STATE_FAILED"
	StateCanceled         = "QUERY_STATE_CANCELED"
	StateExpired          = "QUERY_STATE_EXPIRED"
	StateCompletedPartial = "QUERY_STATE_COMPLETED_PARTIAL"
)

// IsTerminal returns true if the state is a terminal state.
func (s ExecutionStatus) IsTerminal() bool {
	switch s.State {
	case StateCompleted, StateFailed, StateCanceled, StateExpired, StateCompletedPartial:
		return true
	}
	return false
}

// QueryParameter for parameterized queries.
type QueryParameter struct {
	Key         string   `json:"key"`
	Value       string   `json:"value"`
	Type        string   `json:"type"` // text, number, datetime, enum
	EnumOptions []string `json:"enumOptions,omitempty"`
}

// ExecuteRequest body for execute query.
type ExecuteRequest struct {
	QueryParameters map[string]interface{} `json:"query_parameters,omitempty"`
	Performance     string                 `json:"performance,omitempty"` // medium, large
}

// ExecuteSQLRequest body for execute SQL.
type ExecuteSQLRequest struct {
	SQL         string `json:"sql"`
	Performance string `json:"performance,omitempty"`
}

// CSVUploadRequest body for CSV upload.
type CSVUploadRequest struct {
	Data        string `json:"data"`
	TableName   string `json:"table_name"`
	Description string `json:"description,omitempty"`
	IsPrivate   bool   `json:"is_private"`
}

// CSVUploadResponse from upload endpoint.
type CSVUploadResponse struct {
	Success   bool   `json:"success"`
	TableName string `json:"table_name"`
	FullName  string `json:"full_name"`
}

// CreateTableRequest body for table creation.
type CreateTableRequest struct {
	Namespace   string         `json:"namespace"`
	TableName   string         `json:"table_name"`
	Description string         `json:"description,omitempty"`
	IsPrivate   bool           `json:"is_private"`
	Schema      []ColumnSchema `json:"schema"`
}

type ColumnSchema struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable,omitempty"`
}

// SearchRequest body for dataset search.
type SearchRequest struct {
	Query           string   `json:"query"`
	Blockchains     []string `json:"blockchains,omitempty"`
	Categories      []string `json:"categories,omitempty"`
	IncludeSchema   bool     `json:"include_schema,omitempty"`
	IncludeMetadata bool     `json:"include_metadata,omitempty"`
	Limit           int      `json:"limit,omitempty"`
	Offset          int      `json:"offset,omitempty"`
}

// ResultOption configures result fetching.
type ResultOption func(params map[string]string)

func WithLimit(n int) ResultOption {
	return func(p map[string]string) {
		p["limit"] = strconv.Itoa(n)
	}
}

func WithOffset(n int) ResultOption {
	return func(p map[string]string) {
		p["offset"] = strconv.Itoa(n)
	}
}

func WithFilters(f string) ResultOption {
	return func(p map[string]string) {
		p["filters"] = f
	}
}

func WithColumns(cols ...string) ResultOption {
	return func(p map[string]string) {
		p["columns"] = strings.Join(cols, ",")
	}
}

func WithSortBy(s string) ResultOption {
	return func(p map[string]string) {
		p["sort_by"] = s
	}
}

func WithAllowPartialResults() ResultOption {
	return func(p map[string]string) {
		p["allow_partial_results"] = "true"
	}
}
