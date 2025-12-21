#!/bin/bash

# Bmtdblog 日志清理脚本 (Bash版本)
# 清理项目根目录下 slog 目录中的日志文件

# 默认可配置参数
KEEP_DAYS=0
SILENT=false
FORCE=false

# 解析命令行参数
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -d|--keep-days) KEEP_DAYS="$2"; shift ;;
        -s|--silent) SILENT=true ;;
        -f|--force) FORCE=true ;;
        *) echo "未知参数: $1"; exit 1 ;;
    esac
    shift
done

# 获取脚本所在目录并修正路径
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="$SCRIPT_DIR/../../../slog"

if [ "$SILENT" = false ]; then
    echo -e "\e[32m开始清理日志文件...\e[0m"
    echo -e "\e[36m日志目录: $LOG_DIR\e[0m"
fi

# 检查目录是否存在
if [ ! -d "$LOG_DIR" ]; then
    if [ "$SILENT" = false ]; then
        echo -e "\e[31m错误: slog 目录不存在!\e[0m"
    fi
    exit 1
fi

# 获取所有.log文件
LOG_FILES=("$LOG_DIR"/*.log)

# 检查是否有日志文件
if [ ! -f "${LOG_FILES[0]}" ]; then
    if [ "$SILENT" = false ]; then
        echo -e "\e[33m没有找到日志文件。\e[0m"
    fi
else
    # 计算需要保留的日期
    CUTOFF_DATE=$(date -d "$KEEP_DAYS days ago" +"%Y-%m-%d %H:%M:%S")
    
    # 统计旧日志文件数量
    OLD_LOG_COUNT=0
    for file in "${LOG_FILES[@]}"; do
        if [ -f "$file" ]; then
            FILE_DATE=$(stat -c "%y" "$file" | cut -d. -f1)
            if [[ "$FILE_DATE" < "$CUTOFF_DATE" ]]; then
                ((OLD_LOG_COUNT++))
            fi
        fi
    done
    
    if [ "$OLD_LOG_COUNT" -eq 0 ]; then
        if [ "$SILENT" = false ]; then
            echo -e "\e[33m没有需要删除的旧日志文件（已保留最近 $KEEP_DAYS 天的日志）。\e[0m"
        fi
    else
        if [ "$SILENT" = false ]; then
            echo -e "\e[36m找到 $OLD_LOG_COUNT 个旧日志文件，正在删除...\e[0m"
            echo -e "\e[32m保留最近 $KEEP_DAYS 天的日志文件。\e[0m"
        fi
        
        # 删除旧日志文件
        DELETE_COUNT=0
        for file in "${LOG_FILES[@]}"; do
            if [ -f "$file" ]; then
                FILE_DATE=$(stat -c "%y" "$file" | cut -d. -f1)
                if [[ "$FILE_DATE" < "$CUTOFF_DATE" ]]; then
                    rm "$file" 2>/dev/null
                    if [ $? -eq 0 ]; then
                        ((DELETE_COUNT++))
                    fi
                fi
            fi
        done
        
        if [ "$SILENT" = false ]; then
            echo -e "\e[32m成功删除 $DELETE_COUNT 个日志文件!\e[0m"
        fi
    fi
fi

if [ "$SILENT" = false ]; then
    echo -e "\e[32m日志清理完成!\e[0m"
fi