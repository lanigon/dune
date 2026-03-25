# Dune Go SDK 使用指南

## 官方 SDK

包：`github.com/duneanalytics/duneapi-client-go`
GitHub: https://github.com/duneanalytics/duneapi-client-go

```bash
go get github.com/duneanalytics/duneapi-client-go
```

### 基本用法

```go
package main

import (
    "fmt"
    "os"

    "github.com/duneanalytics/duneapi-client-go"
)

func main() {
    client := duneapi.New(os.Getenv("DUNE_API_KEY"))

    results, err := client.ExecuteQuery(1215383)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Got %d rows\n", len(results.Rows))
    for _, row := range results.Rows {
        fmt.Printf("%+v\n", row)
    }
}
```

---

## 自建 HTTP Client（推荐用于自定义需求）

官方 Go SDK 功能有限（主要支持执行端点），以下是直接调用 API 的模式。

### 执行查询

```go
func executeQuery(apiKey string, queryID int) (string, error) {
    url := fmt.Sprintf("https://api.dune.com/api/v1/query/%d/execute", queryID)
    req, _ := http.NewRequest("POST", url, nil)
    req.Header.Set("X-Dune-Api-Key", apiKey)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result struct {
        ExecutionID string `json:"execution_id"`
        State       string `json:"state"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    return result.ExecutionID, nil
}
```

### 执行裸 SQL

```go
func executeSQL(apiKey, sql string) (string, error) {
    payload := map[string]string{"sql": sql, "performance": "medium"}
    body, _ := json.Marshal(payload)

    req, _ := http.NewRequest("POST", "https://api.dune.com/api/v1/sql/execute", bytes.NewReader(body))
    req.Header.Set("X-Dune-Api-Key", apiKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result struct {
        ExecutionID string `json:"execution_id"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    return result.ExecutionID, nil
}
```

### 轮询状态

```go
func waitForExecution(apiKey, executionID string) error {
    url := fmt.Sprintf("https://api.dune.com/api/v1/execution/%s/status", executionID)

    for {
        req, _ := http.NewRequest("GET", url, nil)
        req.Header.Set("X-Dune-Api-Key", apiKey)

        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            return err
        }

        var status struct {
            State              string `json:"state"`
            IsExecutionFinished bool   `json:"is_execution_finished"`
        }
        json.NewDecoder(resp.Body).Decode(&status)
        resp.Body.Close()

        if status.IsExecutionFinished {
            if status.State == "QUERY_STATE_COMPLETED" {
                return nil
            }
            return fmt.Errorf("execution failed: %s", status.State)
        }

        time.Sleep(2 * time.Second)
    }
}
```

### 获取结果

```go
func getResults(apiKey, executionID string) ([]map[string]interface{}, error) {
    url := fmt.Sprintf("https://api.dune.com/api/v1/execution/%s/results", executionID)
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("X-Dune-Api-Key", apiKey)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        Result struct {
            Rows []map[string]interface{} `json:"rows"`
        } `json:"result"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    return result.Result.Rows, nil
}
```

### 获取最新缓存结果

```go
func getLatestResult(apiKey string, queryID int, opts ...QueryOption) ([]map[string]interface{}, error) {
    u, _ := url.Parse(fmt.Sprintf("https://api.dune.com/api/v1/query/%d/results", queryID))
    q := u.Query()
    for _, opt := range opts {
        opt(q)
    }
    u.RawQuery = q.Encode()

    req, _ := http.NewRequest("GET", u.String(), nil)
    req.Header.Set("X-Dune-Api-Key", apiKey)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        Result struct {
            Rows []map[string]interface{} `json:"rows"`
        } `json:"result"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    return result.Result.Rows, nil
}

type QueryOption func(url.Values)

func WithLimit(n int) QueryOption {
    return func(v url.Values) { v.Set("limit", strconv.Itoa(n)) }
}

func WithOffset(n int) QueryOption {
    return func(v url.Values) { v.Set("offset", strconv.Itoa(n)) }
}

func WithFilters(f string) QueryOption {
    return func(v url.Values) { v.Set("filters", f) }
}

func WithColumns(cols ...string) QueryOption {
    return func(v url.Values) { v.Set("columns", strings.Join(cols, ",")) }
}

func WithSortBy(s string) QueryOption {
    return func(v url.Values) { v.Set("sort_by", s) }
}
```

### 上传 CSV

```go
func uploadCSV(apiKey, tableName, description, csvData string) error {
    payload := map[string]interface{}{
        "data":        csvData,
        "table_name":  tableName,
        "description": description,
        "is_private":  false,
    }
    body, _ := json.Marshal(payload)

    req, _ := http.NewRequest("POST", "https://api.dune.com/api/v1/uploads/csv", bytes.NewReader(body))
    req.Header.Set("X-Dune-Api-Key", apiKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        b, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(b))
    }
    return nil
}
```

---

## 完整工作流示例

```go
func main() {
    apiKey := os.Getenv("DUNE_API_KEY")

    // 方式1：获取已保存查询的缓存结果（最省 credits）
    rows, _ := getLatestResult(apiKey, 3493826,
        WithLimit(100),
        WithFilters("blockchain = 'ethereum' AND amount_usd > 10000"),
        WithColumns("tx_hash", "amount_usd", "project"),
        WithSortBy("amount_usd desc"),
    )

    // 方式2：执行裸 SQL
    execID, _ := executeSQL(apiKey, "SELECT * FROM dex.trades WHERE block_time > now() - interval '1' day LIMIT 10")
    waitForExecution(apiKey, execID)
    rows, _ = getResults(apiKey, execID)

    // 方式3：执行已保存查询
    execID, _ = executeQuery(apiKey, 3493826)
    waitForExecution(apiKey, execID)
    rows, _ = getResults(apiKey, execID)
}
```

---

## 最佳实践

1. **优先用 `get_latest_result`**：不触发执行，最省 credits
2. **选择性返回列**：用 `columns` 参数减少数据量和 credit 消耗
3. **过滤时间范围**：减少扫描数据量
4. **使用分区键过滤**：`blockchain` + `block_month` 用于 dex.trades
5. **处理分页**：大结果集用 limit/offset 分页获取
6. **轮询间隔 2-5 秒**：避免触发限速
7. **环境变量存放 API Key**：不要硬编码
