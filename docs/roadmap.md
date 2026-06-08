# 产品化路线图

本文用于说明当前客户端 AI 网关 MVP 的完成度、产品化缺口和建议迭代顺序。

## 当前完成度

| 模块 | 状态 | 说明 |
| --- | --- | --- |
| OpenAI 兼容聊天入口 | 已完成 MVP | 支持 `/v1/chat/completions`、App Token、Trace。 |
| Provider 路由 | 已完成 MVP | 支持本地/云端 Provider、健康状态、启停、探测、模型目录。 |
| Policy Engine | 已完成 MVP | 支持 allow、deny、force_local、deny_cloud_for_sensitive 和 dry-run。 |
| Trace / Audit | 已完成 MVP | 支持 JSONL 持久化、分页、筛选、导出、Trace 关联。 |
| 控制台 | 已完成 MVP | 中文优先、中英文切换、分页、筛选、清空、对象联动、空状态。 |
| 工具调用 | 已完成只读 MVP | 支持内置只读工具、scope 校验、Trace/Audit。 |
| MCP | Manifest MVP | 只加载 Manifest，不执行 MCP Server。 |
| 企业部署说明 | 已补齐初版 | 已有部署、权限审计、失败降级专题文档。 |
| 安全审查清单 | 已补齐初版 | 覆盖 token、管理 API、Trace、Audit、工具、MCP、Provider 策略。 |

## 主要缺口

| 优先级 | 缺口 | 当前风险 |
| --- | --- | --- |
| P0 | 真实浏览器 UI 回归 | 目前只有后端 HTML 锚点测试，缺少截图级布局验证。 |
| P0 | 安全审查清单 | 已补齐初版，后续随真实 MCP / 插件 SDK 继续扩展。 |
| P1 | Provider SDK 边界 | 目前适配器接口可用，但缺少独立 SDK 文档、错误约定和示例模板。 |
| P1 | 插件 / Tool SDK | 目前只读工具能扩展，但缺少插件目录规范、签名/来源校验和沙箱模型。 |
| P1 | 企业集中审计 | 当前 JSONL 适合单机，缺少 SIEM / 日志管道对接方案。 |
| P2 | MCP 真实运行时 | 当前禁止执行外部 MCP，需要补进程模型、授权弹窗、审计字段和沙箱。 |
| P2 | 配额 / 预算 / 速率限制 | 当前未实现用户级限流、Provider 预算和成本统计。 |
| P2 | 安装包与服务管理 | 当前可 `go run` 或构建 daemon，缺少 Windows/macOS 服务安装脚本。 |

## 建议迭代顺序

```mermaid
flowchart LR
  A[当前 MVP] --> B[P0 UI 自动化回归]
  B --> C[P0 安全审查清单]
  C --> D[P1 Provider SDK 文档]
  D --> E[P1 Tool / Plugin SDK]
  E --> F[P1 集中审计方案]
  F --> G[P2 MCP 沙箱运行时]
  G --> H[P2 安装包和企业分发]
```

## 下一阶段验收标准

P0 阶段建议做到：

- 控制台至少有 1 套真实浏览器 smoke 测试，覆盖桌面宽屏和窄屏。
- 安全清单覆盖 token、管理 API、Trace 快照、Audit 导出、MCP、工具调用，并在 README 中可直达。
- `go test ./...`、`go build ./cmd/gateway-daemon` 和 UI smoke 在提交前可复现。
- README 能指向部署、权限审计、失败降级、路线图和错误码文档。

P1 阶段建议做到：

- Provider SDK 文档给出接口、错误码、健康检查、超时和 credential 约定。
- Tool SDK 文档给出 Manifest、scope、只读边界、输入输出 schema 和测试模板。
- 企业审计方案说明 JSONL 到集中日志系统的字段映射和脱敏要求。

## 暂不做的事

- 不开放任意本地命令执行。
- 不启用真实 MCP Server 执行，直到沙箱和授权模型完成。
- 不在控制台展示完整 App Token。
- 不把云端 Provider 作为敏感数据的默认降级目标。
