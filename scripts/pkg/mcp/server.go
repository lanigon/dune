package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/bergtatt/morpheco/scripts/pkg/dune"
	"github.com/bergtatt/morpheco/scripts/pkg/models"
)

// Server is an MCP server that wraps a Dune API client.
type Server struct {
	client *dune.Client
	tools  map[string]ToolHandler
}

type ToolHandler func(args map[string]interface{}) *CallToolResult

// NewServer creates a new MCP server.
func NewServer(client *dune.Client) *Server {
	s := &Server{client: client, tools: make(map[string]ToolHandler)}
	s.registerTools()
	return s
}

func (s *Server) registerTools() {
	s.tools["execute_sql"] = s.handleExecuteSQL
	s.tools["get_latest_result"] = s.handleGetLatestResult
	s.tools["execute_query"] = s.handleExecuteQuery
	s.tools["search_datasets"] = s.handleSearchDatasets
	s.tools["get_execution_status"] = s.handleGetExecutionStatus
}

// Run starts the MCP server on stdio (JSON-RPC 2.0 over stdin/stdout).
func (s *Server) Run() error {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read stdin: %w", err)
		}

		line = trimLine(line)
		if len(line) == 0 {
			continue
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			s.writeError(nil, -32700, "parse error", nil)
			continue
		}

		resp := s.handleRequest(&req)
		if resp != nil {
			s.writeResponse(resp)
		}
	}
}

func (s *Server) handleRequest(req *Request) *Response {
	switch req.Method {
	case "initialize":
		return s.success(req.ID, InitializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities: ServerCapabilities{
				Tools: &ToolsCapability{},
			},
			ServerInfo: ServerInfo{
				Name:    "dune-local",
				Version: "0.1.0",
			},
		})

	case "notifications/initialized":
		return nil // no response for notifications

	case "tools/list":
		return s.success(req.ID, ToolsListResult{Tools: s.toolDefinitions()})

	case "tools/call":
		var params CallToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return s.error(req.ID, -32602, "invalid params", nil)
		}
		handler, ok := s.tools[params.Name]
		if !ok {
			return s.error(req.ID, -32602, fmt.Sprintf("unknown tool: %s", params.Name), nil)
		}
		result := handler(params.Arguments)
		return s.success(req.ID, result)

	case "ping":
		return s.success(req.ID, map[string]interface{}{})

	default:
		// Ignore unknown notifications (method starts with "notifications/")
		if strings.HasPrefix(req.Method, "notifications/") {
			return nil
		}
		return s.error(req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method), nil)
	}
}

func (s *Server) toolDefinitions() []Tool {
	return []Tool{
		{
			Name:        "execute_sql",
			Description: "Execute raw DuneSQL and wait for results. Use for ad-hoc blockchain data queries.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"sql": map[string]interface{}{
						"type":        "string",
						"description": "DuneSQL query to execute",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Max rows to return (default 100)",
					},
				},
				"required": []string{"sql"},
			},
		},
		{
			Name:        "get_latest_result",
			Description: "Get the latest cached result of a saved Dune query. Does not trigger execution. Cheapest way to get data.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query_id": map[string]interface{}{
						"type":        "integer",
						"description": "Dune query ID",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Max rows to return (default 100)",
					},
					"filters": map[string]interface{}{
						"type":        "string",
						"description": "SQL WHERE-like filter expression (e.g. \"blockchain = 'ethereum' AND amount_usd > 1000\")",
					},
					"columns": map[string]interface{}{
						"type":        "string",
						"description": "Comma-separated column names to return",
					},
					"sort_by": map[string]interface{}{
						"type":        "string",
						"description": "ORDER BY expression (e.g. \"amount_usd desc\")",
					},
				},
				"required": []string{"query_id"},
			},
		},
		{
			Name:        "execute_query",
			Description: "Execute a saved Dune query by ID and wait for results.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query_id": map[string]interface{}{
						"type":        "integer",
						"description": "Dune query ID",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Max rows to return (default 100)",
					},
				},
				"required": []string{"query_id"},
			},
		},
		{
			Name:        "search_datasets",
			Description: "Search Dune's data catalog for tables by keyword, blockchain, or category.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query (e.g. \"dex trades\", \"uniswap events\")",
					},
					"blockchains": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Filter by blockchain names (e.g. [\"ethereum\", \"polygon\"])",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "get_execution_status",
			Description: "Check the status of a query execution. Free, does not consume credits.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"execution_id": map[string]interface{}{
						"type":        "string",
						"description": "Execution ID returned by execute_sql or execute_query",
					},
				},
				"required": []string{"execution_id"},
			},
		},
	}
}

// --- Tool handlers ---

