param(
  [string]$Config = "configs/dev.json",
  [string]$BaseUrl = "http://127.0.0.1:18765",
  [string]$OutDir = "artifacts/ui-smoke"
)

$ErrorActionPreference = "Stop"

function Find-Browser {
  $candidates = @(
    "$env:ProgramFiles\Microsoft\Edge\Application\msedge.exe",
    "${env:ProgramFiles(x86)}\Microsoft\Edge\Application\msedge.exe",
    "$env:ProgramFiles\Google\Chrome\Application\chrome.exe",
    "${env:ProgramFiles(x86)}\Google\Chrome\Application\chrome.exe"
  )
  foreach ($candidate in $candidates) {
    if ($candidate -and (Test-Path $candidate)) {
      return $candidate
    }
  }
  throw "Edge or Chrome was not found; cannot run browser UI smoke."
}

function Test-GatewayReady {
  param([string]$Url)
  try {
    $response = Invoke-WebRequest -Uri "${Url}/console" -UseBasicParsing -TimeoutSec 2
    return $response.StatusCode -eq 200
  } catch {
    return $false
  }
}

function Wait-Gateway {
  param([string]$Url)
  for ($i = 0; $i -lt 30; $i++) {
    if (Test-GatewayReady $Url) {
      return
    }
    Start-Sleep -Milliseconds 500
  }
  throw "Gateway did not become ready in time: $Url"
}

function Capture-Screenshot {
  param(
    [string]$Browser,
    [string]$Url,
    [string]$Output,
    [string]$WindowSize
  )
  $parts = $WindowSize.Split(",")
  $expectedWidth = [int]$parts[0]
  $expectedHeight = [int]$parts[1]
  $resolvedOutput = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($Output)
  $userDataDir = Join-Path $env:TEMP ("client-ai-gateway-ui-smoke-" + [guid]::NewGuid().ToString("N"))
  New-Item -ItemType Directory -Path $userDataDir | Out-Null
  try {
    $args = @(
      "--headless=new",
      "--disable-gpu",
      "--hide-scrollbars",
      "--no-first-run",
      "--no-default-browser-check",
      "--user-data-dir=$userDataDir",
      "--window-size=$WindowSize",
      "--screenshot=$resolvedOutput",
      "${Url}/console"
    )
    $process = Start-Process -FilePath $Browser -ArgumentList $args -Wait -PassThru -WindowStyle Hidden
    if ($process.ExitCode -ne 0) {
      throw "Browser screenshot failed, ExitCode=$($process.ExitCode), WindowSize=$WindowSize"
    }
    if (-not (Test-Path $resolvedOutput)) {
      throw "Screenshot was not created: $resolvedOutput"
    }
    $file = Get-Item $resolvedOutput
    if ($file.Length -lt 10000) {
      throw "Screenshot file is too small; page may not have rendered: $resolvedOutput ($($file.Length) bytes)"
    }
    Add-Type -AssemblyName System.Drawing
    $image = [System.Drawing.Image]::FromFile($resolvedOutput)
    try {
      if ($image.Width -ne $expectedWidth -or $image.Height -ne $expectedHeight) {
        throw "Screenshot size mismatch: $resolvedOutput expected ${expectedWidth}x${expectedHeight}, got $($image.Width)x$($image.Height)"
      }
    } finally {
      $image.Dispose()
    }
  } finally {
    Remove-Item -LiteralPath $userDataDir -Recurse -Force -ErrorAction SilentlyContinue
  }
}

function Capture-DOM {
  param(
    [string]$Browser,
    [string]$Url,
    [string]$Output,
    [string]$WindowSize
  )
  $resolvedOutput = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($Output)
  $userDataDir = Join-Path $env:TEMP ("client-ai-gateway-ui-smoke-dom-" + [guid]::NewGuid().ToString("N"))
  New-Item -ItemType Directory -Path $userDataDir | Out-Null
  try {
    $args = @(
      "--headless=new",
      "--disable-gpu",
      "--no-first-run",
      "--no-default-browser-check",
      "--virtual-time-budget=5000",
      "--user-data-dir=$userDataDir",
      "--window-size=$WindowSize",
      "--dump-dom",
      "${Url}/console"
    )
    $stdoutPath = Join-Path $userDataDir "dom.stdout.html"
    $stderrPath = Join-Path $userDataDir "dom.stderr.log"
    $process = Start-Process -FilePath $Browser -ArgumentList $args -Wait -PassThru -NoNewWindow -RedirectStandardOutput $stdoutPath -RedirectStandardError $stderrPath
    if ($process.ExitCode -ne 0) {
      throw "Browser DOM dump failed, ExitCode=$($process.ExitCode), WindowSize=$WindowSize"
    }
    if (Test-Path $stdoutPath) {
      Copy-Item -LiteralPath $stdoutPath -Destination $resolvedOutput -Force
    }
    if (-not (Test-Path $resolvedOutput)) {
      throw "DOM dump was not created: $resolvedOutput"
    }
    $file = Get-Item $resolvedOutput
    if ($file.Length -lt 20000) {
      throw "DOM dump is too small; console may not have rendered: $resolvedOutput ($($file.Length) bytes)"
    }
  } finally {
    Remove-Item -LiteralPath $userDataDir -Recurse -Force -ErrorAction SilentlyContinue
  }
}

