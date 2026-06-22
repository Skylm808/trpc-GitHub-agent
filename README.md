# trpc-GitHub-agent

基于 Wails + React + Go + tRPC-Agent-Go 的 GitHub 开源项目研究桌面 Agent。它面向秋招、开源学习和贡献决策：用户输入技术栈、方向和目标后，应用会生成 GitHub Search query，拉取公开仓库，按确定性规则评分，并导出 Markdown 推荐报告。

## 当前能力

- 自然语言目标转 GitHub Search query。
- GitHub REST API 搜索公开仓库，失败时使用本地 deterministic fixtures。
- SQLite 缓存仓库、评分、报告和查询历史。
- 100 分评分：活跃度、受欢迎度、学习价值、贡献友好度、岗位相关度。
- 区分项目影响力等级和新手友好度。
- 通过 tRPC-Agent-Go runner + function tools 执行项目发现流程。
- React + Ant Design 前端展示项目卡片、评分维度、查询语句和 Markdown 报告。

## Token 说明

- `GITHUB_TOKEN`：GitHub PAT，用于提高公开 REST API rate limit。v1 不做 OAuth，不分析私有仓库。
- `OPENAI_API_KEY`、`ANTHROPIC_API_KEY`、`DEEPSEEK_API_KEY`：后续 LLM provider key，由用户自己提供。当前 demo 只检测配置状态，发现流程仍可无 LLM 运行。

## 架构

```text
React UI
  -> Wails app.go
  -> internal/agent
  -> internal/app
  -> internal/services + internal/clients + internal/store
  -> internal/domain
```

更多说明见 [docs/architecture.md](docs/architecture.md)。

## 本地开发

当前 `go.mod` 使用本地 replace：

```text
replace trpc.group/trpc-go/trpc-agent-go => ../trpc-agent-go
```

因此需要把 `trpc-agent-go` 和本仓库放在同一父目录下。

安装前端依赖后运行：

```bash
cd frontend
npm install
cd ..
/Users/skylm/go/bin/wails dev
```

## 验证

```bash
go test ./...
cd frontend && npm run build
cd ..
/Users/skylm/go/bin/wails build
```
