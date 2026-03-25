package dune

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/bergtatt/morpheco/scripts/pkg/models"
)

const defaultBaseURL = "https://api.dune.com/api/v1"

// Client is a Dune API client.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// New creates a new Dune API client.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type Option func(*Client)

func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = u }
}

func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// ExecuteQuery triggers execution of a saved query.
func (c *Client) ExecuteQuery(queryID int, params map[string]interface{}, performance string) (*models.ExecuteResponse, error) {
	body := models.ExecuteRequest{
		QueryParameters: params,
		Performance:     performance,
	}
	return doJSON[models.ExecuteResponse](c, "POST", fmt.Sprintf("/query/%d/execute", queryID), body)
}

// ExecuteSQL executes raw SQL without saving a query.
func (c *Client) ExecuteSQL(sql string, performance string) (*models.ExecuteResponse, error) {
	body := models.ExecuteSQLRequest{
		SQL:         sql,
		Performance: performance,
	}
	return doJSON[models.ExecuteResponse](c, "POST", "/sql/execute", body)
}

// GetExecutionStatus checks the status of an execution (free, no credits).
func (c *Client) GetExecutionStatus(executionID string) (*models.ExecutionStatus, error) {
	return doJSON[models.ExecutionStatus](c, "GET", fmt.Sprintf("/execution/%s/status", executionID), nil)
}

// GetExecutionResult fetches results for a specific execution.
func (c *Client) GetExecutionResult(executionID string, opts ...models.ResultOption) (*models.QueryResult, error) {
	path := fmt.Sprintf("/execution/%s/results", executionID)
	return c.getWithOpts(path, opts)
}

// GetLatestResult fetches the latest cached result for a query (no execution triggered).
func (c *Client) GetLatestResult(queryID int, opts ...models.ResultOption) (*models.QueryResult, error) {
	path := fmt.Sprintf("/query/%d/results", queryID)
	return c.getWithOpts(path, opts)
}

// CancelExecution cancels a running execution.
func (c *Client) CancelExecution(executionID string) error {
	_, err := doJSON[struct{ Success bool }](c, "POST", fmt.Sprintf("/execution/%s/cancel", executionID), nil)
	return err
}

// WaitForExecution polls until the execution reaches a terminal state.
func (c *Client) WaitForExecution(executionID string, pollInterval time.Duration) (*models.ExecutionStatus, error) {
	if pollInterval == 0 {
		pollInterval = 2 * time.Second
	}
	for {
		status, err := c.GetExecutionStatus(executionID)
		if err != nil {
			return nil, err
		}
		if status.IsTerminal() {
			return status, nil
		}
		time.Sleep(pollInterval)
	}
}

// RunQuery executes a query and waits for results. Convenience method.
func (c *Client) RunQuery(queryID int, opts ...models.ResultOption) (*models.QueryResult, error) {
	exec, err := c.ExecuteQuery(queryID, nil, "medium")
	if err != nil {
		return nil, fmt.Errorf("execute: %w", err)
	}

	status, err := c.WaitForExecution(exec.ExecutionID, 0)
	if err != nil {
		return nil, fmt.Errorf("wait: %w", err)
	}

	if status.State != models.StateCompleted && status.State != models.StateCompletedPartial {
		errMsg := status.State
		if status.Error != nil {
			errMsg = status.Error.Message
		}
		return nil, fmt.Errorf("query failed: %s", errMsg)
	}

	return c.GetExecutionResult(exec.ExecutionID, opts...)
}

// RunSQL executes raw SQL and waits for results. Convenience method.
func (c *Client) RunSQL(sql string, opts ...models.ResultOption) (*models.QueryResult, error) {
	exec, err := c.ExecuteSQL(sql, "medium")
	if err != nil {
		return nil, fmt.Errorf("execute sql: %w", err)
	}

	status, err := c.WaitForExecution(exec.ExecutionID, 0)
	if err != nil {
		return nil, fmt.Errorf("wait: %w", err)
	}

	if status.State != models.StateCompleted && status.State != models.StateCompletedPartial {
		errMsg := status.State
		if status.Error != nil {
			errMsg = status.Error.Message
		}
		return nil, fmt.Errorf("sql failed: %s", errMsg)
	}

	return c.GetExecutionResult(exec.ExecutionID, opts...)
}

// UploadCSV uploads CSV data to create/replace a table.
func (c *Client) UploadCSV(req models.CSVUploadRequest) (*models.CSVUploadResponse, error) {
	return doJSON[models.CSVUploadResponse](c, "POST", "/uploads/csv", req)
}

// SearchDatasets searches the Dune data catalog.
func (c *Client) SearchDatasets(req models.SearchRequest) (json.RawMessage, error) {
	result, err := doJSON[json.RawMessage](c, "POST", "/datasets/search", req)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

// --- internal helpers ---

func (c *Client) getWithOpts(path string, opts []models.ResultOption) (*models.QueryResult, error) {
	params := make(map[string]string)
	for _, opt := range opts {
		opt(params)
	}

	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Dune-Api-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("dune api error (%d): %s", resp.StatusCode, string(body))
	}

	var result models.QueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}

func doJSON[T any](c *Client, method, path string, body interface{}) (*T, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Dune-Api-Key", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("dune api error (%d): %s", resp.StatusCode, string(b))
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}
