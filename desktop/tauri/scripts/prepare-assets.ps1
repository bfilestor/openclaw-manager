$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "../../..")
$tauriDir = Join-Path $root "desktop/tauri"
$resDir = Join-Path $tauriDir "src-tauri/resources"

New-Item -ItemType Directory -Force -Path $resDir | Out-Null

$binLinux = Join-Path $root "src/bin/managerd"
$binWin = Join-Path $root "src/bin/managerd.exe"
if (Test-Path $binLinux) { Copy-Item -Force $binLinux (Join-Path $resDir "managerd") }
if (Test-Path $binWin) { Copy-Item -Force $binWin (Join-Path $resDir "managerd.exe") }

$frontendDist = Join-Path $root "src/frontend/dist"
$targetDist = Join-Path $resDir "frontend-dist"
if (Test-Path $targetDist) { Remove-Item -Recurse -Force $targetDist }
New-Item -ItemType Directory -Force -Path $targetDist | Out-Null
Copy-Item -Recurse -Force (Join-Path $frontendDist "*") $targetDist

Write-Host "Assets prepared under $resDir"
