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
    .tool-table, .mcp-table { min-width: 980px; }
    .issue-table { min-width: 980px; }
    .app-table { min-width: 1040px; }
    .grant-table { min-width: 1040px; }
    .policy-table { min-width: 1060px; }
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
    .tool-table th:nth-child(1), .tool-table td:nth-child(1) { width: 250px; }
    .tool-table th:nth-child(2), .tool-table td:nth-child(2) { width: 100px; }
    .tool-table th:nth-child(3), .tool-table td:nth-child(3) { width: 150px; }
    .tool-table th:nth-child(4), .tool-table td:nth-child(4) { width: 120px; }
    .tool-table th:nth-child(5), .tool-table td:nth-child(5) { width: 190px; }
    .tool-table th:nth-child(6), .tool-table td:nth-child(6) { width: 90px; }
    .mcp-table th:nth-child(1), .mcp-table td:nth-child(1) { width: 220px; }
    .mcp-table th:nth-child(2), .mcp-table td:nth-child(2) { width: 120px; }
    .mcp-table th:nth-child(3), .mcp-table td:nth-child(3) { width: 130px; }
    .mcp-table th:nth-child(4), .mcp-table td:nth-child(4) { width: 260px; }
    .mcp-table th:nth-child(5), .mcp-table td:nth-child(5) { width: 220px; }
    .issue-table th:nth-child(1), .issue-table td:nth-child(1) { width: 110px; }
    .issue-table th:nth-child(2), .issue-table td:nth-child(2) { width: 150px; }
    .issue-table th:nth-child(3), .issue-table td:nth-child(3) { width: 220px; }
    .issue-table th:nth-child(4), .issue-table td:nth-child(4) { width: 420px; }
    .issue-table th:nth-child(5), .issue-table td:nth-child(5) { width: 130px; }
    .app-table th:nth-child(1), .app-table td:nth-child(1) { width: 240px; }
    .app-table th:nth-child(2), .app-table td:nth-child(2) { width: 180px; }
    .app-table th:nth-child(3), .app-table td:nth-child(3) { width: 180px; }
    .app-table th:nth-child(4), .app-table td:nth-child(4) { width: 360px; }
    .policy-table th:nth-child(1), .policy-table td:nth-child(1) { width: 80px; }
    .policy-table th:nth-child(2), .policy-table td:nth-child(2) { width: 220px; }
    .policy-table th:nth-child(3), .policy-table td:nth-child(3) { width: 100px; }
    .policy-table th:nth-child(4), .policy-table td:nth-child(4) { width: 130px; }
    .policy-table th:nth-child(5), .policy-table td:nth-child(5) { width: 420px; }
    .policy-table th:nth-child(6), .policy-table td:nth-child(6) { width: 220px; }
    .grant-table th:nth-child(1), .grant-table td:nth-child(1) { width: 230px; }
    .grant-table th:nth-child(2), .grant-table td:nth-child(2) { width: 130px; }
    .grant-table th:nth-child(3), .grant-table td:nth-child(3) { width: 230px; }
    .grant-table th:nth-child(4), .grant-table td:nth-child(4) { width: 220px; }
    .grant-table th:nth-child(5), .grant-table td:nth-child(5) { width: 240px; }
    .trace-id { font-family: Consolas, "Courier New", monospace; color: var(--blue); }
    .status { display: inline-block; border-radius: 999px; padding: 2px 8px; font-weight: 700; font-size: 12px; }
    .completed, .success, .healthy, .available, .running, .loaded { color: var(--green); background: #e9f7ef; }
    .failed, .denied, .unhealthy { color: var(--red); background: #fdecec; }
    .started, .degraded, .disabled, .not_configured, .unavailable { color: var(--amber); background: #fff4df; }
    .warning { color: var(--amber); background: #fff4df; }
    .info { color: var(--blue); background: #eaf4fb; }
    .provider, .route-item { border: 1px solid var(--line); border-radius: 6px; padding: 10px; margin-bottom: 9px; }
    .provider strong, .route-item strong { display: block; margin-bottom: 3px; }
    .route-item.skipped { background: #fff7ed; }
    .explain-chain { border: 1px solid var(--line); border-radius: 6px; padding: 10px 12px; margin-bottom: 12px; background: #fbfdff; }
    .explain-chain .chain-title { font-weight: 700; margin-bottom: 8px; }
    .chain-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 8px 12px; }
    .chain-cell { min-width: 0; }
    .chain-cell .k { display: block; margin-bottom: 2px; }
    .mcp-tool { margin-top: 8px; padding: 8px; border-top: 1px solid var(--line); }
    .tool-meta { min-height: 20px; }
    .provider-actions button { padding: 5px 8px; font-size: 12px; }
    .filters { display: grid; grid-template-columns: repeat(4, minmax(120px, 1fr)) auto; gap: 8px; align-items: center; }
    .pager { border-top: 1px solid var(--line); padding: 9px 12px; display: flex; align-items: center; justify-content: space-between; gap: 10px; }
    .pager button { padding: 5px 9px; font-size: 12px; }
    .kv { display: grid; grid-template-columns: 130px minmax(0, 1fr); gap: 8px 10px; margin-bottom: 12px; }
    .k { color: var(--muted); }
    .notice { border: 1px solid #d4e5f6; background: #f3f9ff; color: #204b6a; border-radius: 6px; padding: 10px 12px; line-height: 1.55; margin-bottom: 12px; }
    .notice strong { display: block; margin-bottom: 2px; }
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
      .filters { grid-template-columns: 1fr; }
      .actions, .pager-actions { width: 100%; }
      .actions button, .pager-actions button { flex: 1; }
      .kv { grid-template-columns: 1fr; }
      .chain-grid { grid-template-columns: 1fr; }
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
              <h2 data-i18n="issueSummary">运行问题汇总</h2>
              <div class="muted" id="issue-page-summary">第 1 页</div>
            </div>
            <div class="actions">
              <button class="secondary" id="issue-refresh" data-i18n="refresh">刷新</button>
            </div>
          </div>
          <div class="panel-body">
            <div class="muted" id="issue-message" data-i18n="loadingIssues">正在汇总运行问题...</div>
          </div>
          <div class="table-wrap" style="height: 260px;">
            <table class="issue-table">
              <thead>
                <tr>
                  <th data-i18n="severity">级别</th>
                  <th data-i18n="source">来源</th>
                  <th data-i18n="target">目标</th>
                  <th data-i18n="issue">问题</th>
                  <th data-i18n="time">时间</th>
                </tr>
              </thead>
              <tbody id="issue-rows"></tbody>
            </table>
          </div>
          <div class="pager">
            <div class="muted" id="issue-range-summary">0 / 0</div>
            <div class="pager-actions">
              <button class="secondary" id="issue-prev" data-i18n="prev">上一页</button>
              <button class="secondary" id="issue-next" data-i18n="next">下一页</button>
            </div>
          </div>
        </section>
        <section class="panel">
          <div class="panel-head">
            <div>
              <h2 data-i18n="appCatalog">\u5e94\u7528\u4e0e\u6388\u6743</h2>
              <div class="muted" id="app-page-summary">\u7b2c 1 \u9875</div>
            </div>
            <div class="actions">
              <button class="secondary" id="app-export" data-i18n="export">Export</button>
              <button class="secondary" id="app-refresh" data-i18n="refresh">刷新</button>
            </div>
          </div>
          <div class="panel-body form-grid">
            <div class="filters">
              <input id="app-id-filter" data-i18n-placeholder="appFilter" placeholder="App ID" />
              <input id="app-grant-filter" data-i18n-placeholder="grantFilter" placeholder="Grant" />
              <select id="app-quota-filter">
                <option value="" data-i18n="allQuotaStates">All quota states</option>
                <option value="true" data-i18n="quotaEnabled">Quota enabled</option>
                <option value="false" data-i18n="quotaDisabled">Quota disabled</option>
              </select>
              <button class="secondary" id="app-filter-apply" data-i18n="applyFilter">筛选</button>
              <button class="secondary" id="app-filter-clear" data-i18n="clearFilter">清空</button>
            </div>
            <div class="muted" id="app-message">\u6b63\u5728\u52a0\u8f7d\u5e94\u7528\u6388\u6743...</div>
          </div>
          <div class="table-wrap" style="height: 260px;">
            <table class="app-table">
              <thead>
                <tr>
                  <th data-i18n="app">应用</th>
                  <th data-i18n="tokenHint">Token</th>
                  <th data-i18n="quota">Quota</th>
                  <th data-i18n="grants">授权</th>
                </tr>
              </thead>
              <tbody id="app-rows"></tbody>
            </table>
          </div>
          <div class="pager">
            <div class="muted" id="app-range-summary">0 / 0</div>
            <div class="pager-actions">
              <button class="secondary" id="app-prev" data-i18n="prev">上一页</button>
              <button class="secondary" id="app-next" data-i18n="next">下一页</button>
            </div>
          </div>
        </section>
        <section class="panel">
          <div class="panel-head"><h2 data-i18n="configReload">Config Reload</h2></div>
          <div class="panel-body form-grid">
            <button id="config-reload" data-i18n="reloadConfig">Reload Config</button>
            <pre id="config-reload-result" data-i18n="configReloadPlaceholder">Config reload result will appear here.</pre>
          </div>
        </section>
        <section class="panel">
          <div class="panel-head"><h2 data-i18n="routingDryRun">Routing Explain</h2></div>
          <div class="panel-body form-grid">
            <div class="filters">
              <input id="routing-dry-app-id" value="dev-app" data-i18n-placeholder="appFilter" placeholder="App ID" />
              <select id="routing-dry-request-type">
                <option value="chat" data-i18n="chatAction">Chat</option>
                <option value="tool.invoke" data-i18n="toolInvokeAction">Tool Invoke</option>
              </select>
              <input id="routing-dry-model" value="local-small" data-i18n-placeholder="modelFilter" placeholder="Model" />
              <input id="routing-dry-data-labels" data-i18n-placeholder="dataLabelFilter" placeholder="Data Label" />
              <button id="routing-dry-run" data-i18n="explain">Explain</button>
            </div>
          </div>
        </section>
        <section class="panel">
          <div class="panel-head"><h2 data-i18n="policyDryRun">Policy Dry-run</h2></div>
          <div class="panel-body form-grid">
            <div class="filters">
              <input id="policy-dry-app-id" value="dev-app" data-i18n-placeholder="appFilter" placeholder="App ID" />
              <select id="policy-dry-request-type">
                <option value="chat" data-i18n="chatAction">Chat</option>
                <option value="tool.invoke" data-i18n="toolInvokeAction">Tool Invoke</option>
              </select>
              <input id="policy-dry-model" value="local-small" data-i18n-placeholder="modelFilter" placeholder="Model" />
              <select id="policy-dry-provider-class">
                <option value="" data-i18n="allClasses">All classes</option>
                <option value="local" data-i18n="localClass">Local</option>
                <option value="cloud" data-i18n="cloudClass">Cloud</option>
              </select>
              <input id="policy-dry-data-labels" data-i18n-placeholder="dataLabelFilter" placeholder="Data Label" />
              <button id="policy-dry-run" data-i18n="dryRun">Dry-run</button>
            </div>
            <pre id="policy-dry-result" data-i18n="policyDryRunPlaceholder">Policy dry-run result will appear here.</pre>
          </div>
        </section>
        <section class="panel">
          <div class="panel-head">
            <div>
              <h2 data-i18n="providerCatalog">Provider Catalog</h2>
              <div class="muted" id="provider-page-summary">Page 1</div>
            </div>
            <div class="actions">
              <button class="secondary" id="provider-export" data-i18n="export">Export</button>
              <button class="secondary" id="provider-refresh" data-i18n="refresh">Refresh</button>
            </div>
          </div>
          <div class="panel-body form-grid">
            <div class="filters">
              <input id="provider-id-filter" data-i18n-placeholder="providerFilter" placeholder="Provider ID" />
              <select id="provider-class-filter">
                <option value="" data-i18n="allClasses">All classes</option>
                <option value="local" data-i18n="localClass">Local</option>
                <option value="cloud" data-i18n="cloudClass">Cloud</option>
              </select>
              <select id="provider-enabled-filter">
                <option value="" data-i18n="allEnabled">All enabled</option>
                <option value="true" data-i18n="enabled">Enabled</option>
                <option value="false" data-i18n="disabled">Disabled</option>
              </select>
              <select id="provider-runtime-filter">
                <option value="" data-i18n="allRuntimeStatus">All runtime</option>
                <option value="healthy" data-i18n="runtimeHealthy">Healthy</option>
                <option value="degraded" data-i18n="runtimeDegraded">Degraded</option>
                <option value="unhealthy" data-i18n="runtimeUnhealthy">Unhealthy</option>
                <option value="disabled" data-i18n="runtimeDisabled">Disabled</option>
              </select>
              <select id="provider-quota-filter">
                <option value="" data-i18n="allQuotaStates">All quota states</option>
                <option value="true" data-i18n="quotaEnabled">Quota enabled</option>
                <option value="false" data-i18n="quotaDisabled">Quota disabled</option>
              </select>
              <button class="secondary" id="provider-filter-apply" data-i18n="applyFilter">Apply</button>
              <button class="secondary" id="provider-filter-clear" data-i18n="clearFilter">Clear</button>
            </div>
            <div class="muted" id="provider-message" data-i18n="loadingProviders">Loading providers...</div>
          </div>
          <div class="table-wrap" style="height: 320px;">
            <table class="provider-table">
              <thead>
                <tr>
                  <th data-i18n="provider">Provider</th>
                  <th data-i18n="type">Type</th>
                  <th data-i18n="adapter">Adapter</th>
                  <th data-i18n="models">Models</th>
                  <th data-i18n="quota">Quota</th>
                  <th data-i18n="status">Status</th>
                  <th data-i18n="action">Action</th>
                </tr>
              </thead>
              <tbody id="provider-rows"></tbody>
            </table>
          </div>
          <div class="pager">
            <div class="muted" id="provider-range-summary">0 / 0</div>
            <div class="pager-actions">
              <button class="secondary" id="provider-prev" data-i18n="prev">Prev</button>
              <button class="secondary" id="provider-next" data-i18n="next">Next</button>
            </div>
          </div>
        </section>
        <section class="panel">
          <div class="panel-head">
            <div>
              <h2 data-i18n="modelCatalog">Model Catalog</h2>
              <div class="muted" id="model-page-summary">Page 1</div>
            </div>
            <div class="actions">
              <button class="secondary" id="model-export" data-i18n="export">Export</button>
              <button class="secondary" id="model-refresh" data-i18n="refresh">Refresh</button>
            </div>
          </div>
          <div class="panel-body form-grid">
            <div class="filters">
              <input id="model-name-filter" data-i18n-placeholder="modelFilter" placeholder="Model" />
              <input id="model-provider-filter" data-i18n-placeholder="providerFilter" placeholder="Provider ID" />
              <select id="model-class-filter">
                <option value="" data-i18n="allClasses">All classes</option>
                <option value="local" data-i18n="localClass">Local</option>
                <option value="cloud" data-i18n="cloudClass">Cloud</option>
              </select>
              <select id="model-available-filter">
                <option value="" data-i18n="availableOnly">Available only</option>
                <option value="true" data-i18n="available">Available</option>
                <option value="false" data-i18n="unavailable">Unavailable</option>
                <option value="all" data-i18n="allModels">All models</option>
              </select>
              <button class="secondary" id="model-filter-apply" data-i18n="applyFilter">Apply</button>
              <button class="secondary" id="model-filter-clear" data-i18n="clearFilter">Clear</button>
            </div>
            <div class="muted" id="model-message" data-i18n="loadingModels">Loading models...</div>
          </div>
          <div class="table-wrap" style="height: 320px;">
            <table class="model-table">
              <thead>
                <tr>
                  <th data-i18n="model">Model</th>
                  <th data-i18n="provider">Provider</th>
                  <th data-i18n="type">Type</th>
                  <th data-i18n="status">Status</th>
                  <th data-i18n="available">Available</th>
                </tr>
              </thead>
              <tbody id="model-rows"></tbody>
            </table>
          </div>
          <div class="pager">
            <div class="muted" id="model-range-summary">0 / 0</div>
            <div class="pager-actions">
              <button class="secondary" id="model-prev" data-i18n="prev">Prev</button>
              <button class="secondary" id="model-next" data-i18n="next">Next</button>
            </div>
          </div>
        </section>
        <section class="panel">
          <div class="panel-head">
            <div>
              <h2 data-i18n="policyCatalog">Policy Catalog</h2>
              <div class="muted" id="policy-page-summary">Page 1</div>
            </div>
            <div class="actions">
              <button class="secondary" id="policy-export" data-i18n="export">Export</button>
              <button class="secondary" id="policy-refresh" data-i18n="refresh">Refresh</button>
            </div>
          </div>
          <div class="panel-body form-grid">
            <div class="filters">
              <input id="policy-id-filter" data-i18n-placeholder="policyFilter" placeholder="Policy ID" />
              <select id="policy-effect-filter">
                <option value="" data-i18n="allEffects">All effects</option>
                <option value="allow" data-i18n="effectAllow">Allow</option>
                <option value="deny" data-i18n="effectDeny">Deny</option>
                <option value="force_local" data-i18n="effectForceLocal">Force local</option>
                <option value="deny_cloud_for_sensitive" data-i18n="effectDenyCloudSensitive">Deny cloud sensitive</option>
              </select>
              <input id="policy-app-filter" data-i18n-placeholder="appFilter" placeholder="App ID" />
              <input id="policy-model-filter" data-i18n-placeholder="modelFilter" placeholder="Model" />
              <button class="secondary" id="policy-filter-apply" data-i18n="applyFilter">Apply</button>
            </div>
            <div class="filters">
              <input id="policy-request-type-filter" data-i18n-placeholder="requestTypeFilter" placeholder="Request Type" />
              <input id="policy-provider-class-filter" data-i18n-placeholder="providerClassFilter" placeholder="Provider Class" />
              <input id="policy-data-label-filter" data-i18n-placeholder="dataLabelFilter" placeholder="Data Label" />
              <button class="secondary" id="policy-filter-clear" data-i18n="clearFilter">Clear</button>
            </div>
            <div class="muted" id="policy-message" data-i18n="loadingPolicies">Loading policies...</div>
          </div>
          <div class="table-wrap" style="height: 320px;">
            <table class="policy-table">
              <thead>
                <tr>
                  <th data-i18n="evaluationOrder">Order</th>
                  <th data-i18n="policy">Policy</th>
                  <th data-i18n="priority">Priority</th>
                  <th data-i18n="effect">Effect</th>
                  <th data-i18n="condition">Condition</th>
                  <th data-i18n="reason">Reason</th>
                </tr>
              </thead>
              <tbody id="policy-rows"></tbody>
            </table>
          </div>
          <div class="pager">
            <div class="muted" id="policy-range-summary">0 / 0</div>
            <div class="pager-actions">
              <button class="secondary" id="policy-prev" data-i18n="prev">Prev</button>
              <button class="secondary" id="policy-next" data-i18n="next">Next</button>
            </div>
          </div>
          <div class="panel-body" id="policy-detail"><p class="muted" data-i18n="selectPolicy">Select a policy from the table.</p></div>
        </section>
        <section class="panel">
          <div class="panel-head">
            <div>
              <h2 data-i18n="grantCatalog">Grant Catalog</h2>
              <div class="muted" id="grant-page-summary">Page 1</div>
            </div>
            <div class="actions">
              <button class="secondary" id="grant-export" data-i18n="export">Export</button>
              <button class="secondary" id="grant-refresh" data-i18n="refresh">Refresh</button>
            </div>
          </div>
          <div class="panel-body form-grid">
            <div class="filters">
              <input id="grant-id-filter" data-i18n-placeholder="grantFilter" placeholder="Grant" />
              <select id="grant-type-filter">
                <option value="" data-i18n="allGrantTypes">All types</option>
                <option value="core" data-i18n="grantTypeCore">Core</option>
                <option value="tool_broad" data-i18n="grantTypeToolBroad">Tool broad</option>
                <option value="tool_scope" data-i18n="grantTypeToolScope">Tool Scope</option>
                <option value="admin" data-i18n="grantTypeAdmin">Admin</option>
              </select>
              <input id="grant-app-filter" data-i18n-placeholder="appFilter" placeholder="App ID" />
              <input id="grant-tool-filter" data-i18n-placeholder="toolIdPlaceholder" placeholder="Tool ID" />
              <button class="secondary" id="grant-filter-apply" data-i18n="applyFilter">Apply</button>
              <button class="secondary" id="grant-filter-clear" data-i18n="clearFilter">Clear</button>
            </div>
            <div class="muted" id="grant-message" data-i18n="loadingGrants">Loading grant catalog...</div>
          </div>
          <div class="table-wrap" style="height: 300px;">
            <table class="grant-table">
              <thead>
                <tr>
                  <th data-i18n="grants">Grants</th>
                  <th data-i18n="type">Type</th>
                  <th data-i18n="app">App</th>
                  <th data-i18n="tools">Tool Invocation</th>
                  <th data-i18n="description">Description</th>
                </tr>
              </thead>
              <tbody id="grant-rows"></tbody>
            </table>
          </div>
          <div class="pager">
            <div class="muted" id="grant-range-summary">0 / 0</div>
            <div class="pager-actions">
              <button class="secondary" id="grant-prev" data-i18n="prev">Prev</button>
              <button class="secondary" id="grant-next" data-i18n="next">Next</button>
            </div>
          </div>
        </section>
        <section class="panel">
          <div class="panel-head">
            <div>
              <h2 data-i18n="requestTraces">请求追踪</h2>
              <div class="muted" id="summary">正在加载追踪...</div>
            </div>
            <div class="actions">
              <input id="trace-app-filter" data-i18n-placeholder="appFilter" placeholder="App ID" style="max-width: 150px;" />
              <input id="trace-provider-filter" data-i18n-placeholder="providerFilter" placeholder="Provider ID" style="max-width: 170px;" />
              <select id="status-filter" style="max-width: 160px;">
                <option value="" data-i18n="allStatus">全部状态</option>
                <option value="completed" data-i18n="statusCompleted">已完成</option>
                <option value="failed" data-i18n="statusFailed">失败</option>
              </select>
              <button class="secondary" id="trace-filter-apply" data-i18n="applyFilter">筛选</button>
              <button class="secondary" id="trace-filter-clear" data-i18n="clearFilter">清空</button>
              <button class="secondary" id="trace-export" data-i18n="export">导出</button>
            </div>
          </div>
          <div class="panel-body" style="padding-top: 0; padding-bottom: 8px;">
            <div class="muted" id="trace-export-message"></div>
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
            <div class="filters">
              <select id="audit-action-filter">
                <option value="" data-i18n="allActions">全部动作</option>
                <option value="tool.invoke" data-i18n="tools">工具调用</option>
                <option value="policy.dry_run" data-i18n="actionPolicyDryRun">策略试算</option>
                <option value="routing.explain" data-i18n="actionRoutingExplain">路由解释</option>
                <option value="provider.enabled" data-i18n="actionProviderEnabled">Provider 启停</option>
                <option value="provider.probe" data-i18n="actionProviderProbe">Provider 探测</option>
                <option value="config.reload" data-i18n="actionConfigReload">配置重载</option>
              </select>
              <select id="audit-result-filter">
                <option value="" data-i18n="allResults">全部结果</option>
                <option value="success" data-i18n="resultSuccess">成功</option>
                <option value="denied" data-i18n="resultDenied">拒绝</option>
                <option value="failed" data-i18n="resultFailed">失败</option>
              </select>
              <input id="audit-app-filter" data-i18n-placeholder="appFilter" placeholder="App ID" />
              <input id="audit-trace-filter" data-i18n-placeholder="traceFilter" placeholder="Trace ID" />
              <input id="audit-target-filter" data-i18n-placeholder="targetFilter" placeholder="Target" />
              <input id="audit-metadata-key-filter" data-i18n-placeholder="metadataKeyFilter" placeholder="Metadata key" />
              <input id="audit-metadata-value-filter" data-i18n-placeholder="metadataValueFilter" placeholder="Metadata value" />
              <button class="secondary" id="audit-filter-apply" data-i18n="applyFilter">筛选</button>
              <button class="secondary" id="audit-filter-clear" data-i18n="clearFilter">清空</button>
            </div>
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
        <section class="panel">
          <div class="panel-head"><h2 data-i18n="auditDetail">审计详情</h2></div>
          <div class="panel-body" id="audit-detail"><p class="muted" data-i18n="selectAudit">从审计表格中选择一条事件。</p></div>
        </section>
        <section class="panel">
          <div class="panel-head">
            <div>
              <h2 data-i18n="toolCatalog">工具目录</h2>
              <div class="muted" id="tool-page-summary">第 1 页</div>
            </div>
            <div class="actions">
              <button class="secondary" id="tool-export" data-i18n="export">导出</button>
              <button class="secondary" id="tool-refresh" data-i18n="refresh">刷新</button>
            </div>
          </div>
          <div class="panel-body form-grid">
            <div class="filters">
              <select id="tool-origin-filter">
                <option value="" data-i18n="allOrigins">全部来源</option>
                <option value="builtin" data-i18n="builtinTool">内置</option>
                <option value="mcp" data-i18n="mcpTool">MCP</option>
              </select>
              <input id="tool-server-filter" data-i18n-placeholder="mcpServerFilter" placeholder="Server ID" />
              <input id="tool-scope-filter" data-i18n-placeholder="mcpScopeFilter" placeholder="Scope" />
              <select id="tool-enabled-filter">
                <option value="" data-i18n="allEnabled">全部启用状态</option>
                <option value="true" data-i18n="enabled">启用</option>
                <option value="false" data-i18n="disabled">禁用</option>
              </select>
              <button class="secondary" id="tool-filter-apply" data-i18n="applyFilter">筛选</button>
              <button class="secondary" id="tool-filter-clear" data-i18n="clearFilter">清空</button>
            </div>
          </div>
          <div class="table-wrap" style="height: 320px;">
            <table class="tool-table">
              <thead>
                <tr>
                  <th data-i18n="toolName">工具</th>
                  <th data-i18n="origin">来源</th>
                  <th data-i18n="adapter">适配器</th>
                  <th data-i18n="risk">风险</th>
                  <th data-i18n="scope">Scope</th>
                  <th data-i18n="status">状态</th>
                </tr>
              </thead>
              <tbody id="tool-rows"></tbody>
            </table>
          </div>
          <div class="pager">
            <div class="muted" id="tool-range-summary">0 / 0</div>
            <div class="pager-actions">
              <button class="secondary" id="tool-prev" data-i18n="prev">上一页</button>
              <button class="secondary" id="tool-next" data-i18n="next">下一页</button>
            </div>
          </div>
        </section>
        <section class="panel">
          <div class="panel-head">
            <div>
              <h2 data-i18n="mcpCatalog">MCP 目录</h2>
              <div class="muted" id="mcp-page-summary">第 1 页</div>
            </div>
            <div class="actions">
              <button class="secondary" id="mcp-export" data-i18n="export">导出</button>
              <button class="secondary" id="mcp-refresh" data-i18n="refresh">刷新</button>
            </div>
          </div>
          <div class="panel-body form-grid">
            <div class="filters">
              <input id="mcp-server-filter" data-i18n-placeholder="mcpServerFilter" placeholder="Server ID" />
              <input id="mcp-scope-filter" data-i18n-placeholder="mcpScopeFilter" placeholder="Scope" />
              <select id="mcp-enabled-filter">
                <option value="" data-i18n="allEnabled">全部启用状态</option>
                <option value="true" data-i18n="enabled">启用</option>
                <option value="false" data-i18n="disabled">禁用</option>
              </select>
              <button class="secondary" id="mcp-filter-apply" data-i18n="applyFilter">筛选</button>
              <button class="secondary" id="mcp-filter-clear" data-i18n="clearFilter">清空</button>
            </div>
            <div class="muted" id="mcp-catalog">正在加载 MCP 目录...</div>
          </div>
          <div class="table-wrap" style="height: 320px;">
            <table class="mcp-table">
              <thead>
                <tr>
                  <th data-i18n="server">Server</th>
                  <th data-i18n="status">状态</th>
                  <th data-i18n="toolCount">工具数</th>
                  <th data-i18n="tools">工具</th>
                  <th data-i18n="scope">Scope</th>
                </tr>
              </thead>
              <tbody id="mcp-rows"></tbody>
            </table>
          </div>
          <div class="pager">
            <div class="muted" id="mcp-range-summary">0 / 0</div>
            <div class="pager-actions">
              <button class="secondary" id="mcp-prev" data-i18n="prev">上一页</button>
              <button class="secondary" id="mcp-next" data-i18n="next">下一页</button>
            </div>
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
          <div class="panel-head"><h2 data-i18n="accessDryRun">权限试算</h2></div>
          <div class="panel-body form-grid">
            <input id="access-app-id" value="dev-app" data-i18n-placeholder="appFilter" placeholder="App ID" />
            <input id="access-token" value="dev-token" data-i18n-placeholder="tokenPlaceholder" placeholder="Token" />
            <select id="access-action">
              <option value="chat" data-i18n="chatAction">聊天</option>
              <option value="tool.invoke" data-i18n="toolInvokeAction">工具调用</option>
              <option value="admin" data-i18n="adminAction">管理</option>
            </select>
            <input id="access-tool-id" value="gateway.runtime_health" data-i18n-placeholder="toolIdPlaceholder" placeholder="Tool ID" />
            <button id="access-dry-run" data-i18n="dryRun">试算</button>
            <pre id="access-result" data-i18n="accessResultPlaceholder">权限试算结果会显示在这里。</pre>
          </div>
        </section>
        <section class="panel">
          <div class="panel-head">
            <h2 data-i18n="tools">工具调用</h2>
          </div>
          <div class="panel-body form-grid">
            <select id="tool-select"></select>
            <div class="muted tool-meta" id="tool-meta" data-i18n="loadingTools">正在加载工具...</div>
            <input id="tool-token" value="dev-token" />
            <button id="tool-invoke" data-i18n="invokeTool">执行工具</button>
            <div class="actions" id="tool-result-actions"></div>
            <pre id="tool-result" data-i18n="toolResultPlaceholder">工具执行结果会显示在这里。</pre>
          </div>
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
    const traceAppFilter = document.querySelector("#trace-app-filter");
    const traceProviderFilter = document.querySelector("#trace-provider-filter");
    const traceExportMessage = document.querySelector("#trace-export-message");
    const sendResult = document.querySelector("#send-result");
    const auditRows = document.querySelector("#audit-rows");
    const auditMessage = document.querySelector("#audit-message");
    const auditDetail = document.querySelector("#audit-detail");
    const auditActionFilter = document.querySelector("#audit-action-filter");
    const auditResultFilter = document.querySelector("#audit-result-filter");
    const auditAppFilter = document.querySelector("#audit-app-filter");
    const auditTraceFilter = document.querySelector("#audit-trace-filter");
    const auditTargetFilter = document.querySelector("#audit-target-filter");
    const auditMetadataKeyFilter = document.querySelector("#audit-metadata-key-filter");
    const auditMetadataValueFilter = document.querySelector("#audit-metadata-value-filter");
    const routeExplain = document.querySelector("#route-explain");
    const runtimeHealth = document.querySelector("#runtime-health");
    let accessResult = document.querySelector("#access-result");
    const toolSelect = document.querySelector("#tool-select");
    const toolMeta = document.querySelector("#tool-meta");
    const toolResult = document.querySelector("#tool-result");
    const toolResultActions = document.querySelector("#tool-result-actions");
    const toolRows = document.querySelector("#tool-rows");
    const toolOriginFilter = document.querySelector("#tool-origin-filter");
    const toolServerFilter = document.querySelector("#tool-server-filter");
    const toolScopeFilter = document.querySelector("#tool-scope-filter");
    const toolEnabledFilter = document.querySelector("#tool-enabled-filter");
    const appRows = document.querySelector("#app-rows");
    const appMessage = document.querySelector("#app-message");
    const appIDFilter = document.querySelector("#app-id-filter");
    const appGrantFilter = document.querySelector("#app-grant-filter");
    const appQuotaFilter = document.querySelector("#app-quota-filter");
    const grantRows = document.querySelector("#grant-rows");
    const grantMessage = document.querySelector("#grant-message");
    const grantIDFilter = document.querySelector("#grant-id-filter");
    const grantTypeFilter = document.querySelector("#grant-type-filter");
    const grantAppFilter = document.querySelector("#grant-app-filter");
    const grantToolFilter = document.querySelector("#grant-tool-filter");
    const providerRows = document.querySelector("#provider-rows");
    const providerMessage = document.querySelector("#provider-message");
    const providerIDFilter = document.querySelector("#provider-id-filter");
    const providerClassFilter = document.querySelector("#provider-class-filter");
    const providerEnabledFilter = document.querySelector("#provider-enabled-filter");
    const providerRuntimeFilter = document.querySelector("#provider-runtime-filter");
    const providerQuotaFilter = document.querySelector("#provider-quota-filter");
    const modelRows = document.querySelector("#model-rows");
    const modelMessage = document.querySelector("#model-message");
    const modelNameFilter = document.querySelector("#model-name-filter");
    const modelProviderFilter = document.querySelector("#model-provider-filter");
    const modelClassFilter = document.querySelector("#model-class-filter");
    const modelAvailableFilter = document.querySelector("#model-available-filter");
    const policyRows = document.querySelector("#policy-rows");
    const policyMessage = document.querySelector("#policy-message");
    const policyDetail = document.querySelector("#policy-detail");
    const policyIDFilter = document.querySelector("#policy-id-filter");
    const policyEffectFilter = document.querySelector("#policy-effect-filter");
    const policyAppFilter = document.querySelector("#policy-app-filter");
    const policyModelFilter = document.querySelector("#policy-model-filter");
    const policyRequestTypeFilter = document.querySelector("#policy-request-type-filter");
    const policyProviderClassFilter = document.querySelector("#policy-provider-class-filter");
    const policyDataLabelFilter = document.querySelector("#policy-data-label-filter");
    const policyDryAppID = document.querySelector("#policy-dry-app-id");
    const policyDryRequestType = document.querySelector("#policy-dry-request-type");
    const policyDryModel = document.querySelector("#policy-dry-model");
    const policyDryProviderClass = document.querySelector("#policy-dry-provider-class");
    const policyDryDataLabels = document.querySelector("#policy-dry-data-labels");
    let policyDryResult = document.querySelector("#policy-dry-result");
    const routingDryAppID = document.querySelector("#routing-dry-app-id");
    const routingDryRequestType = document.querySelector("#routing-dry-request-type");
    const routingDryModel = document.querySelector("#routing-dry-model");
    const routingDryDataLabels = document.querySelector("#routing-dry-data-labels");
    const configReloadResult = document.querySelector("#config-reload-result");
    const issueRows = document.querySelector("#issue-rows");
    const issueMessage = document.querySelector("#issue-message");
    const mcpCatalog = document.querySelector("#mcp-catalog");
    const mcpRows = document.querySelector("#mcp-rows");
    const mcpServerFilter = document.querySelector("#mcp-server-filter");
    const mcpScopeFilter = document.querySelector("#mcp-scope-filter");
    const mcpEnabledFilter = document.querySelector("#mcp-enabled-filter");
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
    let issueTools = [];
    let toolTotal = 0;
    let toolPage = 1;
    let toolIDFilter = "";
    const toolPageSize = 8;
    let allApps = [];
    let issueApps = [];
    let appTotal = 0;
    let appPage = 1;
    const appPageSize = 8;
    let allGrants = [];
    let grantTotal = 0;
    let grantPage = 1;
    const grantPageSize = 8;
    let allProviders = [];
    let issueProviders = [];
    let providerTotal = 0;
    let providerPage = 1;
    const providerPageSize = 8;
    let allModels = [];
    let issueModels = [];
    let modelTotal = 0;
    let modelPage = 1;
    const modelPageSize = 8;
    let allPolicies = [];
    let policyTotal = 0;
    let policyPage = 1;
    const policyPageSize = 8;
    let mcpTotal = 0;
    let mcpPage = 1;
    const mcpPageSize = 5;
    let allMCPServers = [];
    let issueMCPServers = [];
    let runtimeData = null;
    let allIssues = [];
    let issuePage = 1;
    const issuePageSize = 8;
    let selectedTrace = null;
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
        traceExportSafety: "Trace \u5bfc\u51fa\u5df2\u6309\u5f53\u524d\u7b5b\u9009\u6761\u4ef6\u751f\u6210\uff0c\u8bf7\u6c42\u5feb\u7167\u590d\u7528\u5df2\u4fdd\u5b58\u7684\u8131\u654f/\u622a\u65ad\u7248\u672c\uff0c\u4e0d\u5305\u542b\u5e94\u7528 Token\u3002",
        auditExportSafety: "Audit \u5bfc\u51fa\u9700\u8981\u7ba1\u7406\u5458\u4ee4\u724c\uff0c\u5df2\u6309\u5f53\u524d\u7b5b\u9009\u6761\u4ef6\u751f\u6210\uff1b\u82e5\u901a\u8fc7 trace_id \u590d\u76d8\u8bf7\u6c42\uff0c\u4ecd\u4f7f\u7528 Trace \u5b89\u5168\u5feb\u7167\u3002",
        explainChain: "\u89e3\u91ca\u94fe",
        stage: "\u9636\u6bb5",
        decision: "\u51b3\u7b56",
        nextAction: "\u4e0b\u4e00\u6b65",
        matchedGrant: "\u547d\u4e2d\u6388\u6743",
        missingGrants: "\u7f3a\u5931\u6388\u6743",
        policyVersion: "\u7b56\u7565\u7248\u672c",
        evaluationOrder: "\u987a\u5e8f",
        rulePriority: "\u89c4\u5219\u4f18\u5148\u7ea7",
        condition: "\u6761\u4ef6",
        ruleEvaluations: "\u89c4\u5219\u8bc4\u4f30\u8bca\u65ad",
        matched: "\u547d\u4e2d",
        mismatchFields: "\u672a\u5339\u914d\u5b57\u6bb5",
        yes: "\u662f",
        no: "\u5426",
        candidateCount: "\u5019\u9009\u6570",
        skippedCount: "\u8df3\u8fc7\u6570",
        appFilter: "App ID",
        providerFilter: "Provider ID",
        traceFilter: "Trace ID",
        targetFilter: "Target",
        metadataKeyFilter: "Metadata key",
        metadataValueFilter: "Metadata value",
        allActions: "全部动作",
        allResults: "全部结果",
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
        auditDetail: "审计详情",
        selectAudit: "从审计表格中选择一条事件。",
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
        accessDryRun: "权限试算",
        tokenPlaceholder: "Token",
        toolIdPlaceholder: "Tool ID",
        grantFilter: "Grant",
        appCatalog: "\u5e94\u7528\u4e0e\u6388\u6743",
        tokenHint: "Token \u63d0\u793a",
        quota: "\u914d\u989d",
        allQuotaStates: "\u5168\u90e8\u914d\u989d\u72b6\u6001",
        quotaEnabled: "\u5df2\u542f\u7528\u914d\u989d",
        quotaDisabled: "\u672a\u542f\u7528\u914d\u989d",
        grants: "\u6388\u6743",
        type: "\u7c7b\u578b",
        description: "\u8bf4\u660e",
        grantCatalog: "Grant \u76ee\u5f55",
        allGrantTypes: "\u5168\u90e8\u7c7b\u578b",
        grantTypeCore: "\u6838\u5fc3",
        grantTypeToolBroad: "\u5de5\u5177\u901a\u914d",
        grantTypeToolScope: "\u5de5\u5177 Scope",
        grantTypeAdmin: "\u7ba1\u7406",
        loadingGrants: "\u6b63\u5728\u52a0\u8f7d Grant \u76ee\u5f55...",
        noGrants: "\u6682\u65e0 Grant\u3002",
        loadingApps: "\u6b63\u5728\u52a0\u8f7d\u5e94\u7528\u6388\u6743...",
        noApps: "\u6682\u65e0\u5e94\u7528\u3002",
        noFilterResults: "\u6ca1\u6709\u5339\u914d\u5f53\u524d\u7b5b\u9009\u6761\u4ef6\u7684\u6570\u636e\u3002",
        issueSummary: "\u8fd0\u884c\u95ee\u9898\u6c47\u603b",
        loadingIssues: "\u6b63\u5728\u6c47\u603b\u8fd0\u884c\u95ee\u9898...",
        noIssues: "\u6682\u65e0\u8fd0\u884c\u95ee\u9898\u3002",
        severity: "\u7ea7\u522b",
        source: "\u6765\u6e90",
        target: "\u76ee\u6807",
        issue: "\u95ee\u9898",
        critical: "\u4e25\u91cd",
        warningLevel: "\u8b66\u544a",
        infoLevel: "\u63d0\u793a",
        issueProviderStatus: "Provider \u8fd0\u884c\u72b6\u6001\u4e3a {status}{reason}",
        issueModelUnavailable: "\u6a21\u578b\u5f53\u524d\u4e0d\u53ef\u7528\uff0cProvider={provider}\uff0c\u72b6\u6001={status}",
        issueToolDisabled: "\u5de5\u5177\u5df2\u7981\u7528\uff0cScope={scopes}",
        issueAppQuotaDisabled: "\u5e94\u7528\u5177\u5907 chat \u80fd\u529b\uff0c\u4f46\u672a\u542f\u7528 App RPM \u914d\u989d",
        issueProviderQuotaDisabled: "Provider \u5df2\u542f\u7528\u4e14\u53ef\u8def\u7531\uff0c\u4f46\u672a\u542f\u7528 Provider RPM \u914d\u989d",
        issueMCPServerDisabled: "MCP Server \u5df2\u7981\u7528\uff0c\u5df2\u542f\u7528\u5de5\u5177 {enabled}/{total}",
        issueMCPRuntime: "MCP \u8fd0\u884c\u65f6\u72b6\u6001\u4e3a {status}{reason}",
        issueTraceFailed: "\u6700\u8fd1\u8bf7\u6c42\u5931\u8d25\uff1a{error}",
        issueAuditProblem: "\u6700\u8fd1\u5ba1\u8ba1\u5f02\u5e38\uff1a{action} / {result}{error}",
        issueRange: "{range} / {total} | \u7b2c {page} / {pages} \u9875",
        replayTrace: "\u56de\u586b\u5230\u5feb\u6377\u8bf7\u6c42",
        viewTraceAudit: "\u67e5\u770b\u5173\u8054\u5ba1\u8ba1",
        copyTraceRequest: "\u590d\u5236\u8bf7\u6c42 JSON",
        copyTraceCurl: "\u590d\u5236 curl \u8349\u7a3f",
        traceSnapshotSafetyTitle: "Trace \u5b89\u5168\u5feb\u7167",
        traceSnapshotSafetyBody: "\u672c\u533a\u57df\u53ea\u4f7f\u7528 Trace \u4e2d\u5df2\u4fdd\u5b58\u7684\u8bf7\u6c42\u5feb\u7167\uff1b\u4e0d\u4fdd\u5b58\u5e94\u7528 Token\uff0c\u547d\u4e2d\u654f\u611f\u6807\u7b7e\u7684\u5185\u5bb9\u4f1a\u6309\u914d\u7f6e\u8131\u654f\u6216\u622a\u65ad\u3002",
        traceCopySafetyHint: "\u590d\u5236\u548c\u56de\u586b\u4ec5\u7528\u4e8e\u672c\u5730\u590d\u76d8\uff0ccurl \u8349\u7a3f\u4ec5\u4f7f\u7528 $GATEWAY_TOKEN \u5360\u4f4d\u7b26\u3002",
        traceRequestCopied: "\u5df2\u590d\u5236 Trace \u8bf7\u6c42\u5feb\u7167 JSON\uff0c\u5185\u5bb9\u5df2\u6309\u5feb\u7167\u7b56\u7565\u5904\u7406\u3002",
        traceCurlCopied: "\u5df2\u590d\u5236 curl \u8349\u7a3f\uff0c\u4ee4\u724c\u4f7f\u7528 $GATEWAY_TOKEN \u5360\u4f4d\u7b26\u3002",
        replayFilled: "\u5df2\u56de\u586b Trace \u8bf7\u6c42\u8349\u7a3f\uff0c\u672a\u53d1\u9001\u3002",
        replayUnavailable: "\u8be5 Trace \u6ca1\u6709\u53ef\u56de\u586b\u7684\u8bf7\u6c42\u5feb\u7167\u3002",
        copied: "\u5df2\u590d\u5236\u5230\u526a\u8d34\u677f\u3002",
        copyFailed: "\u590d\u5236\u5931\u8d25\uff1a{error}",
        providerCatalog: "Provider \u76ee\u5f55",
        modelCatalog: "\u6a21\u578b\u76ee\u5f55",
        models: "\u6a21\u578b",
        model: "\u6a21\u578b",
        modelFilter: "\u6a21\u578b",
        allClasses: "\u5168\u90e8\u7c7b\u578b",
        localClass: "\u672c\u5730",
        cloudClass: "\u4e91\u7aef",
        allRuntimeStatus: "\u5168\u90e8\u8fd0\u884c\u72b6\u6001",
        loadingModels: "\u6b63\u5728\u52a0\u8f7d\u6a21\u578b...",
        noProviders: "\u6682\u65e0 Provider\u3002",
        noModels: "\u6682\u65e0\u6a21\u578b\u3002",
        availableOnly: "\u4ec5\u53ef\u7528",
        allModels: "\u5168\u90e8\u6a21\u578b",
        policyCatalog: "\u7b56\u7565\u76ee\u5f55",
        policy: "\u7b56\u7565",
        policyFilter: "\u7b56\u7565 ID",
        effect: "\u6548\u679c",
        priority: "\u4f18\u5148\u7ea7",
        allEffects: "\u5168\u90e8\u6548\u679c",
        effectAllow: "\u5141\u8bb8",
        effectDeny: "\u62d2\u7edd",
        effectForceLocal: "\u5f3a\u5236\u672c\u5730",
        effectDenyCloudSensitive: "\u654f\u611f\u6570\u636e\u7981\u6b62\u4e91\u7aef",
        requestTypeFilter: "\u8bf7\u6c42\u7c7b\u578b",
        requested: "\u8bf7\u6c42\u6a21\u578b",
        final: "\u6700\u7ec8\u6a21\u578b",
        route: "\u8def\u7531",
        providerClassFilter: "Provider \u7c7b\u578b",
        dataLabelFilter: "\u6570\u636e\u6807\u7b7e",
        loadingPolicies: "\u6b63\u5728\u52a0\u8f7d\u7b56\u7565...",
        noPolicies: "\u6682\u65e0\u7b56\u7565\u3002",
        selectPolicy: "\u4ece\u7b56\u7565\u8868\u683c\u4e2d\u9009\u62e9\u4e00\u6761\u89c4\u5219\u3002",
        loadingPolicyDetail: "\u6b63\u5728\u52a0\u8f7d\u7b56\u7565\u8be6\u60c5...",
        copyPolicyJSON: "\u590d\u5236\u7b56\u7565 JSON",
        policyJSONCopied: "\u5df2\u590d\u5236\u7b56\u7565 JSON\u3002",
        fillPolicyDryRun: "\u56de\u586b\u8bd5\u7b97",
        fillAndRunPolicyDryRun: "\u56de\u586b\u5e76\u8bd5\u7b97",
        policyDryRunFilled: "\u5df2\u56de\u586b\u7b56\u7565\u8bd5\u7b97\u6761\u4ef6\uff0c\u672a\u53d1\u9001\u3002",
        policyDryRunFilledAndRunning: "\u5df2\u56de\u586b\u6761\u4ef6\uff0c\u6b63\u5728\u6267\u884c\u8bd5\u7b97\u5e76\u5199\u5165 Audit...",
        anyScope: "\u5168\u90e8",
        policyDryRun: "\u7b56\u7565\u8bd5\u7b97",
        policyDryRunPlaceholder: "\u7b56\u7565\u8bd5\u7b97\u7ed3\u679c\u4f1a\u663e\u793a\u5728\u8fd9\u91cc\u3002",
        policyDryRunning: "\u7b56\u7565\u8bd5\u7b97\u4e2d...",
        routingDryRun: "\u8def\u7531\u89e3\u91ca\u8bd5\u7b97",
        configReload: "\u914d\u7f6e\u91cd\u8f7d",
        reloadConfig: "\u91cd\u8f7d\u914d\u7f6e",
        configReloadPlaceholder: "\u914d\u7f6e\u91cd\u8f7d\u7ed3\u679c\u4f1a\u663e\u793a\u5728\u8fd9\u91cc\u3002",
        configReloading: "\u914d\u7f6e\u91cd\u8f7d\u4e2d...",
        chatAction: "聊天",
        toolInvokeAction: "工具调用",
        adminAction: "管理",
        dryRun: "试算",
        accessResultPlaceholder: "权限试算结果会显示在这里。",
        dryRunning: "试算中...",
        tools: "\u5de5\u5177\u8c03\u7528",
        toolCatalog: "工具目录",
        toolName: "工具",
        origin: "来源",
        adapter: "适配器",
        risk: "风险",
        scope: "Scope",
        server: "Server",
        allOrigins: "全部来源",
        builtinTool: "内置",
        mcpTool: "MCP",
        mcpCatalog: "MCP 目录",
        loadingMCP: "正在加载 MCP 目录...",
        noMCPServers: "暂无 MCP Server。",
        mcpMode: "模式",
        toolCount: "工具数",
        mcpServerFilter: "Server ID",
        mcpScopeFilter: "Scope",
        allEnabled: "全部启用状态",
        applyFilter: "筛选",
        clearFilter: "清空",
        loadingTools: "\u6b63\u5728\u52a0\u8f7d\u5de5\u5177...",
        noTools: "\u6682\u65e0\u53ef\u7528\u5de5\u5177\u3002",
        invokeTool: "\u6267\u884c\u5de5\u5177",
        invokingTool: "\u5de5\u5177\u6267\u884c\u4e2d...",
        openToolTrace: "\u6253\u5f00\u5de5\u5177\u8ffd\u8e2a\u8be6\u60c5",
        toolResultPlaceholder: "\u5de5\u5177\u6267\u884c\u7ed3\u679c\u4f1a\u663e\u793a\u5728\u8fd9\u91cc\u3002",
        toolTokenRequired: "\u9700\u8981\u5177\u6709 tool \u6388\u6743\u7684\u4ee4\u724c\u3002",
        selectToolFirst: "\u8bf7\u5148\u9009\u62e9\u5de5\u5177\u3002",
        readOnlyTool: "\u53ea\u8bfb",
        writeTool: "\u53ef\u5199",
        traceDetail: "追踪详情",
        selectTrace: "从表格中选择一条追踪。",
        loadingTraces: "正在加载追踪...",
        noTraces: "暂无 Trace。",
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
        actionRoutingExplain: "\u8def\u7531\u89e3\u91ca",
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
        catalogRange: "{range} / {total}",
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
        traceExportSafety: "Trace export uses the current filters and the stored safe request snapshot. Redacted or truncated values stay redacted, and app tokens are not included.",
        auditExportSafety: "Audit export requires an admin token and uses the current filters. Request replay by trace_id still relies on the Trace safe snapshot.",
        explainChain: "Explain Chain",
        stage: "Stage",
        decision: "Decision",
        nextAction: "Next Action",
        matchedGrant: "Matched Grant",
        missingGrants: "Missing Grants",
        policyVersion: "Policy Version",
        evaluationOrder: "Order",
        rulePriority: "Rule Priority",
        condition: "Condition",
        ruleEvaluations: "Rule Evaluation Diagnostics",
        matched: "Matched",
        mismatchFields: "Mismatch Fields",
        yes: "Yes",
        no: "No",
        candidateCount: "Candidates",
        skippedCount: "Skipped",
        appFilter: "App ID",
        providerFilter: "Provider ID",
        traceFilter: "Trace ID",
        targetFilter: "Target",
        metadataKeyFilter: "Metadata key",
        metadataValueFilter: "Metadata value",
        allActions: "All actions",
        allResults: "All results",
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
        auditDetail: "Audit Detail",
        selectAudit: "Select an audit event from the table.",
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
        quotaRuntime: "Quota Runtime",
        appQuotaCount: "App Quotas",
        appRpmEnabled: "Enabled App RPM",
        providerQuotaCount: "Provider Quotas",
        providerRpmEnabled: "Enabled Provider RPM",
        totalProviderRpm: "Total Provider RPM",
        totalAppRpm: "Total App RPM",
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
        accessDryRun: "Access Dry-run",
        tokenPlaceholder: "Token",
        toolIdPlaceholder: "Tool ID",
        grantFilter: "Grant",
        appCatalog: "Apps / Grants",
        tokenHint: "Token Hint",
        quota: "Quota",
        allQuotaStates: "All quota states",
        quotaEnabled: "Quota enabled",
        quotaDisabled: "Quota disabled",
        grants: "Grants",
        type: "Type",
        description: "Description",
        grantCatalog: "Grant Catalog",
        allGrantTypes: "All types",
        grantTypeCore: "Core",
        grantTypeToolBroad: "Tool broad",
        grantTypeToolScope: "Tool Scope",
        grantTypeAdmin: "Admin",
        loadingGrants: "Loading grant catalog...",
        noGrants: "No grants.",
        loadingApps: "Loading apps and grants...",
        noApps: "No apps.",
        noFilterResults: "No data matches the current filters.",
        issueSummary: "Runtime Issue Summary",
        loadingIssues: "Summarizing runtime issues...",
        noIssues: "No runtime issues.",
        severity: "Severity",
        source: "Source",
        target: "Target",
        issue: "Issue",
        critical: "Critical",
        warningLevel: "Warning",
        infoLevel: "Info",
        issueProviderStatus: "Provider runtime status is {status}{reason}",
        issueModelUnavailable: "Model is unavailable, provider={provider}, status={status}",
        issueToolDisabled: "Tool is disabled, scopes={scopes}",
        issueAppQuotaDisabled: "App has chat capability but App RPM quota is not enabled",
        issueProviderQuotaDisabled: "Provider is enabled and routable but Provider RPM quota is not enabled",
        issueMCPServerDisabled: "MCP server is disabled, enabled tools {enabled}/{total}",
        issueMCPRuntime: "MCP runtime status is {status}{reason}",
        issueTraceFailed: "Recent request failed: {error}",
        issueAuditProblem: "Recent audit problem: {action} / {result}{error}",
        issueRange: "{range} of {total} | Page {page} / {pages}",
        replayTrace: "Fill Quick Request",
        viewTraceAudit: "View Related Audit",
        copyTraceRequest: "Copy Request JSON",
        copyTraceCurl: "Copy curl Draft",
        traceSnapshotSafetyTitle: "Trace Safe Snapshot",
        traceSnapshotSafetyBody: "This panel only uses the request snapshot already stored in Trace. App tokens are never stored; data that matches sensitive labels is redacted or truncated by configuration.",
        traceCopySafetyHint: "Copy and fill actions are for local replay only. The curl draft uses the $GATEWAY_TOKEN placeholder.",
        traceRequestCopied: "Trace request snapshot JSON copied; content follows the configured snapshot safety policy.",
        traceCurlCopied: "curl draft copied with the $GATEWAY_TOKEN placeholder.",
        replayFilled: "Trace request draft filled, not sent.",
        replayUnavailable: "This trace has no request snapshot to fill.",
        copied: "Copied to clipboard.",
        copyFailed: "Copy failed: {error}",
        providerCatalog: "Provider Catalog",
        modelCatalog: "Model Catalog",
        models: "Models",
        model: "Model",
        modelFilter: "Model",
        allClasses: "All classes",
        localClass: "Local",
        cloudClass: "Cloud",
        allRuntimeStatus: "All runtime",
        loadingModels: "Loading models...",
        noProviders: "No providers.",
        noModels: "No models.",
        availableOnly: "Available only",
        allModels: "All models",
        policyCatalog: "Policy Catalog",
        policy: "Policy",
        policyFilter: "Policy ID",
        effect: "Effect",
        priority: "Priority",
        allEffects: "All effects",
        effectAllow: "Allow",
        effectDeny: "Deny",
        effectForceLocal: "Force local",
        effectDenyCloudSensitive: "Deny cloud sensitive",
        requestTypeFilter: "Request Type",
        requested: "Requested Model",
        final: "Final Model",
        route: "Route",
        providerClassFilter: "Provider Class",
        dataLabelFilter: "Data Label",
        loadingPolicies: "Loading policies...",
        noPolicies: "No policies.",
        selectPolicy: "Select a policy from the table.",
        loadingPolicyDetail: "Loading policy detail...",
        copyPolicyJSON: "Copy Policy JSON",
        policyJSONCopied: "Policy JSON copied.",
        fillPolicyDryRun: "Fill Dry-run",
        fillAndRunPolicyDryRun: "Fill And Run",
        policyDryRunFilled: "Policy dry-run fields filled, not sent.",
        policyDryRunFilledAndRunning: "Policy dry-run fields filled; running dry-run and writing Audit...",
        anyScope: "Any",
        policyDryRun: "Policy Dry-run",
        policyDryRunPlaceholder: "Policy dry-run result will appear here.",
        policyDryRunning: "Running policy dry-run...",
        routingDryRun: "Routing Explain",
        configReload: "Config Reload",
        reloadConfig: "Reload Config",
        configReloadPlaceholder: "Config reload result will appear here.",
        configReloading: "Reloading config...",
        chatAction: "Chat",
        toolInvokeAction: "Tool Invoke",
        adminAction: "Admin",
        dryRun: "Dry-run",
        accessResultPlaceholder: "Access dry-run result will appear here.",
        dryRunning: "Dry-running...",
        tools: "Tool Invocation",
        toolCatalog: "Tool Catalog",
        toolName: "Tool",
        origin: "Origin",
        adapter: "Adapter",
        risk: "Risk",
        scope: "Scope",
        server: "Server",
        allOrigins: "All origins",
        builtinTool: "Built-in",
        mcpTool: "MCP",
        mcpCatalog: "MCP Catalog",
        loadingMCP: "Loading MCP catalog...",
        noMCPServers: "No MCP servers.",
        mcpMode: "Mode",
        toolCount: "Tools",
        mcpServerFilter: "Server ID",
        mcpScopeFilter: "Scope",
        allEnabled: "All enabled states",
        applyFilter: "Apply",
        clearFilter: "Clear",
        loadingTools: "Loading tools...",
        noTools: "No tools available.",
        invokeTool: "Invoke Tool",
        invokingTool: "Invoking tool...",
        openToolTrace: "Open Tool Trace Detail",
        toolResultPlaceholder: "Tool result will appear here.",
        toolTokenRequired: "A token with the tool grant is required.",
        selectToolFirst: "Select a tool first.",
        readOnlyTool: "read-only",
        writeTool: "write",
        traceDetail: "Trace Detail",
        selectTrace: "Select a trace from the table.",
        loadingTraces: "Loading traces...",
        noTraces: "No traces.",
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
        actionRoutingExplain: "routing explain",
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
        catalogRange: "{range} of {total}",
        loadConsoleFailed: "Failed to load console: "
      }
    };

    document.querySelector("#refresh").addEventListener("click", loadAll);
    document.querySelector("#issue-refresh").addEventListener("click", loadAll);
    document.querySelector("#audit-refresh").addEventListener("click", loadAudit);
    document.querySelector("#trace-filter-apply").addEventListener("click", () => { tracePage = 1; loadTraces(); });
    document.querySelector("#trace-filter-clear").addEventListener("click", clearTraceFilters);
    document.querySelector("#audit-filter-apply").addEventListener("click", () => { auditPage = 1; loadAudit(); });
    document.querySelector("#audit-filter-clear").addEventListener("click", clearAuditFilters);
    document.querySelector("#trace-export").addEventListener("click", exportTraces);
    document.querySelector("#audit-export").addEventListener("click", exportAudit);
    document.querySelector("#lang-toggle").addEventListener("click", () => setLang(lang === "zh" ? "en" : "zh"));
    document.querySelector("#sample").addEventListener("click", () => sendQuick("success"));
    document.querySelector("#send").addEventListener("click", () => sendQuick(document.querySelector("#mode").value));
    document.querySelector("#explain").addEventListener("click", () => explainRouting(document.querySelector("#mode").value));
    document.querySelector("#access-dry-run").addEventListener("click", accessDryRun);
    document.querySelector("#tool-invoke").addEventListener("click", invokeTool);
    document.querySelector("#tool-export").addEventListener("click", exportTools);
    document.querySelector("#tool-refresh").addEventListener("click", loadTools);
    document.querySelector("#tool-filter-apply").addEventListener("click", () => { toolIDFilter = ""; toolPage = 1; loadTools(); });
    document.querySelector("#tool-filter-clear").addEventListener("click", clearToolFilters);
    document.querySelector("#app-export").addEventListener("click", exportApps);
    document.querySelector("#app-refresh").addEventListener("click", loadApps);
    document.querySelector("#app-filter-apply").addEventListener("click", () => { appPage = 1; loadApps(); });
    document.querySelector("#app-filter-clear").addEventListener("click", clearAppFilters);
    document.querySelector("#grant-export").addEventListener("click", exportGrants);
    document.querySelector("#grant-refresh").addEventListener("click", loadGrants);
    document.querySelector("#grant-filter-apply").addEventListener("click", () => { grantPage = 1; loadGrants(); });
    document.querySelector("#grant-filter-clear").addEventListener("click", clearGrantFilters);
    document.querySelector("#provider-export").addEventListener("click", exportProviderCatalog);
    document.querySelector("#provider-refresh").addEventListener("click", loadProviderCatalog);
    document.querySelector("#provider-filter-apply").addEventListener("click", () => { providerPage = 1; loadProviderCatalog(); });
    document.querySelector("#provider-filter-clear").addEventListener("click", clearProviderFilters);
    document.querySelector("#model-export").addEventListener("click", exportModelCatalog);
    document.querySelector("#model-refresh").addEventListener("click", loadModelCatalog);
    document.querySelector("#model-filter-apply").addEventListener("click", () => { modelPage = 1; loadModelCatalog(); });
    document.querySelector("#model-filter-clear").addEventListener("click", clearModelFilters);
    document.querySelector("#policy-export").addEventListener("click", exportPolicyCatalog);
    document.querySelector("#policy-refresh").addEventListener("click", loadPolicyCatalog);
    document.querySelector("#policy-filter-apply").addEventListener("click", () => { policyPage = 1; loadPolicyCatalog(); });
    document.querySelector("#policy-filter-clear").addEventListener("click", clearPolicyFilters);
    document.querySelector("#policy-dry-run").addEventListener("click", policyDryRun);
    document.addEventListener("click", event => {
      const button = event.target.closest("button[data-chain-policy-id]");
      if (!button) return;
      loadPolicyDetail(button.dataset.chainPolicyId);
    });
    document.addEventListener("click", event => {
      const button = event.target.closest("button[data-trace-policy-id]");
      if (!button) return;
      loadPolicyDetail(button.dataset.tracePolicyId);
    });
    document.addEventListener("click", event => {
      const button = event.target.closest("button[data-trace-link-id]");
      if (!button) return;
      filterTraceByID(button.dataset.traceLinkId);
    });
    document.addEventListener("click", event => {
      const button = event.target.closest("button[data-issue-trace-id]");
      if (!button) return;
      loadDetail(button.dataset.issueTraceId);
    });
    document.addEventListener("click", event => {
      const button = event.target.closest("button[data-issue-audit-trace-id], button[data-issue-audit-target]");
      if (!button) return;
      auditTraceFilter.value = button.dataset.issueAuditTraceId || "";
      auditTargetFilter.value = button.dataset.issueAuditTarget || "";
      auditPage = 1;
      loadAudit();
    });
    document.addEventListener("click", event => {
      const button = event.target.closest("button[data-issue-provider-id]");
      if (!button) return;
      filterProviderByID(button.dataset.issueProviderId, button.dataset.issueProviderQuotaEnabled || "");
    });
    document.addEventListener("click", event => {
      const button = event.target.closest("button[data-issue-app-id]");
      if (!button) return;
      filterAppByID(button.dataset.issueAppId, button.dataset.issueAppQuotaEnabled || "");
    });
    document.addEventListener("click", event => {
      const button = event.target.closest("button[data-provider-link-id]");
      if (!button) return;
      filterProviderByID(button.dataset.providerLinkId);
    });
    document.addEventListener("click", event => {
      const button = event.target.closest("button[data-provider-class-link]");
      if (!button) return;
      filterProviderByClass(button.dataset.providerClassLink);
    });
    document.addEventListener("click", event => {
      const button = event.target.closest("button[data-policy-filter-key]");
      if (!button) return;
      filterPolicyByField(button.dataset.policyFilterKey, button.dataset.policyFilterValue || "");
    });
    document.addEventListener("click", event => {
      const button = event.target.closest("button[data-issue-model]");
      if (!button) return;
      filterModelByID(button.dataset.issueModel, button.dataset.issueModelProvider || "");
    });
    document.addEventListener("click", event => {
      const button = event.target.closest("button[data-model-link-id]");
      if (!button) return;
      filterModelByID(button.dataset.modelLinkId, button.dataset.modelLinkProvider || "");
    });
    document.addEventListener("click", event => {
      const button = event.target.closest("button[data-app-link-id]");
      if (!button) return;
      filterAppByID(button.dataset.appLinkId);
    });
    document.addEventListener("click", event => {
      const button = event.target.closest("button[data-grant-link-id]");
      if (!button) return;
      filterGrantByID(button.dataset.grantLinkId);
    });
    document.addEventListener("click", event => {
      const button = event.target.closest("button[data-tool-link-id]");
      if (!button) return;
      filterToolByID(button.dataset.toolLinkId);
    });
    document.addEventListener("click", event => {
      const button = event.target.closest("button[data-mcp-server-link-id]");
      if (!button) return;
      filterMCPServerByID(button.dataset.mcpServerLinkId);
    });
    document.addEventListener("click", event => {
      const button = event.target.closest("button[data-audit-metadata-key]");
      if (!button) return;
      filterAuditByMetadata(button.dataset.auditMetadataKey, button.dataset.auditMetadataValue || "");
    });
    document.addEventListener("click", event => {
      const button = event.target.closest("button[data-issue-tool-id]");
      if (!button) return;
      toolIDFilter = button.dataset.issueToolId;
      toolEnabledFilter.value = "false";
      toolPage = 1;
      loadTools();
    });
    document.addEventListener("click", event => {
      const button = event.target.closest("button[data-issue-mcp-server-id], button[data-issue-mcp-runtime]");
      if (!button) return;
      mcpServerFilter.value = button.dataset.issueMcpServerId || "";
      mcpPage = 1;
      loadMCPCatalog();
    });
    document.querySelector("#routing-dry-run").addEventListener("click", routingDryRun);
    document.querySelector("#config-reload").addEventListener("click", configReload);
    toolSelect.addEventListener("change", renderSelectedTool);
    document.querySelector("#mcp-filter-apply").addEventListener("click", () => { mcpPage = 1; loadMCPCatalog(); });
    document.querySelector("#mcp-filter-clear").addEventListener("click", clearMCPFilters);
    document.querySelector("#mcp-export").addEventListener("click", exportMCPCatalog);
    document.querySelector("#mcp-refresh").addEventListener("click", loadMCPCatalog);
    document.querySelector("#trace-prev").addEventListener("click", () => { tracePage = Math.max(1, tracePage - 1); loadTraces(); });
    document.querySelector("#trace-next").addEventListener("click", () => { tracePage += 1; loadTraces(); });
    document.querySelector("#audit-prev").addEventListener("click", () => { auditPage = Math.max(1, auditPage - 1); loadAudit(); });
    document.querySelector("#audit-next").addEventListener("click", () => { auditPage += 1; loadAudit(); });
    document.querySelector("#tool-prev").addEventListener("click", () => { toolPage = Math.max(1, toolPage - 1); loadTools(); });
    document.querySelector("#tool-next").addEventListener("click", () => { toolPage += 1; loadTools(); });
    document.querySelector("#app-prev").addEventListener("click", () => { appPage = Math.max(1, appPage - 1); loadApps(); });
    document.querySelector("#app-next").addEventListener("click", () => { appPage += 1; loadApps(); });
    document.querySelector("#grant-prev").addEventListener("click", () => { grantPage = Math.max(1, grantPage - 1); loadGrants(); });
    document.querySelector("#grant-next").addEventListener("click", () => { grantPage += 1; loadGrants(); });
    document.querySelector("#provider-prev").addEventListener("click", () => { providerPage = Math.max(1, providerPage - 1); loadProviderCatalog(); });
    document.querySelector("#provider-next").addEventListener("click", () => { providerPage += 1; loadProviderCatalog(); });
    document.querySelector("#model-prev").addEventListener("click", () => { modelPage = Math.max(1, modelPage - 1); loadModelCatalog(); });
    document.querySelector("#model-next").addEventListener("click", () => { modelPage += 1; loadModelCatalog(); });
    document.querySelector("#policy-prev").addEventListener("click", () => { policyPage = Math.max(1, policyPage - 1); loadPolicyCatalog(); });
    document.querySelector("#policy-next").addEventListener("click", () => { policyPage += 1; loadPolicyCatalog(); });
    document.querySelector("#issue-prev").addEventListener("click", () => { issuePage = Math.max(1, issuePage - 1); renderIssues(); });
    document.querySelector("#issue-next").addEventListener("click", () => { issuePage += 1; renderIssues(); });
    document.querySelector("#mcp-prev").addEventListener("click", () => { mcpPage = Math.max(1, mcpPage - 1); loadMCPCatalog(); });
    document.querySelector("#mcp-next").addEventListener("click", () => { mcpPage += 1; loadMCPCatalog(); });
    statusFilter.addEventListener("change", () => { tracePage = 1; loadTraces(); });

    const zhFallback = {
      quotaRuntime: "\u914d\u989d\u8fd0\u884c\u65f6",
      quota: "\u914d\u989d",
      allQuotaStates: "\u5168\u90e8\u914d\u989d\u72b6\u6001",
      quotaEnabled: "\u5df2\u542f\u7528\u914d\u989d",
      quotaDisabled: "\u672a\u542f\u7528\u914d\u989d",
      appQuotaCount: "\u5e94\u7528\u914d\u989d",
      appRpmEnabled: "\u542f\u7528 App RPM",
      totalAppRpm: "\u603b App RPM",
      providerQuotaCount: "Provider \u914d\u989d",
      providerRpmEnabled: "\u542f\u7528 Provider RPM",
      totalProviderRpm: "\u603b Provider RPM"
    };

    function t(key, vars = {}) {
      let value = (i18n[lang] && i18n[lang][key]) || i18n.zh[key] || zhFallback[key] || key;
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
      document.querySelectorAll("[data-i18n-placeholder]").forEach(node => {
        node.setAttribute("placeholder", t(node.dataset.i18nPlaceholder));
      });
      if (!allTraces.length) summary.textContent = t("loadingTraces");
      renderTraces();
      renderAudit();
      renderTools();
      renderProviderCatalog();
      renderModelCatalog();
      renderPolicyCatalog();
      renderIssues();
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
        "routing.explain": t("actionRoutingExplain"),
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
    function hasQueryParams(query) {
      for (const key of query.keys()) {
        if (key !== "limit" && key !== "offset") return true;
      }
      return false;
    }
    function emptyText(defaultKey, query) {
      return hasQueryParams(query) ? t("noFilterResults") : t(defaultKey);
    }
    async function loadAll() {
      await Promise.all([loadTraces(), loadRuntimeHealth(), loadProviders(), loadModels(), loadProviderCatalog(), loadModelCatalog(), loadPolicyCatalog(), loadAudit(), loadApps(), loadGrants(), loadTools(), loadMCPCatalog()]);
    }
    function traceQuery(limit = tracePageSize, offset = (tracePage - 1) * tracePageSize) {
      const query = new URLSearchParams({ limit: String(limit), offset: String(offset) });
      if (statusFilter.value) query.set("status", statusFilter.value);
      if (traceAppFilter.value.trim()) query.set("app_id", traceAppFilter.value.trim());
      if (traceProviderFilter.value.trim()) query.set("provider_id", traceProviderFilter.value.trim());
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
      syncIssues();
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
      traceExportMessage.textContent = t("traceExportSafety");
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
          "<td class=\"trace-id\">" + renderTraceLink(item.trace_id) + "</td>" +
          "<td>" + renderAppLink(item.app_id) + "</td>" +
          "<td>" + renderModelLink(item.requested_model, item.provider_id) + "</td>" +
          "<td>" + renderProviderLink(item.provider_id) + "</td>" +
          "<td>" + renderModelLink(item.final_model, item.provider_id) + "</td>" +
          "<td>" + renderPolicyLink(item.policy && item.policy.rule_id) + "</td>" +
          "<td>" + ((item.fallbacks || []).length) + "</td>" +
          "<td>" + esc(time(item.started_at)) + "</td>" +
          "<td title=\"" + esc(item.error) + "\">" + esc(item.error) + "</td>" +
        "</tr>"
      ).join("");
      rows.querySelectorAll("tr").forEach(row => row.addEventListener("click", event => {
        if (event.target.closest("button")) return;
        loadDetail(row.dataset.trace);
      }));
      if (!allTraces.length) {
        rows.innerHTML = "<tr><td colspan=\"10\" class=\"muted\">" + emptyText("noTraces", traceQuery()) + "</td></tr>";
      }
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
    function syncIssues(resetPage = false) {
      allIssues = buildIssues();
      if (resetPage) issuePage = 1;
      renderIssues();
    }
    function buildIssues() {
      const issues = [];
      const now = new Date().toISOString();
      const addIssue = (severity, source, target, message, createdAt = "", ref = {}) => {
        issues.push({ severity, source, target, message, created_at: createdAt || now, ref });
      };
      if (runtimeData && runtimeData.mcp_runtime) {
        const mcp = runtimeData.mcp_runtime;
        if (["degraded", "unavailable", "not_configured"].includes(mcp.status)) {
          addIssue(mcp.status === "unavailable" ? "critical" : "warning", "mcp_runtime", mcp.mode || "-", t("issueMCPRuntime", {
            status: labelRuntime(mcp.status),
            reason: mcp.reason ? " / " + mcp.reason : ""
          }), "", { mcp_runtime: true });
        }
      }
      issueProviders.forEach(item => {
        const status = item.runtime_status || (item.configured_healthy ? "healthy" : "unhealthy");
        if (["unhealthy", "degraded", "disabled"].includes(status) || item.enabled === false) {
          addIssue(status === "unhealthy" ? "critical" : "warning", "provider", item.id, t("issueProviderStatus", {
            status: labelRuntime(status || "disabled"),
            reason: item.degraded_reason ? " / " + item.degraded_reason : ""
          }), item.updated_at, { provider_id: item.id });
        }
        if (item.enabled !== false && ["healthy", "degraded"].includes(status) && !(item.quota && item.quota.enabled)) {
          addIssue("info", "provider", item.id, t("issueProviderQuotaDisabled"), item.updated_at, { provider_id: item.id, quota_enabled: "false" });
        }
      });
      issueModels.forEach(item => {
        if (!item.available) {
          addIssue("warning", "model", item.model, t("issueModelUnavailable", {
            provider: item.provider_id || "-",
            status: labelRuntime(item.runtime_status || "unavailable")
          }), "", { model: item.model, provider_id: item.provider_id });
        }
      });
      issueApps.forEach(item => {
        if ((item.grants || []).includes("chat") && !(item.quota && item.quota.enabled)) {
          addIssue("info", "app", item.id, t("issueAppQuotaDisabled"), "", { app_id: item.id, quota_enabled: "false" });
        }
      });
      issueTools.forEach(item => {
        if (!item.enabled) {
          addIssue("info", "tool", item.id, t("issueToolDisabled", {
            scopes: (item.scopes || []).join(", ") || "-"
          }), "", { tool_id: item.id });
        }
      });
      issueMCPServers.forEach(server => {
        if (!server.enabled) {
          addIssue("info", "mcp_server", server.id, t("issueMCPServerDisabled", {
            enabled: server.enabled_tools || 0,
            total: server.tool_count || 0
          }), "", { mcp_server_id: server.id });
        }
      });
      allTraces.filter(item => item.status === "failed").forEach(item => {
        addIssue("critical", "trace", item.trace_id, t("issueTraceFailed", { error: item.error || "-" }), item.started_at, { trace_id: item.trace_id });
      });
      allAuditEvents.filter(item => item.result === "denied" || item.result === "failed").forEach(item => {
        addIssue(item.result === "failed" ? "critical" : "warning", "audit", item.target || item.trace_id || "-", t("issueAuditProblem", {
          action: labelAction(item.action),
          result: labelResult(item.result),
          error: item.error ? " / " + item.error : ""
        }), item.created_at, { trace_id: item.trace_id, target: item.target });
      });
      const severityRank = { critical: 0, warning: 1, info: 2 };
      return issues.sort((a, b) => {
        const severityDiff = (severityRank[a.severity] ?? 9) - (severityRank[b.severity] ?? 9);
        if (severityDiff !== 0) return severityDiff;
        return String(b.created_at || "").localeCompare(String(a.created_at || ""));
      });
    }
    function labelSeverity(value) {
      return ({ critical: t("critical"), warning: t("warningLevel"), info: t("infoLevel") })[value] || value || "";
    }
    function renderIssues() {
      const total = allIssues.length;
      const totalPages = pageCount(total, issuePageSize);
      issuePage = clampPage(issuePage, totalPages);
      const range = pageRange(total, issuePage, issuePageSize);
      document.querySelector("#issue-page-summary").textContent = t("page", { page: issuePage, total: totalPages });
      document.querySelector("#issue-range-summary").textContent = t("issueRange", { range: range.label, total, page: issuePage, pages: totalPages });
      document.querySelector("#issue-prev").disabled = issuePage <= 1;
      document.querySelector("#issue-next").disabled = issuePage >= totalPages;
      issueMessage.textContent = total ? "" : t("noIssues");
      const pageItems = allIssues.slice((issuePage - 1) * issuePageSize, issuePage * issuePageSize);
      issueRows.innerHTML = pageItems.map(item =>
        "<tr>" +
          "<td><span class=\"status " + esc(item.severity === "critical" ? "failed" : item.severity) + "\">" + esc(labelSeverity(item.severity)) + "</span></td>" +
          "<td>" + esc(item.source) + "</td>" +
          "<td title=\"" + esc(item.target || "-") + "\">" + renderIssueTarget(item) + "</td>" +
          "<td title=\"" + esc(item.message) + "\">" + esc(item.message) + "</td>" +
          "<td>" + esc(time(item.created_at)) + "</td>" +
        "</tr>"
      ).join("");
      if (!pageItems.length) {
        issueRows.innerHTML = "<tr><td colspan=\"5\" class=\"muted\">" + t("noIssues") + "</td></tr>";
      }
    }
    function renderIssueTarget(item) {
      if (!item || !item.target) return "-";
      if (item.source === "trace" && item.ref && item.ref.trace_id) {
        return "<button class=\"link-button\" data-issue-trace-id=\"" + esc(item.ref.trace_id) + "\">" + esc(item.target) + "</button>";
      }
      if (item.source === "audit" && item.ref && (item.ref.trace_id || item.ref.target)) {
        return "<button class=\"link-button\" data-issue-audit-trace-id=\"" + esc(item.ref.trace_id || "") + "\" data-issue-audit-target=\"" + esc(item.ref.target || "") + "\">" + esc(item.target) + "</button>";
      }
      if (item.source === "provider" && item.ref && item.ref.provider_id) {
        return "<button class=\"link-button\" data-issue-provider-id=\"" + esc(item.ref.provider_id) + "\" data-issue-provider-quota-enabled=\"" + esc(item.ref.quota_enabled || "") + "\">" + esc(item.target) + "</button>";
      }
      if (item.source === "app" && item.ref && item.ref.app_id) {
        return "<button class=\"link-button\" data-issue-app-id=\"" + esc(item.ref.app_id) + "\" data-issue-app-quota-enabled=\"" + esc(item.ref.quota_enabled || "") + "\">" + esc(item.target) + "</button>";
      }
      if (item.source === "model" && item.ref && item.ref.model) {
        return "<button class=\"link-button\" data-issue-model=\"" + esc(item.ref.model) + "\" data-issue-model-provider=\"" + esc(item.ref.provider_id || "") + "\">" + esc(item.target) + "</button>";
      }
      if (item.source === "tool" && item.ref && item.ref.tool_id) {
        return "<button class=\"link-button\" data-issue-tool-id=\"" + esc(item.ref.tool_id) + "\">" + esc(item.target) + "</button>";
      }
      if (item.source === "mcp_server" && item.ref && item.ref.mcp_server_id) {
        return "<button class=\"link-button\" data-issue-mcp-server-id=\"" + esc(item.ref.mcp_server_id) + "\">" + esc(item.target) + "</button>";
      }
      if (item.source === "mcp_runtime" && item.ref && item.ref.mcp_runtime) {
        return "<button class=\"link-button\" data-issue-mcp-runtime=\"true\">" + esc(item.target) + "</button>";
      }
      return esc(item.target);
    }
    async function loadRuntimeHealth() {
      runtimeHealth.textContent = t("loadingRuntime");
      try {
        const res = await fetch("/gateway/v1/runtime/health");
        const data = await res.json();
        if (!res.ok) {
          runtimeData = null;
          runtimeHealth.textContent = t("failedPrefix") + (data.error && data.error.message || res.status);
          syncIssues();
          return;
        }
        runtimeData = data;
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
            "<div class=\"k\">" + t("quotaRuntime") + "</div><div>" + renderQuotaRuntime(data.quota_runtime || {}) + "</div>" +
            "<div class=\"k\">" + t("modelRuntime") + "</div><div>" + esc(labelRuntime(data.model_runtime && data.model_runtime.status)) + "</div>" +
            "<div class=\"k\">" + t("mcpRuntime") + "</div><div>" + renderComponentHealth(data.mcp_runtime || {}) + "</div>" +
          "</div>";
      } catch (err) {
        runtimeData = null;
        runtimeHealth.textContent = t("failedPrefix") + err.message;
      }
      syncIssues();
    }
    function renderQuotaRuntime(quota) {
      const modeText = quota.mode ? " / " + esc(quota.mode) : "";
      const appQuotaText = " / " + esc(t("appQuotaCount")) + ": " + esc(quota.app_quota_count || 0);
      const appRPMText = " / " + esc(t("appRpmEnabled")) + ": " + esc(quota.enabled_app_rpm || 0);
      const totalRPMText = quota.total_app_rpm !== undefined ? " / " + esc(t("totalAppRpm")) + ": " + esc(quota.total_app_rpm || 0) : "";
      const providerQuotaText = quota.provider_quota_count !== undefined ? " / " + esc(t("providerQuotaCount")) + ": " + esc(quota.provider_quota_count || 0) : "";
      const providerRPMText = quota.enabled_provider_rpm !== undefined ? " / " + esc(t("providerRpmEnabled")) + ": " + esc(quota.enabled_provider_rpm || 0) : "";
      const totalProviderRPMText = quota.total_provider_rpm !== undefined ? " / " + esc(t("totalProviderRpm")) + ": " + esc(quota.total_provider_rpm || 0) : "";
      const reasonText = quota.reason ? " / " + esc(quota.reason) : "";
      return esc(labelRuntime(quota.status)) + modeText + appQuotaText + appRPMText + totalRPMText + providerQuotaText + providerRPMText + totalProviderRPMText + reasonText;
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
          "<div class=\"muted\">" + renderProviderLink(item.id) + " / " + renderProviderClassLink(item.class) + " / " + esc(item.adapter || "mock") + "</div>" +
          "<div class=\"muted\">" + (item.enabled === false ? t("disabled") : t("enabled")) + " / " + t("runtime") + ": " + esc(labelRuntime(item.runtime_status) || (item.healthy ? t("healthy") : t("unhealthy"))) + "</div>" +
          (item.degraded_reason ? "<div class=\"muted\">" + esc(item.degraded_reason) + "</div>" : "") +
          "<div>" + renderModelLinks(item.models || [], item.id) + "</div>" +
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
        const query = mcpCatalogQuery();
        query.set("limit", String(mcpPageSize));
        query.set("offset", String((mcpPage - 1) * mcpPageSize));
        const suffix = query.toString() ? "?" + query.toString() : "";
        const issueQuery = mcpCatalogQuery();
        issueQuery.set("limit", "500");
        issueQuery.set("offset", "0");
        const issueSuffix = issueQuery.toString() ? "?" + issueQuery.toString() : "";
        const [res, issueRes] = await Promise.all([
          fetch("/gateway/v1/mcp/servers" + suffix),
          fetch("/gateway/v1/mcp/servers" + issueSuffix)
        ]);
        const data = await res.json();
        const issueData = await issueRes.json();
        if (!res.ok) {
          allMCPServers = [];
          issueMCPServers = [];
          mcpTotal = 0;
          mcpCatalog.textContent = t("failedPrefix") + (data.error && data.error.message || res.status);
          renderMCPServers([]);
          syncIssues();
          return;
        }
        issueMCPServers = issueRes.ok ? (issueData.servers || []) : [];
        const servers = data.servers || [];
        allMCPServers = servers;
        mcpTotal = data.total || servers.length;
        if (!servers.length) {
          mcpCatalog.textContent = t("mcpMode") + ": " + (data.mode || "-") + " / " + emptyText("noMCPServers", mcpCatalogQuery());
          renderMCPServers([]);
          syncIssues();
          return;
        }
        mcpCatalog.textContent = t("mcpMode") + ": " + (data.mode || "-") + " / " + (data.enabled ? t("enabled") : t("disabled"));
        renderMCPServers(servers);
        syncIssues();
      } catch (err) {
        allMCPServers = [];
        issueMCPServers = [];
        mcpTotal = 0;
        mcpCatalog.textContent = t("failedPrefix") + err.message;
        renderMCPServers([]);
        syncIssues();
      }
    }
    function mcpCatalogQuery() {
      const query = new URLSearchParams();
      if (mcpServerFilter.value.trim()) query.set("server_id", mcpServerFilter.value.trim());
      if (mcpScopeFilter.value.trim()) query.set("scope", mcpScopeFilter.value.trim());
      if (mcpEnabledFilter.value) query.set("enabled", mcpEnabledFilter.value);
      return query;
    }
    function renderMCPServers(servers) {
      const totalPages = pageCount(mcpTotal, mcpPageSize);
      mcpPage = clampPage(mcpPage, totalPages);
      const range = pageRange(mcpTotal, mcpPage, mcpPageSize);
      document.querySelector("#mcp-page-summary").textContent = t("page", { page: mcpPage, total: totalPages });
      document.querySelector("#mcp-range-summary").textContent = t("catalogRange", { range: range.label, total: mcpTotal });
      document.querySelector("#mcp-prev").disabled = mcpPage <= 1;
      document.querySelector("#mcp-next").disabled = mcpPage >= totalPages;
      mcpRows.innerHTML = servers.map(server => {
        const tools = server.tools || [];
        const names = tools.map(tool => tool.name || tool.id).join(", ") || "-";
        const scopes = [...new Set(tools.flatMap(tool => tool.scopes || []))];
        const scopeText = scopes.join(", ") || "-";
        return "<tr>" +
          "<td><strong>" + esc(server.name || server.id) + "</strong><div class=\"muted\">" + renderMCPServerLink(server.id) + "</div></td>" +
          "<td><span class=\"status " + (server.enabled ? "healthy" : "disabled") + "\">" + (server.enabled ? t("enabled") : t("disabled")) + "</span></td>" +
          "<td>" + esc(server.enabled_tools || 0) + " / " + esc(server.tool_count || 0) + "</td>" +
          "<td title=\"" + esc(names) + "\">" + esc(names) + "</td>" +
          "<td title=\"" + esc(scopeText) + "\">" + renderScopeGrantLinks(scopes) + "</td>" +
        "</tr>";
      }).join("");
      if (!servers.length) {
        mcpRows.innerHTML = "<tr><td colspan=\"5\" class=\"muted\">" + emptyText("noMCPServers", mcpCatalogQuery()) + "</td></tr>";
      }
    }
    function exportMCPCatalog() {
      const query = mcpCatalogQuery();
      const suffix = query.toString() ? "?" + query.toString() : "";
      downloadURL("/gateway/v1/mcp/servers/export" + suffix, "mcp-servers.jsonl");
    }
    async function loadTools() {
      toolMeta.textContent = t("loadingTools");
      try {
        const query = toolCatalogQuery();
        query.set("limit", String(toolPageSize));
        query.set("offset", String((toolPage - 1) * toolPageSize));
        const issueQuery = toolCatalogQuery();
        issueQuery.set("limit", "500");
        issueQuery.set("offset", "0");
        const [res, issueRes] = await Promise.all([
          fetch("/gateway/v1/tools?" + query.toString()),
          fetch("/gateway/v1/tools?" + issueQuery.toString())
        ]);
        const data = await res.json();
        const issueData = await issueRes.json();
        if (!res.ok) {
          allTools = [];
          issueTools = [];
          toolTotal = 0;
          toolSelect.innerHTML = "";
          toolMeta.textContent = t("failedPrefix") + (data.error && data.error.message || res.status);
          renderTools();
          syncIssues();
          return;
        }
        allTools = data.tools || [];
        issueTools = issueRes.ok ? (issueData.tools || []) : [];
        toolTotal = data.total || allTools.length;
        renderTools();
      } catch (err) {
        allTools = [];
        issueTools = [];
        toolTotal = 0;
        toolSelect.innerHTML = "";
        toolMeta.textContent = t("failedPrefix") + err.message;
        renderTools();
        syncIssues();
      }
    }
    function toolCatalogQuery() {
      const query = new URLSearchParams();
      if (toolIDFilter) query.set("tool_id", toolIDFilter);
      if (toolOriginFilter.value) query.set("origin", toolOriginFilter.value);
      if (toolServerFilter.value.trim()) query.set("server_id", toolServerFilter.value.trim());
      if (toolScopeFilter.value.trim()) query.set("scope", toolScopeFilter.value.trim());
      if (toolEnabledFilter.value) query.set("enabled", toolEnabledFilter.value);
      return query;
    }
    function renderTools() {
      const totalPages = pageCount(toolTotal, toolPageSize);
      toolPage = clampPage(toolPage, totalPages);
      const range = pageRange(toolTotal, toolPage, toolPageSize);
      document.querySelector("#tool-page-summary").textContent = t("page", { page: toolPage, total: totalPages });
      document.querySelector("#tool-range-summary").textContent = t("catalogRange", { range: range.label, total: toolTotal });
      document.querySelector("#tool-prev").disabled = toolPage <= 1;
      document.querySelector("#tool-next").disabled = toolPage >= totalPages;
      toolRows.innerHTML = allTools.map(item =>
        "<tr data-tool=\"" + esc(item.id) + "\">" +
          "<td><strong>" + esc(item.name || item.id) + "</strong><div class=\"muted\">" + renderToolLink(item.id) + "</div></td>" +
          "<td>" + esc(item.origin || "builtin") + (item.server_id ? "<div class=\"muted\">" + renderMCPServerLink(item.server_id) + "</div>" : "") + "</td>" +
          "<td>" + esc(item.adapter) + "</td>" +
          "<td>" + esc(item.risk_level || "-") + "<div class=\"muted\">" + (item.read_only ? t("readOnlyTool") : t("writeTool")) + "</div></td>" +
          "<td title=\"" + esc((item.scopes || []).join(", ") || "-") + "\">" + renderScopeGrantLinks(item.scopes || []) + "</td>" +
          "<td><span class=\"status " + (item.enabled ? "healthy" : "disabled") + "\">" + (item.enabled ? t("enabled") : t("disabled")) + "</span></td>" +
        "</tr>"
      ).join("");
      toolRows.querySelectorAll("tr[data-tool]").forEach(row => row.addEventListener("click", () => {
        toolSelect.value = row.dataset.tool;
        renderSelectedTool();
      }));
      if (!allTools.length) {
        const text = emptyText("noTools", toolCatalogQuery());
        toolRows.innerHTML = "<tr><td colspan=\"6\" class=\"muted\">" + text + "</td></tr>";
        toolSelect.innerHTML = "";
        toolMeta.textContent = text;
        document.querySelector("#tool-invoke").disabled = true;
        syncIssues();
        return;
      }
      const selected = toolSelect.value || allTools[0].id;
      toolSelect.innerHTML = allTools.map(item =>
        "<option value=\"" + esc(item.id) + "\"" + (item.id === selected ? " selected" : "") + ">" + esc(item.name || item.id) + "</option>"
      ).join("");
      document.querySelector("#tool-invoke").disabled = false;
      renderSelectedTool();
      syncIssues();
    }
    function renderSelectedTool() {
      const tool = allTools.find(item => item.id === toolSelect.value);
      if (!tool) {
        toolMeta.textContent = t("noTools");
        return;
      }
      toolMeta.textContent = tool.id + " / " + (tool.origin || "builtin") + (tool.server_id ? ":" + tool.server_id : "") + " / " + tool.adapter + " / " + (tool.read_only ? t("readOnlyTool") : t("writeTool")) + " / " + (tool.risk_level || "-") + " / " + ((tool.scopes || []).join(", ") || "-");
    }
    function exportTools() {
      const query = toolCatalogQuery();
      const suffix = query.toString() ? "?" + query.toString() : "";
      downloadURL("/gateway/v1/tools/export" + suffix, "tools.jsonl");
    }
    function exportApps() {
      const query = appCatalogQuery();
      exportAdminJSONL("/gateway/v1/apps/export", query, "apps.jsonl", appMessage);
    }
    async function loadApps() {
      const token = adminToken();
      if (!token) {
        allApps = [];
        issueApps = [];
        appTotal = 0;
        appMessage.textContent = t("adminTokenRequired");
        renderApps();
        syncIssues();
        return;
      }
      appMessage.textContent = t("loadingApps");
      try {
        const query = appCatalogQuery();
        query.set("limit", String(appPageSize));
        query.set("offset", String((appPage - 1) * appPageSize));
        const issueQuery = new URLSearchParams({ limit: "500", offset: "0" });
        const [res, issueRes] = await Promise.all([
          fetch("/gateway/v1/apps?" + query.toString(), {
            headers: { "Authorization": "Bearer " + token }
          }),
          fetch("/gateway/v1/apps?" + issueQuery.toString(), {
            headers: { "Authorization": "Bearer " + token }
          })
        ]);
        const data = await res.json();
        const issueData = await issueRes.json();
        if (!res.ok) {
          allApps = [];
          issueApps = [];
          appTotal = 0;
          appMessage.textContent = t("failedPrefix") + (data.error && data.error.message || res.status);
          renderApps();
          syncIssues();
          return;
        }
        allApps = data.apps || [];
        issueApps = issueRes.ok ? (issueData.apps || []) : [];
        appTotal = data.total || allApps.length;
        appMessage.textContent = appTotal ? "" : emptyText("noApps", appCatalogQuery());
        renderApps();
        syncIssues();
      } catch (err) {
        allApps = [];
        issueApps = [];
        appTotal = 0;
        appMessage.textContent = t("failedPrefix") + err.message;
        renderApps();
        syncIssues();
      }
    }
    function appCatalogQuery() {
      const query = new URLSearchParams();
      if (appIDFilter.value.trim()) query.set("app_id", appIDFilter.value.trim());
      if (appGrantFilter.value.trim()) query.set("grant", appGrantFilter.value.trim());
      if (appQuotaFilter.value) query.set("quota_enabled", appQuotaFilter.value);
      return query;
    }
    function renderApps() {
      const totalPages = pageCount(appTotal, appPageSize);
      appPage = clampPage(appPage, totalPages);
      const range = pageRange(appTotal, appPage, appPageSize);
      document.querySelector("#app-page-summary").textContent = t("page", { page: appPage, total: totalPages });
      document.querySelector("#app-range-summary").textContent = t("catalogRange", { range: range.label, total: appTotal });
      document.querySelector("#app-prev").disabled = appPage <= 1;
      document.querySelector("#app-next").disabled = appPage >= totalPages;
      appRows.innerHTML = allApps.map(item =>
        "<tr>" +
          "<td><strong>" + esc(item.name || item.id) + "</strong><div class=\"muted\">" + renderAppLink(item.id) + "</div></td>" +
          "<td class=\"trace-id\">" + esc(item.token_hint || "-") + "</td>" +
          "<td>" + renderAppQuota(item.quota || {}) + "</td>" +
          "<td title=\"" + esc((item.grants || []).join(", ")) + "\">" + renderGrantLinks(item.grants || []) + "</td>" +
        "</tr>"
      ).join("");
      if (!allApps.length) {
        appRows.innerHTML = "<tr><td colspan=\"4\" class=\"muted\">" + emptyText("noApps", appCatalogQuery()) + "</td></tr>";
      }
    }
    function renderAppQuota(quota) {
      if (!quota || !quota.enabled) {
        return "<span class=\"status disabled\">" + esc(t("disabled")) + "</span>";
      }
      return "<span class=\"status healthy\">" + esc(t("enabled")) + "</span><div class=\"muted\">RPM " + esc(quota.requests_per_minute || 0) + "</div>";
    }
    function renderProviderQuota(quota) {
      if (!quota || !quota.enabled) {
        return "<span class=\"status disabled\">" + esc(t("disabled")) + "</span>";
      }
      return "<span class=\"status healthy\">" + esc(t("enabled")) + "</span><div class=\"muted\">RPM " + esc(quota.requests_per_minute || 0) + "</div>";
    }
    function exportGrants() {
      const query = grantCatalogQuery();
      exportAdminJSONL("/gateway/v1/grants/export", query, "grants.jsonl", grantMessage);
    }
    async function loadProviderCatalog() {
      providerMessage.textContent = t("loadingProviders");
      try {
        const query = providerCatalogQuery();
        query.set("limit", String(providerPageSize));
        query.set("offset", String((providerPage - 1) * providerPageSize));
        const issueQuery = new URLSearchParams({ limit: "500", offset: "0" });
        const [res, issueRes] = await Promise.all([
          fetch("/gateway/v1/providers?" + query.toString()),
          fetch("/gateway/v1/providers?" + issueQuery.toString())
        ]);
        const data = await res.json();
        const issueData = await issueRes.json();
        if (!res.ok) {
          allProviders = [];
          issueProviders = [];
          providerTotal = 0;
          providerMessage.textContent = t("failedPrefix") + (data.error && data.error.message || res.status);
          renderProviderCatalog();
          syncIssues();
          return;
        }
        allProviders = data.providers || [];
        issueProviders = issueRes.ok ? (issueData.providers || []) : [];
        providerTotal = data.total || allProviders.length;
        providerMessage.textContent = providerTotal ? "" : emptyText("noProviders", providerCatalogQuery());
        renderProviderCatalog();
        syncIssues();
      } catch (err) {
        allProviders = [];
        issueProviders = [];
        providerTotal = 0;
        providerMessage.textContent = t("failedPrefix") + err.message;
        renderProviderCatalog();
        syncIssues();
      }
    }
    function providerCatalogQuery() {
      const query = new URLSearchParams();
      if (providerIDFilter.value.trim()) query.set("provider_id", providerIDFilter.value.trim());
      if (providerClassFilter.value) query.set("class", providerClassFilter.value);
      if (providerEnabledFilter.value) query.set("enabled", providerEnabledFilter.value);
      if (providerRuntimeFilter.value) query.set("runtime_status", providerRuntimeFilter.value);
      if (providerQuotaFilter.value) query.set("quota_enabled", providerQuotaFilter.value);
      return query;
    }
    function renderProviderCatalog() {
      const totalPages = pageCount(providerTotal, providerPageSize);
      providerPage = clampPage(providerPage, totalPages);
      const range = pageRange(providerTotal, providerPage, providerPageSize);
      document.querySelector("#provider-page-summary").textContent = t("page", { page: providerPage, total: totalPages });
      document.querySelector("#provider-range-summary").textContent = t("catalogRange", { range: range.label, total: providerTotal });
      document.querySelector("#provider-prev").disabled = providerPage <= 1;
      document.querySelector("#provider-next").disabled = providerPage >= totalPages;
      providerRows.innerHTML = allProviders.map(item => {
        const runtimeStatus = item.runtime_status || (item.configured_healthy ? "healthy" : "unhealthy");
        return "<tr>" +
          "<td><strong>" + esc(item.name || item.id) + "</strong><div class=\"muted\">" + renderProviderLink(item.id) + "</div></td>" +
          "<td>" + renderProviderClassLink(item.class) + "</td>" +
          "<td>" + esc(item.adapter || "mock") + "</td>" +
          "<td title=\"" + esc((item.models || []).join(", ")) + "\">" + renderModelLinks(item.models || [], item.id) + "</td>" +
          "<td>" + renderProviderQuota(item.quota || {}) + "</td>" +
          "<td><span class=\"status " + esc(runtimeStatus) + "\">" + esc(labelRuntime(runtimeStatus)) + "</span><div class=\"muted\">" + (item.enabled === false ? t("disabled") : t("enabled")) + "</div></td>" +
          "<td><div class=\"provider-actions\">" +
            "<button class=\"secondary\" data-provider=\"" + esc(item.id) + "\" data-action=\"probe\">" + t("probe") + "</button>" +
            "<button class=\"secondary\" data-provider=\"" + esc(item.id) + "\" data-action=\"toggle\" data-enabled=\"" + (item.enabled === false ? "true" : "false") + "\">" + (item.enabled === false ? t("enable") : t("disable")) + "</button>" +
          "</div></td>" +
        "</tr>";
      }).join("");
      providerRows.querySelectorAll("[data-action='probe']").forEach(button => button.addEventListener("click", () => probeProvider(button.dataset.provider)));
      providerRows.querySelectorAll("[data-action='toggle']").forEach(button => button.addEventListener("click", () => setProviderEnabled(button.dataset.provider, button.dataset.enabled === "true")));
      if (!allProviders.length) {
        providerRows.innerHTML = "<tr><td colspan=\"7\" class=\"muted\">" + emptyText("noProviders", providerCatalogQuery()) + "</td></tr>";
      }
    }
    function exportProviderCatalog() {
      const query = providerCatalogQuery();
      const suffix = query.toString() ? "?" + query.toString() : "";
      downloadURL("/gateway/v1/providers/export" + suffix, "providers.jsonl");
    }
    async function loadModelCatalog() {
      modelMessage.textContent = t("loadingModels");
      try {
        const query = modelCatalogQuery();
        query.set("limit", String(modelPageSize));
        query.set("offset", String((modelPage - 1) * modelPageSize));
        const issueQuery = new URLSearchParams({ limit: "500", offset: "0", all: "true" });
        const [res, issueRes] = await Promise.all([
          fetch("/gateway/v1/models?" + query.toString()),
          fetch("/gateway/v1/models?" + issueQuery.toString())
        ]);
        const data = await res.json();
        const issueData = await issueRes.json();
        if (!res.ok) {
          allModels = [];
          issueModels = [];
          modelTotal = 0;
          modelMessage.textContent = t("failedPrefix") + (data.error && data.error.message || res.status);
          renderModelCatalog();
          syncIssues();
          return;
        }
        allModels = data.models || [];
        issueModels = issueRes.ok ? (issueData.models || []) : [];
        modelTotal = data.total || allModels.length;
        modelMessage.textContent = modelTotal ? "" : emptyText("noModels", modelCatalogQuery());
        renderModelCatalog();
        syncIssues();
      } catch (err) {
        allModels = [];
        issueModels = [];
        modelTotal = 0;
        modelMessage.textContent = t("failedPrefix") + err.message;
        renderModelCatalog();
        syncIssues();
      }
    }
    function modelCatalogQuery() {
      const query = new URLSearchParams();
      if (modelNameFilter.value.trim()) query.set("model", modelNameFilter.value.trim());
      if (modelProviderFilter.value.trim()) query.set("provider_id", modelProviderFilter.value.trim());
      if (modelClassFilter.value) query.set("provider_class", modelClassFilter.value);
      if (modelAvailableFilter.value === "all") query.set("all", "true");
      if (modelAvailableFilter.value === "true" || modelAvailableFilter.value === "false") {
        query.set("all", "true");
        query.set("available", modelAvailableFilter.value);
      }
      return query;
    }
    function renderModelCatalog() {
      const totalPages = pageCount(modelTotal, modelPageSize);
      modelPage = clampPage(modelPage, totalPages);
      const range = pageRange(modelTotal, modelPage, modelPageSize);
      document.querySelector("#model-page-summary").textContent = t("page", { page: modelPage, total: totalPages });
      document.querySelector("#model-range-summary").textContent = t("catalogRange", { range: range.label, total: modelTotal });
      document.querySelector("#model-prev").disabled = modelPage <= 1;
      document.querySelector("#model-next").disabled = modelPage >= totalPages;
      modelRows.innerHTML = allModels.map(item =>
        "<tr>" +
          "<td class=\"trace-id\">" + renderModelLink(item.model, item.provider_id) + "</td>" +
          "<td>" + renderProviderLink(item.provider_id) + "</td>" +
          "<td>" + esc(item.provider_class || "-") + "</td>" +
          "<td><span class=\"status " + esc(item.runtime_status || (item.available ? "healthy" : "unhealthy")) + "\">" + esc(labelRuntime(item.runtime_status) || "-") + "</span><div class=\"muted\">" + (item.enabled ? t("enabled") : t("disabled")) + "</div></td>" +
          "<td><span class=\"status " + (item.available ? "healthy" : "unhealthy") + "\">" + (item.available ? t("available") : t("unavailable")) + "</span></td>" +
        "</tr>"
      ).join("");
      if (!allModels.length) {
        modelRows.innerHTML = "<tr><td colspan=\"5\" class=\"muted\">" + emptyText("noModels", modelCatalogQuery()) + "</td></tr>";
      }
    }
    function exportModelCatalog() {
      const query = modelCatalogQuery();
      const suffix = query.toString() ? "?" + query.toString() : "";
      downloadURL("/gateway/v1/models/export" + suffix, "models.jsonl");
    }
    async function loadPolicyCatalog() {
      const token = adminToken();
      if (!token) {
        allPolicies = [];
        policyTotal = 0;
        policyMessage.textContent = t("adminTokenRequired");
        renderPolicyCatalog();
        syncIssues();
        return;
      }
      policyMessage.textContent = t("loadingPolicies");
      try {
        const query = policyCatalogQuery();
        query.set("limit", String(policyPageSize));
        query.set("offset", String((policyPage - 1) * policyPageSize));
        const res = await fetch("/gateway/v1/policies?" + query.toString(), {
          headers: { "Authorization": "Bearer " + token }
        });
        const data = await res.json();
        if (!res.ok) {
          allPolicies = [];
          policyTotal = 0;
          policyMessage.textContent = t("failedPrefix") + (data.error && data.error.message || res.status);
          renderPolicyCatalog();
          syncIssues();
          return;
        }
        allPolicies = data.policies || [];
        policyTotal = data.total || allPolicies.length;
        policyMessage.textContent = policyTotal ? "" : emptyText("noPolicies", policyCatalogQuery());
        renderPolicyCatalog();
        syncIssues();
      } catch (err) {
        allPolicies = [];
        policyTotal = 0;
        policyMessage.textContent = t("failedPrefix") + err.message;
        renderPolicyCatalog();
        syncIssues();
      }
    }
    function policyCatalogQuery() {
      const query = new URLSearchParams();
      if (policyIDFilter.value.trim()) query.set("policy_id", policyIDFilter.value.trim());
      if (policyEffectFilter.value) query.set("effect", policyEffectFilter.value);
      if (policyAppFilter.value.trim()) query.set("app_id", policyAppFilter.value.trim());
      if (policyModelFilter.value.trim()) query.set("model", policyModelFilter.value.trim());
      if (policyRequestTypeFilter.value.trim()) query.set("request_type", policyRequestTypeFilter.value.trim());
      if (policyProviderClassFilter.value.trim()) query.set("provider_class", policyProviderClassFilter.value.trim());
      if (policyDataLabelFilter.value.trim()) query.set("data_label", policyDataLabelFilter.value.trim());
      return query;
    }
    function renderPolicyCatalog() {
      const totalPages = pageCount(policyTotal, policyPageSize);
      policyPage = clampPage(policyPage, totalPages);
      const range = pageRange(policyTotal, policyPage, policyPageSize);
      document.querySelector("#policy-page-summary").textContent = t("page", { page: policyPage, total: totalPages });
      document.querySelector("#policy-range-summary").textContent = t("catalogRange", { range: range.label, total: policyTotal });
      document.querySelector("#policy-prev").disabled = policyPage <= 1;
      document.querySelector("#policy-next").disabled = policyPage >= totalPages;
      policyRows.innerHTML = allPolicies.map(item => {
        const scope = [
          "app: " + labelList(item.app_ids),
          "request: " + labelList(item.request_types),
          "model: " + labelList(item.models),
          "provider: " + labelList(item.provider_classes),
          "label: " + labelList(item.data_labels)
        ].join(" / ");
        return "<tr>" +
          "<td>" + esc(item.evaluation_order || "") + "</td>" +
          "<td><button class=\"link-button\" data-policy-id=\"" + esc(item.id) + "\">" + esc(item.id) + "</button></td>" +
          "<td>" + esc(item.priority == null ? 0 : item.priority) + "</td>" +
          "<td title=\"" + esc(labelEffectSemantics(item.effect_semantics)) + "\"><span class=\"status " + esc(item.effect || "healthy") + "\">" + esc(labelPolicyEffect(item.effect)) + "</span></td>" +
          "<td title=\"" + esc(scope) + "\">" + esc(item.condition_summary || scope) + "</td>" +
          "<td title=\"" + esc(item.reason || "") + "\">" + esc(item.reason || "") + "</td>" +
        "</tr>";
      }).join("");
      if (!allPolicies.length) {
        policyRows.innerHTML = "<tr><td colspan=\"6\" class=\"muted\">" + emptyText("noPolicies", policyCatalogQuery()) + "</td></tr>";
      }
      policyRows.querySelectorAll("button[data-policy-id]").forEach(button => {
        button.addEventListener("click", () => loadPolicyDetail(button.dataset.policyId));
      });
    }
    function labelList(values) {
      return (values && values.length) ? values.join(", ") : t("anyScope");
    }
    function renderPolicyValues(key, values) {
      if (!values || !values.length) return t("anyScope");
      return values.map(value => renderPolicyValue(key, value)).join(", ");
    }
    function renderPolicyValue(key, value) {
      const text = String(value);
      if (key === "app_id") return renderAppLink(text);
      if (key === "model") return renderModelLink(text);
      if (key === "provider_class") return renderProviderClassLink(text);
      return renderPolicyFilterLink(key, text);
    }
    function renderPolicyFilterLink(key, value) {
      return "<button class=\"link-button\" data-policy-filter-key=\"" + esc(key) + "\" data-policy-filter-value=\"" + esc(value) + "\">" + esc(value) + "</button>";
    }
    function labelPolicyEffect(value) {
      return ({
        allow: t("effectAllow"),
        deny: t("effectDeny"),
        force_local: t("effectForceLocal"),
        deny_cloud_for_sensitive: t("effectDenyCloudSensitive")
      })[value] || value || "";
    }
    function labelEffectSemantics(value) {
      if (!value) return "";
      return [
        t("decision") + ": " + (value.allowed ? t("allowed") : t("blocked")),
        t("cloud") + ": " + (value.allow_cloud ? t("allowed") : t("blocked")),
        t("nextAction") + ": " + (value.force_local ? t("effectForceLocal") : t("continue")),
        value.description || ""
      ].filter(Boolean).join(" / ");
    }
    async function loadPolicyDetail(policyID) {
      const token = adminToken();
      if (!token) {
        policyDetail.innerHTML = "<p class=\"muted\">" + t("adminTokenRequired") + "</p>";
        return;
      }
      policyDetail.innerHTML = "<p class=\"muted\">" + t("loadingPolicyDetail") + "</p>";
      try {
        const res = await fetch("/gateway/v1/policies/" + encodeURIComponent(policyID), {
          headers: { "Authorization": "Bearer " + token }
        });
        const data = await res.json();
        if (!res.ok) {
          policyDetail.innerHTML = "<p class=\"muted\">" + esc(t("failedPrefix") + (data.error && data.error.message || res.status)) + "</p>";
          return;
        }
        renderPolicyDetail(data.policy);
      } catch (err) {
        policyDetail.innerHTML = "<p class=\"muted\">" + esc(t("failedPrefix") + err.message) + "</p>";
      }
    }
    function renderPolicyDetail(item) {
      if (!item) {
        policyDetail.innerHTML = "<p class=\"muted\">" + t("selectPolicy") + "</p>";
        return;
      }
      const cells = [
        [t("evaluationOrder"), esc(item.evaluation_order == null ? "" : item.evaluation_order)],
        [t("policy"), renderPolicyLink(item.id)],
        [t("priority"), esc(item.priority == null ? 0 : item.priority)],
        [t("effect"), esc(labelPolicyEffect(item.effect))],
        [t("condition"), esc(item.condition_summary || "")],
        [t("reason"), esc(item.reason || "")],
        [t("app"), renderPolicyValues("app_id", item.app_ids)],
        [t("requestTypeFilter"), renderPolicyValues("request_type", item.request_types)],
        [t("model"), renderPolicyValues("model", item.models)],
        [t("provider"), renderPolicyValues("provider_class", item.provider_classes)],
        [t("dataLabelFilter"), renderPolicyValues("data_label", item.data_labels)],
        [t("cloud"), esc(item.effect_semantics && item.effect_semantics.allow_cloud ? t("allowed") : t("blocked"))],
        [t("nextAction"), esc(item.effect_semantics && item.effect_semantics.force_local ? t("effectForceLocal") : t("continue"))]
      ].filter(item => item[1] !== undefined && item[1] !== null && item[1] !== "");
      policyDetail.innerHTML = "<div class=\"explain-chain\">" +
        "<div class=\"chain-title\">" + esc(item.id) + "</div>" +
        "<div class=\"actions\" style=\"margin-bottom: 12px;\">" +
          "<button class=\"secondary\" id=\"policy-copy-json\" data-i18n=\"copyPolicyJSON\">" + t("copyPolicyJSON") + "</button>" +
          "<button class=\"secondary\" id=\"policy-fill-dry-run\" data-i18n=\"fillPolicyDryRun\">" + t("fillPolicyDryRun") + "</button>" +
          "<button class=\"secondary\" id=\"policy-fill-run\" data-i18n=\"fillAndRunPolicyDryRun\">" + t("fillAndRunPolicyDryRun") + "</button>" +
          "<span class=\"muted\" id=\"policy-copy-status\"></span>" +
        "</div>" +
        "<div class=\"chain-grid\">" +
          cells.map(cell => "<div class=\"chain-cell\"><span class=\"k\">" + esc(cell[0]) + "</span><div>" + String(cell[1]) + "</div></div>").join("") +
        "</div>" +
        "<pre>" + esc(JSON.stringify(item, null, 2)) + "</pre>" +
      "</div>";
      document.querySelector("#policy-copy-json").addEventListener("click", () => copyPolicyJSON(item));
      document.querySelector("#policy-fill-dry-run").addEventListener("click", () => fillPolicyDryRunFromPolicy(item));
      document.querySelector("#policy-fill-run").addEventListener("click", () => fillAndRunPolicyDryRunFromPolicy(item));
    }
    async function copyPolicyJSON(item) {
      await copyText(JSON.stringify(item, null, 2), t("policyJSONCopied"), document.querySelector("#policy-copy-status"));
    }
    function fillPolicyDryRunFromPolicy(item) {
      policyDryAppID.value = firstPolicyValue(item.app_ids, policyDryAppID.value);
      policyDryRequestType.value = firstPolicyValue(item.request_types, "chat");
      policyDryModel.value = firstPolicyValue(item.models, policyDryModel.value);
      policyDryProviderClass.value = firstPolicyValue(item.provider_classes, policyDryProviderClass.value);
      policyDryDataLabels.value = (item.data_labels || []).join(",");
      document.querySelector("#policy-copy-status").textContent = t("policyDryRunFilled");
    }
    async function fillAndRunPolicyDryRunFromPolicy(item) {
      fillPolicyDryRunFromPolicy(item);
      document.querySelector("#policy-copy-status").textContent = t("policyDryRunFilledAndRunning");
      await policyDryRun();
    }
    function firstPolicyValue(values, fallback) {
      return values && values.length ? values[0] : fallback;
    }
    function renderExplainChain(chain) {
      if (!chain) return "";
      const cells = [
        ["stage", t("stage"), chain.stage],
        ["decision", t("decision"), chain.decision],
        ["reason", t("reason"), chain.reason],
        ["policy_rule_id", t("policyRule"), chain.policy_rule_id],
        ["rule_priority", t("rulePriority"), chain.rule_priority],
        ["policy_version", t("policyVersion"), chain.policy_version],
        ["condition", t("condition"), chain.condition],
        ["allow_cloud", t("cloud"), chain.allow_cloud == null ? "" : (chain.allow_cloud ? t("allowed") : t("blocked"))],
        ["matched_grant", t("matchedGrant"), chain.matched_grant],
        ["missing_grants", t("missingGrants"), labelChainList(chain.missing_grants)],
        ["tool_id", t("toolName"), chain.tool_id],
        ["candidate_count", t("candidateCount"), chain.candidate_count],
        ["skipped_count", t("skippedCount"), chain.skipped_count],
        ["next_action", t("nextAction"), chain.next_action]
      ].filter(item => item[2] !== undefined && item[2] !== null && item[2] !== "");
      if (!cells.length) return "";
      return "<div class=\"explain-chain\">" +
        "<div class=\"chain-title\">" + t("explainChain") + "</div>" +
        "<div class=\"chain-grid\">" +
          cells.map(item => "<div class=\"chain-cell\"><span class=\"k\">" + esc(item[1]) + "</span><div>" + renderChainValue(item[0], item[2]) + "</div></div>").join("") +
        "</div>" +
      "</div>";
    }
    function renderChainValue(key, value) {
      if (key === "policy_rule_id" && value) {
        return renderPolicyLink(String(value));
      }
      if (key === "matched_grant" && value) {
        return renderGrantLink(String(value));
      }
      if (key === "missing_grants" && value) {
        return renderGrantLinks(Array.isArray(value) ? value : String(value).split(",").map(item => item.trim()).filter(Boolean));
      }
      if (key === "tool_id" && value) {
        return renderToolLink(String(value));
      }
      return esc(String(value));
    }
    function renderPolicyLink(policyID) {
      if (!policyID) return "-";
      return "<button class=\"link-button\" data-chain-policy-id=\"" + esc(policyID) + "\">" + esc(policyID) + "</button>";
    }
    function renderProviderLink(providerID) {
      if (!providerID) return "-";
      return "<button class=\"link-button\" data-provider-link-id=\"" + esc(providerID) + "\">" + esc(providerID) + "</button>";
    }
    function renderModelLink(model, providerID = "") {
      if (!model) return "-";
      return "<button class=\"link-button\" data-model-link-id=\"" + esc(model) + "\" data-model-link-provider=\"" + esc(providerID || "") + "\">" + esc(model) + "</button>";
    }
    function renderModelLinks(models, providerID = "") {
      if (!models || !models.length) return "-";
      return models.map(model => renderModelLink(model, providerID)).join(", ");
    }
    function renderProviderClassLink(providerClass) {
      if (!providerClass) return "-";
      return "<button class=\"link-button\" data-provider-class-link=\"" + esc(providerClass) + "\">" + esc(providerClass) + "</button>";
    }
    function renderAppLink(appID) {
      if (!appID) return "-";
      return "<button class=\"link-button\" data-app-link-id=\"" + esc(appID) + "\">" + esc(appID) + "</button>";
    }
    function renderAppLinks(appIDs) {
      if (!appIDs || !appIDs.length) return "-";
      return appIDs.map(item => renderAppLink(item)).join(", ");
    }
    function renderGrantLink(grantID) {
      if (!grantID) return "-";
      return "<button class=\"link-button\" data-grant-link-id=\"" + esc(grantID) + "\">" + esc(grantID) + "</button>";
    }
    function renderGrantLinks(grants) {
      if (!grants || !grants.length) return "-";
      return grants.map(item => renderGrantLink(item)).join(", ");
    }
    function renderScopeGrantLinks(scopes) {
      if (!scopes || !scopes.length) return "-";
      return scopes.map(scope => renderGrantLink("tool:" + scope)).join(", ");
    }
    function renderToolLink(toolID) {
      if (!toolID) return "-";
      return "<button class=\"link-button\" data-tool-link-id=\"" + esc(toolID) + "\">" + esc(toolID) + "</button>";
    }
    function renderToolLinks(toolIDs) {
      if (!toolIDs || !toolIDs.length) return "-";
      return toolIDs.map(item => renderToolLink(item)).join(", ");
    }
    function renderMCPServerLink(serverID) {
      if (!serverID) return "-";
      return "<button class=\"link-button\" data-mcp-server-link-id=\"" + esc(serverID) + "\">" + esc(serverID) + "</button>";
    }
    function renderMCPServerLinks(serverIDs) {
      if (!serverIDs || !serverIDs.length) return "-";
      return serverIDs.map(item => renderMCPServerLink(item)).join(", ");
    }
    function renderTraceLink(traceID, label = "") {
      if (!traceID) return "-";
      return "<button class=\"link-button\" data-trace-link-id=\"" + esc(traceID) + "\">" + esc(label || traceID) + "</button>";
    }
    function renderAuditTarget(event) {
      if (!event || !event.target) return "-";
      const target = event.target;
      const metadata = event.metadata || {};
      if (event.action === "tool.invoke" || metadata.tool_id) {
        return renderToolLink(metadata.tool_id || target);
      }
      if (event.action === "provider.enabled" || event.action === "provider.probe") {
        return renderProviderLink(target);
      }
      if (metadata.policy_rule_id) {
        return renderPolicyLink(metadata.policy_rule_id);
      }
      if (event.action === "routing.explain" || event.action === "policy.dry_run") {
        return renderModelLink(target, metadata.provider_id || "");
      }
      return esc(target);
    }
    function labelChainList(values) {
      if (!values) return "";
      return Array.isArray(values) ? values.join(", ") : String(values);
    }
    function renderPolicyEvaluations(evaluations) {
      if (!evaluations || !evaluations.length) return "";
      return "<div class=\"explain-chain\">" +
        "<div class=\"chain-title\">" + t("ruleEvaluations") + "</div>" +
        "<div class=\"table-wrap\"><table class=\"policy-table\"><thead><tr>" +
          "<th data-i18n=\"evaluationOrder\">" + t("evaluationOrder") + "</th>" +
          "<th data-i18n=\"policy\">" + t("policy") + "</th>" +
          "<th data-i18n=\"priority\">" + t("priority") + "</th>" +
          "<th data-i18n=\"matched\">" + t("matched") + "</th>" +
          "<th data-i18n=\"mismatchFields\">" + t("mismatchFields") + "</th>" +
          "<th data-i18n=\"condition\">" + t("condition") + "</th>" +
        "</tr></thead><tbody>" +
          evaluations.map(item => "<tr>" +
            "<td>" + esc(item.evaluation_order || "") + "</td>" +
            "<td class=\"trace-id\">" + renderPolicyLink(item.rule_id || "") + "</td>" +
            "<td>" + esc(item.priority == null ? 0 : item.priority) + "</td>" +
            "<td>" + esc(item.matched ? t("yes") : t("no")) + "</td>" +
            "<td title=\"" + esc(labelChainList(item.mismatch_fields)) + "\">" + esc(labelChainList(item.mismatch_fields) || "-") + "</td>" +
            "<td title=\"" + esc(item.condition_summary || "") + "\">" + esc(item.condition_summary || "") + "</td>" +
          "</tr>").join("") +
        "</tbody></table></div>" +
      "</div>";
    }
    function renderJSONPanel(node, html) {
      node.outerHTML = html;
      return document.querySelector("#" + node.id);
    }
    function exportPolicyCatalog() {
      const query = policyCatalogQuery();
      exportAdminJSONL("/gateway/v1/policies/export", query, "policies.jsonl", policyMessage);
    }
    async function policyDryRun() {
      policyDryResult.textContent = t("policyDryRunning");
      const labels = policyDryDataLabels.value.split(",").map(item => item.trim()).filter(Boolean);
      const body = {
        app_id: policyDryAppID.value.trim(),
        request_type: policyDryRequestType.value,
        model: policyDryModel.value.trim(),
        provider_class: policyDryProviderClass.value,
        data_labels: labels
      };
      try {
        const res = await fetch("/gateway/v1/policy/dry-run", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(body)
        });
        const data = await res.json();
        if (!res.ok) {
          policyDryResult.textContent = t("failedPrefix") + (data.error && data.error.message || res.status);
          return;
        }
        policyDryResult = renderJSONPanel(policyDryResult, "<div id=\"policy-dry-result\">" + renderExplainChain(data.explain_chain) + renderPolicyEvaluations(data.decision && data.decision.rule_evaluations) + "<pre>" + esc(JSON.stringify(data, null, 2)) + "</pre></div>");
        loadAudit();
      } catch (err) {
        policyDryResult.textContent = t("failedPrefix") + err.message;
      }
    }
    async function loadGrants() {
      const token = adminToken();
      if (!token) {
        allGrants = [];
        grantTotal = 0;
        grantMessage.textContent = t("adminTokenRequired");
        renderGrants();
        return;
      }
      grantMessage.textContent = t("loadingGrants");
      try {
        const query = grantCatalogQuery();
        query.set("limit", String(grantPageSize));
        query.set("offset", String((grantPage - 1) * grantPageSize));
        const res = await fetch("/gateway/v1/grants?" + query.toString(), {
          headers: { "Authorization": "Bearer " + token }
        });
        const data = await res.json();
        if (!res.ok) {
          allGrants = [];
          grantTotal = 0;
          grantMessage.textContent = t("failedPrefix") + (data.error && data.error.message || res.status);
          renderGrants();
          return;
        }
        allGrants = data.grants || [];
        grantTotal = data.total || allGrants.length;
        grantMessage.textContent = grantTotal ? "" : emptyText("noGrants", grantCatalogQuery());
        renderGrants();
      } catch (err) {
        allGrants = [];
        grantTotal = 0;
        grantMessage.textContent = t("failedPrefix") + err.message;
        renderGrants();
      }
    }
    function grantCatalogQuery() {
      const query = new URLSearchParams();
      if (grantIDFilter.value.trim()) query.set("grant", grantIDFilter.value.trim());
      if (grantTypeFilter.value) query.set("type", grantTypeFilter.value);
      if (grantAppFilter.value.trim()) query.set("app_id", grantAppFilter.value.trim());
      if (grantToolFilter.value.trim()) query.set("tool_id", grantToolFilter.value.trim());
      return query;
    }
    function renderGrants() {
      const totalPages = pageCount(grantTotal, grantPageSize);
      grantPage = clampPage(grantPage, totalPages);
      const range = pageRange(grantTotal, grantPage, grantPageSize);
      document.querySelector("#grant-page-summary").textContent = t("page", { page: grantPage, total: totalPages });
      document.querySelector("#grant-range-summary").textContent = t("catalogRange", { range: range.label, total: grantTotal });
      document.querySelector("#grant-prev").disabled = grantPage <= 1;
      document.querySelector("#grant-next").disabled = grantPage >= totalPages;
      grantRows.innerHTML = allGrants.map(item => {
        const apps = (item.apps || []).join(", ") || "-";
        const tools = (item.tools || []).join(", ") || "-";
        const servers = (item.servers || []).join(", ");
        const toolText = servers ? tools + " / " + servers : tools;
        return "<tr>" +
          "<td class=\"trace-id\">" + renderGrantLink(item.id) + "</td>" +
          "<td>" + esc(item.type || "-") + "</td>" +
          "<td title=\"" + esc(apps) + "\">" + renderAppLinks(item.apps || []) + "</td>" +
          "<td title=\"" + esc(toolText) + "\">" + renderToolLinks(item.tools || []) + (item.servers && item.servers.length ? " / " + renderMCPServerLinks(item.servers) : "") + "</td>" +
          "<td title=\"" + esc(item.description || "") + "\">" + esc(item.description || "") + "</td>" +
        "</tr>";
      }).join("");
      if (!allGrants.length) {
        grantRows.innerHTML = "<tr><td colspan=\"5\" class=\"muted\">" + emptyText("noGrants", grantCatalogQuery()) + "</td></tr>";
      }
    }
    async function invokeTool() {
      const toolID = toolSelect.value;
      const token = document.querySelector("#tool-token").value.trim();
      toolResultActions.innerHTML = "";
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
        renderToolTraceAction(data.trace_id || (data.error && data.error.trace_id));
        await Promise.all([loadAudit(), loadTraces()]);
      } catch (err) {
        toolResult.textContent = t("failedPrefix") + err.message;
      }
    }
    function renderToolTraceAction(traceID) {
      toolResultActions.innerHTML = "";
      if (!traceID) return;
      toolResultActions.innerHTML = "<button class=\"secondary\" id=\"tool-open-trace\" data-trace-id=\"" + esc(traceID) + "\">" + t("openToolTrace") + "</button>";
      document.querySelector("#tool-open-trace").addEventListener("click", () => loadDetail(traceID));
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
        await Promise.all([loadProviders(), loadModels(), loadProviderCatalog(), loadModelCatalog(), loadAudit()]);
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
        await Promise.all([loadProviders(), loadModels(), loadProviderCatalog(), loadModelCatalog(), loadAudit()]);
      } catch (err) {
        auditMessage.textContent = t("failedPrefix") + err.message;
      }
    }
    async function configReload() {
      configReloadResult.textContent = t("configReloading");
      try {
        const data = await providerRequest("/gateway/v1/config/reload", { method: "POST" });
        configReloadResult.textContent = JSON.stringify(data, null, 2);
        await loadAll();
      } catch (err) {
        configReloadResult.textContent = t("failedPrefix") + err.message;
      }
    }
    async function loadDetail(traceId) {
      const res = await fetch("/gateway/v1/traces/" + encodeURIComponent(traceId));
      const trace = await res.json();
      selectedTrace = trace;
      detail.innerHTML =
        "<div class=\"kv\">" +
          "<div class=\"k\">" + t("traceId") + "</div><div class=\"trace-id\">" + renderTraceLink(trace.trace_id) + "</div>" +
          "<div class=\"k\">" + t("status") + "</div><div><span class=\"status " + esc(trace.status) + "\">" + esc(labelStatus(trace.status)) + "</span></div>" +
          "<div class=\"k\">" + t("provider") + "</div><div>" + renderProviderLink(trace.provider_id) + "</div>" +
          "<div class=\"k\">" + t("policyRule") + "</div><div>" + renderTracePolicy(trace.policy) + "</div>" +
          "<div class=\"k\">" + t("fallbacks") + "</div><div>" + ((trace.fallbacks || []).length) + "</div>" +
        "</div>" +
        renderTraceSummary(trace) +
        "<div class=\"actions\" style=\"margin-bottom: 12px;\">" +
          "<button class=\"secondary\" id=\"trace-fill-quick\" data-i18n=\"replayTrace\">" + t("replayTrace") + "</button>" +
          "<button class=\"secondary\" id=\"trace-view-audit\" data-i18n=\"viewTraceAudit\">" + t("viewTraceAudit") + "</button>" +
          "<button class=\"secondary\" id=\"trace-copy-request\" data-i18n=\"copyTraceRequest\">" + t("copyTraceRequest") + "</button>" +
          "<button class=\"secondary\" id=\"trace-copy-curl\" data-i18n=\"copyTraceCurl\">" + t("copyTraceCurl") + "</button>" +
        "</div>" +
        "<div class=\"notice\">" +
          "<strong data-i18n=\"traceSnapshotSafetyTitle\">" + t("traceSnapshotSafetyTitle") + "</strong>" +
          "<div data-i18n=\"traceSnapshotSafetyBody\">" + t("traceSnapshotSafetyBody") + "</div>" +
          "<div class=\"muted\" data-i18n=\"traceCopySafetyHint\">" + t("traceCopySafetyHint") + "</div>" +
        "</div>" +
        "<pre>" + esc(JSON.stringify(trace, null, 2)) + "</pre>";
      document.querySelector("#trace-fill-quick").addEventListener("click", fillQuickRequestFromTrace);
      document.querySelector("#trace-view-audit").addEventListener("click", filterAuditBySelectedTrace);
      document.querySelector("#trace-copy-request").addEventListener("click", copyTraceRequestJSON);
      document.querySelector("#trace-copy-curl").addEventListener("click", copyTraceCurlDraft);
    }
    function renderTracePolicy(policy) {
      if (!policy || !policy.rule_id) return "-";
      return "<button class=\"link-button\" data-trace-policy-id=\"" + esc(policy.rule_id) + "\">" + esc(policy.rule_id) + "</button> / " + esc(policy.explanation || "");
    }
    function renderTraceSummary(trace) {
      if (!trace) return "";
      const cells = [
        [t("requestTypeFilter"), esc(trace.request_type || "-")],
        [t("apps"), renderAppLink(trace.app_id)],
        [t("toolName"), renderToolLink(trace.tool_id)],
        [t("requested"), renderModelLink(trace.requested_model, trace.provider_id)],
        [t("final"), renderModelLink(trace.final_model, trace.provider_id)],
        [t("provider"), renderProviderLink(trace.provider_id)]
      ];
      const routeCells = (trace.routes || []).map(item =>
        [t("route"), renderProviderLink(item.provider_id) + " / " + renderModelLink(item.model, item.provider_id) + (item.reason ? " / " + esc(item.reason) : "")]
      );
      const fallbackCells = (trace.fallbacks || []).map(item =>
        [t("fallbacks"), renderProviderLink(item.from_provider_id) + " / " + esc(item.action || "-") + (item.reason ? " / " + esc(item.reason) : "")]
      );
      return "<div class=\"explain-chain\"><div class=\"chain-title\">Trace</div><div class=\"chain-grid\">" +
        cells.concat(routeCells, fallbackCells)
          .filter(item => item[1] && item[1] !== "-")
          .map(item => "<div class=\"chain-cell\"><span class=\"k\">" + esc(item[0]) + "</span><div>" + item[1] + "</div></div>")
          .join("") +
        "</div></div>";
    }
    async function filterAuditBySelectedTrace() {
      if (!selectedTrace || !selectedTrace.trace_id) return;
      auditTraceFilter.value = selectedTrace.trace_id;
      auditPage = 1;
      await loadAudit();
    }
    async function filterAuditByMetadata(key, value) {
      if (!key) return;
      auditMetadataKeyFilter.value = key;
      auditMetadataValueFilter.value = value || "";
      auditPage = 1;
      await loadAudit();
    }
    function clearTraceFilters() {
      traceAppFilter.value = "";
      traceProviderFilter.value = "";
      statusFilter.value = "";
      tracePage = 1;
      loadTraces();
    }
    function clearAuditFilters() {
      auditActionFilter.value = "";
      auditResultFilter.value = "";
      auditAppFilter.value = "";
      auditTraceFilter.value = "";
      auditTargetFilter.value = "";
      auditMetadataKeyFilter.value = "";
      auditMetadataValueFilter.value = "";
      auditPage = 1;
      loadAudit();
    }
    function clearToolFilters() {
      toolIDFilter = "";
      toolOriginFilter.value = "";
      toolServerFilter.value = "";
      toolScopeFilter.value = "";
      toolEnabledFilter.value = "";
      toolPage = 1;
      loadTools();
    }
    function clearMCPFilters() {
      mcpServerFilter.value = "";
      mcpScopeFilter.value = "";
      mcpEnabledFilter.value = "";
      mcpPage = 1;
      loadMCPCatalog();
    }
    function clearAppFilters() {
      appIDFilter.value = "";
      appGrantFilter.value = "";
      appQuotaFilter.value = "";
      appPage = 1;
      loadApps();
    }
    function clearGrantFilters() {
      grantIDFilter.value = "";
      grantTypeFilter.value = "";
      grantAppFilter.value = "";
      grantToolFilter.value = "";
      grantPage = 1;
      loadGrants();
    }
    function clearProviderFilters() {
      providerIDFilter.value = "";
      providerClassFilter.value = "";
      providerEnabledFilter.value = "";
      providerRuntimeFilter.value = "";
      providerQuotaFilter.value = "";
      providerPage = 1;
      loadProviderCatalog();
    }
    function clearModelFilters() {
      modelNameFilter.value = "";
      modelProviderFilter.value = "";
      modelClassFilter.value = "";
      modelAvailableFilter.value = "";
      modelPage = 1;
      loadModelCatalog();
    }
    function clearPolicyFilters() {
      policyIDFilter.value = "";
      policyEffectFilter.value = "";
      policyAppFilter.value = "";
      policyModelFilter.value = "";
      policyRequestTypeFilter.value = "";
      policyProviderClassFilter.value = "";
      policyDataLabelFilter.value = "";
      policyPage = 1;
      loadPolicyCatalog();
    }
    function filterProviderByID(providerID, quotaEnabled = "") {
      if (!providerID) return;
      providerIDFilter.value = providerID;
      providerQuotaFilter.value = quotaEnabled || "";
      providerPage = 1;
      loadProviderCatalog();
    }
    function filterProviderByClass(providerClass) {
      if (!providerClass) return;
      providerIDFilter.value = "";
      providerClassFilter.value = providerClass;
      providerPage = 1;
      loadProviderCatalog();
    }
    function filterModelByID(model, providerID = "") {
      if (!model) return;
      modelNameFilter.value = model;
      modelProviderFilter.value = providerID || "";
      modelAvailableFilter.value = "all";
      modelPage = 1;
      loadModelCatalog();
    }
    function filterAppByID(appID, quotaEnabled = "") {
      if (!appID) return;
      appIDFilter.value = appID;
      appQuotaFilter.value = quotaEnabled || "";
      appPage = 1;
      loadApps();
    }
    function filterGrantByID(grantID) {
      if (!grantID) return;
      grantIDFilter.value = grantID;
      grantTypeFilter.value = "";
      grantAppFilter.value = "";
      grantToolFilter.value = "";
      grantPage = 1;
      loadGrants();
    }
    function filterToolByID(toolID) {
      if (!toolID) return;
      toolIDFilter = toolID;
      toolOriginFilter.value = "";
      toolServerFilter.value = "";
      toolScopeFilter.value = "";
      toolEnabledFilter.value = "";
      toolPage = 1;
      loadTools();
    }
    function filterMCPServerByID(serverID) {
      if (!serverID) return;
      mcpServerFilter.value = serverID;
      mcpPage = 1;
      loadMCPCatalog();
      toolIDFilter = "";
      toolSelect.value = "";
      toolOriginFilter.value = "mcp";
      toolServerFilter.value = serverID;
      toolPage = 1;
      loadTools();
    }
    function filterTraceByID(traceID) {
      if (!traceID) return;
      loadDetail(traceID);
    }
    function filterPolicyByField(key, value) {
      if (!key || !value) return;
      if (key === "request_type") policyRequestTypeFilter.value = value;
      if (key === "data_label") policyDataLabelFilter.value = value;
      policyPage = 1;
      loadPolicyCatalog();
    }
    function traceRequestBody() {
      const request = selectedTrace && selectedTrace.request;
      if (!request || !request.model) return null;
      const body = {
        model: request.model,
        messages: request.messages || []
      };
      if (request.metadata && Object.keys(request.metadata).length) body.metadata = request.metadata;
      if (request.data_labels && request.data_labels.length) body.data_labels = request.data_labels;
      return body;
    }
    function fillQuickRequestFromTrace() {
      const request = selectedTrace && selectedTrace.request;
      if (!request || !request.model) {
        sendResult.textContent = t("replayUnavailable");
        return;
      }
      document.querySelector("#model").value = request.model || selectedTrace.requested_model || "";
      const userMessage = (request.messages || []).find(item => item.role === "user") || (request.messages || [])[0];
      if (userMessage) {
        document.querySelector("#prompt").value = userMessage.content || "";
      }
      const metadata = request.metadata || {};
      const labels = request.data_labels || [];
      let mode = "success";
      if (metadata.fail_provider) mode = "fallback";
      if (labels.includes("sensitive")) mode = "blocked";
      document.querySelector("#mode").value = mode;
      routingDryModel.value = request.model || "";
      routingDryDataLabels.value = labels.join(",");
      routingDryRequestType.value = selectedTrace.request_type || "chat";
      sendResult.textContent = t("replayFilled");
    }
    async function copyTraceRequestJSON() {
      const body = traceRequestBody();
      if (!body) {
        sendResult.textContent = t("replayUnavailable");
        return;
      }
      await copyText(JSON.stringify(body, null, 2), t("traceRequestCopied"));
    }
    async function copyTraceCurlDraft() {
      const body = traceRequestBody();
      if (!body) {
        sendResult.textContent = t("replayUnavailable");
        return;
      }
      const json = JSON.stringify(body, null, 2);
      const curl = "curl -X POST http://127.0.0.1:8080/v1/chat/completions \\\n" +
        "  -H \"Authorization: Bearer $GATEWAY_TOKEN\" \\\n" +
        "  -H \"Content-Type: application/json\" \\\n" +
        "  --data '" + json.replaceAll("'", "'\\''") + "'";
      await copyText(curl, t("traceCurlCopied"));
    }
    async function copyText(value, successMessage, targetNode) {
      const target = targetNode || sendResult;
      try {
        if (navigator.clipboard && window.isSecureContext) {
          await navigator.clipboard.writeText(value);
        } else {
          const area = document.createElement("textarea");
          area.value = value;
          area.style.position = "fixed";
          area.style.left = "-9999px";
          document.body.appendChild(area);
          area.focus();
          area.select();
          document.execCommand("copy");
          area.remove();
        }
        target.textContent = successMessage || t("copied");
      } catch (err) {
        target.textContent = t("copyFailed", { error: err.message });
      }
    }
    async function loadAudit() {
      const token = adminToken();
      if (!token) {
          auditRows.innerHTML = "";
          auditMessage.textContent = t("adminTokenRequired");
          syncIssues();
          return;
      }
      try {
        auditMessage.textContent = t("loadingAudit");
        const query = auditQuery(auditPageSize, (auditPage - 1) * auditPageSize);
        const res = await fetch("/gateway/v1/audit/events?" + query.toString(), {
          headers: { "Authorization": "Bearer " + token }
        });
        const data = await res.json();
        if (!res.ok) {
          auditRows.innerHTML = "";
          auditMessage.textContent = t("failedPrefix") + (data.error && data.error.message || res.status);
          syncIssues();
          return;
        }
        allAuditEvents = data.events || [];
        auditTotal = data.total || allAuditEvents.length;
        auditMessage.textContent = auditTotal ? "" : emptyText("noAuditEvents", auditQuery());
        renderAudit();
        syncIssues();
      } catch (err) {
        auditRows.innerHTML = "";
        auditMessage.textContent = t("failedPrefix") + err.message;
        syncIssues();
      }
    }
    function auditQuery(limit = auditPageSize, offset = (auditPage - 1) * auditPageSize) {
      const query = new URLSearchParams({ limit: String(limit), offset: String(offset) });
      if (auditActionFilter.value) query.set("action", auditActionFilter.value);
      if (auditResultFilter.value) query.set("result", auditResultFilter.value);
      if (auditAppFilter.value.trim()) query.set("app_id", auditAppFilter.value.trim());
      if (auditTraceFilter.value.trim()) query.set("trace_id", auditTraceFilter.value.trim());
      if (auditTargetFilter.value.trim()) query.set("target", auditTargetFilter.value.trim());
      if (auditMetadataKeyFilter.value.trim()) query.set("metadata_key", auditMetadataKeyFilter.value.trim());
      if (auditMetadataValueFilter.value.trim()) query.set("metadata_value", auditMetadataValueFilter.value.trim());
      return query;
    }
    function exportAudit() {
      const token = adminToken();
      if (!token) {
        auditMessage.textContent = t("adminTokenRequired");
        return;
      }
      auditMessage.textContent = t("auditExportSafety");
      fetch("/gateway/v1/audit/events/export?" + auditQuery(500, 0).toString(), {
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
    function exportAdminJSONL(path, query, filename, messageNode) {
      const token = adminToken();
      if (!token) {
        messageNode.textContent = t("adminTokenRequired");
        return;
      }
      const suffix = query.toString() ? "?" + query.toString() : "";
      fetch(path + suffix, {
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
          downloadURL(url, filename);
          setTimeout(() => URL.revokeObjectURL(url), 1000);
        })
        .catch(err => {
          messageNode.textContent = t("failedPrefix") + err.message;
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
        "<tr data-audit=\"" + esc(item.id || "") + "\" data-trace=\"" + esc(item.trace_id || "") + "\" title=\"" + esc(item.error || item.trace_id || "") + "\">" +
          "<td>" + esc(labelAction(item.action)) + "</td>" +
          "<td><span class=\"status " + esc(item.result) + "\">" + esc(labelResult(item.result)) + "</span></td>" +
          "<td class=\"trace-id\">" + renderTraceLink(item.trace_id, shortTraceID(item.trace_id)) + "</td>" +
          "<td>" + renderAppLink(item.app_id) + " / " + renderAuditTarget(item) + "</td>" +
          "<td>" + esc(time(item.created_at)) + "</td>" +
          "<td>" + esc(item.duration_ms == null ? "-" : item.duration_ms + "ms") + "</td>" +
        "</tr>"
      ).join("");
      if (!allAuditEvents.length) {
        auditRows.innerHTML = "<tr><td colspan=\"6\" class=\"muted\">" + emptyText("noAuditEvents", auditQuery()) + "</td></tr>";
      }
      auditRows.querySelectorAll("tr[data-audit]").forEach(row => {
        row.addEventListener("click", event => {
          if (event.target.closest("button")) return;
          showAuditDetail(row.dataset.audit, row.dataset.trace);
        });
      });
    }
    function showAuditDetail(auditID, traceID) {
      const event = allAuditEvents.find(item => item.id === auditID);
      if (!event) {
        auditDetail.innerHTML = "<p class=\"muted\">" + t("selectAudit") + "</p>";
        return;
      }
      auditDetail.innerHTML =
        "<div class=\"kv\">" +
          "<div class=\"k\">" + t("action") + "</div><div>" + esc(labelAction(event.action)) + "</div>" +
          "<div class=\"k\">" + t("result") + "</div><div><span class=\"status " + esc(event.result) + "\">" + esc(labelResult(event.result)) + "</span></div>" +
          "<div class=\"k\">" + t("traceId") + "</div><div class=\"trace-id\">" + renderTraceLink(event.trace_id) + "</div>" +
          "<div class=\"k\">" + t("appTarget") + "</div><div>" + renderAppLink(event.app_id) + " / " + renderAuditTarget(event) + "</div>" +
          "<div class=\"k\">" + t("policyRule") + "</div><div>" + renderAuditPolicy(event.metadata) + "</div>" +
          "<div class=\"k\">" + t("time") + "</div><div>" + esc(time(event.created_at)) + "</div>" +
        "</div>" +
        renderExplainChain(event.metadata && event.metadata.explain_chain) +
        renderAuditMetadataSummary(event.metadata) +
        "<pre>" + esc(JSON.stringify(event, null, 2)) + "</pre>";
      if (traceID) {
        loadDetail(traceID);
      }
    }
    function renderAuditPolicy(metadata) {
      const policyID = metadata && metadata.policy_rule_id;
      if (!policyID) return "-";
      return renderPolicyLink(policyID);
    }
    function renderAuditMetadataSummary(metadata) {
      if (!metadata || !Object.keys(metadata).length) return "";
      const keys = ["tool_id", "provider_id", "policy_rule_id", "matched_grant", "missing_grants", "required_scopes", "origin", "server_id"];
      const cells = keys
        .filter(key => metadata[key] !== undefined && metadata[key] !== null && metadata[key] !== "")
        .map(key => "<div class=\"chain-cell\"><span class=\"k\">" + esc(key) + "</span><div>" + renderAuditMetadataValue(key, metadata[key]) + "</div></div>");
      if (!cells.length) return "";
      return "<div class=\"explain-chain\"><div class=\"chain-title\">Metadata</div><div class=\"chain-grid\">" + cells.join("") + "</div></div>";
    }
    function renderAuditMetadataValue(key, value) {
      const values = Array.isArray(value) ? value : [value];
      return values.map(item => {
        const text = String(item);
        const filter = renderAuditMetadataFilter(key, text);
        let linked = esc(text);
        if (key === "tool_id") linked = renderToolLink(text);
        if (key === "provider_id") linked = renderProviderLink(text);
        if (key === "policy_rule_id") linked = renderPolicyLink(text);
        if (key === "matched_grant" || key === "missing_grants") linked = renderGrantLink(text);
        if (key === "required_scopes") linked = renderGrantLink("tool:" + text);
        if (key === "server_id") linked = renderMCPServerLink(text);
        return linked + " " + filter;
      }).join(", ");
    }
    function renderAuditMetadataFilter(key, value) {
      return "<button class=\"link-button\" data-audit-metadata-key=\"" + esc(key) + "\" data-audit-metadata-value=\"" + esc(value) + "\">#</button>";
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
        sendResult.innerHTML = res.ok
          ? t("okPrefix") + renderTraceLink(data.trace_id)
          : t("failedPrefix") + renderTraceLink(data.error && data.error.trace_id);
      } catch (err) {
        sendResult.textContent = t("failedPrefix") + err.message;
      }
      await loadTraces();
    }
    function requestLabelsForMode(mode) {
      return mode === "blocked" ? ["sensitive"] : [];
    }
    async function explainRouting(mode) {
      await runRoutingExplain({
        app_id: "dev-app",
        request_type: "chat",
        model: document.querySelector("#model").value,
        data_labels: requestLabelsForMode(mode)
      });
    }
    async function routingDryRun() {
      const labels = routingDryDataLabels.value.split(",").map(item => item.trim()).filter(Boolean);
      await runRoutingExplain({
        app_id: routingDryAppID.value.trim(),
        request_type: routingDryRequestType.value,
        model: routingDryModel.value.trim(),
        data_labels: labels
      });
    }
    async function runRoutingExplain(body) {
      routeExplain.innerHTML = "<p class=\"muted\">" + t("explainLoading") + "</p>";
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
            "<div class=\"k\">" + t("rule") + "</div><div>" + renderPolicyLink(data.policy && data.policy.rule_id) + "</div>" +
            "<div class=\"k\">" + t("cloud") + "</div><div>" + (data.policy && data.policy.allow_cloud ? t("allowed") : t("blocked")) + "</div>" +
            "<div class=\"k\">" + t("reason") + "</div><div>" + esc(data.policy && data.policy.explanation) + "</div>" +
          "</div>" +
          renderExplainChain(data.explain_chain) +
          "<h2>" + t("candidates") + "</h2>" +
          renderRouteItems(data.candidates || [], false) +
          "<h2 style=\"margin-top:12px;\">" + t("skipped") + "</h2>" +
          renderRouteItems(data.skipped || [], true);
        loadAudit();
      } catch (err) {
        routeExplain.textContent = t("failedPrefix") + err.message;
      }
    }
    async function accessDryRun() {
      accessResult.textContent = t("dryRunning");
      const body = {
        app_id: document.querySelector("#access-app-id").value.trim(),
        token: document.querySelector("#access-token").value.trim(),
        action: document.querySelector("#access-action").value,
        tool_id: document.querySelector("#access-tool-id").value.trim()
      };
      try {
        const res = await fetch("/gateway/v1/access/dry-run", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(body)
        });
        const data = await res.json();
        accessResult = renderJSONPanel(accessResult, "<div id=\"access-result\">" + renderExplainChain(data.explain_chain) + "<pre>" + esc(JSON.stringify(data, null, 2)) + "</pre></div>");
        await loadAudit();
      } catch (err) {
        accessResult.textContent = t("failedPrefix") + err.message;
      }
    }
    function renderRouteItems(items, skipped) {
      if (!items.length) return "<p class=\"muted\">" + t("none") + "</p>";
      return items.map(item => {
        const provider = skipped ? { id: item.provider_id, class: item.class } : (item.provider || {});
        return "<div class=\"route-item" + (skipped ? " skipped" : "") + "\">" +
          "<strong>" + renderProviderLink(provider.id || "") + "</strong>" +
          "<div class=\"muted\">" + esc(provider.class || "-") + (item.model ? " / " + renderModelLink(item.model, provider.id || "") : "") + "</div>" +
          "<div>" + esc(item.reason || "") + "</div>" +
        "</div>";
      }).join("");
    }
    setLang(lang);
    loadAll().catch(err => { summary.textContent = t("loadConsoleFailed") + err.message; });
  </script>
</body>
</html>`
