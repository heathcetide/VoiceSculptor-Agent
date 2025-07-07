#!/bin/bash

# 设置默认的环境变量
ENV_FILE=".env"
if [ -f "$ENV_FILE" ]; then
    export $(cat $ENV_FILE | grep -v '^#' | xargs)
fi

# 检查 Go 环境
echo "检查 Go 环境..."
go version
if [ $? -ne 0 ]; then
    echo "Go 环境未安装，正在尝试安装 Go..."

    # 自动安装 Go（适用于 Linux 系统，使用官方方式）
    GO_VERSION="1.18.3"  # 设置需要安装的 Go 版本
    OS=$(uname -s)
    ARCH=$(uname -m)

    if [ "$OS" == "Linux" ]; then
        # 根据系统架构下载对应的 Go 版本
        if [ "$ARCH" == "x86_64" ]; then
            wget https://golang.org/dl/go$GO_VERSION.linux-amd64.tar.gz
            sudo tar -C /usr/local -xzf go$GO_VERSION.linux-amd64.tar.gz
            rm go$GO_VERSION.linux-amd64.tar.gz
            echo "Go 安装完成！"
        else
            echo "当前系统架构不支持自动安装 Go，请手动安装 Go。"
            exit 1
        fi
    else
        echo "当前操作系统不支持自动安装 Go，请手动安装 Go。"
        exit 1
    fi

    # 更新环境变量
    echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc
    source ~/.bashrc

    # 再次检查 Go 是否安装成功
    go version
    if [ $? -ne 0 ]; then
        echo "Go 安装失败！"
        exit 1
    fi
else
    echo "Go 环境已安装！"
fi

# 检查是否已经克隆仓库
if [ ! -d ".git" ]; then
    echo "未检测到 Git 仓库，正在克隆仓库..."
    # TODO 填写你的仓库地址
    git clone https://github.com/your-username/your-repository.git
    cd your-repository
else
    echo "已检测到 Git 仓库，拉取最新代码..."
    git pull origin main
    if [ $? -ne 0 ]; then
        echo "拉取代码失败！"
        exit 1
    fi
fi

# 运行 Go 依赖下载
echo "下载 Go 依赖..."
go mod tidy
if [ $? -ne 0 ]; then
    echo "Go 依赖下载失败！"
    exit 1
fi

# 编译 Go 项目
echo "编译项目..."
go build -o VoiceSculptor ./cmd/server/main.go
if [ $? -ne 0 ]; then
    echo "编译失败！"
    exit 1
fi

# 迁移数据库（如果有数据库迁移脚本）
echo "开始数据库迁移..."
./scripts/migrate_db.sh
if [ $? -ne 0 ]; then
    echo "数据库迁移失败！"
    exit 1
fi

# 启动应用
echo "启动应用..."
./VoiceSculptor &
if [ $? -ne 0 ]; then
    echo "应用启动失败！"
    exit 1
fi

# 输出部署成功信息
echo "部署完成！应用已启动。"
