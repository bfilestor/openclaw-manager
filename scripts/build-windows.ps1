param(
  [switch]$Frontend,
  [switch]$Backend,
  [switch]$All
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$src = Join-Path $root "src"
$frontendDir = Join-Path $src "frontend"
$targetHome = Join-Path $env:USERPROFILE ".openclaw-manager"
$targetBin = Join-Path $targetHome "managerd.exe"

if (-not $Frontend -and -not $Backend -and -not $All) {
  $All = $true
}
if ($All) {
  $Frontend = $true
  $Backend = $true
}

if ($Backend) {
  Write-Host "[build] backend"
  Push-Location $src
  try {
    go clean -cache
    $env:CGO_ENABLED = "0"
    go build -o (Join-Path $src "bin/managerd.exe") ./cmd/server
  } finally {
    Pop-Location
  }
}

if ($Frontend) {
  Write-Host "[build] frontend"
  Push-Location $frontendDir
  try {
    pnpm run build
  } finally {
    Pop-Location
  }
}

$built = Join-Path $src "bin/managerd.exe"
if (-not (Test-Path $built)) {
  throw "Backend binary not found: $built"
}

New-Item -ItemType Directory -Force -Path $targetHome | Out-Null
Copy-Item -Force $built $targetBin

Write-Host "Build completed. Binary copied to $targetBin"
Write-Host "If service is installed, run scripts/install-service.ps1 to refresh configuration."
