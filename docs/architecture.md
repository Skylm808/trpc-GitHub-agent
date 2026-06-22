# 架构说明

## 为什么不是 MVC

这个项目不是传统 Web CRUD 应用，核心流程是“用户目标 -> GitHub 查询 -> 仓库画像 -> 评分 -> 报告 -> 多轮 Agent 细化”。如果直接套 MVC，Controller 很容易堆满 GitHub API、评分、缓存和 Agent 调用逻辑。

因此当前采用轻量分层：

```text
Wails UI
  -> app.go
  -> internal/agent
  -> internal/app
  -> internal/services + internal/clients + internal/store
  -> internal/domain
```

## 为什么不是完整 DDD

DDD 适合复杂业务边界、多人长期维护、多上下文协作的系统。当前 v1 demo 更需要快速验证产品价值和开源学习价值，完整 DDD 会引入过多概念成本。

当前保留 DDD 中有用的部分：

- `domain` 保存核心数据结构。
- `services` 保存确定性领域规则，例如评分和查询规划。
- `clients` 隔离第三方 API。
- `store` 隔离持久化。

但不引入 Entity、Aggregate、Repository interface 工厂等重型结构，等项目复杂度真正上来再升级。

## tRPC-Agent-Go 的位置

tRPC-Agent-Go 是 Agent runtime，不是业务代码目录模板。本项目不会复制 `trpc-agent-go` 仓库自身的目录结构，而是在 `internal/agent` 中使用它：

- `runner.Runner`：承载一次 Agent 调用。
- `agent.Agent`：当前先实现确定性 `open_source_project_researcher`。
- `tool/function`：把搜索、评分、报告、记忆封装成 callable tools。

当前 v1 demo 没有强制依赖 LLM key。没有 LLM key 时，确定性 Agent 仍通过 tRPC-Agent-Go runner 调用工具，返回结构化结果；有 LLM key 后，可以把 `internal/agent` 中的确定性 Agent 替换或升级为 LLMAgent/GraphAgent。

## 当前调用链

```text
React UI
  -> Wails DiscoverProjects(userInput, limit)
  -> internal/agent.Runner.DiscoverProjects
  -> tRPC-Agent-Go runner.Run
  -> researchAgent.Run
  -> search_repositories tool
  -> internal/app.DiscoveryService
  -> GitHub REST API / fallback fixtures
  -> scoring + report + SQLite cache
```

## 第三方 Client 边界

GitHub 和 LLM provider 都属于第三方 client，保留在 `internal/clients`：

- 这是桌面应用，不计划把这些 client 暴露成公共 SDK。
- `internal` 能防止外部误依赖未稳定 API。
- 后续支持 OpenAI、Anthropic/Claude、DeepSeek 时，应新增 `internal/clients/llm/*`，再由 `internal/agent` 选择 provider。

## 后续演进顺序

1. 增加前端高级筛选：最小 star、最近活跃时间、语言、难度。
2. 接入 LLM provider 配置，但保留确定性 fallback。
3. 加仓库详情分析：README、目录树、Issue、PR 单仓报告。
4. 增加多轮问答 session 和用户长期 memory。
5. 封装 `open_source_project_researcher` Skill 工作流。
