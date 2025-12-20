# Bmtdblog 日志清理脚本 - 静默版本
# 直接删除 slog 目录下的所有 .log 文件，无需确认

param(
    [switch]$Force = $false
)

# 获取脚本所在目录
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$logDir = Join-Path $scriptDir "slog"

Write-Host "开始清理日志文件..." -ForegroundColor Green

# 检查目录是否存在
if (-not (Test-Path $logDir)) {
    Write-Host "错误: slog 目录不存在!" -ForegroundColor Red
    exit 1
}

# 获取所有 .log 文件
$logFiles = Get-ChildItem -Path $logDir -Filter "*.log" -File

if ($logFiles.Count -eq 0) {
    Write-Host "没有找到日志文件。" -ForegroundColor Yellow
} else {
    Write-Host "找到 $($logFiles.Count) 个日志文件，正在删除..." -ForegroundColor Cyan
    
    try {
        # 直接删除所有 .log 文件
        Remove-Item -Path (Join-Path $logDir "*.log") -Force
        Write-Host "成功删除 $($logFiles.Count) 个日志文件!" -ForegroundColor Green
    }
    catch {
        Write-Host "删除文件时出错: $($_.Exception.Message)" -ForegroundColor Red
        exit 1
    }
}

Write-Host "日志清理完成!" -ForegroundColor Green