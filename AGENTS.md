# trpc-GitHub-agent 开发约定

## 项目定位

这是一个基于 Wails + React + Go + tRPC-Agent-Go 的 GitHub 开源项目研究桌面 Agent。目标用户是准备秋招、学习开源项目、寻找贡献入口的个人开发者。

## 代码组织

- `app.go` 只保留 Wails 生命周期和前端可调用方法转发。
- `internal/agent` 封装 tRPC-Agent-Go Runner、Agent 和 Tools。
- `internal/app` 编排项目发现、画像、评分、报告和缓存。
- `internal/clients` 放第三方 API client，例如 GitHub 和后续 LLM provider。
- `internal/services` 放确定性业务服务，例如 query planner、scoring、report。
- `internal/store/sqlite` 放 SQLite schema、迁移和缓存读写。
- `internal/domain` 放跨层共享的数据结构。

## 注释规范

- 导出的类型、构造函数、Wails 方法、Agent tools 必须写中文注释。
- 复杂私有函数需要在关键分支前写简短中文意图注释。
- 不写机械注释，例如“遍历数组”“返回结果”。
- 注释解释业务目的、边界、权衡，不重复代码字面含义。

## 设计边界

- v1 使用 SQLite 作为主存储，JSON 只用于报告导出或调试快照。
- v1 不引入向量数据库；v2 可考虑 sqlite-vec。
- GitHub Token 指 GitHub PAT，只用于提高公开 API rate limit。
- LLM Token 指 OpenAI、Anthropic/Claude、DeepSeek 等 provider API key，由用户自己配置。
- 评分必须保持确定性规则，不把排序完全交给 LLM。
- tRPC-Agent-Go 负责 Agent orchestration、tools、runner、session/memory 演进；GitHub、评分、报告等业务逻辑不塞进框架层。

## 验证要求

修改 Go 后端后至少运行：

```bash
go test ./...
```

修改前端或 Wails 绑定后补充：

```bash
npm run build
/Users/skylm/go/bin/wails build
```
