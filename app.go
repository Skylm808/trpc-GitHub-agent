package main

import (
	"context"

	researchagent "trpc-GitHub-agent/internal/agent"
	appsvc "trpc-GitHub-agent/internal/app"
	gh "trpc-GitHub-agent/internal/clients/github"
	"trpc-GitHub-agent/internal/config"
	"trpc-GitHub-agent/internal/domain"
	"trpc-GitHub-agent/internal/store/sqlite"
)

type App struct {
	ctx       context.Context
	store     *sqlite.SQLiteStore
	discovery *appsvc.DiscoveryService
	agent     *researchagent.Runner
	storePath string
}

// NewApp 创建 Wails 根对象，并初始化可在无 SQLite 时运行的发现服务。
func NewApp() *App {
	discovery := appsvc.NewDiscoveryService(gh.NewClient(), nil)
	return &App{discovery: discovery, agent: researchagent.NewRunner(discovery, nil)}
}

// startup 在 Wails 启动时打开本地 SQLite 缓存。
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	sqliteStore, err := sqlite.Open("")
	if err == nil {
		a.store = sqliteStore
		a.storePath, _ = sqlite.DefaultPath()
		a.discovery.SetStore(sqliteStore)
		a.agent = researchagent.NewRunner(a.discovery, sqliteStore)
	}
}

// shutdown 关闭本地 SQLite 连接。
func (a *App) shutdown(ctx context.Context) {
	if a.store != nil {
		_ = a.store.Close()
	}
}

// DiscoverProjects 暴露给前端：根据用户背景和目标发现、评分并生成项目报告。
func (a *App) DiscoverProjects(userInput string, limit int) (domain.DiscoveryResult, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.agent.DiscoverProjects(ctx, userInput, limit)
}

// StorePath 返回当前 SQLite 缓存文件路径，便于用户定位本地缓存。
func (a *App) StorePath() string {
	return a.storePath
}

// SettingsStatus 返回 GitHub 与 LLM provider key 的本地配置状态。
func (a *App) SettingsStatus() config.SettingsStatus {
	return config.LoadSettingsStatus()
}
