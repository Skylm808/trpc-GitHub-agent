package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	appsvc "trpc-GitHub-agent/internal/app"
	"trpc-GitHub-agent/internal/domain"
	"trpc-GitHub-agent/internal/store/sqlite"

	coreagent "trpc.group/trpc-go/trpc-agent-go/agent"
	"trpc.group/trpc-go/trpc-agent-go/event"
	"trpc.group/trpc-go/trpc-agent-go/model"
	trpcrunner "trpc.group/trpc-go/trpc-agent-go/runner"
	"trpc.group/trpc-go/trpc-agent-go/tool"
)

const (
	researchAgentName = "open_source_project_researcher"
	appName           = "trpc-github-agent"
)

// Runner 通过 tRPC-Agent-Go runner 执行项目研究工作流。
type Runner struct {
	framework trpcrunner.Runner
}

// NewRunner 创建一个确定性 Agent runner；后续可在这里切换到 LLM Agent。
func NewRunner(discovery *appsvc.DiscoveryService, store *sqlite.SQLiteStore) *Runner {
	toolset := NewToolset(discovery, store)
	agent := &researchAgent{
		name:  researchAgentName,
		tools: toolset.Tools(),
	}
	return &Runner{
		framework: trpcrunner.NewRunner(appName, agent),
	}
}

// DiscoverProjects 通过框架 runner 执行一次项目发现，并反序列化 Agent 返回的结构化结果。
func (r *Runner) DiscoverProjects(ctx context.Context, userInput string, limit int) (domain.DiscoveryResult, error) {
	if r == nil || r.framework == nil {
		return domain.DiscoveryResult{}, fmt.Errorf("agent runner is not configured")
	}
	payload, err := json.Marshal(SearchRepositoriesInput{UserInput: userInput, Limit: limit})
	if err != nil {
		return domain.DiscoveryResult{}, err
	}
	events, err := r.framework.Run(ctx, "local-user", "default-session", model.NewUserMessage(string(payload)))
	if err != nil {
		return domain.DiscoveryResult{}, err
	}

	var content string
	for evt := range events {
		if evt == nil || evt.Response == nil {
			continue
		}
		if evt.Error != nil {
			return domain.DiscoveryResult{}, fmt.Errorf("agent run failed: %s", evt.Error.Message)
		}
		if len(evt.Choices) > 0 {
			content = evt.Choices[0].Message.Content
		}
	}
	if strings.TrimSpace(content) == "" {
		return domain.DiscoveryResult{}, fmt.Errorf("agent returned an empty result")
	}
	var result domain.DiscoveryResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return domain.DiscoveryResult{}, fmt.Errorf("decode agent discovery result: %w", err)
	}
	return result, nil
}

type researchAgent struct {
	name  string
	tools []tool.Tool
}

func (a *researchAgent) Info() coreagent.Info {
	return coreagent.Info{
		Name:        a.name,
		Description: "GitHub open-source project researcher for recruiting and learning workflows.",
	}
}

func (a *researchAgent) Tools() []tool.Tool {
	return a.tools
}

func (a *researchAgent) SubAgents() []coreagent.Agent {
	return nil
}

func (a *researchAgent) FindSubAgent(name string) coreagent.Agent {
	return nil
}

func (a *researchAgent) Run(ctx context.Context, invocation *coreagent.Invocation) (<-chan *event.Event, error) {
	ch := make(chan *event.Event, 1)
	go func() {
		defer close(ch)
		if invocation == nil {
			ch <- event.NewErrorEvent("", a.name, model.ErrorTypeFlowError, "nil invocation")
			return
		}
		result, err := a.callSearchTool(ctx, []byte(invocation.Message.Content))
		if err != nil {
			ch <- event.NewErrorEvent(invocation.InvocationID, a.name, model.ErrorTypeFlowError, err.Error())
			return
		}
		body, err := json.Marshal(result)
		if err != nil {
			ch <- event.NewErrorEvent(invocation.InvocationID, a.name, model.ErrorTypeFlowError, err.Error())
			return
		}
		ch <- event.NewResponseEvent(invocation.InvocationID, a.name, &model.Response{
			Object: model.ObjectTypeChatCompletion,
			Done:   true,
			Choices: []model.Choice{
				{Index: 0, Message: model.NewAssistantMessage(string(body))},
			},
		})
	}()
	return ch, nil
}

func (a *researchAgent) callSearchTool(ctx context.Context, args []byte) (domain.DiscoveryResult, error) {
	for _, candidate := range a.tools {
		declaration := candidate.Declaration()
		if declaration == nil || declaration.Name != "search_repositories" {
			continue
		}
		callable, ok := candidate.(tool.CallableTool)
		if !ok {
			return domain.DiscoveryResult{}, fmt.Errorf("search_repositories is not callable")
		}
		result, err := callable.Call(ctx, args)
		if err != nil {
			return domain.DiscoveryResult{}, err
		}
		discoveryResult, ok := result.(domain.DiscoveryResult)
		if !ok {
			return domain.DiscoveryResult{}, fmt.Errorf("unexpected search_repositories result type %T", result)
		}
		return discoveryResult, nil
	}
	return domain.DiscoveryResult{}, fmt.Errorf("search_repositories tool is not registered")
}
