# Initialize SEU OJ MySQL schema and demo data via Go (no mysql CLI required).
# Reads ../config/config.yaml automatically.
#
# Usage (PowerShell):
#   cd seu-oj-backend/database
#   powershell -ExecutionPolicy Bypass -File .\init.ps1
#   powershell -ExecutionPolicy Bypass -File .\init.ps1 -Force

param(
    [switch]$Force
)

$ErrorActionPreference = "Stop"
$BackendDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Push-Location $BackendDir
try {
    $args = @()
    if ($Force) { $args += "--force" }
    & go run ./cmd/db-init @args
    if ($LASTEXITCODE -ne 0) { throw "db-init failed" }
}
finally {
    Pop-Location
}

Write-Host "Next: run from repo root -> .\scripts\dev.ps1 or bash scripts/dev.sh"
