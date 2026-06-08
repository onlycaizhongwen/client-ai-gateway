# 安装包与服务管理设计草案

本文定义客户端 AI 网关从本地 `go run` 走向企业桌面分发时的安装包、服务管理、升级、回滚和运维边界。当前版本尚未提供正式安装器或服务安装脚本。

## 目标形态

| 平台 | 推荐形态 | 说明 |
| --- | --- | --- |
| Windows | Windows Service 或登录后用户态托管进程 | 企业桌面优先服务化；开发环境可手动运行。 |
| macOS | `launchd` LaunchAgent | 以当前用户身份运行，避免默认 root。 |
| Linux | systemd user service | 开发机或受控终端可选。 |

默认监听仍应保持 `127.0.0.1`，不要在安装包里默认暴露到局域网。

## 目录布局

Windows 建议：

```text
C:\Program Files\ClientAIGateway\
  gateway-daemon.exe
  configs\
    gateway.json
  VERSION
C:\ProgramData\ClientAIGateway\
  data\
    traces.jsonl
    audit.jsonl
  logs\
```

macOS 建议：

```text
/Applications/ClientAIGateway/
  gateway-daemon
  configs/gateway.json
~/Library/Application Support/ClientAIGateway/
  data/traces.jsonl
  data/audit.jsonl
  logs/
```

原则：

- 程序目录只读。
- 配置由企业配置系统下发。
- Trace/Audit 写入数据目录。
- 日志和审计文件进入企业采集策略。

## Windows 服务草案

服务参数建议：

| 项 | 建议 |
| --- | --- |
| Service Name | `ClientAIGateway` |
| Display Name | `Client AI Gateway` |
| Binary | `gateway-daemon.exe -config C:\Program Files\ClientAIGateway\configs\gateway.json` |
| Account | 低权限本地账号或当前用户托管，不默认 LocalSystem。 |
| Startup | Automatic delayed 或企业策略控制。 |
| Recovery | 首次/二次失败自动重启，连续失败后告警。 |

安装脚本必须检查：

- 配置文件存在且可读。
- `listen_addr` 不为 `0.0.0.0`，除非显式企业审批。
- 数据目录可写。
- 旧进程已停止。
- 新版本 `/healthz` 可用。

## macOS launchd 草案

LaunchAgent 建议：

```xml
<key>Label</key>
<string>com.client-ai-gateway.daemon</string>
<key>ProgramArguments</key>
<array>
  <string>/Applications/ClientAIGateway/gateway-daemon</string>
  <string>-config</string>
  <string>/Applications/ClientAIGateway/configs/gateway.json</string>
</array>
<key>RunAtLoad</key>
<true/>
<key>KeepAlive</key>
<true/>
```

不要默认使用 LaunchDaemon/root，除非企业明确要求并完成权限审查。

## 升级与回滚

推荐流程：

1. 停止服务。
2. 备份当前二进制、配置和 VERSION。
3. 替换二进制。
4. 校验配置加载。
5. 启动服务。
6. 调用 `/healthz` 和 `/gateway/v1/runtime/health`。
7. 失败则恢复旧二进制并重启。

升级包必须包含：

- 版本号。
- 构建时间。
- 校验和。
- 变更说明。
- 最小兼容配置版本。

## 健康检查

安装后至少验证：

```powershell
curl http://127.0.0.1:18765/healthz
curl http://127.0.0.1:18765/gateway/v1/runtime/health
```

企业部署还应验证：

- 控制台 `/console` 可打开。
- Trace/Audit 路径可写。
- Provider health 正常或有明确 degraded reason。
- MCP runtime 保持 manifest-only。
- Audit 采集 Agent 能读取审计文件。

## 日志与审计

建议：

- stdout/stderr 写入平台服务日志。
- Audit JSONL 使用 `audit_store_path`，进入集中审计采集。
- Trace JSONL 使用 `trace_store_path`，按企业留存策略采集或保留本机。
- 安装/升级/卸载动作写入企业终端管理日志。

## 卸载要求

卸载应区分：

| 操作 | 行为 |
| --- | --- |
| 移除程序 | 删除二进制和服务注册。 |
| 保留数据 | 默认保留 Trace/Audit，便于审计追溯。 |
| 清理数据 | 需要管理员显式选择，并记录终端管理日志。 |
| 移除配置 | 企业配置系统应同步回收。 |

## 安全要求

- 不在安装包中写入真实 App Token 或 Provider API Key。
- 示例 token 仅用于开发包。
- 安装脚本不得打印完整 token。
- 服务账号权限最小化。
- 默认只监听 `127.0.0.1`。
- 升级包必须校验签名或 checksum。

## 验收门槛

正式提供安装脚本前至少需要：

- Windows 安装、启动、停止、卸载脚本。
- macOS LaunchAgent 模板。
- 配置路径覆盖测试。
- 升级失败回滚演练。
- 服务运行下的 UI smoke。
- 安装后健康检查脚本。
- 企业部署文档和 README 更新。

## 当前决策

当前版本保持：

- 不提供正式安装器。
- 不自动注册系统服务。
- 不修改系统目录。
- 不写入真实企业 token。
- 只提供服务化设计和后续验收门槛。
