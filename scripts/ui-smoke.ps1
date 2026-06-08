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
  Write-Host "UI smoke passed: $OutDir"
} finally {
  if ($startedProcess -and -not $startedProcess.HasExited) {
    Stop-Process -Id $startedProcess.Id -Force
  }
}
