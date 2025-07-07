@echo off
REM 默认运行环境
set MODE=development

REM 解析命令行参数
:parse_args
if "%1"=="" goto end_args
if "%1"=="-mode" (
    set MODE=%2
    shift
    shift
    goto parse_args
)
echo 未知参数: %1
exit /b 1
:end_args

REM 设置环境变量并启动应用
set APP_ENV=%MODE%
go run cmd/server/main.go -mode=%MODE% 