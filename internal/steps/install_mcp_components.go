package steps

import (
	"fmt"

	"github.com/KevG1t/SpecAI/internal/agents"
	"github.com/KevG1t/SpecAI/internal/components/mcp"
	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/planner"
)

// StepInjectMCPComponents injects Notion and/or Jira MCP server configurations for each
// IDE in ctx.IDEs when those components appear in the resolved plan.
// Any auth guidance strings are appended to ctx.AuthGuidance for display on completion screen.
type StepInjectMCPComponents struct {
	ctx  *InstallContext
	plan planner.ResolvedPlan
}

// NewStepInjectMCPComponents returns a new StepInjectMCPComponents.
func NewStepInjectMCPComponents(ctx *InstallContext, plan planner.ResolvedPlan) *StepInjectMCPComponents {
	return &StepInjectMCPComponents{ctx: ctx, plan: plan}
}

func (s *StepInjectMCPComponents) ID() string {
	return "Inyectando servidores MCP (Notion, Jira)"
}

func (s *StepInjectMCPComponents) Run() error {
	hasNotion := containsComponent(s.plan, model.ComponentNotion)
	hasJira := containsComponent(s.plan, model.ComponentJira)

	if !hasNotion && !hasJira {
		return nil
	}

	for _, ide := range s.ctx.IDEs {
		adapter, err := agents.NewAdapter(ide.AgentID())
		if err != nil {
			// Unknown agent — skip.
			continue
		}

		if hasNotion {
			_, guidance, err := mcp.InjectNotion(s.ctx.HomeDir, adapter)
			if err != nil {
				return fmt.Errorf("%s notion: %w", ide.AgentID(), err)
			}
			if guidance != "" {
				s.ctx.AuthGuidance = append(s.ctx.AuthGuidance, guidance)
			}
		}

		if hasJira {
			_, guidance, err := mcp.InjectJira(s.ctx.HomeDir, adapter)
			if err != nil {
				return fmt.Errorf("%s jira: %w", ide.AgentID(), err)
			}
			if guidance != "" {
				s.ctx.AuthGuidance = append(s.ctx.AuthGuidance, guidance)
			}
		}
	}
	return nil
}

func containsComponent(plan planner.ResolvedPlan, id model.ComponentID) bool {
	for _, c := range plan.OrderedComponents {
		if c == id {
			return true
		}
	}
	return false
}
