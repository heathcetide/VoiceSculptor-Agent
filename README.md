# Go 服务端项目

VoiceSculptor 是一款创新的客户服务和语音交互平台，旨在为企业提供高度灵活和定制化的语音客服解决方案。它结合了最新的技术和深度学习，赋能用户构建自己的虚拟客服角色，实现更智能、个性化的客户服务体验。

# 使用开发环境
./scripts/start.sh -mode=development

# 使用测试环境
./scripts/start.sh -mode=test

# 使用生产环境
./scripts/start.sh -mode=production


# 使用开发环境
scripts\start.bat -mode development

# 使用测试环境
scripts\start.bat -mode test

# 使用生产环境
scripts\start.bat -mode production
---

## 📦 技术栈

- **Go**: 主要编程语言
- **Gin/Gorilla Mux/Chi**（选择其一）: Web 路由框架
- **GORM**: ORM 库，用于数据库操作
- **PostgreSQL/MySQL/MongoDB**（根据项目需要选择）: 数据库支持
- **Redis**: 缓存服务
- **JWT/OAuth2**: 认证与授权机制
- **Logrus/Zap**: 日志记录
- **Viper**: 配置管理
- **Docker**: 容器化部署
- **Makefile**: 构建和运行命令简化
- **Unit Tests**: 使用 Go testing 包进行单元测试

---

## 🧰 功能模块

- 用户认证 (登录、注册、JWT)
- 用户管理
- API 接口文档 (Swagger 或者 Gin Swagger UI)
- 数据持久化 (数据库 CRUD)
- Redis 缓存优化
- 错误处理及中间件
- 环境配置管理
- 单元测试覆盖率报告

---

## 📁 项目结构
```
/VoiceSculptor                # 项目根目录
│
├── cmd/                      # 启动文件
│   └── server/               # 启动服务的具体实现
│       └── main.go           # 主入口文件，启动整个应用
│   └── worker/               # 启动服务的具体实现
│       └── main.go           # 主入口文件，启动整个应用
│
├── configs/                  # 配置文件
│   ├── config.go             # 配置管理（如读取配置、环境变量）
│   └── example_config.yaml  # 配置样例文件
│
├── internal/                 # 内部逻辑模块（私有，不对外暴露）
│   ├── constants/            # 常量定义
│   ├── handler/              # HTTP 路由和请求处理
│   │   └── webhook.go        # 处理 Webhook 请求
│   │   └── customer.go       # 处理客服相关业务
│   │   └── api.go            # 处理 API 接口业务
│   │   └── auth.go           # 处理认证和授权
│   │
│   ├── service/              # 核心业务服务逻辑
│   │   └── call.go           # 处理呼叫相关的逻辑
│   │   └── customer_service.go # 客服角色逻辑
│   │   └── client.go         # 与客户端（如 RustPBX）交互
│   │
│   ├── model/                # 数据结构和模型定义
│   │   └── call.go           # 呼叫模型
│   │   └── customer.go       # 客服角色模型
│   │
│   ├── util/                 # 工具类目录
│   │   ├── signals.go        # 处理信号的工具类
│   │   ├── base.go           # 基础工具类
│   │   ├── caches.go         # 缓存工具类
│   │   ├── configs.go        # 配置管理工具类
│   │   └── json.go           # JSON 处理工具类
│   │
│   ├── middleware/           # 中间件
│   │   └── auth_middleware.go # 认证中间件
│   │   └── logging_middleware.go # 日志中间件
│   │
│   └── repository/           # 数据存取层
│       └── call_repository.go # 呼叫数据存取
│       └── customer_repository.go # 客服角色数据存取
│
├── pkg/                      # 公共的库和第三方依赖
│   ├── client/               # 客户端相关
│   │   └── rustpbx_client.go # RustPBX 客户端
│   └── api/                  # API 公共模块
│       └── response.go       # API 响应格式
│
├── scripts/                  # 脚本文件（数据库迁移、自动化部署等）
│   └── init_db.sh            # 数据库初始化脚本
│   └── deploy.sh             # 部署脚本
│   └── migrate_db.sh         # 数据库迁移脚本
│
├── .gitignore                # Git 忽略文件
├── README.md                 # 项目说明文档
├── go.mod                    # Go 模块管理文件
└── go.sum                    # Go 依赖文件
```

---

## 🔧 开发环境搭建

### 前提条件

- [Go 1.20+](https://golang.org/dl/)
- [Docker](https://www.docker.com/)
- [Make](https://www.gnu.org/software/make/)
- 数据库（如 PostgreSQL）

### 安装依赖
```bash
go mod tidy
```

### 启动服务
```bash
make run
```
或者使用 Docker：
```bash
docker-compose up
```

---

## 🚀 生产部署

### 构建二进制文件

```bash 
make build
```

### 使用 Docker 部署

```bash
docker build -t your-project-name . docker run -p 8080:8080 your-project-name
```

---

## 📝 API 文档

API 文档可通过访问 `/swagger/index.html` 查看（需集成 Swagger 并生成文档）。

---

## 🧪 单元测试

运行所有单元测试：


---

## 📊 监控 & 日志

- 日志输出路径：`logs/app.log`
- Prometheus 集成支持（可选）
- Grafana 可视化监控面板（可选）

---

## 📈 版本历史

- v1.0.0 - 初始版本，基础功能完成
- v1.1.0 - 新增用户权限模块
- v1.2.0 - 集成 Redis 缓存提升性能
- v2.0.0 - 重构代码结构，引入接口抽象层

---

## 📬 联系方式

如有问题或建议，请联系：
- 邮箱: example@example.com
- GitHub: [https://github.com/yourusername/yourprojectname](https://github.com/yourusername/yourprojectname)

---

## 📜 License

MIT License
