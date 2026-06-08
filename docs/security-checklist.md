# 安全审查清单

本文用于在演示、交付、企业试点或发布前快速检查客户端 AI 网关的安全边界。每一项都应在环境配置、自动化测试或人工验收中有明确证据。

## 身份与 Token

| 检查项 | 要求 | 证据 |
| --- | --- | --- |
| App Token 不为空 | `configs/*.json` 中每个 App 必须有独立 token。 | 配置加载校验会拒绝空 token。 |
| Token 不重复 | 不同 App 不能共用 token。 | 配置加载校验会拒绝重复 token。 |
| 控制台不展示完整 token | App 目录只返回 `token_hint`。 | `GET /gateway/v1/apps` 不包含完整 token。 |
| 示例 token 不用于生产 | `dev-token`、`admin-token` 仅用于本地开发。 | 企业配置应由配置系统下发。 |

## 管理 API

| API | 要求 |
| --- | --- |
| `GET /gateway/v1/audit/events` | 必须携带 `admin` Grant。 |
| `GET /gateway/v1/audit/events/export` | 必须携带 `admin` Grant。 |
| `POST /gateway/v1/config/reload` | 必须携带 `admin` Grant。 |
| `POST /gateway/v1/providers/{id}/enabled` | 必须携带 `admin` Grant。 |
| `POST /gateway/v1/providers/{id}/probe` | 必须携带 `admin` Grant。 |
| `GET /gateway/v1/apps` / `GET /gateway/v1/grants` | 必须携带 `admin` Grant。 |

验收方式：

- 无 token 请求应返回 `unauthorized`。
- 普通 `chat` / `tool` App 不应访问管理 API。
- 管理操作应写入 Audit，包含 `app_id`、`action`、`target`、`result`。

## Trace 快照

| 检查项 | 要求 |
| --- | --- |
| 不保存 App Token | Trace 请求快照只保存请求体安全子集。 |
| 敏感标签脱敏 | 命中 `trace_redact_labels` 时，消息内容和 metadata value 写为 `[redacted]`。 |
| 长文本截断 | `trace_snapshot_max_chars` 控制单值最大长度。 |
| 导出复用安全快照 | Trace 导出不重新读取原始请求，不恢复敏感内容。 |

推荐配置：

```json
{
  "trace_snapshot_enabled": true,
  "trace_redact_labels": ["sensitive"],
  "trace_snapshot_max_chars": 4000
}
```

## Audit 导出

| 检查项 | 要求 |
| --- | --- |
| 管理员授权 | Audit 查询和导出必须要求 `admin` Grant。 |
| 可过滤导出 | 导出应复用当前筛选条件，降低泄露面。 |
| 关联 Trace | 通过 `trace_id` 串联请求链路，但请求内容仍以 Trace 安全快照为准。 |
| metadata 可解释 | 权限拒绝、工具调用、策略试算应写入 `explain_chain` 或缺失授权信息。 |

## 工具调用

| 检查项 | 要求 |
| --- | --- |
| 默认只读 | 当前 MVP 只允许 `read_only=true` 工具。 |
| scope 最小权限 | 优先使用 `tool:<scope>`，避免直接授予 `tool`。 |
| 缺失 scope 失败关闭 | 返回 `tool_scope_denied`，并写入 Trace/Audit。 |
| 非只读失败关闭 | 返回 `tool_denied`。 |
| 未注册 adapter 失败关闭 | 返回 `tool_unavailable` 或 `not_found`。 |
| 不执行任意命令 | 当前无任意本地命令执行入口。 |

## MCP Manifest

| 检查项 | 要求 |
| --- | --- |
| 只允许 Manifest | `mcp_runtime.mode` 只能为 `manifest_only` 或关闭。 |
| 不启动外部进程 | `stdio`、`direct`、`sandboxed` 等真实执行模式应被配置校验拒绝。 |
| 工具必须只读 | MCP Tool 必须 `read_only=true`。 |
| 必须声明 scope | MCP Tool 的 `scopes` 不能为空。 |
| 禁止沙箱占位执行 | 在沙箱运行时完成前，`sandbox_required=true` 会被拒绝。 |

## Provider 与云端策略

| 检查项 | 要求 |
| --- | --- |
| 云端 Provider 显式标记 | 云端上游必须配置 `class="cloud"`。 |
| 敏感数据禁止云端降级 | 敏感场景应配置 `deny_cloud_for_sensitive`。 |
| 本地优先策略 | 指定 App 或模型可配置 `force_local`。 |
| Credential 不进错误响应 | Provider API Key 缺失或失败不应回显密钥内容。 |
| 不健康 Provider 不参与路由 | `unhealthy` / `disabled` Provider 不应被 Router 选择。 |

## 发布前命令

```powershell
go test ./internal/access
go test ./...
go build ./cmd/gateway-daemon
git diff --check
```

## 交付前人工确认

- README 指向部署、权限审计、失败降级、路线图、安全清单和错误码文档。
- 控制台 `/console` 不暴露完整 App Token。
- 管理员 token 不写入 Trace。
- 敏感请求的 Trace 快照已脱敏。
- MCP 仍处于 Manifest-only，不会执行外部命令。
- 企业试点配置不复用 `dev-token` / `admin-token`。

