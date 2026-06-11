# SEU OJ one-shot dev launcher (Windows PowerShell)
#
# Usage (from repo root):
#   .\scripts\dev.ps1
#   .\scripts\dev.ps1 -SetupOnly
#   .\scripts\dev.ps1 -ForceDbInit
#
# MySQL settings are read from seu-oj-backend/config/config.yaml by default.
# CLI params / DB_* env vars override config.yaml when provided.

param(
    [switch]$SetupOnly,
    [switch]$ForceDbInit,
    [switch]$SkipDockerPull,
    [string]$MySQLHost = "",
    [string]$MySQLPort = "",
    [string]$MySQLUser = "",
    [string]$MySQLPassword = "",
    [string]$Database = ""
)

$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "config-utils.ps1")

$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$BackendDir = Join-Path $RepoRoot "seu-oj-backend"
$ConfigExample = Join-Path $BackendDir "config\config.example.yaml"
$ConfigFile = Join-Path $BackendDir "config\config.yaml"
$CodeMirrorDir = Join-Path $RepoRoot "seu-oj-frontend\CodeMirror"
$CodeMirrorMarker = Join-Path $CodeMirrorDir "node_modules\codemirror\dist\index.js"

function Write-Step([string]$Message) {
    Write-Host ""
    Write-Host "==> $Message" -ForegroundColor Cyan
}

function Write-Ok([string]$Message) {
    Write-Host "    [OK] $Message" -ForegroundColor Green
}

function Write-Skip([string]$Message) {
    Write-Host "    [SKIP] $Message" -ForegroundColor DarkYellow
}

function Write-Warn([string]$Message) {
    Write-Host "    [WARN] $Message" -ForegroundColor Yellow
}

function Write-Fail([string]$Message) {
    Write-Host "    [FAIL] $Message" -ForegroundColor Red
}

function Test-CommandExists([string]$Name) {
    return [bool](Get-Command $Name -ErrorAction SilentlyContinue)
}

function Ensure-ConfigFile {
    Write-Step "Check backend config"
    if (-not (Test-Path $ConfigExample)) {
        throw "Missing config template: $ConfigExample"
    }
    if (-not (Test-Path $ConfigFile)) {
        Copy-Item $ConfigExample $ConfigFile
        Write-Ok "Created config/config.yaml from template"
        Write-Warn "Edit config/config.yaml (database, jwt_secret) before demo"
    }
    else {
        Write-Skip "config/config.yaml already exists"
    }
}

function Initialize-DatabaseSettings {
    $configPath = Get-SeuOjConfigPath -BackendDir $BackendDir
    $yamlDb = Read-DatabaseConfigFromYaml -ConfigPath $configPath
    $resolved = Resolve-DatabaseSettings -YamlConfig $yamlDb -OverrideHost $MySQLHost -OverridePort $MySQLPort -OverrideUser $MySQLUser -OverridePassword $MySQLPassword -OverrideName $Database

    $script:MySQLHost = $resolved.Host
    $script:MySQLPort = $resolved.Port
    $script:MySQLUser = $resolved.User
    $script:MySQLPassword = $resolved.Password
    $script:Database = $resolved.Name
    $script:ConfigPath = $configPath

    Write-Step "Database settings (from config.yaml unless overridden)"
    Write-Host "    host: $MySQLHost"
    Write-Host "    port: $MySQLPort"
    Write-Host "    user: $MySQLUser"
    Write-Host "    name: $Database"
    Write-Host "    source: $configPath"
}

function Invoke-DbInitTool {
    param([string[]]$Args)
    Push-Location $BackendDir
    try {
        & go run ./cmd/db-init @Args
        if ($LASTEXITCODE -ne 0) {
            throw "db-init failed with exit code $LASTEXITCODE"
        }
    }
    finally {
        Pop-Location
    }
}

function Test-DatabaseInitialized {
    Push-Location $BackendDir
    try {
        & go run ./cmd/db-init --check 2>$null | Out-Null
        return ($LASTEXITCODE -eq 0)
    }
    catch {
        return $false
    }
    finally {
        Pop-Location
    }
}

function Ensure-CodeMirrorDeps {
    Write-Step "Check CodeMirror frontend deps"
    if (-not (Test-Path (Join-Path $CodeMirrorDir "package.json"))) {
        Write-Warn "CodeMirror/package.json not found, skip npm install"
        return
    }
    if (Test-Path $CodeMirrorMarker) {
        Write-Skip "CodeMirror node_modules ready"
        return
    }
    if (-not (Test-CommandExists "npm")) {
        throw "npm is required to install CodeMirror deps"
    }
    Write-Host "    Running npm install (first run may take a while)..."
    Push-Location $CodeMirrorDir
    try {
        & npm install --no-fund --no-audit
        if ($LASTEXITCODE -ne 0) {
            throw "npm install failed"
        }
        Write-Ok "CodeMirror deps installed"
    }
    finally {
        Pop-Location
    }
}