func (s *Server) handleExecuteSQL(args map[string]interface{}) *CallToolResult {
	sql, _ := args["sql"].(string)
	if sql == "" {
		return ErrorResult("sql is required")
	}

	limit := 100
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	result, err := s.client.RunSQL(sql, models.WithLimit(limit))
	if err != nil {
		return ErrorResult(fmt.Sprintf("execute sql failed: %v", err))
	}
	return s.formatQueryResult(result)
}

func (s *Server) handleGetLatestResult(args map[string]interface{}) *CallToolResult {
	queryID, err := getIntArg(args, "query_id")
	if err != nil {
		return ErrorResult("query_id (integer) is required")
	}

	var opts []models.ResultOption
	if limit, ok := args["limit"].(float64); ok {
		opts = append(opts, models.WithLimit(int(limit)))
	} else {
		opts = append(opts, models.WithLimit(100))
	}
	if filters, ok := args["filters"].(string); ok && filters != "" {
		opts = append(opts, models.WithFilters(filters))
	}
	if columns, ok := args["columns"].(string); ok && columns != "" {
		opts = append(opts, models.WithColumns(columns))
	}
	if sortBy, ok := args["sort_by"].(string); ok && sortBy != "" {
		opts = append(opts, models.WithSortBy(sortBy))
	}

	result, err := s.client.GetLatestResult(queryID, opts...)
	if err != nil {
		return ErrorResult(fmt.Sprintf("get latest result failed: %v", err))
	}
	return s.formatQueryResult(result)
}

func (s *Server) handleExecuteQuery(args map[string]interface{}) *CallToolResult {
	queryID, err := getIntArg(args, "query_id")
	if err != nil {
		return ErrorResult("query_id (integer) is required")
	}

	var opts []models.ResultOption
	if limit, ok := args["limit"].(float64); ok {
		opts = append(opts, models.WithLimit(int(limit)))
	} else {
		opts = append(opts, models.WithLimit(100))
	}

	result, err := s.client.RunQuery(queryID, opts...)
	if err != nil {
		return ErrorResult(fmt.Sprintf("execute query failed: %v", err))
	}
	return s.formatQueryResult(result)
}

func (s *Server) handleSearchDatasets(args map[string]interface{}) *CallToolResult {
	query, _ := args["query"].(string)
	if query == "" {
		return ErrorResult("query is required")
	}

	req := models.SearchRequest{
		Query:           query,
		IncludeMetadata: true,
		Limit:           20,
	}
	if chains, ok := args["blockchains"].([]interface{}); ok {
		for _, c := range chains {
			if s, ok := c.(string); ok {
				req.Blockchains = append(req.Blockchains, s)
			}
		}
	}

	raw, err := s.client.SearchDatasets(req)
	if err != nil {
		return ErrorResult(fmt.Sprintf("search datasets failed: %v", err))
	}
	return SuccessResult(string(raw))
}

func (s *Server) handleGetExecutionStatus(args map[string]interface{}) *CallToolResult {
	execID, _ := args["execution_id"].(string)
	if execID == "" {
		return ErrorResult("execution_id is required")
	}

	status, err := s.client.GetExecutionStatus(execID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("get status failed: %v", err))
	}

	b, _ := json.MarshalIndent(status, "", "  ")
	return SuccessResult(string(b))
}

// --- helpers ---

func (s *Server) formatQueryResult(result *models.QueryResult) *CallToolResult {
	if result == nil {
		return ErrorResult("no result")
	}
	if result.Error != nil {
		return ErrorResult(fmt.Sprintf("query error: %s", result.Error.Message))
	}

	b, _ := json.MarshalIndent(result, "", "  ")
	return SuccessResult(string(b))
}

func (s *Server) success(id json.RawMessage, result interface{}) *Response {
	return &Response{JSONRPC: "2.0", ID: id, Result: result}
}

func (s *Server) error(id json.RawMessage, code int, msg string, data interface{}) *Response {
	return &Response{JSONRPC: "2.0", ID: id, Error: &RPCError{Code: code, Message: msg, Data: data}}
}

func (s *Server) writeResponse(resp *Response) {
	b, _ := json.Marshal(resp)
	fmt.Fprintf(os.Stdout, "%s\n", b)
}

func (s *Server) writeError(id json.RawMessage, code int, msg string, data interface{}) {
	resp := &Response{JSONRPC: "2.0", ID: id, Error: &RPCError{Code: code, Message: msg, Data: data}}
	s.writeResponse(resp)
}

func trimLine(b []byte) []byte {
	for len(b) > 0 && (b[len(b)-1] == '\n' || b[len(b)-1] == '\r') {
		b = b[:len(b)-1]
	}
	return b
}

func getIntArg(args map[string]interface{}, key string) (int, error) {
	v, ok := args[key]
	if !ok {
		return 0, fmt.Errorf("%s is required", key)
	}
	switch val := v.(type) {
	case float64:
		return int(val), nil
	case string:
		return strconv.Atoi(val)
	default:
		return 0, fmt.Errorf("%s must be an integer", key)
	}
}
