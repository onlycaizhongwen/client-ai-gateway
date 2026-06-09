# Tool / Plugin SDK 边界

本文说明如何为客户端 AI 网关新增只读工具，以及工具 Manifest、scope、输入输出 schema、审计和测试约定。当前版本仍处于只读 MVP，不支持任意本地命令、写操作或真实 MCP Server 执行。

## 当前定位

Tool / Plugin SDK 面向三类扩展：

- 内置只读工具：由 `internal/tools` 注册并可通过 `/gateway/v1/tools/:id/invoke` 执行。
- MCP Manifest 工具：只进入目录和权限体系，`origin=mcp`，当前不可执行。
- 未来插件包：建议按本文的目录、签名、权限和沙箱约定演进。

当前真实可执行 adapter 只有 `runtime-health`。新增可执行工具必须先进入仓库内部实现，不能通过配置加载任意二进制或脚本。

## 能力矩阵

| 扩展形态 | 当前状态 | 是否可执行 | 来源约束 | 适用场景 |
| --- | --- | --- | --- | --- |
| 内置只读工具 | 已支持 | 是 | 仓库代码编译进 daemon | 读取网关运行状态、低风险本地上下文。 |
| MCP Manifest 工具 | 已支持目录化 | 否 | 配置文件声明，只加载 manifest | 提前梳理企业桌面工具目录、scope 和审计边界。 |
| 外部插件包 | 设计约定 | 否 | 未来要求签名、checksum、来源 allowlist | 企业分发、第三方工具生态。 |
| 可写工具 | 未支持 | 否 | 必须先有沙箱、授权弹窗和回滚 | 文件写入、系统设置、自动化修复等高风险动作。 |
| 任意本地命令 | 明确禁止 | 否 | 不允许通过配置开放 | 避免客户端网关变成无边界命令执行器。 |

当前工具能力只覆盖 read-only、request/response、JSON 输入输出。流式工具输出、长任务、后台任务、可取消任务队列和用户交互式授权仍属于后续设计范围。

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

外部插件包 `plugin.json` 建议模板：

```json
{
  "id": "vendor.desktop-context",
  "name": "Desktop Context",
  "version": "0.1.0",
  "publisher": "Vendor Inc.",
  "homepage": "https://example.com/plugins/desktop-context",
  "entry": {
    "type": "native",
    "path": "bin/plugin.exe"
  },
  "tools": [
    {
      "id": "vendor.desktop_context.list",
      "name": "List Desktop Context",
      "description": "Read a summarized desktop context.",
      "read_only": true,
      "risk_level": "low",
      "scopes": ["desktop.read"],
      "input_schema": {
        "type": "object",
        "additionalProperties": false
      },
      "output_schema": {
        "type": "object"
      },
      "sandbox_required": true,
      "enabled": true
    }
  ],
  "signature": {
    "algorithm": "ed25519",
    "key_id": "vendor-2026-01",
    "value": "<base64-signature>"
  }
}
```

注意：上述模板是未来插件包格式，不代表当前 daemon 会加载或执行 `entry.path`。

## 来源、签名与校验设计

未来启用外部插件前，建议按以下顺序失败关闭：

1. 只扫描管理员配置的插件目录，不递归任意用户目录。
2. 校验 `plugin.json` schema、插件 `id`、工具 `id`、版本号和 publisher。
3. 校验 `checksums.txt`，确认 `plugin.json`、schemas 和二进制文件未被篡改。
4. 校验签名，签名主体应覆盖 manifest、checksum 文件和入口文件摘要。
5. 校验来源 allowlist，例如 `publisher`、`key_id`、插件 `id` 前缀。
6. 校验权限：所有工具必须有 scope，且 scope 不得越过企业策略允许范围。
7. 校验沙箱：凡是外部二进制或 MCP 真实执行，必须 `sandbox_required=true`。
8. 任一环节失败时不注册工具，并写入 Audit / runtime health issue。

建议的信任模型：

| 级别 | 来源 | 默认策略 |
| --- | --- | --- |
| `builtin` | 随 daemon 编译 | 可注册，只读工具可执行。 |
| `enterprise_signed` | 企业签名和 allowlist | 可进入目录，是否执行取决于沙箱能力。 |
| `vendor_signed` | 第三方签名 | 默认只进入待审批状态。 |
| `unsigned` | 无签名或校验失败 | 不注册，不执行。 |

当前版本没有实现签名校验，因此所有外部插件都应视为不可加载。

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

## 沙箱阶段模型

沙箱能力建议分阶段交付：