function Ensure-Database {
    Write-Step "Check MySQL database"

    $needsInit = $ForceDbInit -or (-not (Test-DatabaseInitialized))
    if (-not $needsInit) {
        Write-Skip "Database '$Database' already initialized"
        return
    }

    if ($ForceDbInit) {
        Write-Host "    Force re-initializing database via Go db-init..."
        Invoke-DbInitTool -Args @("--force")
    }
    else {
        Write-Host "    Database not initialized, importing schema and seed via Go db-init..."
        Invoke-DbInitTool -Args @()
    }
    Write-Ok "Database initialization completed"
}

function Test-ExternalServices {
    Write-Step "Check external services"

    try {
        Invoke-DbInitTool -Args @("--ping") | Out-Null
        Write-Ok "MySQL reachable ($MySQLHost`:$MySQLPort/$Database)"
    }
    catch {
        Write-Fail ("MySQL unreachable: " + $_.Exception.Message)
        Write-Warn "Verify database settings in config/config.yaml"
    }

    $redis = Read-RedisAddrFromYaml -ConfigPath $ConfigPath
    if (Test-CommandExists "redis-cli") {
        $redisPing = & redis-cli -h $redis.Host -p $redis.Port ping 2>&1
        if ($redisPing -eq "PONG") {
            Write-Ok ("Redis reachable (" + $redis.Host + ":" + $redis.Port + ")")
        }
        else {
            Write-Warn "Redis did not respond PONG; submit enqueue may fail"
        }
    }
    else {
        $redisOpen = Test-NetConnection -ComputerName $redis.Host -Port $redis.Port -WarningAction SilentlyContinue -ErrorAction SilentlyContinue
        if ($redisOpen.TcpTestSucceeded) {
            Write-Ok ("Redis port open (" + $redis.Host + ":" + $redis.Port + ")")
        }
        else {
            Write-Warn ("Cannot verify Redis at " + $redis.Host + ":" + $redis.Port)
        }
    }

    if (Test-CommandExists "docker") {
        $prevEap = $ErrorActionPreference
        $ErrorActionPreference = "SilentlyContinue"
        & docker info 1>$null 2>$null
        $dockerRunning = ($LASTEXITCODE -eq 0)
        $ErrorActionPreference = $prevEap
        if ($dockerRunning) {
            Write-Ok "Docker available"
            if (-not $SkipDockerPull) {
                $ErrorActionPreference = "SilentlyContinue"
                & docker image inspect gcc:13 1>$null 2>$null
                $hasImage = ($LASTEXITCODE -eq 0)
                $ErrorActionPreference = $prevEap
                if (-not $hasImage) {
                    Write-Host "    Pulling judge image gcc:13..."
                    $ErrorActionPreference = "SilentlyContinue"
                    & docker pull gcc:13 1>$null
                    $pullOk = ($LASTEXITCODE -eq 0)
                    $ErrorActionPreference = $prevEap
                    if ($pullOk) {
                        Write-Ok "gcc:13 image ready"
                    }
                    else {
                        Write-Warn "Failed to pull gcc:13; Run/Submit may not work"
                    }
                }
                else {
                    Write-Skip "gcc:13 image already exists"
                }
            }
        }
        else {
            Write-Warn "Docker is not running; Run/Submit will not work"
        }
    }
    else {
        Write-Warn "docker command not found; Run/Submit will not work"
    }

    if (-not (Test-CommandExists "go")) {
        throw "go command not found"
    }
    Write-Ok ("Go installed: " + (go version))
}

function Start-DevProcesses {
    Write-Step "Start dev services (opening two new terminal windows)"

    $webCmd = '$Host.UI.RawUI.WindowTitle = ''SEU OJ Web :8080''; Set-Location ''' + $BackendDir + '''; Write-Host ''SEU OJ Web - http://127.0.0.1:8080/'' -ForegroundColor Green; go run .'
    $workerCmd = '$Host.UI.RawUI.WindowTitle = ''SEU OJ Judge Worker''; Set-Location ''' + $BackendDir + '''; Write-Host ''SEU OJ Judge Worker'' -ForegroundColor Green; go run ./cmd/judge-worker'

    Start-Process powershell -ArgumentList @("-NoExit", "-Command", $webCmd) | Out-Null
    Start-Sleep -Milliseconds 800
    Start-Process powershell -ArgumentList @("-NoExit", "-Command", $workerCmd) | Out-Null

    Write-Ok "Web server window opened (port 8080)"
    Write-Ok "Judge worker window opened"
    Write-Host ""
    Write-Host "Open in browser: http://127.0.0.1:8080/" -ForegroundColor Green
    Write-Host "Demo accounts: docs/demo-checklist.md (password 123456)"
    Write-Host "Stop: Ctrl+C in each service window, or run .\scripts\stop-dev.ps1"
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Magenta
Write-Host "  SEU OJ dev.ps1" -ForegroundColor Magenta
Write-Host "========================================" -ForegroundColor Magenta

Ensure-ConfigFile
Initialize-DatabaseSettings
Test-ExternalServices
Ensure-Database
Ensure-CodeMirrorDeps

if ($SetupOnly) {
    Write-Step "SetupOnly mode"
    Write-Ok "Setup finished. Run .\scripts\dev.ps1 again to start services."
    exit 0
}

Start-DevProcesses
