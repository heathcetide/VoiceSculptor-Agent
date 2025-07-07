#!/bin/bash

# 确保数据库连接已配置
if [ -z "$DB_DRIVER" ] || [ -z "$DB_DSN" ]; then
    echo "数据库连接信息未配置，请检查环境变量 DB_DRIVER 和 DB_DSN。"
    exit 1
fi

echo "正在初始化数据库..."

# 检查数据库类型（支持 PostgreSQL, MySQL 等）
if [ "$DB_DRIVER" == "postgres" ]; then
    echo "使用 PostgreSQL 数据库初始化..."

    # 连接并创建数据库（如果不存在）
    psql $DB_DSN -c "CREATE DATABASE IF NOT EXISTS my_database;" # 如果是 PostgreSQL
    if [ $? -ne 0 ]; then
        echo "数据库创建失败！"
        exit 1
    fi

elif [ "$DB_DRIVER" == "mysql" ]; then
    echo "使用 MySQL 数据库初始化..."

    # 连接并创建数据库（如果不存在）
    mysql $DB_DSN -e "CREATE DATABASE IF NOT EXISTS my_database;"
    if [ $? -ne 0 ]; then
        echo "数据库创建失败！"
        exit 1
    fi

else
    echo "不支持的数据库类型：$DB_DRIVER"
    exit 1
fi

# 执行数据库迁移
echo "开始执行数据库迁移..."
./scripts/migrate_db.sh
if [ $? -ne 0 ]; then
    echo "数据库迁移失败！"
    exit 1
fi

# 填充初始数据（如果有）
echo "填充初始数据..."
./scripts/seed_db.sh
if [ $? -ne 0 ]; then
    echo "初始数据填充失败！"
    exit 1
fi

echo "数据库初始化完成！"
