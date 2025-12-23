@echo off
REM Redis邮件队列启动脚本 (Windows版本)
REM 使用优化的持久化配置启动Redis

set REDIS_CONF=..\conf\redis-email-queue.conf
set REDIS_LOG=..\logs\redis-email-queue.log

REM 创建日志目录
if not exist "logs" mkdir logs

echo 🚀 启动Redis邮件队列服务...
echo 配置文件: %REDIS_CONF%
echo 日志文件: %REDIS_LOG%

REM 检查Redis是否已安装
redis-server --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Redis未安装或不在PATH中
    echo 💡 请下载Redis for Windows: https://github.com/tporadowski/redis/releases
    pause
    exit /b 1
)

REM 启动Redis
echo 正在启动Redis服务器...
start /B redis-server %REDIS_CONF%

REM 等待启动
timeout /t 3 /nobreak >nul

REM 检查Redis是否启动成功
redis-cli ping >nul 2>&1
if errorlevel 1 (
    echo ❌ Redis启动失败或连接失败
    pause
    exit /b 1
)

echo ✅ Redis启动成功！
echo 📊 检查持久化配置...

REM 检查AOF配置
for /f "tokens=2" %%i in ('redis-cli CONFIG GET appendonly') do set AOF_STATUS=%%i
echo    AOF持久化: %AOF_STATUS%

if "%AOF_STATUS%"=="yes" (
    for /f "tokens=2" %%i in ('redis-cli CONFIG GET appendfsync') do set FSYNC_STATUS=%%i
    echo    同步策略: %FSYNC_STATUS%
)

REM 检查混合持久化
for /f "tokens=2" %%i in ('redis-cli CONFIG GET aof-use-rdb-preamble') do set HYBRID_STATUS=%%i
echo    混合持久化: %HYBRID_STATUS%

echo.
echo 🎯 Redis邮件队列服务已启动，数据持久化已配置！
echo 💡 使用 'redis-cli shutdown' 停止服务
echo 💡 运行 'go run test_redis_persistence.go' 验证配置
echo.
pause