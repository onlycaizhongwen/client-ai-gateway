# 本地待推送变更摘要

当前本地 `main` 分支包含一组尚未推送到远端的配额治理和审计增强提交。

## 已完成

- Provider 级 RPM 限流：Provider 候选调用前检查配额，超限跳过候选，所有候选不可用时返回 `rate_limited`。
- 配额 Trace 结构化：`quota_checked` / `quota_rejected` 事件携带 `quota.subject`、`id`、`window`、`limit`、`remaining`、`reset_at`。
- 控制台配额指标：Trace 聚合展示配额拒绝、App 配额拒绝、Provider 配额拒绝和配额检查数量。
- App / Provider 配额管理：控制台和管理 API 支持调整 `requests_per_minute`，`0` 表示关闭 RPM 配额。
- 配置写回保护：运行时配置更新串行化，减少控制台并发保存时互相覆盖的风险。
- 审计 before/after：App / Provider 配额和 Provider 启停记录 old/new metadata，并在控制台审计列表和详情中友好展示。

## 验证命令

```powershell
go test ./...
go build ./cmd/gateway-daemon
powershell -ExecutionPolicy Bypass -File scripts/ui-smoke.ps1
git diff --check
```

## 暂未覆盖

- 未启用真实 MCP Server 执行。
- 未开放任意本地命令执行。
- 未实现 Provider 预算、token/day 记账和成本统计。
- 未做跨进程同时编辑同一配置文件的锁保护。