function Assert-FileContains {
  param(
    [string]$Path,
    [string[]]$Needles
  )
  $content = Get-Content -Path $Path -Raw -Encoding UTF8
  foreach ($needle in $Needles) {
    if (-not $content.Contains($needle)) {
      throw "Expected UI smoke artifact '$Path' to contain '$needle'."
    }
  }
}

function New-Text {
  param([int[]]$CodePoints)
  return -join ($CodePoints | ForEach-Object { [char]$_ })
}

$browser = Find-Browser
New-Item -ItemType Directory -Path $OutDir -Force | Out-Null

$startedProcess = $null
if (-not (Test-GatewayReady $BaseUrl)) {
  $daemonExe = Join-Path $OutDir "gateway-daemon-ui-smoke.exe"
  & go build -o $daemonExe ./cmd/gateway-daemon
  if ($LASTEXITCODE -ne 0) {
    throw "go build failed for UI smoke daemon."
  }
  $startedProcess = Start-Process -FilePath $daemonExe -ArgumentList @("-config", $Config) -PassThru -WindowStyle Hidden
}

try {
  Wait-Gateway $BaseUrl
  Capture-Screenshot -Browser $browser -Url $BaseUrl -Output (Join-Path $OutDir "console-desktop.png") -WindowSize "1440,1000"
  Capture-Screenshot -Browser $browser -Url $BaseUrl -Output (Join-Path $OutDir "console-narrow.png") -WindowSize "390,900"

  $desktopDOM = Join-Path $OutDir "console-desktop.dom.html"
  $narrowDOM = Join-Path $OutDir "console-narrow.dom.html"
  Capture-DOM -Browser $browser -Url $BaseUrl -Output $desktopDOM -WindowSize "1440,1000"
  Capture-DOM -Browser $browser -Url $BaseUrl -Output $narrowDOM -WindowSize "390,900"

  $zhTitle = (New-Text @(0x5ba2, 0x6237, 0x7aef)) + " AI " + (New-Text @(0x7f51, 0x5173, 0x63a7, 0x5236, 0x53f0))
  $zhIssues = New-Text @(0x8fd0, 0x884c, 0x95ee, 0x9898, 0x6c47, 0x603b)
  $zhApps = New-Text @(0x5e94, 0x7528, 0x4e0e, 0x6388, 0x6743)
  $zhProviders = "Provider " + (New-Text @(0x76ee, 0x5f55))
  $zhTraces = New-Text @(0x8bf7, 0x6c42, 0x8ffd, 0x8e2a)
  $zhAudit = New-Text @(0x5ba1, 0x8ba1, 0x4e8b, 0x4ef6)
  $zhQuotaRejected = New-Text @(0x914d, 0x989d, 0x62d2, 0x7edd)
  $zhUsageSummary = New-Text @(0x4f7f, 0x7528, 0x91cf, 0x6c47, 0x603b)
  $zhPrev = New-Text @(0x4e0a, 0x4e00, 0x9875)
  $zhNext = New-Text @(0x4e0b, 0x4e00, 0x9875)

  Assert-FileContains -Path $desktopDOM -Needles @(
    $zhTitle,
    $zhIssues,
    $zhApps,
    $zhProviders,
    $zhTraces,
    $zhAudit,
    $zhQuotaRejected,
    $zhUsageSummary,
    $zhPrev,
    $zhNext,
    "English"
  )
  Assert-FileContains -Path $narrowDOM -Needles @(
    $zhTitle,
    $zhIssues,
    $zhApps,
    $zhTraces,
    $zhAudit,
    $zhUsageSummary,
    $zhPrev,
    $zhNext
  )
  Write-Host "UI smoke passed: $OutDir"
} finally {
  if ($startedProcess -and -not $startedProcess.HasExited) {
    Stop-Process -Id $startedProcess.Id -Force
  }
}
