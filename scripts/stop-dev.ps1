# 停止 SEU OJ 开发服务（按窗口标题匹配 go 进程）
#
# 用法：.\scripts\stop-dev.ps1
#
# 说明：dev.ps1 会在标题含 "SEU OJ Web" / "SEU OJ Judge Worker" 的窗口中运行 go。
#       本脚本尝试结束对应 PowerShell 窗口及其子进程。

$ErrorActionPreference = "SilentlyContinue"

function Stop-ProcessesByWindowTitle([string]$TitlePart) {
    $procs = Get-CimInstance Win32_Process |
        Where-Object { $_.Name -match '^powershell(\.exe)?$' -and $_.CommandLine -and $_.CommandLine -like "*$TitlePart*" }

    foreach ($proc in $procs) {
        Write-Host "Stopping PID $($proc.ProcessId) ($TitlePart)..."
        Stop-Process -Id $proc.ProcessId -Force -ErrorAction SilentlyContinue
    }
}

Write-Host "Stopping SEU OJ dev services..."
Stop-ProcessesByWindowTitle "SEU OJ Web"
Stop-ProcessesByWindowTitle "SEU OJ Judge Worker"

# 兜底：释放 8080 端口（若仍有监听）
try {
    $conn = Get-NetTCPConnection -LocalPort 8080 -State Listen -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($conn) {
        $ownerPid = $conn.OwningProcess
        Write-Host "Stopping process on port 8080 (PID $ownerPid)..."
        Stop-Process -Id $ownerPid -Force -ErrorAction SilentlyContinue
    }
} catch {}

Write-Host "Done."
