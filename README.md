# 客户端 AI 网关

这是一个面向 AI PC、本地开发环境、企业桌面 AI 能力底座的客户端 AI 网关原型。当前实现提供 OpenAI 兼容聊天入口、Provider 路由与健康管理、策略试算、Trace/Audit 可观测能力、只读工具调用、中文优先的本地控制台。

## 当前能力

- OpenAI 兼容接口：`POST /v1/chat/completions`
- 本地控制台：`GET /console`
- Trace 查询、详情、导出和保留策略
- Audit 查询、分页、导出、按 `trace_id` 关联
- Provider 健康监控、启停、探测、模型目录
- Runtime Health 状态接口
- Policy dry-run 和 Routing explain
- OpenAI-compatible Provider 适配器
- 只读工具运行时 MVP
- 工具调用 Trace 化、Audit 关联、权限 scope 校验
- 中文优先、中英文切换控制台

## 启动

```powershell
go run ./cmd/gateway-daemon -config ./configs/dev.json
```

默认监听：

```text
127.0.0.1:18765
```

健康检查：

```powershell
curl http://127.0.0.1:18765/healthz
```

控制台地址：

```text
http://127.0.0.1:18765/console
```

## 快速聊天请求

```powershell
curl -X POST http://127.0.0.1:18765/v1/chat/completions `
  -H "Authorization: Bearer dev-token" `
  -H "Content-Type: application/json" `
  -d "{\"model\":\"local-small\",\"messages\":[{\"role\":\"user\",\"content\":\"你好\"}]}"
```

响应会包含 `trace_id`，可用于排障和审计关联。

## Trace

```powershell
curl http://127.0.0.1:18765/gateway/v1/traces
curl "http://127.0.0.1:18765/gateway/v1/traces?limit=20&offset=0&status=completed"
curl http://127.0.0.1:18765/gateway/v1/traces/{trace_id}
curl "http://127.0.0.1:18765/gateway/v1/traces/export?limit=500&status=failed" -o traces.jsonl
```

Trace 列表返回 `traces`、`total`、`offset`、`limit`。支持按 `status`、`app_id`、`provider_id` 过滤。

默认 Trace 存储为 `data/traces.jsonl`。可通过配置项 `trace_retention_max` 控制最多保留条数；`0` 或不配置表示不裁剪。

## Audit

```powershell
curl "http://127.0.0.1:18765/gateway/v1/audit/events?limit=20&offset=0" `
  -H "Authorization: Bearer admin-token"

curl "http://127.0.0.1:18765/gateway/v1/audit/events?trace_id={trace_id}" `
  -H "Authorization: Bearer admin-token"

curl "http://127.0.0.1:18765/gateway/v1/audit/events/export?limit=500&action=tool.invoke" `
  -H "Authorization: Bearer admin-token" `
  -o audit-events.jsonl
```

Audit 支持 `action`、`result`、`app_id`、`trace_id`、`limit`、`offset` 查询。默认持久化到 `data/audit.jsonl`，可通过 `audit_store_path` 调整路径，通过 `audit_retention_max` 控制保留条数。

## Provider 与模型目录

```powershell
curl http://127.0.0.1:18765/gateway/v1/providers
curl http://127.0.0.1:18765/gateway/v1/models
curl http://127.0.0.1:18765/gateway/v1/runtime/health
```

Provider 列表包含静态配置和运行时健康字段：`runtime_status`、`degraded_reason`、`last_checked_at`。当前状态包括：

- `healthy`
- `degraded`
- `unhealthy`
- `disabled`

模型目录会聚合 Provider 中声明的模型，并返回 Provider 元信息和可用性。使用 `?all=true` 可查看不可用模型。

## Provider 管理

```powershell
curl -X POST http://127.0.0.1:18765/gateway/v1/providers/local-mock/enabled `
  -H "Authorization: Bearer admin-token" `
  -H "Content-Type: application/json" `
  -d "{\"enabled\":false}"

curl -X POST http://127.0.0.1:18765/gateway/v1/providers/local-mock/probe `
  -H "Authorization: Bearer admin-token"
