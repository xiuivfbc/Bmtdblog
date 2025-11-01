@echo off
REM SSL证书管理工具 - Windows版本
REM 支持生成私钥和自签名证书

setlocal enabledelayedexpansion

set "DEFAULT_DOMAIN=localhost"
set "DEFAULT_KEYFILE=server.key"
set "DEFAULT_CERTFILE=server.crt"
set "DEFAULT_DAYS=365"

REM 显示帮助信息
if "%1"=="help" goto :show_help
if "%1"=="" goto :show_help

REM 解析参数
set "COMMAND=%1"
set "DOMAIN=%DEFAULT_DOMAIN%"
set "KEYFILE=%DEFAULT_KEYFILE%"
set "CERTFILE=%DEFAULT_CERTFILE%"
set "DAYS=%DEFAULT_DAYS%"

REM 解析命令行选项
:parse_args
shift
if "%1"=="" goto :execute_command
if "%1"=="-d" (
    set "DOMAIN=%2"
    shift
    goto :parse_args
)
if "%1"=="-k" (
    set "KEYFILE=%2"
    shift
    goto :parse_args
)
if "%1"=="-c" (
    set "CERTFILE=%2"
    shift
    goto :parse_args
)
shift
goto :parse_args

:execute_command
if "%COMMAND%"=="key" goto :generate_key
if "%COMMAND%"=="self" goto :generate_self_signed
if "%COMMAND%"=="config" goto :show_config
echo 错误: 未知命令 %COMMAND%
goto :show_help

:show_help
echo SSL证书管理工具 - Windows版本
echo.
echo 用法:
echo   %0 [命令] [选项]
echo.
echo 命令:
echo   key       生成RSA私钥
echo   self      生成自签名证书
echo   config    显示配置示例
echo   help      显示帮助信息
echo.
echo 选项:
echo   -d        域名 (默认: %DEFAULT_DOMAIN%)
echo   -k        私钥文件名 (默认: %DEFAULT_KEYFILE%)
echo   -c        证书文件名 (默认: %DEFAULT_CERTFILE%)
echo.
echo 示例:
echo   %0 key -k mysite.key
echo   %0 self -d example.com -k mysite.key -c mysite.crt
echo.
echo 注意: 需要安装OpenSSL工具
echo 下载地址: https://slproweb.com/products/Win32OpenSSL.html
goto :end

:generate_key
echo 生成RSA私钥...
echo 文件: %KEYFILE%
echo.

REM 检查OpenSSL是否可用
openssl version >nul 2>&1
if errorlevel 1 (
    echo 错误: 未找到OpenSSL命令
    echo 请从 https://slproweb.com/products/Win32OpenSSL.html 下载安装OpenSSL
    goto :end
)

REM 检查文件是否存在
if exist "%KEYFILE%" (
    echo 警告: 私钥文件 %KEYFILE% 已存在
    set /p "confirm=是否覆盖? (y/N): "
    if not "!confirm!"=="y" if not "!confirm!"=="Y" (
        echo 操作已取消
        goto :end
    )
)

REM 生成私钥
openssl genrsa -out "%KEYFILE%" 2048

if errorlevel 0 (
    echo 私钥生成成功: %KEYFILE%
    echo.
    echo 私钥信息:
    openssl rsa -in "%KEYFILE%" -text -noout | findstr /C:"Private-Key"
    echo.
    echo 注意事项:
    echo 1. 请妥善保管此私钥文件
    echo 2. 不要将私钥文件上传到公共代码仓库
    echo 3. 定期备份私钥文件
) else (
    echo 错误: 私钥生成失败
)
goto :end

:generate_self_signed
echo 生成自签名SSL证书...
echo 域名: %DOMAIN%
echo 私钥: %KEYFILE%
echo 证书: %CERTFILE%
echo 有效期: %DAYS% 天
echo.

REM 检查OpenSSL是否可用
openssl version >nul 2>&1
if errorlevel 1 (
    echo 错误: 未找到OpenSSL命令
    echo 请从 https://slproweb.com/products/Win32OpenSSL.html 下载安装OpenSSL
    goto :end
)

REM 检查私钥是否存在
if not exist "%KEYFILE%" (
    echo 私钥文件 %KEYFILE% 不存在，将自动生成...
    openssl genrsa -out "%KEYFILE%" 2048
    if errorlevel 1 (
        echo 错误: 私钥生成失败
        goto :end
    )
    echo 私钥生成成功: %KEYFILE%
)

REM 生成自签名证书
openssl req -new -x509 -key "%KEYFILE%" -out "%CERTFILE%" -days %DAYS% -subj "/C=CN/ST=State/L=City/O=Organization/OU=OrgUnit/CN=%DOMAIN%"

if errorlevel 0 (
    echo 证书生成成功: %CERTFILE%
    echo.
    echo 证书信息:
    openssl x509 -in "%CERTFILE%" -text -noout | findstr /C:"Subject:" /C:"Not Before" /C:"Not After"
    echo.
    echo 配置文件设置:
    echo [tls]
    echo enabled = true
    echo auto_cert = false
    echo cert_file = "%CERTFILE%"
    echo key_file = "%KEYFILE%"
    echo.
    echo 注意: 自签名证书仅适用于开发测试环境
    echo       生产环境请使用权威CA签发的证书
) else (
    echo 错误: 证书生成失败
)
goto :end

:show_config
echo # HTTPS配置示例
echo.
echo ## 1. 使用自动Let's Encrypt证书 (推荐生产环境)
echo [tls]
echo enabled = true
echo auto_cert = true
echo domain = "yourdomain.com"
echo email = "admin@yourdomain.com"
echo cert_dir = "certs"
echo.
echo ## 2. 使用手动证书
echo [tls]
echo enabled = true
echo auto_cert = false
echo cert_file = "server.crt"
echo key_file = "server.key"
echo.
echo ## 3. 禁用HTTPS (仅HTTP)
echo [tls]
echo enabled = false
echo.
echo # 注意事项:
echo # - 自动证书需要域名解析到服务器IP
echo # - 自动证书需要开放80和443端口  
echo # - 手动证书需要自行获取和更新
echo # - 开发测试可使用自签名证书
goto :end

:end
endlocal