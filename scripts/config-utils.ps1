# Shared helpers for reading seu-oj-backend/config/config.yaml in PowerShell scripts.

function Get-SeuOjBackendDir {
    param([string]$StartDir = $PSScriptRoot)
    $candidate = Resolve-Path (Join-Path $StartDir "..\seu-oj-backend") -ErrorAction SilentlyContinue
    if ($candidate -and (Test-Path (Join-Path $candidate "config"))) {
        return $candidate.Path
    }
    $candidate = Resolve-Path (Join-Path $StartDir "..") -ErrorAction SilentlyContinue
    if ($candidate -and (Test-Path (Join-Path $candidate "config"))) {
        return $candidate.Path
    }
    throw "Cannot locate seu-oj-backend directory from $StartDir"
}

function Get-SeuOjConfigPath {
    param([string]$BackendDir)
    $configFile = Join-Path $BackendDir "config\config.yaml"
    if (Test-Path $configFile) {
        return $configFile
    }
    $exampleFile = Join-Path $BackendDir "config\config.example.yaml"
    if (Test-Path $exampleFile) {
        return $exampleFile
    }
    throw "Missing config/config.yaml and config.example.yaml under $BackendDir"
}

function Get-YamlSectionBlock {
    param(
        [string]$Content,
        [string]$Section
    )
    $pattern = "(?ms)^$Section`:\s*\r?\n(.*?)(?=^[A-Za-z0-9_]+:\s*(?:`"|'|\S)|\z)"
    if ($Content -match $pattern) {
        return $matches[1]
    }
    return $null
}

function Get-YamlKeyValue {
    param(
        [string]$SectionBlock,
        [string]$Key
    )
    if (-not $SectionBlock) {
        return $null
    }
    if ($SectionBlock -match "(?m)^\s*$Key\s*:\s*""([^""]*)""\s*$") {
        return $matches[1]
    }
    if ($SectionBlock -match "(?m)^\s*$Key\s*:\s*'([^']*)'\s*$") {
        return $matches[1]
    }
    if ($SectionBlock -match "(?m)^\s*$Key\s*:\s*(\S+)\s*$") {
        return $matches[1]
    }
    return $null
}

function Read-DatabaseConfigFromYaml {
    param([string]$ConfigPath)

    $result = @{
        Host     = "127.0.0.1"
        Port     = "3306"
        User     = "root"
        Password = ""
        Name     = "seu_oj"
    }

    if (-not (Test-Path $ConfigPath)) {
        return $result
    }

    $content = Get-Content -Path $ConfigPath -Raw -Encoding UTF8
    $dbBlock = Get-YamlSectionBlock -Content $content -Section "database"
    if (-not $dbBlock) {
        return $result
    }

    foreach ($key in @("host", "port", "user", "password", "name")) {
        $value = Get-YamlKeyValue -SectionBlock $dbBlock -Key $key
        if ($null -ne $value -and $value -ne "") {
            switch ($key) {
                "host" { $result.Host = $value }
                "port" { $result.Port = $value }
                "user" { $result.User = $value }
                "password" { $result.Password = $value }
                "name" { $result.Name = $value }
            }
        }
    }

    return $result
}

function Read-RedisAddrFromYaml {
    param([string]$ConfigPath)

    $hostName = "127.0.0.1"
    $port = 6379
    if (-not (Test-Path $ConfigPath)) {
        return @{ Host = $hostName; Port = $port }
    }

    $content = Get-Content -Path $ConfigPath -Raw -Encoding UTF8
    $redisBlock = Get-YamlSectionBlock -Content $content -Section "redis"
    $addr = Get-YamlKeyValue -SectionBlock $redisBlock -Key "addr"
    if ($addr -match "^(.+):(\d+)$") {
        $hostName = $matches[1]
        $port = [int]$matches[2]
    }
    return @{ Host = $hostName; Port = $port }
}

function Resolve-DatabaseSettings {
    param(
        [hashtable]$YamlConfig,
        [string]$OverrideHost = "",
        [string]$OverridePort = "",
        [string]$OverrideUser = "",
        [string]$OverridePassword = "",
        [string]$OverrideName = ""
    )

    function Pick-Value {
        param([string]$Override, [string]$EnvName, [string]$YamlValue, [string]$Default)
        if ($Override -ne "") { return $Override }
        $envValue = [Environment]::GetEnvironmentVariable($EnvName)
        if ($envValue) { return $envValue }
        if ($YamlValue -ne "") { return $YamlValue }
        return $Default
    }

    return @{
        Host     = Pick-Value $OverrideHost "DB_HOST" $YamlConfig.Host "127.0.0.1"
        Port     = Pick-Value $OverridePort "DB_PORT" $YamlConfig.Port "3306"
        User     = Pick-Value $OverrideUser "DB_USER" $YamlConfig.User "root"
        Password = if ($OverridePassword -ne "") { $OverridePassword } elseif ([Environment]::GetEnvironmentVariable("DB_PASSWORD")) { [Environment]::GetEnvironmentVariable("DB_PASSWORD") } else { $YamlConfig.Password }
        Name     = Pick-Value $OverrideName "DB_NAME" $YamlConfig.Name "seu_oj"
    }
}
