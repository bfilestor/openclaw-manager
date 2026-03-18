param(
  [string]$ServiceName = "OpenClawManager",
  [string]$Listen = "0.0.0.0:18799"
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$managerHome = Join-Path $env:USERPROFILE ".openclaw-manager"
$configPath = Join-Path $managerHome "config.toml"
$binaryPath = Join-Path $managerHome "managerd.exe"

if (-not (Test-Path $binaryPath)) {
  throw "managerd.exe not found at $binaryPath. Run scripts/build-windows.ps1 first."
}

if (-not (Test-Path $configPath)) {
  throw "config.toml not found at $configPath. Please create it first."
}

$binArg = '"' + $binaryPath + '" --config "' + $configPath + '" --static-dir "' + (Join-Path $root "src/frontend/dist") + '"'

$exists = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($null -eq $exists) {
  sc.exe create $ServiceName binPath= $binArg start= auto | Out-Host
} else {
  sc.exe config $ServiceName binPath= $binArg start= auto | Out-Host
}

try {
  sc.exe stop $ServiceName | Out-Host
  Start-Sleep -Seconds 1
} catch {
}
sc.exe start $ServiceName | Out-Host
sc.exe query $ServiceName | Out-Host

Write-Host "Service installed/updated: $ServiceName"
