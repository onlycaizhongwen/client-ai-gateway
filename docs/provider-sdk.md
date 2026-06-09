# Provider SDK 边界

本文说明如何为客户端 AI 网关新增 Provider Adapter，以及适配器必须遵守的接口、错误、凭证、健康检查和测试约定。

## 适配器位置

Provider Adapter 位于 `internal/adapters`。当前已有：

- `mock`：本地测试和演示用 Provider。
- `openai-compatible`：面向 LM Studio、Ollama OpenAI-compatible endpoint、企业云模型网关等上游。

配置入口位于 `configs/*.json` 的 `providers[]`。

## 能力矩阵

当前 Provider SDK 仍是仓库内部接口，能力以“声明 + 运行时探测”为主，不做自动推断。新增 Provider 前建议先按下表确认适配范围：

| 能力 | 当前契约 | 控制面影响 | 新 Provider 要求 |
| --- | --- | --- | --- |
| Chat Completions | 已支持 | `/v1/chat/completions`、Trace、路由解释 | 必须实现 `Provider.Chat`。 |
| Streaming | 未支持 | API 固定非流式 | 不要向上游开启 `stream=true`。 |
| Embeddings | 未支持 | 暂无入口和 grant | 不要在 `models` 中承诺 embedding-only 能力。 |
| Vision / 多模态 | 未支持 | 请求模型只有文本 messages | 如上游支持，也只能按文本 chat 接入。 |
| Tool Calling | 未支持 | 工具由网关内置 Tool runtime 管理 | Provider 不应自行执行工具。 |
| Health Check | 可选但推荐 | Provider health、模型目录、路由可用性 | 企业/云端 Provider 应实现 `HealthChecker`。 |
| Credential | 已支持 env 引用 | Trace/Audit 不暴露密钥 | 使用 `api_key_env`，不要把密钥写入配置。 |
| Rate Limit | 已支持 Provider RPM | 路由候选超限跳过 | 配额由网关控制，Provider 不要自行改全局状态。 |
| Cost / Budget | 未支持 | 暂无成本报表 | 可返回 token usage，不要伪造成本金额。 |

Provider 配置中的 `models` 目前只表示“该 Provider 可被路由候选选中”，不等同于完整模型能力清单。后续如果引入 vision、embedding、tool-call 等能力，应扩展独立 capability 字段，而不是复用 `models` 字符串做隐式约定。

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

自定义 Provider 配置模板：

```json
{
  "id": "enterprise-llm",
  "name": "Enterprise LLM",
  "class": "cloud",
  "adapter": "enterprise-compatible",
  "base_url": "https://llm.example.internal",
  "api_key_env": "ENTERPRISE_LLM_API_KEY",
  "models": ["enterprise-chat"],
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

最小 Adapter 模板：

```go
package adapters

import (
    "context"
    "fmt"
    "net/http"
    "strings"
)

type EnterpriseCompatibleConfig struct {
    ID        string
    BaseURL   string
    APIKey    string
    APIKeyEnv string
    Client    *http.Client
}

type EnterpriseCompatibleProvider struct {
    id        string
    baseURL   string
    apiKey    string
    apiKeyEnv string
    client    *http.Client
}

func NewEnterpriseCompatibleProvider(cfg EnterpriseCompatibleConfig) (*EnterpriseCompatibleProvider, error) {
    if strings.TrimSpace(cfg.ID) == "" {
        return nil, fmt.Errorf("provider id is required")
    }
    if strings.TrimSpace(cfg.BaseURL) == "" {
        return nil, fmt.Errorf("base url is required")
    }
    client := cfg.Client
    if client == nil {
        client = http.DefaultClient
    }
    return &EnterpriseCompatibleProvider{
        id:        cfg.ID,
        baseURL:   strings.TrimRight(cfg.BaseURL, "/"),
        apiKey:    cfg.APIKey,
        apiKeyEnv: strings.TrimSpace(cfg.APIKeyEnv),
        client:    client,
    }, nil
}

