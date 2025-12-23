# Bmtdblog 日志清理脚本
# 清理项目根目录下 slog 目录中的日志文件

param(
    [int]$KeepDays = 0,                # 保留最近多少天的日志
    [switch]$Force = $false,           # 是否强制删除
    [switch]$Compress = $false,        # 是否压缩后删除（暂未实现）
    [switch]$Silent = $false           # 静默模式，不输出详细信息
)

# 获取脚本所在目录
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
# 修复路径：slog目录在项目根目录下，而不是deploy目录下
$logDir = Join-Path $scriptDir ".." "slog"

# 确保路径是绝对路径
$logDir = Resolve-Path $logDir

if (-not $Silent) {
    Write-Host "开始清理日志文件..." -ForegroundColor Green
    Write-Host "日志目录: $logDir" -ForegroundColor Cyan
}

# 检查目录是否存在
if (-not (Test-Path $logDir)) {
    if (-not $Silent) {
        Write-Host "错误: slog 目录不存在!" -ForegroundColor Red
    }
    exit 1
}

# 获取所有 .log 文件
$logFiles = Get-ChildItem -Path $logDir -Filter "*.log" -File

if ($logFiles.Count -eq 0) {
    if (-not $Silent) {
        Write-Host "没有找到日志文件。" -ForegroundColor Yellow
    }
} else {
    # 计算需要保留的日期
    $cutoffDate = (Get-Date).AddDays(-$KeepDays)
    
    # 筛选出需要删除的旧日志
    $oldLogFiles = $logFiles | Where-Object {$_.LastWriteTime -lt $cutoffDate}
    
    if ($oldLogFiles.Count -eq 0) {
        if (-not $Silent) {
            Write-Host "没有需要删除的旧日志文件（已保留最近 $KeepDays 天的日志）。" -ForegroundColor Yellow
        }
    } else {
        if (-not $Silent) {
            Write-Host "找到 $($oldLogFiles.Count) 个旧日志文件，正在删除..." -ForegroundColor Cyan
            Write-Host "保留最近 $KeepDays 天的日志文件。" -ForegroundColor Green
        }
        
        try {
            # 删除旧日志文件
            $oldLogFiles | Remove-Item -Force:$Force
            if (-not $Silent) {
                Write-Host "成功删除 $($oldLogFiles.Count) 个日志文件!" -ForegroundColor Green
            }
        }
        catch {
            if (-not $Silent) {
                Write-Host "删除文件时出错: $($_.Exception.Message)" -ForegroundColor Red
            }
            exit 1
        }
    }
}

if (-not $Silent) {
    Write-Host "日志清理完成!" -ForegroundColor Green
}