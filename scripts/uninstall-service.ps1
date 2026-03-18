param(
  [string]$ServiceName = "OpenClawManager"
)

$ErrorActionPreference = "Stop"

$exists = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($null -eq $exists) {
  Write-Host "Service not found: $ServiceName"
  exit 0
}

try {
  sc.exe stop $ServiceName | Out-Host
  Start-Sleep -Seconds 1
} catch {
}

sc.exe delete $ServiceName | Out-Host
Write-Host "Service removed: $ServiceName"
