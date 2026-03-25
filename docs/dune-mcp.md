# Dune MCP 使用指南

## 什么是 MCP

MCP（Model Context Protocol）是连接 AI 应用与外部系统的开源标准协议。Dune MCP 让 AI Agent 可以直接访问 Dune 的数据发现、查询执行和可视化能力。

## Dune 官方 MCP Server

- **Endpoint:** `https://api.dune.com/mcp/v1`
- **认证（Header）:** `x-dune-api-key: <your-key>`
- **认证（Query）:** `?api_key=<your-key>`

### 可用工具（12 个）

| 类别 | 工具 | 说明 |
|------|------|------|
| **Discovery** | `searchDocs` | 搜索 Dune 文档 |
| **Discovery** | `searchTables` | 按协议/链/分类搜索表 |
| **Discovery** | `listBlockchains` | 列出已索引的区块链 |
| **Discovery** | `searchTablesByContractAddress` | 按合约地址查找 decoded 表 |
| **Query** | `createDuneQuery` | 创建并保存新查询 |
| **Query** | `getDuneQuery` | 获取已有查询的 SQL 和元数据 |
| **Query** | `updateDuneQuery` | 更新查询的 SQL/标题/描述/标签/参数 |
| **Query** | `executeQueryById` | 执行已保存查询，返回 execution ID |
| **Query** | `getExecutionResults` | 获取执行状态和结果 |
| **Visualization** | `generateVisualization` | 从查询结果生成图表/计数器/表格 |
| **Account** | `getUsage` | 查看当前 billing 周期 credit 用量 |

## 接入各客户端

### Claude Code

```bash
claude mcp add --scope user --transport http dune https://api.dune.com/mcp/v1 \
  --header "x-dune-api-key: <dune-api-key>"
```

### Codex

```bash
codex mcp add dune_prod --url "https://api.dune.com/mcp/v1?api_key=<dune_api_key>"
```

**已知问题：** Codex 默认 tool timeout 60s，长时间查询会断连。解决方案：

```toml
[mcp_servers.dune]
url = "https://api.dune.com/mcp/v1?api_key=<dune_api_key>"
tool_timeout_sec = 300
```

### Cursor

```json
{
  "mcpServers": {
    "dune": {
      "url": "https://api.dune.com/mcp/v1",
      "headers": {
        "X-DUNE-API-KEY": "<api_key>"
      }
    }
  }
}
```

## Dune CLI & Skills

### 安装 CLI

```bash
curl -sSfL https://dune.com/cli/install.sh | sh
```

自动完成：安装 CLI + 运行 `dune auth` + 安装 Agent Skill。

### 手动认证

```bash
dune auth
# 或
export DUNE_API_KEY=<your-key>
# 或
dune query run 12345 --api-key <your-key>
```

配置存储：`~/.config/dune/config.yaml`

### CLI 用法

```bash
# 执行裸 SQL（JSON 输出，适合 agent 消费）
dune query run-sql --sql "SELECT * FROM ethereum.transactions LIMIT 5" -o json

# 搜索数据集
dune dataset search "dex trades ethereum"

# 查看帮助
dune --help
```

### Agent Skill 安装

```bash
npx skills add duneanalytics/skills
```

各 agent 的 skill 目录：

| Agent | 目录 |
|-------|------|
| Claude Code | `~/.claude/skills/` |
| Cursor | `~/.cursor/skills/` |
| OpenCode | `~/.config/opencode/skills/` |
| Codex | `~/.codex/skills/` |

### Skill 工作流程

安装后自动触发（当对话涉及区块链数据时）：
1. `dune dataset search` 发现正确的表
2. 检查列 schema 构建正确 SQL
3. 使用分区过滤写高效 DuneSQL
4. 执行查询并解析结果
5. 内置错误恢复策略

## MCP 协议实现（本地 Server）

我们在 `scripts/cmd/mcp/` 实现了本地 MCP server，包装了自建的 Dune client，
支持通过 stdio transport 与任何 MCP client 交互。

### 支持的工具

- `execute_sql` — 执行裸 SQL 并等待结果
- `get_latest_result` — 获取已保存查询的缓存结果
- `execute_query` — 执行已保存查询并等待结果
- `search_datasets` — 搜索 Dune 数据目录

### 用法

```bash
export DUNE_API_KEY=xxx
go run ./cmd/mcp

# Claude Code 接入本地 MCP server
claude mcp add --scope project --transport stdio dune-local -- go run ./cmd/mcp
```
