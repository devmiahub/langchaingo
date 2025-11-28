package agents

import (
	"context"

	"github.com/devmiahub/langchaingo/chains"
	"github.com/devmiahub/langchaingo/schema"
	"github.com/devmiahub/langchaingo/tools"
)

// Agent is the interface all agents must implement.
type Agent interface {
	// Plan Given an input and previous steps decide what to do next. Returns
	// either actions or a finish. Options can be passed to configure LLM
	// parameters like temperature, max tokens, etc.
	Plan(ctx context.Context, intermediateSteps []schema.AgentStep, inputs map[string]string, options ...chains.ChainCallOption) ([]schema.AgentAction, *schema.AgentFinish, error) //nolint:lll
	GetInputKeys() []string
	GetOutputKeys() []string
	GetTools() []tools.Tool
}