```

Provider 启停会写回配置文件并重新加载运行时快照。Provider 探测只更新运行时健康状态。

## 配置重载

```powershell
curl -X POST http://127.0.0.1:18765/gateway/v1/config/reload `
  -H "Authorization: Bearer admin-token"
```

配置重载会重新读取配置文件，重建 Provider adapter、Policy engine、Router、Pipeline 和 Provider health monitor，然后原子替换运行时快照。重载需要 `admin` grant。

## 策略试算与路由解释

Policy dry-run：

```powershell
curl -X POST http://127.0.0.1:18765/gateway/v1/policy/dry-run `
  -H "Content-Type: application/json" `
  -d "{\"app_id\":\"dev-app\",\"request_type\":\"chat\",\"data_labels\":[\"sensitive\"],\"model\":\"local-small\"}"
```

Routing explain：

```powershell
curl -X POST http://127.0.0.1:18765/gateway/v1/routing/explain `
  -H "Content-Type: application/json" `
  -d "{\"app_id\":\"dev-app\",\"request_type\":\"chat\",\"model\":\"local-small\",\"data_labels\":[\"sensitive\"]}"
```

策略规则支持轻量匹配 DSL。空匹配字段表示匹配全部。支持字段：

- `app_ids`
- `request_types`
- `models`
- `provider_classes`
- `data_labels`

支持效果：

- `allow`：允许请求，可走本地或云端。
- `deny`：路由前拒绝请求。
- `force_local`：允许请求，但禁止云端 Provider。
- `deny_cloud_for_sensitive`：敏感数据禁止云端降级的兼容效果。

示例：

```json
{
  "id": "desktop-local-only",
  "effect": "force_local",
  "reason": "桌面 AI 请求优先留在本地运行时",
  "app_ids": ["desktop-app"],
  "request_types": ["chat"],
  "models": ["local-large"],
  "provider_classes": ["cloud"]
}
```

## 工具调用

查看工具：

```powershell
curl http://127.0.0.1:18765/gateway/v1/tools
```

调用内置只读工具：

```powershell
curl -X POST http://127.0.0.1:18765/gateway/v1/tools/gateway.runtime_health/invoke `
  -H "Authorization: Bearer dev-token" `
  -H "Content-Type: application/json" `
  -d "{}"
