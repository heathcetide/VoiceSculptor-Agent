#!/bin/bash

# 默认运行环境
MODE="development"

# 解析命令行参数
while [[ $# -gt 0 ]]; do
  case "$1" in
    -mode)
      MODE="$2"
      shift 2
      ;;
    *)
      echo "未知参数: $1"
      exit 1
      ;;
  esac
done

# 设置环境变量并启动应用
export APP_ENV=$MODE
go run cmd/server/main.go -mode=$MODE 