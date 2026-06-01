package access

const consoleHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>客户端 AI 网关控制台</title>
  <style>
    :root {
      --bg: #f5f7fa;
      --panel: #fff;
      --ink: #172033;
      --muted: #667085;
      --line: #d8dee9;
      --head: #eef3f8;
      --blue: #2368a2;
      --green: #1d7a42;
      --red: #b42318;
      --amber: #a66300;
      --code: #eef2f7;
    }
    * { box-sizing: border-box; }
    body { margin: 0; background: var(--bg); color: var(--ink); font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Microsoft YaHei", Arial, sans-serif; font-size: 14px; }
    header { min-height: 66px; padding: 12px 24px; background: var(--panel); border-bottom: 1px solid var(--line); display: flex; align-items: center; justify-content: space-between; gap: 16px; }
    h1 { margin: 0; font-size: 21px; line-height: 1.2; }
    h2 { margin: 0; font-size: 16px; }
    main { padding: 18px 22px 32px; }
    button { border: 1px solid #1d6fa5; background: var(--blue); color: #fff; border-radius: 6px; padding: 8px 12px; cursor: pointer; white-space: nowrap; }
    button.secondary { background: #fff; color: var(--blue); }
    button:disabled { opacity: .45; cursor: not-allowed; }
    input, select { border: 1px solid var(--line); border-radius: 6px; padding: 8px 10px; width: 100%; min-width: 0; }
    .muted { color: var(--muted); }
    .cards { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 12px; margin-bottom: 14px; }
    .card, .panel { background: var(--panel); border: 1px solid var(--line); border-radius: 8px; }
    .card { padding: 13px 15px; }
    .metric { margin-top: 5px; font-size: 25px; font-weight: 750; }
    .layout { display: grid; grid-template-columns: minmax(900px, 1fr) 360px; gap: 14px; align-items: start; }
    .main-stack, .side, .form-grid { display: grid; gap: 14px; }
    .panel-head { padding: 13px 15px; border-bottom: 1px solid var(--line); display: flex; align-items: center; justify-content: space-between; gap: 12px; }
    .panel-body { padding: 13px 15px; }
    .actions, .pager-actions, .provider-actions { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
    .table-wrap { overflow: auto; height: 610px; }
    .audit-table-wrap { overflow: auto; max-height: 360px; border: 1px solid var(--line); border-radius: 6px; }
    table { width: 100%; border-collapse: collapse; table-layout: fixed; }
    .trace-table { min-width: 1280px; }
    .audit-table { min-width: 760px; font-size: 13px; }
    th, td { border-bottom: 1px solid var(--line); padding: 9px 10px; text-align: left; vertical-align: top; }
    th { position: sticky; top: 0; background: var(--head); z-index: 1; font-weight: 700; }
    td { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
    tr { cursor: pointer; }
    tr:hover { background: #f7fbff; }
    .trace-table th:nth-child(1), .trace-table td:nth-child(1) { width: 92px; }
    .trace-table th:nth-child(2), .trace-table td:nth-child(2) { width: 190px; }
    .trace-table th:nth-child(3), .trace-table td:nth-child(3) { width: 100px; }
    .trace-table th:nth-child(4), .trace-table td:nth-child(4) { width: 145px; }
    .trace-table th:nth-child(5), .trace-table td:nth-child(5) { width: 130px; }
    .trace-table th:nth-child(6), .trace-table td:nth-child(6) { width: 130px; }
    .trace-table th:nth-child(7), .trace-table td:nth-child(7) { width: 180px; }
    .trace-table th:nth-child(8), .trace-table td:nth-child(8) { width: 90px; }
    .trace-table th:nth-child(9), .trace-table td:nth-child(9) { width: 180px; }
    .trace-table th:nth-child(10), .trace-table td:nth-child(10) { width: 240px; }
    .audit-table th:nth-child(1), .audit-table td:nth-child(1) { width: 150px; }
    .audit-table th:nth-child(2), .audit-table td:nth-child(2) { width: 90px; }
    .audit-table th:nth-child(3), .audit-table td:nth-child(3) { width: 130px; }
    .audit-table th:nth-child(4), .audit-table td:nth-child(4) { width: 190px; }
    .audit-table th:nth-child(5), .audit-table td:nth-child(5) { width: 150px; }
    .audit-table th:nth-child(6), .audit-table td:nth-child(6) { width: 80px; }
    .trace-id { font-family: Consolas, "Courier New", monospace; color: var(--blue); }
    .status { display: inline-block; border-radius: 999px; padding: 2px 8px; font-weight: 700; font-size: 12px; }
    .completed, .success, .healthy, .available, .running, .loaded { color: var(--green); background: #e9f7ef; }
    .failed, .denied, .unhealthy { color: var(--red); background: #fdecec; }
    .started, .degraded, .disabled, .not_configured, .unavailable { color: var(--amber); background: #fff4df; }
    .provider, .route-item { border: 1px solid var(--line); border-radius: 6px; padding: 10px; margin-bottom: 9px; }
    .provider strong, .route-item strong { display: block; margin-bottom: 3px; }
    .route-item.skipped { background: #fff7ed; }
    .mcp-tool { margin-top: 8px; padding: 8px; border-top: 1px solid var(--line); }
    .tool-meta { min-height: 20px; }
    .provider-actions button { padding: 5px 8px; font-size: 12px; }
    .pager { border-top: 1px solid var(--line); padding: 9px 12px; display: flex; align-items: center; justify-content: space-between; gap: 10px; }
    .pager button { padding: 5px 9px; font-size: 12px; }
    .kv { display: grid; grid-template-columns: 130px minmax(0, 1fr); gap: 8px 10px; margin-bottom: 12px; }
    .k { color: var(--muted); }
    pre { margin: 0; padding: 12px; background: var(--code); border: 1px solid var(--line); border-radius: 6px; overflow: auto; max-height: 430px; font-size: 12px; }
    @media (max-width: 1280px) {
      main { padding: 14px; }
      .cards { grid-template-columns: repeat(2, minmax(0, 1fr)); }
      .layout { grid-template-columns: 1fr; }
      .table-wrap { height: 430px; }
    }
    @media (max-width: 720px) {
      header { align-items: flex-start; flex-direction: column; }
      .cards { grid-template-columns: 1fr; }
      .panel-head, .pager { align-items: flex-start; flex-direction: column; }
      .actions, .pager-actions { width: 100%; }
      .actions button, .pager-actions button { flex: 1; }
      .kv { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>
  <header>
    <div>
      <h1 data-i18n="title">客户端 AI 网关控制台</h1>
      <div class="muted" data-i18n="subtitle">本地模型/工具网关 · 追踪 · 策略 · 降级</div>
    </div>
    <div class="actions">
      <button class="secondary" id="lang-toggle">English</button>
      <button class="secondary" id="sample" data-i18n="sendTest">发送测试请求</button>
      <button id="refresh" data-i18n="refresh">刷新</button>
    </div>
  </header>
  <main>
    <section class="cards">
      <div class="card"><div class="muted" data-i18n="totalTraces">追踪总数</div><div class="metric" id="m-total">0</div></div>
      <div class="card"><div class="muted" data-i18n="completed">已完成</div><div class="metric" id="m-completed">0</div></div>
      <div class="card"><div class="muted" data-i18n="failed">失败</div><div class="metric" id="m-failed">0</div></div>
      <div class="card"><div class="muted" data-i18n="fallbackAttempts">降级次数</div><div class="metric" id="m-fallbacks">0</div></div>
    </section>
    <div class="layout">
      <div class="main-stack">
        <section class="panel">
          <div class="panel-head">
            <div>
              <h2 data-i18n="requestTraces">请求追踪</h2>
              <div class="muted" id="summary">正在加载追踪...</div>
            </div>
            <div class="actions">
              <select id="status-filter" style="max-width: 160px;">
                <option value="" data-i18n="allStatus">全部状态</option>
                <option value="completed" data-i18n="statusCompleted">已完成</option>
                <option value="failed" data-i18n="statusFailed">失败</option>
              </select>
              <button class="secondary" id="trace-export" data-i18n="export">导出</button>
            </div>
          </div>
          <div class="table-wrap">
            <table class="trace-table">
              <thead>
                <tr>
                  <th data-i18n="status">状态</th>
                  <th data-i18n="traceId">追踪 ID</th>
                  <th data-i18n="app">应用</th>
                  <th data-i18n="requestedModel">请求模型</th>
                  <th data-i18n="provider">Provider</th>
                  <th data-i18n="finalModel">最终模型</th>
                  <th data-i18n="policyRule">策略规则</th>
                  <th data-i18n="fallbacks">降级</th>
                  <th data-i18n="startedAt">开始时间</th>
                  <th data-i18n="error">错误</th>
                </tr>
              </thead>
              <tbody id="rows"></tbody>
            </table>
          </div>
          <div class="pager">
            <div class="muted" id="trace-page-summary">第 1 页</div>
            <div class="pager-actions">
              <button class="secondary" id="trace-prev" data-i18n="prev">上一页</button>
              <button class="secondary" id="trace-next" data-i18n="next">下一页</button>
            </div>
          </div>
        </section>
        <section class="panel">
          <div class="panel-head">
            <h2 data-i18n="auditEvents">审计事件</h2>
            <div class="actions">
              <button class="secondary" id="audit-export" data-i18n="export">导出</button>
              <button class="secondary" id="audit-refresh" data-i18n="refresh">刷新</button>
            </div>
          </div>
          <div class="panel-body form-grid">
            <input id="admin-token" value="admin-token" />
            <div class="audit-table-wrap">
              <table class="audit-table">
                <thead>
                  <tr>
                    <th data-i18n="action">动作</th>
                    <th data-i18n="result">结果</th>
                    <th data-i18n="traceId">追踪 ID</th>
                    <th data-i18n="appTarget">应用 / 目标</th>
                    <th data-i18n="time">时间</th>
                    <th data-i18n="duration">耗时</th>
                  </tr>
                </thead>
                <tbody id="audit-rows"></tbody>
              </table>
            </div>
            <div class="pager">
              <div class="muted" id="audit-page-summary">第 1 页</div>
              <div class="pager-actions">
                <button class="secondary" id="audit-prev" data-i18n="prev">上一页</button>
                <button class="secondary" id="audit-next" data-i18n="next">下一页</button>
              </div>
            </div>
            <div id="audit-message" class="muted"></div>
          </div>
        </section>
      </div>
      <aside class="side">
        <section class="panel">
          <div class="panel-head"><h2 data-i18n="runtimeStatus">运行时状态</h2></div>
          <div class="panel-body" id="runtime-health">正在加载运行时状态...</div>
        </section>
        <section class="panel">
          <div class="panel-head"><h2 data-i18n="providersHealth">Provider / 健康状态</h2></div>
          <div class="panel-body" id="providers">正在加载 Provider...</div>
        </section>
        <section class="panel">
          <div class="panel-head"><h2 data-i18n="quickRequest">快捷请求</h2></div>
          <div class="panel-body form-grid">
            <input id="model" value="local-small" list="model-options" />
            <datalist id="model-options"></datalist>
            <input id="prompt" value="hello from console" />
            <select id="mode">
              <option value="success" data-i18n="modeSuccess">成功请求</option>
              <option value="fallback" data-i18n="modeFallback">本地失败后降级到云端</option>
              <option value="blocked" data-i18n="modeBlocked">敏感数据阻止云端降级</option>
            </select>
            <div class="actions">
              <button id="send" data-i18n="send">发送</button>
              <button class="secondary" id="explain" data-i18n="explain">解释路由</button>
            </div>
            <div class="muted" id="send-result">就绪。</div>
          </div>
        </section>
        <section class="panel">
          <div class="panel-head"><h2 data-i18n="routingExplain">路由解释</h2></div>
          <div class="panel-body" id="route-explain"><p class="muted" data-i18n="routingExplainHint">点击“解释路由”预览策略和 Provider 路由。</p></div>
        </section>
        <section class="panel">
          <div class="panel-head"><h2 data-i18n="tools">工具调用</h2></div>
          <div class="panel-body form-grid">
            <select id="tool-select"></select>
            <div class="muted tool-meta" id="tool-meta" data-i18n="loadingTools">正在加载工具...</div>
            <input id="tool-token" value="dev-token" />
            <button id="tool-invoke" data-i18n="invokeTool">执行工具</button>
            <pre id="tool-result" data-i18n="toolResultPlaceholder">工具执行结果会显示在这里。</pre>
          </div>
        </section>
        <section class="panel">
          <div class="panel-head"><h2 data-i18n="mcpCatalog">MCP 目录</h2></div>
          <div class="panel-body" id="mcp-catalog">正在加载 MCP 目录...</div>
        </section>
        <section class="panel">
          <div class="panel-head"><h2 data-i18n="traceDetail">追踪详情</h2></div>
          <div class="panel-body" id="detail"><p class="muted" data-i18n="selectTrace">从表格中选择一条追踪。</p></div>
        </section>
      </aside>
    </div>
  </main>
  <script>
    const rows = document.querySelector("#rows");
    const detail = document.querySelector("#detail");
    const summary = document.querySelector("#summary");
    const statusFilter = document.querySelector("#status-filter");
    const sendResult = document.querySelector("#send-result");
    const auditRows = document.querySelector("#audit-rows");
    const auditMessage = document.querySelector("#audit-message");
    const routeExplain = document.querySelector("#route-explain");
    const runtimeHealth = document.querySelector("#runtime-health");
    const toolSelect = document.querySelector("#tool-select");
    const toolMeta = document.querySelector("#tool-meta");
    const toolResult = document.querySelector("#tool-result");
    const mcpCatalog = document.querySelector("#mcp-catalog");
    let allTraces = [];
    let traceTotal = 0;
    let traceStats = { total: 0, completed: 0, failed: 0, fallbacks: 0 };
    let tracePage = 1;
    const tracePageSize = 20;
    let allAuditEvents = [];
    let auditTotal = 0;
    let auditPage = 1;
    const auditPageSize = 8;
    let allTools = [];
    let lang = localStorage.getItem("gatewayConsoleLang") || "zh";

    const i18n = {
      zh: {
        title: "客户端 AI 网关控制台",
        subtitle: "本地模型/工具网关 · 追踪 · 策略 · 降级",
        sendTest: "发送测试请求",
        refresh: "刷新",
        export: "导出",
        totalTraces: "追踪总数",
        completed: "已完成",
        failed: "失败",
        fallbackAttempts: "降级次数",
        requestTraces: "请求追踪",
        allStatus: "全部状态",
        status: "状态",
        traceId: "追踪 ID",
        app: "应用",
        requestedModel: "请求模型",
        provider: "Provider",
        finalModel: "最终模型",
        policyRule: "策略规则",
        fallbacks: "降级",
        startedAt: "开始时间",
        error: "错误",
        prev: "上一页",
        next: "下一页",
        auditEvents: "审计事件",
        action: "动作",
        result: "结果",
        appTarget: "应用 / 目标",
        time: "时间",
        duration: "\u8017\u65f6",
        runtimeStatus: "运行时状态",
        daemon: "网关进程",
        config: "配置",
        stores: "存储",
        providerMonitor: "Provider 监控",
        modelRuntime: "模型运行时",
        mcpRuntime: "MCP 运行时",
        listenAddr: "监听地址",
        policyVersion: "策略版本",
        apps: "应用数",
        providers: "Provider 数",
        reloads: "重载次数",
        traceStore: "Trace 存储",
        auditStore: "Audit 存储",
        lastReloadedAt: "最近重载",
        notConfigured: "未配置",
        configured: "已配置",
        loaded: "已加载",
        availableStatus: "可用",
        running: "运行中",
        unavailable: "不可用",
        mcpCounts: "Server / Tool",
        providersHealth: "Provider / 健康状态",
        quickRequest: "快捷请求",
        modeSuccess: "成功请求",
        modeFallback: "本地失败后降级到云端",
        modeBlocked: "敏感数据阻止云端降级",
        send: "发送",
        explain: "解释路由",
        routingExplain: "路由解释",
        routingExplainHint: "点击“解释路由”预览策略和 Provider 路由。",
        tools: "\u5de5\u5177\u8c03\u7528",
        mcpCatalog: "MCP 目录",
        loadingMCP: "正在加载 MCP 目录...",
        noMCPServers: "暂无 MCP Server。",
        mcpMode: "模式",
        toolCount: "工具数",
        loadingTools: "\u6b63\u5728\u52a0\u8f7d\u5de5\u5177...",
        noTools: "\u6682\u65e0\u53ef\u7528\u5de5\u5177\u3002",
        invokeTool: "\u6267\u884c\u5de5\u5177",
        invokingTool: "\u5de5\u5177\u6267\u884c\u4e2d...",
        toolResultPlaceholder: "\u5de5\u5177\u6267\u884c\u7ed3\u679c\u4f1a\u663e\u793a\u5728\u8fd9\u91cc\u3002",
        toolTokenRequired: "\u9700\u8981\u5177\u6709 tool \u6388\u6743\u7684\u4ee4\u724c\u3002",
        selectToolFirst: "\u8bf7\u5148\u9009\u62e9\u5de5\u5177\u3002",
        readOnlyTool: "\u53ea\u8bfb",
        writeTool: "\u53ef\u5199",
        traceDetail: "追踪详情",
        selectTrace: "从表格中选择一条追踪。",
        loadingTraces: "正在加载追踪...",
        loadingProviders: "正在加载 Provider...",
        loadingAudit: "正在加载审计事件...",
        loadingRuntime: "正在加载运行时状态...",
        ready: "就绪。",
        adminTokenRequired: "需要管理员令牌。",
        noAuditEvents: "暂无审计事件。",
        failedPrefix: "失败：",
        statusCompleted: "已完成",
        statusFailed: "失败",
        statusStarted: "已开始",
        resultSuccess: "成功",
        resultDenied: "拒绝",
        resultFailed: "失败",
        actionConfigReload: "配置重载",
        actionPolicyDryRun: "策略试算",
        actionProviderEnabled: "Provider 启停",
        actionProviderProbe: "Provider 探测",
        runtimeHealthy: "健康",
        runtimeDegraded: "降级",
        runtimeUnhealthy: "异常",
        runtimeDisabled: "已禁用",
        sending: "发送中...",
        okPrefix: "成功：",
        probingPrefix: "正在探测 ",
        enablingPrefix: "正在启用 ",
        disablingPrefix: "正在禁用 ",
        explainLoading: "正在解释路由...",
        rule: "规则",
        cloud: "云端",
        reason: "原因",
        candidates: "候选 Provider",
        skipped: "跳过的 Provider",
        none: "无。",
        allowed: "允许",
        blocked: "阻止",
        enabled: "启用",
        disabled: "禁用",
        runtime: "运行时",
        healthy: "健康",
        unhealthy: "异常",
        available: "可用",
        enable: "启用",
        disable: "禁用",
        probe: "探测",
        page: "第 {page} / {total} 页",
        range: "{range} / {total}，最新优先",
        auditRange: "{range} / {total} | 第 {page} / {pages} 页",
        loadConsoleFailed: "控制台加载失败："
      },
      en: {
        title: "Client AI Gateway Console",
        subtitle: "Local model/tool gateway · trace · policy · fallback",
        sendTest: "Send Test Request",
        refresh: "Refresh",
        export: "Export",
        totalTraces: "Total Traces",
        completed: "Completed",
        failed: "Failed",
        fallbackAttempts: "Fallback Attempts",
        requestTraces: "Request Traces",
        allStatus: "All status",
        status: "Status",
        traceId: "Trace ID",
        app: "App",
        requestedModel: "Requested Model",
        provider: "Provider",
        finalModel: "Final Model",
        policyRule: "Policy Rule",
        fallbacks: "Fallbacks",
        startedAt: "Started At",
        error: "Error",
        prev: "Prev",
        next: "Next",
        auditEvents: "Audit Events",
        action: "Action",
        result: "Result",
        appTarget: "App / Target",
        time: "Time",
        duration: "Duration",
        runtimeStatus: "Runtime Status",
        daemon: "Daemon",
        config: "Config",
        stores: "Stores",
        providerMonitor: "Provider Monitor",
        modelRuntime: "Model Runtime",
        mcpRuntime: "MCP Runtime",
        listenAddr: "Listen Address",
        policyVersion: "Policy Version",
        apps: "Apps",
        providers: "Providers",
        reloads: "Reloads",
        traceStore: "Trace Store",
        auditStore: "Audit Store",
        lastReloadedAt: "Last Reloaded At",
        notConfigured: "not configured",
        configured: "configured",
        loaded: "loaded",
        availableStatus: "available",
        running: "running",
        unavailable: "unavailable",
        mcpCounts: "Servers / Tools",
        providersHealth: "Providers / Health",
        quickRequest: "Quick Request",
        modeSuccess: "success",
        modeFallback: "fallback local -> cloud",
        modeBlocked: "sensitive blocks cloud fallback",
        send: "Send",
        explain: "Explain",
        routingExplain: "Routing Explain",
        routingExplainHint: "Run Explain to preview policy and provider routing.",
        tools: "Tool Invocation",
        mcpCatalog: "MCP Catalog",
        loadingMCP: "Loading MCP catalog...",
        noMCPServers: "No MCP servers.",
        mcpMode: "Mode",
        toolCount: "Tools",
        loadingTools: "Loading tools...",
        noTools: "No tools available.",
        invokeTool: "Invoke Tool",
        invokingTool: "Invoking tool...",
        toolResultPlaceholder: "Tool result will appear here.",
        toolTokenRequired: "A token with the tool grant is required.",
        selectToolFirst: "Select a tool first.",
        readOnlyTool: "read-only",
        writeTool: "write",
        traceDetail: "Trace Detail",
        selectTrace: "Select a trace from the table.",
        loadingTraces: "Loading traces...",
        loadingProviders: "Loading providers...",
        loadingAudit: "Loading audit events...",
        loadingRuntime: "Loading runtime status...",
        ready: "Ready.",
        adminTokenRequired: "Admin token required.",
        noAuditEvents: "No audit events.",
        failedPrefix: "Failed: ",
        statusCompleted: "completed",
        statusFailed: "failed",
        statusStarted: "started",
        resultSuccess: "success",
        resultDenied: "denied",
        resultFailed: "failed",
        actionConfigReload: "config reload",
        actionPolicyDryRun: "policy dry-run",
        actionProviderEnabled: "provider enabled",
        actionProviderProbe: "provider probe",
        runtimeHealthy: "healthy",
        runtimeDegraded: "degraded",
        runtimeUnhealthy: "unhealthy",
        runtimeDisabled: "disabled",
        sending: "Sending...",
        okPrefix: "Success: ",
        probingPrefix: "Probing ",
        enablingPrefix: "Enabling ",
        disablingPrefix: "Disabling ",
        explainLoading: "Explaining route...",
        rule: "Rule",
        cloud: "Cloud",
        reason: "Reason",
        candidates: "Candidate Providers",
        skipped: "Skipped Providers",
        none: "None.",
        allowed: "allowed",
        blocked: "blocked",
        enabled: "enabled",
        disabled: "disabled",
        runtime: "runtime",
        healthy: "healthy",
        unhealthy: "unhealthy",
        available: "available",
        enable: "Enable",
        disable: "Disable",
        probe: "Probe",
        page: "Page {page} / {total}",
        range: "{range} of {total}, newest first",
        auditRange: "{range} of {total} | Page {page} / {pages}",
        loadConsoleFailed: "Failed to load console: "
      }
    };

    document.querySelector("#refresh").addEventListener("click", loadAll);
    document.querySelector("#audit-refresh").addEventListener("click", loadAudit);
    document.querySelector("#trace-export").addEventListener("click", exportTraces);
    document.querySelector("#audit-export").addEventListener("click", exportAudit);
    document.querySelector("#lang-toggle").addEventListener("click", () => setLang(lang === "zh" ? "en" : "zh"));
    document.querySelector("#sample").addEventListener("click", () => sendQuick("success"));
    document.querySelector("#send").addEventListener("click", () => sendQuick(document.querySelector("#mode").value));
    document.querySelector("#explain").addEventListener("click", () => explainRouting(document.querySelector("#mode").value));
    document.querySelector("#tool-invoke").addEventListener("click", invokeTool);
    toolSelect.addEventListener("change", renderSelectedTool);
    document.querySelector("#trace-prev").addEventListener("click", () => { tracePage = Math.max(1, tracePage - 1); loadTraces(); });
    document.querySelector("#trace-next").addEventListener("click", () => { tracePage += 1; loadTraces(); });
    document.querySelector("#audit-prev").addEventListener("click", () => { auditPage = Math.max(1, auditPage - 1); loadAudit(); });
    document.querySelector("#audit-next").addEventListener("click", () => { auditPage += 1; loadAudit(); });
    statusFilter.addEventListener("change", () => { tracePage = 1; loadTraces(); });

    function t(key, vars = {}) {
      let value = (i18n[lang] && i18n[lang][key]) || i18n.zh[key] || key;
      Object.entries(vars).forEach(([name, replacement]) => {
        value = value.replaceAll("{" + name + "}", replacement);
      });
      return value;
    }
    function setLang(nextLang) {
      lang = nextLang;
      localStorage.setItem("gatewayConsoleLang", lang);
      document.documentElement.lang = lang === "zh" ? "zh-CN" : "en";
      document.title = t("title");
      document.querySelector("#lang-toggle").textContent = lang === "zh" ? "English" : "中文";
      document.querySelectorAll("[data-i18n]").forEach(node => {
        node.textContent = t(node.dataset.i18n);
      });
      if (!allTraces.length) summary.textContent = t("loadingTraces");
      renderTraces();
      renderAudit();
      renderTools();
    }
    function esc(value) {
      return String(value ?? "").replace(/[&<>"']/g, ch => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&#39;" }[ch]));
    }
    function time(value) {
      if (!value) return "";
      return new Date(value).toLocaleString();
    }
    function shortTraceID(value) {
      if (!value) return "-";
      return value.length > 16 ? value.slice(0, 16) + "..." : value;
    }
    function labelStatus(value) {
      return ({ completed: t("statusCompleted"), failed: t("statusFailed"), started: t("statusStarted") })[value] || value || "";
    }
    function labelResult(value) {
      return ({ success: t("resultSuccess"), denied: t("resultDenied"), failed: t("resultFailed") })[value] || value || "";
    }
    function labelAction(value) {
      return ({
        "config.reload": t("actionConfigReload"),
        "policy.dry_run": t("actionPolicyDryRun"),
        "provider.enabled": t("actionProviderEnabled"),
        "provider.probe": t("actionProviderProbe"),
        "tool.invoke": t("tools")
      })[value] || value || "";
    }
    function labelRuntime(value) {
      return ({
        healthy: t("runtimeHealthy"),
        degraded: t("runtimeDegraded"),
        unhealthy: t("runtimeUnhealthy"),
        disabled: t("runtimeDisabled"),
        not_configured: t("notConfigured"),
        configured: t("configured"),
        loaded: t("loaded"),
        available: t("availableStatus"),
        running: t("running"),
        unavailable: t("unavailable")
      })[value] || value || "";
    }
    async function loadAll() {
      await Promise.all([loadTraces(), loadRuntimeHealth(), loadProviders(), loadModels(), loadAudit(), loadTools(), loadMCPCatalog()]);
    }
    function traceQuery(limit = tracePageSize, offset = (tracePage - 1) * tracePageSize) {
      const query = new URLSearchParams({ limit: String(limit), offset: String(offset) });
      if (statusFilter.value) query.set("status", statusFilter.value);
      return query;
    }
    async function loadTraces() {
      const [pageRes, statsRes] = await Promise.all([
        fetch("/gateway/v1/traces?" + traceQuery().toString()),
        fetch("/gateway/v1/traces?limit=500")
      ]);
      const data = await pageRes.json();
      const statsData = await statsRes.json();
      allTraces = data.traces || [];
      traceTotal = data.total || allTraces.length;
      const statsTraces = statsData.traces || [];
      traceStats = {
        total: statsData.total ?? statsTraces.length,
        completed: statsTraces.filter(item => item.status === "completed").length,
        failed: statsTraces.filter(item => item.status === "failed").length,
        fallbacks: statsTraces.reduce((sum, item) => sum + ((item.fallbacks || []).length), 0)
      };
      document.querySelector("#m-total").textContent = traceStats.total;
      document.querySelector("#m-completed").textContent = traceStats.completed;
      document.querySelector("#m-failed").textContent = traceStats.failed;
      document.querySelector("#m-fallbacks").textContent = traceStats.fallbacks;
      renderTraces();
    }
    function downloadURL(url, filename = "") {
      const link = document.createElement("a");
      link.href = url;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      link.remove();
    }
    function exportTraces() {
      downloadURL("/gateway/v1/traces/export?" + traceQuery(500, 0).toString(), "traces.jsonl");
    }
    function renderTraces() {
      const totalPages = pageCount(traceTotal, tracePageSize);
      tracePage = clampPage(tracePage, totalPages);
      const range = pageRange(traceTotal, tracePage, tracePageSize);
      summary.textContent = t("range", { range: range.label, total: traceTotal });
      document.querySelector("#trace-page-summary").textContent = t("page", { page: tracePage, total: totalPages });
      document.querySelector("#trace-prev").disabled = tracePage <= 1;
      document.querySelector("#trace-next").disabled = tracePage >= totalPages;
      rows.innerHTML = allTraces.map(item =>
        "<tr data-trace=\"" + esc(item.trace_id) + "\">" +
          "<td><span class=\"status " + esc(item.status) + "\">" + esc(labelStatus(item.status)) + "</span></td>" +
          "<td class=\"trace-id\">" + esc(item.trace_id) + "</td>" +
          "<td>" + esc(item.app_id) + "</td>" +
          "<td>" + esc(item.requested_model) + "</td>" +
          "<td>" + esc(item.provider_id) + "</td>" +
          "<td>" + esc(item.final_model) + "</td>" +
          "<td>" + esc(item.policy && item.policy.rule_id) + "</td>" +
          "<td>" + ((item.fallbacks || []).length) + "</td>" +
          "<td>" + esc(time(item.started_at)) + "</td>" +
          "<td title=\"" + esc(item.error) + "\">" + esc(item.error) + "</td>" +
        "</tr>"
      ).join("");
      rows.querySelectorAll("tr").forEach(row => row.addEventListener("click", () => loadDetail(row.dataset.trace)));
    }
    function pageCount(total, size) {
      return Math.max(1, Math.ceil(total / size));
    }
    function clampPage(page, totalPages) {
      return Math.max(1, Math.min(page, totalPages));
    }
    function pageRange(total, page, size) {
      if (!total) return { start: 0, end: 0, label: "0" };
      const start = (page - 1) * size + 1;
      const end = Math.min(total, start + size - 1);
      return { start, end, label: start + "-" + end };
    }
    async function loadRuntimeHealth() {
      runtimeHealth.textContent = t("loadingRuntime");
      try {
        const res = await fetch("/gateway/v1/runtime/health");
        const data = await res.json();
        if (!res.ok) {
          runtimeHealth.textContent = t("failedPrefix") + (data.error && data.error.message || res.status);
          return;
        }
        runtimeHealth.innerHTML =
          "<div class=\"kv\">" +
            "<div class=\"k\">" + t("status") + "</div><div><span class=\"status " + esc(data.status) + "\">" + esc(labelRuntime(data.status)) + "</span></div>" +
            "<div class=\"k\">" + t("daemon") + "</div><div>" + esc(labelRuntime(data.daemon && data.daemon.status)) + "</div>" +
            "<div class=\"k\">" + t("listenAddr") + "</div><div>" + esc(data.config && data.config.listen_addr) + "</div>" +
            "<div class=\"k\">" + t("policyVersion") + "</div><div>" + esc(data.config && data.config.policy_version) + "</div>" +
            "<div class=\"k\">" + t("apps") + "</div><div>" + esc(data.config && data.config.app_count) + "</div>" +
            "<div class=\"k\">" + t("providers") + "</div><div>" + esc(data.config && data.config.provider_count) + "</div>" +
            "<div class=\"k\">" + t("reloads") + "</div><div>" + esc(data.config && data.config.reload_count) + "</div>" +
            "<div class=\"k\">" + t("lastReloadedAt") + "</div><div>" + esc(time(data.config && data.config.last_reloaded_at)) + "</div>" +
            "<div class=\"k\">" + t("traceStore") + "</div><div>" + esc(data.stores && data.stores.trace && data.stores.trace.path) + "</div>" +
            "<div class=\"k\">" + t("auditStore") + "</div><div>" + esc(data.stores && data.stores.audit && data.stores.audit.path) + "</div>" +
            "<div class=\"k\">" + t("providerMonitor") + "</div><div>" + renderProviderMonitor(data.provider_monitor || {}) + "</div>" +
            "<div class=\"k\">" + t("modelRuntime") + "</div><div>" + esc(labelRuntime(data.model_runtime && data.model_runtime.status)) + "</div>" +
            "<div class=\"k\">" + t("mcpRuntime") + "</div><div>" + renderComponentHealth(data.mcp_runtime || {}) + "</div>" +
          "</div>";
      } catch (err) {
        runtimeHealth.textContent = t("failedPrefix") + err.message;
      }
    }
    function renderComponentHealth(component) {
      const countText = component.server_count !== undefined
        ? " / " + t("mcpCounts") + ": " + esc(component.enabled_servers || 0) + "/" + esc(component.server_count || 0) + " · " + esc(component.enabled_tools || 0) + "/" + esc(component.tool_count || 0)
        : "";
      const modeText = component.mode ? " / " + esc(component.mode) : "";
      const reasonText = component.reason ? " / " + esc(component.reason) : "";
      return esc(labelRuntime(component.status)) + modeText + countText + reasonText;
    }
    function renderProviderMonitor(monitor) {
      return esc(labelRuntime(monitor.status)) + " / " +
        esc(t("healthy")) + ": " + esc(monitor.healthy || 0) + " / " +
        esc(t("runtimeDegraded")) + ": " + esc(monitor.degraded || 0) + " / " +
        esc(t("unhealthy")) + ": " + esc(monitor.unhealthy || 0) + " / " +
        esc(t("disabled")) + ": " + esc(monitor.disabled || 0);
    }
    async function loadProviders() {
      const res = await fetch("/gateway/v1/providers");
      const data = await res.json();
      const providers = data.providers || [];
      document.querySelector("#providers").innerHTML = providers.map(item =>
        "<div class=\"provider\">" +
          "<strong>" + esc(item.name || item.id) + "</strong>" +
          "<div class=\"muted\">" + esc(item.id) + " / " + esc(item.class) + " / " + esc(item.adapter || "mock") + "</div>" +
          "<div class=\"muted\">" + (item.enabled === false ? t("disabled") : t("enabled")) + " / " + t("runtime") + ": " + esc(labelRuntime(item.runtime_status) || (item.healthy ? t("healthy") : t("unhealthy"))) + "</div>" +
          (item.degraded_reason ? "<div class=\"muted\">" + esc(item.degraded_reason) + "</div>" : "") +
          "<div>" + esc((item.models || []).join(", ")) + "</div>" +
          "<div class=\"provider-actions\">" +
            "<button class=\"secondary\" data-provider=\"" + esc(item.id) + "\" data-action=\"probe\">" + t("probe") + "</button>" +
            "<button class=\"secondary\" data-provider=\"" + esc(item.id) + "\" data-action=\"toggle\" data-enabled=\"" + (item.enabled === false ? "true" : "false") + "\">" + (item.enabled === false ? t("enable") : t("disable")) + "</button>" +
          "</div>" +
        "</div>"
      ).join("");
      document.querySelectorAll("[data-action='probe']").forEach(button => button.addEventListener("click", () => probeProvider(button.dataset.provider)));
      document.querySelectorAll("[data-action='toggle']").forEach(button => button.addEventListener("click", () => setProviderEnabled(button.dataset.provider, button.dataset.enabled === "true")));
    }
    async function loadModels() {
      const res = await fetch("/gateway/v1/models");
      const data = await res.json();
      const options = data.models || [];
      document.querySelector("#model-options").innerHTML = options.map(item =>
        "<option value=\"" + esc(item.model) + "\">" + esc(item.provider_id) + " / " + esc(labelRuntime(item.runtime_status) || t("available")) + "</option>"
      ).join("");
    }
    async function loadMCPCatalog() {
      mcpCatalog.textContent = t("loadingMCP");
      try {
        const res = await fetch("/gateway/v1/mcp/servers");
        const data = await res.json();
        if (!res.ok) {
          mcpCatalog.textContent = t("failedPrefix") + (data.error && data.error.message || res.status);
          return;
        }
        const servers = data.servers || [];
        if (!servers.length) {
          mcpCatalog.innerHTML = "<div class=\"muted\">" + t("mcpMode") + ": " + esc(data.mode || "-") + " / " + t("noMCPServers") + "</div>";
          return;
        }
        mcpCatalog.innerHTML =
          "<div class=\"muted\">" + t("mcpMode") + ": " + esc(data.mode || "-") + " / " + (data.enabled ? t("enabled") : t("disabled")) + "</div>" +
          servers.map(server =>
            "<div class=\"provider\">" +
              "<strong>" + esc(server.name || server.id) + "</strong>" +
              "<div class=\"muted\">" + esc(server.id) + " / " + (server.enabled ? t("enabled") : t("disabled")) + " / " + t("toolCount") + ": " + esc(server.enabled_tools || 0) + "/" + esc(server.tool_count || 0) + "</div>" +
              (server.tools || []).map(tool =>
                "<div class=\"mcp-tool\">" +
                  "<strong>" + esc(tool.name || tool.id) + "</strong>" +
                  "<div class=\"muted\">" + esc(tool.id) + " / " + esc(tool.risk_level || "-") + " / " + (tool.read_only ? t("readOnlyTool") : t("writeTool")) + " / " + ((tool.scopes || []).map(esc).join(", ") || "-") + "</div>" +
                  (tool.description ? "<div class=\"muted\">" + esc(tool.description) + "</div>" : "") +
                "</div>"
              ).join("") +
            "</div>"
          ).join("");
      } catch (err) {
        mcpCatalog.textContent = t("failedPrefix") + err.message;
      }
    }
    async function loadTools() {
      toolMeta.textContent = t("loadingTools");
      try {
        const res = await fetch("/gateway/v1/tools");
        const data = await res.json();
        if (!res.ok) {
          allTools = [];
          toolSelect.innerHTML = "";
          toolMeta.textContent = t("failedPrefix") + (data.error && data.error.message || res.status);
          return;
        }
        allTools = data.tools || [];
        renderTools();
      } catch (err) {
        allTools = [];
        toolSelect.innerHTML = "";
        toolMeta.textContent = t("failedPrefix") + err.message;
      }
    }
    function renderTools() {
      if (!allTools.length) {
        toolSelect.innerHTML = "";
        toolMeta.textContent = t("noTools");
        document.querySelector("#tool-invoke").disabled = true;
        return;
      }
      const selected = toolSelect.value || allTools[0].id;
      toolSelect.innerHTML = allTools.map(item =>
        "<option value=\"" + esc(item.id) + "\"" + (item.id === selected ? " selected" : "") + ">" + esc(item.name || item.id) + "</option>"
      ).join("");
      document.querySelector("#tool-invoke").disabled = false;
      renderSelectedTool();
    }
    function renderSelectedTool() {
      const tool = allTools.find(item => item.id === toolSelect.value);
      if (!tool) {
        toolMeta.textContent = t("noTools");
        return;
      }
      toolMeta.textContent = tool.id + " / " + (tool.origin || "builtin") + (tool.server_id ? ":" + tool.server_id : "") + " / " + tool.adapter + " / " + (tool.read_only ? t("readOnlyTool") : t("writeTool")) + " / " + (tool.risk_level || "-") + " / " + ((tool.scopes || []).join(", ") || "-");
    }
    async function invokeTool() {
      const toolID = toolSelect.value;
      const token = document.querySelector("#tool-token").value.trim();
      if (!toolID) {
        toolResult.textContent = t("selectToolFirst");
        return;
      }
      if (!token) {
        toolResult.textContent = t("toolTokenRequired");
        return;
      }
      toolResult.textContent = t("invokingTool");
      try {
        const res = await fetch("/gateway/v1/tools/" + encodeURIComponent(toolID) + "/invoke", {
          method: "POST",
          headers: { "Authorization": "Bearer " + token, "Content-Type": "application/json" },
          body: JSON.stringify({})
        });
        const data = await res.json();
        toolResult.textContent = JSON.stringify(data, null, 2);
        await Promise.all([loadAudit(), loadTraces()]);
      } catch (err) {
        toolResult.textContent = t("failedPrefix") + err.message;
      }
    }
    function adminToken() {
      return document.querySelector("#admin-token").value.trim();
    }
    async function providerRequest(path, options = {}) {
      const token = adminToken();
      if (!token) throw new Error(t("adminTokenRequired"));
      const res = await fetch(path, {
        ...options,
        headers: {
          "Authorization": "Bearer " + token,
          ...(options.headers || {})
        }
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error && data.error.message || String(res.status));
      return data;
    }
    async function probeProvider(providerId) {
      auditMessage.textContent = t("probingPrefix") + providerId + "...";
      try {
        await providerRequest("/gateway/v1/providers/" + encodeURIComponent(providerId) + "/probe", { method: "POST" });
        await Promise.all([loadProviders(), loadModels(), loadAudit()]);
      } catch (err) {
        auditMessage.textContent = t("failedPrefix") + err.message;
      }
    }
    async function setProviderEnabled(providerId, enabled) {
      auditMessage.textContent = (enabled ? t("enablingPrefix") : t("disablingPrefix")) + providerId + "...";
      try {
        await providerRequest("/gateway/v1/providers/" + encodeURIComponent(providerId) + "/enabled", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ enabled })
        });
        await Promise.all([loadProviders(), loadModels(), loadAudit()]);
      } catch (err) {
        auditMessage.textContent = t("failedPrefix") + err.message;
      }
    }
    async function loadDetail(traceId) {
      const res = await fetch("/gateway/v1/traces/" + encodeURIComponent(traceId));
      const trace = await res.json();
      detail.innerHTML =
        "<div class=\"kv\">" +
          "<div class=\"k\">" + t("traceId") + "</div><div class=\"trace-id\">" + esc(trace.trace_id) + "</div>" +
          "<div class=\"k\">" + t("status") + "</div><div><span class=\"status " + esc(trace.status) + "\">" + esc(labelStatus(trace.status)) + "</span></div>" +
          "<div class=\"k\">" + t("provider") + "</div><div>" + esc(trace.provider_id) + "</div>" +
          "<div class=\"k\">" + t("policyRule") + "</div><div>" + esc(trace.policy && trace.policy.rule_id) + " / " + esc(trace.policy && trace.policy.explanation) + "</div>" +
          "<div class=\"k\">" + t("fallbacks") + "</div><div>" + ((trace.fallbacks || []).length) + "</div>" +
        "</div>" +
        "<pre>" + esc(JSON.stringify(trace, null, 2)) + "</pre>";
    }
    async function loadAudit() {
      const token = adminToken();
      if (!token) {
        auditRows.innerHTML = "";
        auditMessage.textContent = t("adminTokenRequired");
        return;
      }
      try {
        auditMessage.textContent = t("loadingAudit");
        const offset = (auditPage - 1) * auditPageSize;
        const query = new URLSearchParams({ limit: String(auditPageSize), offset: String(offset) });
        const res = await fetch("/gateway/v1/audit/events?" + query.toString(), {
          headers: { "Authorization": "Bearer " + token }
        });
        const data = await res.json();
        if (!res.ok) {
          auditRows.innerHTML = "";
          auditMessage.textContent = t("failedPrefix") + (data.error && data.error.message || res.status);
          return;
        }
        allAuditEvents = data.events || [];
        auditTotal = data.total || allAuditEvents.length;
        auditMessage.textContent = auditTotal ? "" : t("noAuditEvents");
        renderAudit();
      } catch (err) {
        auditRows.innerHTML = "";
        auditMessage.textContent = t("failedPrefix") + err.message;
      }
    }
    function exportAudit() {
      const token = adminToken();
      if (!token) {
        auditMessage.textContent = t("adminTokenRequired");
        return;
      }
      fetch("/gateway/v1/audit/events/export?limit=500", {
        headers: { "Authorization": "Bearer " + token }
      })
        .then(async res => {
          if (!res.ok) {
            const data = await res.json();
            throw new Error(data.error && data.error.message || String(res.status));
          }
          return res.blob();
        })
        .then(blob => {
          const url = URL.createObjectURL(blob);
          downloadURL(url, "audit-events.jsonl");
          setTimeout(() => URL.revokeObjectURL(url), 1000);
        })
        .catch(err => {
          auditMessage.textContent = t("failedPrefix") + err.message;
        });
    }
    function renderAudit() {
      const totalPages = pageCount(auditTotal, auditPageSize);
      auditPage = clampPage(auditPage, totalPages);
      const range = pageRange(auditTotal, auditPage, auditPageSize);
      document.querySelector("#audit-page-summary").textContent = t("auditRange", { range: range.label, total: auditTotal, page: auditPage, pages: totalPages });
      document.querySelector("#audit-prev").disabled = auditPage <= 1;
      document.querySelector("#audit-next").disabled = auditPage >= totalPages;
      auditRows.innerHTML = allAuditEvents.map(item =>
        "<tr data-trace=\"" + esc(item.trace_id || "") + "\" title=\"" + esc(item.error || item.trace_id || "") + "\">" +
          "<td>" + esc(labelAction(item.action)) + "</td>" +
          "<td><span class=\"status " + esc(item.result) + "\">" + esc(labelResult(item.result)) + "</span></td>" +
          "<td class=\"trace-id\">" + esc(shortTraceID(item.trace_id)) + "</td>" +
          "<td>" + esc(item.app_id || "-") + " / " + esc(item.target || "-") + "</td>" +
          "<td>" + esc(time(item.created_at)) + "</td>" +
          "<td>" + esc(item.duration_ms == null ? "-" : item.duration_ms + "ms") + "</td>" +
        "</tr>"
      ).join("");
      auditRows.querySelectorAll("tr[data-trace]").forEach(row => {
        if (row.dataset.trace) row.addEventListener("click", () => loadDetail(row.dataset.trace));
      });
    }
    async function sendQuick(mode) {
      const body = {
        model: document.querySelector("#model").value,
        messages: [{ role: "user", content: document.querySelector("#prompt").value }],
        metadata: {},
        data_labels: []
      };
      if (mode === "fallback" || mode === "blocked") body.metadata.fail_provider = "local-mock";
      if (mode === "blocked") body.data_labels = ["sensitive"];
      sendResult.textContent = t("sending");
      try {
        const res = await fetch("/v1/chat/completions", {
          method: "POST",
          headers: { "Authorization": "Bearer dev-token", "Content-Type": "application/json" },
          body: JSON.stringify(body)
        });
        const data = await res.json();
        sendResult.textContent = res.ok ? (t("okPrefix") + data.trace_id) : (t("failedPrefix") + (data.error && data.error.trace_id));
      } catch (err) {
        sendResult.textContent = t("failedPrefix") + err.message;
      }
      await loadTraces();
    }
    function requestLabelsForMode(mode) {
      return mode === "blocked" ? ["sensitive"] : [];
    }
    async function explainRouting(mode) {
      routeExplain.innerHTML = "<p class=\"muted\">" + t("explainLoading") + "</p>";
      const body = {
        app_id: "dev-app",
        request_type: "chat",
        model: document.querySelector("#model").value,
        data_labels: requestLabelsForMode(mode)
      };
      try {
        const res = await fetch("/gateway/v1/routing/explain", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(body)
        });
        const data = await res.json();
        if (!res.ok) {
          routeExplain.textContent = t("failedPrefix") + (data.error && data.error.message || res.status);
          return;
        }
        routeExplain.innerHTML =
          "<div class=\"kv\">" +
            "<div class=\"k\">" + t("rule") + "</div><div>" + esc(data.policy && data.policy.rule_id) + "</div>" +
            "<div class=\"k\">" + t("cloud") + "</div><div>" + (data.policy && data.policy.allow_cloud ? t("allowed") : t("blocked")) + "</div>" +
            "<div class=\"k\">" + t("reason") + "</div><div>" + esc(data.policy && data.policy.explanation) + "</div>" +
          "</div>" +
          "<h2>" + t("candidates") + "</h2>" +
          renderRouteItems(data.candidates || [], false) +
          "<h2 style=\"margin-top:12px;\">" + t("skipped") + "</h2>" +
          renderRouteItems(data.skipped || [], true);
      } catch (err) {
        routeExplain.textContent = t("failedPrefix") + err.message;
      }
    }
    function renderRouteItems(items, skipped) {
      if (!items.length) return "<p class=\"muted\">" + t("none") + "</p>";
      return items.map(item => {
        const provider = skipped ? { id: item.provider_id, class: item.class } : (item.provider || {});
        return "<div class=\"route-item" + (skipped ? " skipped" : "") + "\">" +
          "<strong>" + esc(provider.id || "-") + "</strong>" +
          "<div class=\"muted\">" + esc(provider.class || "-") + (item.model ? " / " + esc(item.model) : "") + "</div>" +
          "<div>" + esc(item.reason || "") + "</div>" +
        "</div>";
      }).join("");
    }
    setLang(lang);
    loadAll().catch(err => { summary.textContent = t("loadConsoleFailed") + err.message; });
  </script>
</body>
</html>`
