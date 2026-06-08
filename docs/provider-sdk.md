# Provider SDK 边界

本文说明如何为客户端 AI 网关新增 Provider Adapter，以及适配器必须遵守的接口、错误、凭证、健康检查和测试约定。

## 适配器位置

Provider Adapter 位于 `internal/adapters`。当前已有：

- `mock`：本地测试和演示用 Provider。
- `openai-compatible`：面向 LM Studio、Ollama OpenAI-compatible endpoint、企业云模型网关等上游。

配置入口位于 `configs/*.json` 的 `providers[]`。

## 核心接口

```go
type Provider interface {
    ID() string
    Chat(context.Context, ChatInput) (Result, error)
}

type HealthChecker interface {
    CheckHealth(context.Context) error
}
```

`HealthChecker` 是可选接口。未实现时 Provider health 会进入 `degraded`，但不一定阻止启动。企业 Provider 推荐实现健康检查，便于控制台和路由判断。

## 输入输出模型

```go
type ChatInput struct {
    TraceID  string
    Model    string
    Messages []Message
    FailMode string
}

type Result struct {
    ProviderID string
    Model      string
    Content    string
    Usage      Usage
}
```

约定：

- `ProviderID` 必须返回实际处理请求的 Provider ID。
- `Model` 优先返回上游响应模型；为空时可回退为请求模型。
- `Usage` 不可得时保持 0，不要伪造成本数据。
- Adapter 不应保存 App Token；调用上游时只使用 Provider 自己的凭证。

## 配置字段

Provider 常用配置：

| 字段 | 说明 |
| --- | --- |
| `id` | Provider 唯一 ID。 |
| `name` | 控制台展示名称。 |
| `class` | `local` 或 `cloud`，策略会用它判断云端边界。 |
| `adapter` | 适配器名称，例如 `openai-compatible`。为空时默认 `mock`。 |
| `base_url` | 上游根地址。 |
| `api_key_env` | 从环境变量读取 API Key，不直接写入配置。 |
| `models` | Provider 声明支持的模型。 |
| `healthy` | 静态健康开关。 |
| `enabled` | 启停开关。 |

OpenAI-compatible 示例：

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

## 凭证约定

- 使用 `api_key_env` 从环境变量读取密钥。
- 缺少环境变量时 daemon 可启动，但 Provider health 应显示异常。
- 错误响应和 Trace/Audit 不应包含密钥值。
- 本地 Provider 可不配置 `api_key_env`。

## 超时与健康检查

当前 OpenAI-compatible adapter：

- Chat HTTP client 默认超时：60 秒。
- Provider health monitor 单次检查默认超时：3 秒。
- Monitor 默认间隔：30 秒。
- 连续失败 3 次后进入 `unhealthy`，之前为 `degraded`。

`CheckHealth` 推荐调用上游轻量健康端点，例如 `/healthz`。如果上游没有健康端点，可后续扩展为模型列表或极轻量请求，但必须避免昂贵推理。

## 错误约定

Provider adapter 应尽量返回 `ProviderError`，并使用稳定错误码：

| Adapter code | API code |
| --- | --- |
| `missing_credential` | `provider_missing_credential` |
| `connection_failed` | `provider_connection_failed` |
| `timeout` | `provider_timeout` |
| `unauthorized` | `provider_unauthorized` |
| `rate_limited` | `provider_rate_limited` |
| `upstream_status` | `provider_upstream_status` |
| `invalid_response` | `provider_invalid_response` |
| `empty_response` | `provider_empty_response` |

未分类错误会进入 `provider_failed`。错误消息必须描述问题，不得包含 API Key、Authorization header 或完整请求体。

## 注册流程

新增 Provider Adapter 的步骤：

1. 在 `internal/adapters` 新增 adapter 实现。
2. 实现 `Provider`，推荐同时实现 `HealthChecker`。
3. 在 `NewProvider` 中增加 `adapter` 名称分支。
4. 在 `config.Provider.Validate` 相关逻辑中补充配置校验。
5. 在 `docs/error-codes.md` 中补充新增稳定错误码。
6. 在 README 或本文件中增加配置示例。

## 测试要求

至少补充：

- Adapter 单元测试：成功请求、上游 401/403、429、非 2xx、超时、无效 JSON、空 choices。
- 凭证测试：`api_key_env` 配置但环境变量为空时返回 `missing_credential`。
- 健康检查测试：成功、超时、失败状态分类。
- 配置测试：缺少必填字段时加载失败。
- Access / Pipeline 测试：Provider 错误能转成稳定 API 错误并写入 Trace。

提交前运行：

```powershell
go test ./internal/adapters
go test ./internal/providerhealth
go test ./internal/access
go test ./...
go build ./cmd/gateway-daemon
```

## 当前限制

- 还没有独立发布的外部 Provider SDK 包，当前以仓库内部接口为准。
- 未实现流式响应。
- 未实现 Provider 级预算、配额或速率限制。
- 未实现模型能力矩阵，例如 vision、embedding、tool-call capability。

