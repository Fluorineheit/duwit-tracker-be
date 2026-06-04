$ErrorActionPreference = "Stop"

$root = Resolve-Path -LiteralPath (Join-Path $PSScriptRoot "..")
$envFile = Join-Path $root ".env"
$port = "8080"

if (Test-Path -LiteralPath $envFile) {
    foreach ($line in Get-Content -LiteralPath $envFile) {
        if ($line -match "^\s*APP_PORT\s*=\s*(.+?)\s*$") {
            $port = $Matches[1].Trim()
            break
        }
    }
}

$listeners = Get-NetTCPConnection -LocalPort ([int]$port) -State Listen -ErrorAction SilentlyContinue

foreach ($listener in $listeners) {
    $process = Get-Process -Id $listener.OwningProcess -ErrorAction SilentlyContinue
    if ($null -eq $process) {
        continue
    }

    $isGoBuildApi = $process.ProcessName -eq "api" -and $process.Path -match "\\go-build|\\Temp\\go-build"
    if (-not $isGoBuildApi) {
        Write-Host "Port $port is used by $($process.ProcessName) (PID $($process.Id)); leaving it alone."
        continue
    }

    Write-Host "Stopping old API process on port $port (PID $($process.Id))..."
    Stop-Process -Id $process.Id -Force
}

$still = Get-NetTCPConnection -LocalPort ([int]$port) -State Listen -ErrorAction SilentlyContinue
foreach ($conn in $still) {
    Stop-Process -Id $conn.OwningProcess -Force -ErrorAction SilentlyContinue
}

Set-Location -LiteralPath $root
go run ./cmd/api
