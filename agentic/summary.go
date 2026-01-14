package agentic

import (
	"context"

	"github.com/basenana/friday/core/agents/summarize"
	fridayapi "github.com/basenana/friday/core/api"
	"github.com/basenana/plugin/api"
	"github.com/basenana/plugin/logger"
	"github.com/basenana/plugin/types"
	"go.uber.org/zap"
)

const (
	summaryPluginName    = "summary"
	summaryPluginVersion = "1.0.0"
)

var SummaryPluginSpec = types.PluginSpec{
	Name:           summaryPluginName,
	Version:        summaryPluginVersion,
	Type:           types.TypeProcess,
	RequiredConfig: LLMRequiredConfig(),
}

type SummaryPlugin struct {
	logger      *zap.SugaredLogger
	workingPath string
	jobID       string
	config      map[string]string
}

func (p *SummaryPlugin) Name() string           { return summaryPluginName }
func (p *SummaryPlugin) Type() types.PluginType { return types.TypeProcess }
func (p *SummaryPlugin) Version() string        { return summaryPluginVersion }

func (p *SummaryPlugin) Run(ctx context.Context, request *api.Request) (*api.Response, error) {
	message := api.GetStringParameter("message", request, "")
	if message == "" {
		p.logger.Warnw("message parameter is required")
		return api.NewFailedResponse("message parameter is required"), nil
	}

	systemPrompt := api.GetStringParameter("system_prompt", request, "")

	p.logger.Infow("summary plugin started", "message_len", len(message), "has_system_prompt", systemPrompt != "")

	llm, err := NewLLMClient(p.config)
	if err != nil {
		p.logger.Warnw("create LLM client failed", "error", err)
		return api.NewFailedResponse(err.Error()), nil
	}

	agent := summarize.New("summary", "Summary Agent", llm, summarize.Option{
		SystemPrompt: systemPrompt,
	})

	resp := agent.Chat(ctx, &fridayapi.Request{
		Session:     NewSession(p.jobID),
		UserMessage: message,
	})

	content, _, err := CollectResponse(ctx, resp)
	if err != nil {
		p.logger.Warnw("collect response failed", "error", err)
		return api.NewFailedResponse(err.Error()), nil
	}

	p.logger.Infow("summary plugin completed", "result_len", len(content))
	return api.NewResponseWithResult(map[string]any{
		"result": content,
	}), nil
}

func NewSummaryPlugin(ps types.PluginCall) types.Plugin {
	return &SummaryPlugin{
		logger:      logger.NewPluginLogger(summaryPluginName, ps.JobID),
		workingPath: ps.WorkingPath,
		jobID:       ps.JobID,
		config:      ps.Config,
	}
}
