# 产品化路线图

本文用于说明当前客户端 AI 网关 MVP 的完成度、产品化缺口和建议迭代顺序。

## 当前完成度

| 模块 | 状态 | 说明 |
| --- | --- | --- |
| OpenAI 兼容聊天入口 | 已完成 MVP | 支持 `/v1/chat/completions`、App Token、Trace。 |
| Provider 路由 | 已完成 MVP | 支持本地/云端 Provider、健康状态、启停、探测、模型目录。 |
| Policy Engine | 已完成 MVP | 支持 allow、deny、force_local、deny_cloud_for_sensitive 和 dry-run。 |
| Trace / Audit | 已完成 MVP | 支持 JSONL 持久化、分页、筛选、导出、Trace 关联。 |
| 控制台 | 已完成 MVP | 中文优先、中英文切换、分页、筛选、清空、对象联动、空状态，并有 headless 浏览器 smoke。 |
| 工具调用 | 已完成只读 MVP | 支持内置只读工具、scope 校验、Trace/Audit。 |
| MCP | Manifest MVP | 只加载 Manifest，不执行 MCP Server。 |
| 企业部署说明 | 已补齐初版 | 已有部署、权限审计、失败降级专题文档。 |
| 安全审查清单 | 已补齐初版 | 覆盖 token、管理 API、Trace、Audit、工具、MCP、Provider 策略。 |
| Provider SDK 边界 | 已补齐初版 | 已有接口、配置、凭证、健康检查、错误码和测试要求。 |
| Tool / Plugin SDK 边界 | 已补齐初版 | 已有 Manifest、scope、只读边界、输入输出 schema、Trace/Audit 和测试要求。 |
| 企业集中审计 | 已补齐初版 | 已有 JSONL 到 SIEM / SOC 的字段映射、脱敏、采集和告警建议。 |
| MCP 真实运行时设计 | 已补齐草案 | 已有进程模型、沙箱、授权、审计字段和失败关闭门槛。 |
| 配额预算与限流 | 已完成 App + Provider RPM MVP | 已有 App 请求前 requests_per_minute 限流和 Provider 候选级 RPM 跳过；Provider 预算、token/day 和成本统计仍是设计草案。 |
| 安装包与服务管理设计 | 已补齐草案 | 已有 Windows Service、macOS launchd、目录布局、升级回滚和验收门槛。 |

## 主要缺口

| 优先级 | 缺口 | 当前风险 |
| --- | --- | --- |
| P0 | 真实浏览器 UI 回归 | 已补最小 headless smoke，后续需要 Playwright 级交互断言和像素回归。 |
| P0 | 安全审查清单 | 已补齐初版，后续随真实 MCP / 插件 SDK 继续扩展。 |
| P1 | Provider SDK 边界 | 已补齐初版，后续需要示例模板和能力矩阵。 |
| P1 | 插件 / Tool SDK | 已补齐初版，后续需要签名/来源校验、示例模板和真实沙箱模型。 |
| P1 | 企业集中审计 | 已补齐初版，后续需要内置 exporter、时间游标和企业租户字段。 |
| P2 | MCP 真实运行时 | 已补设计草案，当前仍禁止执行外部 MCP；后续进入沙箱运行时实现。 |
| P2 | 配额 / 预算 / 速率限制 | 已完成 App + Provider RPM MVP，后续补 Provider 预算、token/day 记账和控制台配额面板。 |
| P2 | 安装包与服务管理 | 已补设计草案，当前仍未提供正式安装器或系统服务安装脚本。 |

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

- Provider SDK 文档给出接口、错误码、健康检查、超时和 credential 约定，并在 README 中可直达。
- Tool SDK 文档给出 Manifest、scope、只读边界、输入输出 schema 和测试模板，并在 README 中可直达。
- 企业审计方案说明 JSONL 到集中日志系统的字段映射和脱敏要求，并在 README 中可直达。

## 暂不做的事

- 不开放任意本地命令执行。
- 不启用真实 MCP Server 执行，直到沙箱和授权模型完成。
- 不在控制台展示完整 App Token。
- 不把云端 Provider 作为敏感数据的默认降级目标。