| 阶段 | 能力 | 允许动作 | 退出条件 |
| --- | --- | --- | --- |
| `manifest_only` | 只加载目录 | 不执行 | 当前已采用，MCP 占位工具返回 `tool_unavailable`。 |
| `builtin_readonly` | 编译进 daemon 的只读工具 | 执行低风险读取 | 当前 `runtime-health` 使用该模式。 |
| `external_readonly_sandbox` | 外部只读插件 + 沙箱 | 只读、限时、限资源 | 需要签名、checksum、进程隔离和审计字段。 |
| `write_with_approval` | 可写工具 + 用户/企业授权 | 白名单写操作 | 需要授权弹窗、回滚、审批记录和更强审计。 |

在 `external_readonly_sandbox` 前，不应通过配置引入 `command`、`args`、脚本路径或任意本地二进制执行。

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

最小内置工具模板：

```go
package tools

import (
    "context"
    "strings"
)

type DesktopSummaryTool struct {
    manifest Manifest
    readSummary func(context.Context) (map[string]any, error)
}

func NewDesktopSummaryTool(manifest Manifest, readSummary func(context.Context) (map[string]any, error)) *DesktopSummaryTool {
    return &DesktopSummaryTool{manifest: manifest, readSummary: readSummary}
}

func (t *DesktopSummaryTool) ID() string {
    return t.manifest.ID
}

func (t *DesktopSummaryTool) Manifest() Manifest {
    return t.manifest
}

func (t *DesktopSummaryTool) Invoke(ctx context.Context, input Input) (Result, error) {
    if input.ToolID != t.manifest.ID {
        return Result{}, NewError(ErrCodeFailed, "tool id mismatch", nil)
    }
    if t.readSummary == nil {
        return Result{}, NewError(ErrCodeUnavailable, "desktop summary reader is unavailable", nil)
    }
    summary, err := t.readSummary(ctx)
    if err != nil {
        return Result{}, NewError(ErrCodeFailed, "read desktop summary failed", err)
    }
    if raw, ok := summary["authorization"].(string); ok && strings.TrimSpace(raw) != "" {
        delete(summary, "authorization")
    }
    return Result{
        ToolID: input.ToolID,
        AppID:  input.AppID,
        Output: summary,
    }, nil
}
```

`NewRegistryFromConfig` 注册分支模板：

```go
case "desktop-summary":
    registry.Register(NewDesktopSummaryTool(ManifestFromConfig(toolCfg), readDesktopSummary))
```

配置模板：

```json
{
  "id": "gateway.desktop_summary",
  "name": "Desktop Summary",
  "adapter": "desktop-summary",
  "description": "Read a sanitized desktop summary.",
  "read_only": true,
  "risk_level": "medium",
  "scopes": ["desktop.read"],
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

Registry 测试模板：

```go
func TestDesktopSummaryToolManifestAndInvoke(t *testing.T) {
    tool := NewDesktopSummaryTool(Manifest{
        ID:        "gateway.desktop_summary",
        Adapter:   "desktop-summary",
        ReadOnly:  true,
        RiskLevel: "medium",
        Scopes:    []string{"desktop.read"},
    }, func(context.Context) (map[string]any, error) {
        return map[string]any{
            "active_app": "IDE",
            "authorization": "Bearer secret",
        }, nil
    })

    result, err := tool.Invoke(context.Background(), Input{
        AppID:  "dev-app",
        ToolID: "gateway.desktop_summary",
    })
    if err != nil {
        t.Fatalf("invoke: %v", err)
    }
    output := result.Output.(map[string]any)
    if output["active_app"] != "IDE" {
        t.Fatalf("unexpected output: %+v", output)
    }
    if _, ok := output["authorization"]; ok {
        t.Fatalf("sensitive field leaked: %+v", output)
    }
}
```

上线检查清单：

- Manifest `read_only=true`、`sandbox_required=false`、`scopes` 非空且最小化。
- 工具只返回必要字段，不返回 token、API Key、Authorization header、完整路径或大体积原始内容。
- `Invoke` 尊重 `context.Context`，依赖不可用时返回 `tool_unavailable`。
- 新 adapter 已补 Config、Registry、Access、Trace/Audit 测试。
- README 或本文有调用示例、scope 说明和失败模式说明。
- 如工具来自外部插件包，当前版本必须保持不可执行，直到签名与沙箱实现完成。

## 当前限制

- 没有独立发布的外部插件 SDK 包。
- 已补签名、来源校验和 checksum enforcement 设计，但尚未实现。
- 已补沙箱阶段模型，但尚未实现真实沙箱进程。
- 没有可写工具授权和回滚机制。
- MCP 仍是 Manifest-only，占位工具不可执行。
