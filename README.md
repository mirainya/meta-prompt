# Meta Prompt

AI 驱动的提示词自动推演系统。输入自然语言需求，经过四阶段智能流水线（分析 → 架构 → 编写 → 审核），自动生成结构化、高质量的提示词工作流。

## 核心特性

- **四阶段推演流水线**：Analyzer 理解意图 → Architect 设计蓝图 → Writer 逐组编写 → Reviewer 审核优化
- **多模型支持**：同时接入 Claude / OpenAI / Gemini，按需切换
- **可编辑元提示词**：每个阶段的系统提示词可版本化管理，支持自定义模板
- **用户系统**：注册登录、积分额度、管理后台
- **实时进度**：推演过程实时展示当前阶段，支持中途取消

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go, Gin, GORM |
| 前端 | React, TypeScript, Tailwind CSS, Vite |
| 数据库 | PostgreSQL |
| 认证 | JWT (HS256) |
| 部署 | 单二进制（前端 embed 嵌入） |

## 快速开始

### 环境要求

- Go 1.26+
- Node.js 20+
- PostgreSQL 15+

### 配置

复制并编辑配置文件：

```yaml
# config.yaml
server:
  port: 9874

database:
  host: localhost
  port: 5432
  user: meta_prompt
  password: your_password
  dbname: meta_prompt
  sslmode: disable

llm:
  claude:
    api_key: your_claude_key
    model: claude-sonnet-4-6
    max_tokens: 4096
  openai:
    api_key: your_openai_key
    model: gpt-4o
    max_tokens: 4096
  gemini:
    api_key: your_gemini_key
    model: gemini-2.5-pro
    max_tokens: 4096

defaults:
  llm_provider: claude
  temperature: 0.7

auth:
  jwt_secret: your_jwt_secret
```

### 构建运行

```bash
# 构建前端
cd web && npm install && npm run build && cd ..

# 编译并启动
go build -o meta-prompt ./cmd/server/
./meta-prompt
```

访问 `http://localhost:9874`

## 项目结构

```
├── cmd/server/          # 入口，路由注册，静态文件服务
├── internal/
│   ├── config/          # 配置加载 (Viper)
│   ├── handler/         # HTTP 处理器
│   ├── llm/             # LLM Provider 抽象层 (Claude/OpenAI/Gemini)
│   ├── middleware/       # JWT 认证、管理员鉴权
│   ├── model/           # 数据模型
│   ├── service/         # 业务逻辑（Pipeline、Auth）
│   └── store/           # 数据访问层
├── web/                 # React 前端
├── database/migrations/ # 数据库迁移文件
└── config.yaml          # 配置文件
```

## 推演流程

```
用户输入需求
    │
    ▼
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│ Analyzer │───→│ Architect│───→│  Writer  │───→│ Reviewer │
│ 需求分析  │    │ 架构设计  │    │ 提示词编写│    │ 审核优化  │
└──────────┘    └──────────┘    └──────────┘    └──────────┘
                                                     │
                                                     ▼
                                              结构化提示词工作流
```

每个阶段使用独立的**元提示词模板**驱动，模板可在管理后台编辑和版本管理。

## License

MIT
