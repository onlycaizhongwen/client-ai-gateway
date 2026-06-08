# Tool / Plugin SDK 边界

本文说明如何为客户端 AI 网关新增只读工具，以及工具 Manifest、scope、输入输出 schema、审计和测试约定。当前版本仍处于只读 MVP，不支持任意本地命令、写操作或真实 MCP Server 执行。

## 当前定位

Tool / Plugin SDK 面向三类扩展：

- 内置只读工具：由 `internal/tools` 注册并可通过 `/gateway/v1/tools/:id/invoke` 执行。
- MCP Manifest 工具：只进入目录和权限体系，`origin=mcp`，当前不可执行。
- 未来插件包：建议按本文的目录、签名、权限和沙箱约定演进。

当前真实可执行 adapter 只有 `runtime-health`。新增可执行工具必须先进入仓库内部实现，不能通过配置加载任意二进制或脚本。

## 插件目录建议

未来外部插件包建议使用如下结构：

```text
plugins/
  vendor.plugin-id/
    plugin.json
    README.md
    checksums.txt
    bin/
      plugin.exe
    schemas/
      input.json
      output.json
```

当前 MVP 不读取该目录。它只作为后续企业分发和签名校验的目录约定。

## Manifest 字段

内置工具配置位于 `configs/*.json` 的 `tools[]`。MCP 占位工具位于 `mcp_runtime.servers[].tools[]`。

| 字段 | 说明 |
| --- | --- |
| `id` | 工具唯一 ID，建议使用反向域名或产品前缀，例如 `gateway.runtime_health`。 |
| `name` | 控制台展示名称。 |
| `adapter` | 内置工具适配器名称，当前仅支持 `runtime-health`。 |
| `description` | 工具用途说明。 |
| `read_only` | 当前必须为 `true`。 |
| `risk_level` | `low`、`medium`、`high`，默认建议显式填写。 |
| `scopes` | 最小权限 scope，不能为空。 |
| `input_schema` | JSON Schema 风格输入描述。 |
| `output_schema` | JSON Schema 风格输出描述。 |
| `sandbox_required` | 当前必须为 `false`，真实沙箱运行时完成后才能启用。 |
| `enabled` | 启停开关，缺省为启用。 |

示例：

```json
{
  "id": "gateway.runtime_health",
  "name": "Runtime Health",
  "adapter": "runtime-health",
  "description": "Read gateway runtime health snapshot.",
  "read_only": true,
  "risk_level": "low",
  "scopes": ["runtime.read"],
  "input_schema": {
    "type": "object",
    "additionalProperties": false
  },
  "output_schema": {
    "type": "object"
  },
  "sandbox_required": false,
  "enabled": true
}
```

## 核心接口

工具实现位于 `internal/tools`，稳定契约如下：

```go
type Tool interface {
    ID() string
    Manifest() Manifest
    Invoke(context.Context, Input) (Result, error)
}

type Input struct {
    AppID     string
    ToolID    string
    Arguments map[string]any
}

type Result struct {
    ToolID     string
    TraceID    string
    AppID      string
    DurationMS int64
    Output     any
}
```

约定：

- `ID()` 必须与 Manifest `id` 一致。
- `Manifest()` 必须返回完整权限、风险和 schema 信息。
- `Invoke()` 只接收已通过 Access 层授权的请求，但仍应自行校验输入。
- 工具不得读取或返回 App Token、Provider API Key、Authorization header。
- 工具不得绕过 `context.Context`，超时或取消后应尽快返回。

## Scope 与权限

工具调用需要应用具备以下任一 grant：

- `tool`：工具通配权限，适合开发或高信任应用。
- `tool:<scope>`：细粒度权限，例如 `tool:runtime.read`。

配置校验要求：

- `scopes` 不能为空。
- scope 只能包含小写字母、数字、`.`、`_`、`-`。
- 同一个工具内 scope 不能重复。
- 缺少权限时返回 `tool_scope_denied`，并写入 Trace/Audit。