```

当前 Phase 2 只开放只读工具。内置 `gateway.runtime_health` 会通过工具调用链路返回运行时健康快照。

工具调用要求应用具备：

- 宽权限：`tool`
- 或细粒度权限：例如 `tool:runtime.read`

工具响应包含：

- `trace_id`
- `app_id`
- `duration_ms`
- `output`

每次工具调用都会写入 Trace，并产生同一个 `trace_id` 关联的 `tool.invoke` Audit 事件。

## 工具 Manifest

当前工具配置字段：

- `id`
- `name`
- `adapter`
- `description`
- `read_only`
- `risk_level`
- `scopes`
- `input_schema`
- `output_schema`
- `sandbox_required`
- `enabled`

当前 MVP 约束：

- `read_only` 必须为 `true`
- `scopes` 不能为空
- `risk_level` 支持 `low`、`medium`、`high`
- `sandbox_required` 暂时必须为 `false`

## 新增只读工具

工具扩展走 `internal/tools` 包的稳定契约：

```go
type Tool interface {
    ID() string
    Manifest() tools.Manifest
    Invoke(context.Context, tools.Input) (tools.Result, error)
}
```

新增一个只读工具的步骤：

1. 在 `internal/tools` 中实现 `Tool` 接口。
2. 在 `NewRegistryFromConfig` 中按 adapter 名称注册工具。
3. 在 `configs/dev.json` 的 `tools` 数组中增加工具 manifest。
4. 为调用方 app 添加 `tool:<scope>` 或 broad `tool` grant。
5. 增加 registry contract test 和 access HTTP test。

工具错误应返回 `tools.Error`，并带稳定 `Code`。当前内置错误码：

- `tool_unavailable`
- `tool_failed`

Access 层会把工具错误码映射成 HTTP 错误响应、Trace 和 Audit 事件；新增只读工具通常不需要修改 access 层。

## MCP 运行时占位

当前版本支持在配置中声明 MCP server 和只读工具 manifest，用于提前打通企业桌面工具目录、权限 scope、风险等级和运行时健康展示。

重要边界：

- 只加载 manifest，不启动 MCP server。
- 不执行外部命令、不读取 command 字段。
- MCP 工具会出现在 `GET /gateway/v1/tools`，`origin` 为 `mcp`，`adapter` 为 `mcp-placeholder`。
- MCP 工具目前不会注册为可执行适配器，调用会返回稳定错误码 `tool_unavailable` 并写入 Trace/Audit。
- 所有 MCP 工具必须 `read_only=true`，且 `sandbox_required=false`。
- `mcp_runtime.mode` 当前只支持 `manifest_only` 或 `disabled`；`stdio`、`direct`、`sandboxed` 等真实执行模式会在配置加载时被拒绝。

查看 MCP 目录：

```powershell
curl http://127.0.0.1:18765/gateway/v1/mcp/servers
```

支持查询参数：

- `server_id`：只看指定 MCP server
- `scope`：只看包含指定 scope 的工具
- `enabled`：`true` 只看启用工具，`false` 只看禁用工具

响应包含：

- `enabled`
- `mode`
- `servers[].enabled`
- `servers[].tool_count`
- `servers[].enabled_tools`
- `servers[].tools[]`

配置示例：

```json
{
  "mcp_runtime": {
    "enabled": true,
    "mode": "manifest_only",
    "servers": [
      {
        "id": "desktop-context",
        "name": "Desktop Context MCP Placeholder",
        "enabled": true,
        "tools": [
          {
            "id": "mcp.desktop.list_context",
            "name": "Desktop Context List",
            "read_only": true,
            "risk_level": "low",
            "scopes": ["desktop.read"],
            "sandbox_required": false,
            "enabled": true
          }
        ]
      }
    ]
  }
}
```

运行时健康接口会返回 `mcp_runtime.status`、`mode`、server/tool 总数和启用数量。后续接入真实 MCP 适配器前，需要先补沙箱进程模型、授权弹窗、审计字段和 Provider SDK 边界。

## OpenAI-Compatible Provider

Provider 在 `configs/dev.json` 中注册。未配置 `adapter` 时默认使用内置 `mock` adapter。

OpenAI-compatible Provider 使用标准 `/v1/chat/completions` 上游路径。

示例：

```json
{
  "id": "local-openai",
  "name": "Local OpenAI Compatible",
  "class": "local",
  "adapter": "openai-compatible",
  "base_url": "http://127.0.0.1:1234",
  "api_key_env": "LOCAL_OPENAI_API_KEY",
  "models": ["local-small"],
  "healthy": true,
  "enabled": true
}
```

如果设置了 `api_key_env`，网关会从对应环境变量读取 API Key。环境变量缺失或为空时，daemon 仍可启动，但 Provider health 会报告 `missing_credential`，并通过正常健康监控进入 degraded/unhealthy。

LM Studio 示例：

```json
{
  "id": "lm-studio",
  "name": "LM Studio",
  "class": "local",
  "adapter": "openai-compatible",
  "base_url": "http://127.0.0.1:1234",
  "models": ["local-small"],
  "healthy": true,
  "enabled": true
}
```

Ollama OpenAI-compatible 示例：

```json
{
  "id": "ollama",
  "name": "Ollama",
  "class": "local",
  "adapter": "openai-compatible",
  "base_url": "http://127.0.0.1:11434",
  "models": ["llama3.1"],
  "healthy": true,
  "enabled": true
}
```

云端 OpenAI-compatible 示例：

```powershell
$env:OPENAI_API_KEY = "sk-..."
```

```json
{
  "id": "openai-cloud",
  "name": "OpenAI Compatible Cloud",
  "class": "cloud",
  "adapter": "openai-compatible",
  "base_url": "https://api.example.com",
  "api_key_env": "OPENAI_API_KEY",
  "models": ["gpt-compatible"],
  "healthy": true,
  "enabled": true
}
```

## 错误码

稳定错误码说明见：

```text
docs/error-codes.md
```

## 测试

```powershell
go test ./...
```