func (p *EnterpriseCompatibleProvider) ID() string {
    return p.id
}

func (p *EnterpriseCompatibleProvider) Chat(ctx context.Context, input ChatInput) (Result, error) {
    if p.apiKeyEnv != "" && strings.TrimSpace(p.apiKey) == "" {
        return Result{}, &ProviderError{
            ProviderID: p.id,
            Code:       ErrorMissingCredential,
            Message:    "api key environment variable " + p.apiKeyEnv + " is not set",
        }
    }
    // 1. Translate ChatInput into the upstream request.
    // 2. Set Authorization from p.apiKey when present.
    // 3. Classify upstream errors as ProviderError with stable codes.
    // 4. Return actual ProviderID, final model and usage from upstream.
    return Result{}, &ProviderError{ProviderID: p.id, Code: ErrorUpstreamStatus, Message: "not implemented"}
}

func (p *EnterpriseCompatibleProvider) CheckHealth(ctx context.Context) error {
    // Use a cheap health endpoint, never an expensive inference request.
    return nil
}
```

`NewProvider` 注册分支模板：

```go
case "enterprise-compatible":
    apiKey := ""
    if provider.APIKeyEnv != "" {
        apiKey = os.Getenv(provider.APIKeyEnv)
    }
    return NewEnterpriseCompatibleProvider(EnterpriseCompatibleConfig{
        ID:        provider.ID,
        BaseURL:   provider.BaseURL,
        APIKey:    apiKey,
        APIKeyEnv: provider.APIKeyEnv,
        Client:    &http.Client{Timeout: 60 * time.Second},
    })
```

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

Adapter 测试模板：

```go
func TestEnterpriseCompatibleChatSuccess(t *testing.T) {
    upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/v1/chat/completions" {
            t.Fatalf("unexpected path %s", r.URL.Path)
        }
        w.Header().Set("Content-Type", "application/json")
        _, _ = w.Write([]byte(`{
          "model":"enterprise-chat",
          "choices":[{"message":{"role":"assistant","content":"ok"}}],
          "usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}
        }`))
    }))
    defer upstream.Close()

    provider, err := NewEnterpriseCompatibleProvider(EnterpriseCompatibleConfig{
        ID:      "enterprise",
        BaseURL: upstream.URL,
        Client:  upstream.Client(),
    })
    if err != nil {
        t.Fatalf("new provider: %v", err)
    }
    result, err := provider.Chat(context.Background(), ChatInput{
        TraceID: "trace-test",
        Model:   "enterprise-chat",
        Messages: []Message{{Role: "user", Content: "hello"}},
    })
    if err != nil {
        t.Fatalf("chat: %v", err)
    }
    if result.ProviderID != "enterprise" || result.Model != "enterprise-chat" || result.Content != "ok" {
        t.Fatalf("unexpected result: %+v", result)
    }
}
```

上线检查清单：

- Adapter 名称稳定，配置示例和 `NewProvider` 分支一致。
- 所有上游错误都归一为 `ProviderError` 或明确可接受的普通错误。
- Trace/Audit、日志和错误消息不包含 API Key、Authorization header 或完整请求体。
- `CheckHealth` 使用低成本探测，不触发真实推理账单。
- `go test ./internal/adapters ./internal/providerhealth ./internal/access` 覆盖成功、认证失败、限流、超时和无效响应。
- README 或部署文档说明环境变量、Provider `class` 和敏感数据策略影响。

## 当前限制

- 还没有独立发布的外部 Provider SDK 包，当前以仓库内部接口为准。
- 未实现流式响应。
- 已实现 Provider RPM 配额，但未实现 Provider 级预算、token/day 记账和成本统计。
- 已补文档级能力矩阵，但未实现运行时 capability 字段，例如 vision、embedding、tool-call capability。