推荐 scope 命名：

| 类型 | 示例 | 说明 |
| --- | --- | --- |
| Runtime | `runtime.read` | 读取网关状态。 |
| Desktop | `desktop.read` | 读取桌面上下文，当前只用于 MCP Manifest。 |
| Developer | `dev.read` | 读取本地开发环境信息。 |
| Enterprise | `enterprise.policy.read` | 读取企业策略或目录信息。 |

## 只读边界

当前 MVP 的硬边界：

- `read_only` 必须为 `true`。
- `sandbox_required` 必须为 `false`。
- 不执行任意本地命令。
- 不启动真实 MCP Server。
- 不允许写文件、改注册表、改系统设置、发起破坏性操作。
- 不把完整敏感输入写入 Trace snapshot 或 Audit metadata。

如果未来引入可写工具，必须先完成沙箱进程模型、授权弹窗、签名校验、审计扩展和回滚机制。

## 错误约定

工具错误应返回 `tools.Error`：

```go
return tools.Result{}, tools.NewError(tools.ErrCodeUnavailable, "runtime health is unavailable", err)
```

当前稳定错误码：

| Code | 说明 |
| --- | --- |
| `tool_unavailable` | 工具注册表不可用、依赖不可用、MCP manifest-only 工具不可执行。 |
| `tool_failed` | 工具执行失败或未分类错误。 |
| `tool_denied` | 工具不满足只读 MVP 安全边界。 |
| `tool_scope_denied` | App Token 缺少所需 scope。 |

Access 层会把错误写入 HTTP 响应、Trace 和 `tool.invoke` Audit 事件。

## Trace 与 Audit

每次工具调用都会生成 Trace：

- `request_type=tool`
- `tool_id`
- `app_id`
- `status`
- `duration_ms`
- `tool_invocation_started`
- `tool_invocation_completed` 或 `tool_invocation_failed`

Audit 事件使用：

- `action=tool.invoke`
- `target=<tool_id>`
- `result=success|failed|denied`
- `metadata.required_scopes`
- `metadata.matched_grant`
- `metadata.missing_grants`
- `metadata.origin`
- `metadata.sandbox_required`

工具输出可以返回给调用方，但不应把大体积或敏感输出完整塞入 Audit metadata。

## 注册流程

新增内置只读工具：

1. 在 `internal/tools` 实现 `Tool`。
2. 在 `NewRegistryFromConfig` 中增加 adapter 分支。
3. 在 `config.Config.Validate` 中允许新 adapter，并保持只读、scope、risk、sandbox 校验。
4. 在 `configs/dev.json` 增加工具 Manifest。
5. 给需要调用的 app 增加 `tool:<scope>` grant。
6. 如新增稳定错误码，同步更新 `docs/error-codes.md`。
7. 在 README 或本文补充调用示例。

## 测试要求

至少补充：

- Registry 测试：Manifest 字段、禁用工具、未知 adapter。
- Config 测试：`read_only=false`、空 scope、非法 scope、重复 scope、非法 risk、`sandbox_required=true`。
- Access 测试：无 token、缺 scope、非只读拒绝、MCP manifest-only 不可执行、成功调用写入 Trace/Audit。
- Tool 测试：输入校验、依赖不可用、context cancel/timeout、输出 schema 关键字段。
- 安全测试：错误和输出不包含 token、API Key、Authorization header。

提交前运行：

```powershell
go test ./internal/tools
go test ./internal/config
go test ./internal/access
go test ./...
go build ./cmd/gateway-daemon
```

## 当前限制

- 没有独立发布的外部插件 SDK 包。
- 没有插件签名、来源校验和 checksum enforcement。
- 没有真实沙箱进程模型。
- 没有可写工具授权和回滚机制。
- MCP 仍是 Manifest-only，占位工具不可执行。
