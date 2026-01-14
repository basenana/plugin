package agentic

import (
	"context"

	"github.com/basenana/friday/core/agents/research"
	fridayapi "github.com/basenana/friday/core/api"
	"github.com/basenana/plugin/api"
	"github.com/basenana/plugin/logger"
	"github.com/basenana/plugin/types"
	"go.uber.org/zap"
)

const (
	researchPluginName    = "research"
	researchPluginVersion = "1.0.0"
)

var ResearchPluginSpec = types.PluginSpec{
	Name:    researchPluginName,
	Version: researchPluginVersion,
	Type:    types.TypeProcess,
	RequiredConfig: append(LLMRequiredConfig(),
		"friday_websearch_type", // WebSearch type: pse (Google Programmable Search Engine)
		"friday_pse_engine_id",  // Google PSE Engine ID (required when websearch_type=pse)
		"friday_pse_api_key",    // Google PSE API Key (required when websearch_type=pse)
	),
}

type ResearchPlugin struct {
	logger      *zap.SugaredLogger
	workingPath string
	jobID       string
	config      map[string]string
}

func (p *ResearchPlugin) Name() string           { return researchPluginName }
func (p *ResearchPlugin) Type() types.PluginType { return types.TypeProcess }
func (p *ResearchPlugin) Version() string        { return researchPluginVersion }

func (p *ResearchPlugin) Run(ctx context.Context, request *api.Request) (*api.Response, error) {
	message := api.GetStringParameter("message", request, "")
	if message == "" {
		p.logger.Warnw("message parameter is required")
		return api.NewFailedResponse("message parameter is required"), nil
	}

	systemPrompt := api.GetStringParameter("system_prompt", request, "")

	websearchType := p.config["friday_websearch_type"]
	p.logger.Infow("research plugin started", "message_len", len(message), "has_system_prompt", systemPrompt != "", "websearch_type", websearchType)

	llm, err := NewLLMClient(p.config)
	if err != nil {
		p.logger.Warnw("create LLM client failed", "error", err)
		return api.NewFailedResponse(err.Error()), nil
	}

	rsTools := FileAccessTools(p.workingPath)

	// Check for websearch_type config and add corresponding tools
	switch p.config["friday_websearch_type"] {
	case "pse":
		engineID := p.config["friday_pse_engine_id"]
		apiKey := p.config["friday_pse_api_key"]
		if engineID != "" && apiKey != "" {
			rsTools = append(rsTools, NewPSEWebSearchTool(engineID, apiKey)...)
			p.logger.Infow("PSE web search tool added", "engine_id", engineID)
		}
	}

	agent := research.New("research", "Research Agent", llm, research.Option{
		SystemPrompt: systemPrompt,
		Tools:        rsTools,
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

	p.logger.Infow("research plugin completed", "result_len", len(content))
	return api.NewResponseWithResult(map[string]any{
		"result": content,
	}), nil
}

func NewResearchPlugin(ps types.PluginCall) types.Plugin {
	return &ResearchPlugin{
		logger:      logger.NewPluginLogger(researchPluginName, ps.JobID),
		workingPath: ps.WorkingPath,
		jobID:       ps.JobID,
		config:      ps.Config,
	}
}
