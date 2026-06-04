package screens

import (
	"fmt"
	"strings"
	"testing"

	"github.com/KevG1t/specai/internal/model"
	"github.com/KevG1t/specai/internal/planner"
	"github.com/KevG1t/specai/internal/versions"
)

func TestRenderDependencyTreePiOnlySddMemoryPlanShowsComponentAndPiInstallCopy(t *testing.T) {
	selection := model.Selection{
		Agents:     []model.AgentID{model.AgentPi},
		Preset:     model.PresetFullModism,
		Components: []model.ComponentID{model.ComponentSddMemory},
	}
	plan := planner.ResolvedPlan{
		Agents:            []model.AgentID{model.AgentPi},
		OrderedComponents: []model.ComponentID{model.ComponentSddMemory},
	}

	out := RenderDependencyTree(plan, selection, 0)

	if strings.Contains(out, "No components selected yet.") {
		t.Fatalf("RenderDependencyTree() showed generic empty copy for Pi-only SddMemory plan; output:\n%s", out)
	}
	for _, want := range []string{
		"Components to install",
		"sdd-memory",
		"Pi agent support will be installed.",
		"pi install npm:specai-pi",
		"pi install npm:sdd-memory-kevg1t",
		"pi install npm:pi-mcp-adapter",
		fmt.Sprintf("npm exec --yes --package sdd-memory-kevg1t@%s -- pi-sdd-memory init", versions.SDDMemory),
		"pi install npm:pi-subagents",
		"pi install npm:pi-intercom",
		"pi install npm:@juicesharp/rpiv-ask-user-question",
		"pi install npm:pi-web-access",
		"pi install npm:pi-lens",
		"pi install npm:@juicesharp/rpiv-todo",
		"pi install npm:pi-btw",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("RenderDependencyTree() missing %q for Pi-only plan; output:\n%s", want, out)
		}
	}
}

func TestRenderDependencyTreeGenericEmptyPlanKeepsExistingCopy(t *testing.T) {
	selection := model.Selection{Preset: model.PresetFullModism}

	out := RenderDependencyTree(planner.ResolvedPlan{}, selection, 0)

	if !strings.Contains(out, "No components selected yet.") {
		t.Fatalf("RenderDependencyTree() missing generic empty copy; output:\n%s", out)
	}
	if strings.Contains(out, "Pi agent support will be installed.") {
		t.Fatalf("RenderDependencyTree() showed Pi copy for generic empty plan; output:\n%s", out)
	}
}

func TestRenderDependencyTreeMixedPiEmptyPlanShowsPiInstallCopy(t *testing.T) {
	selection := model.Selection{
		Agents: []model.AgentID{model.AgentPi, model.AgentOpenCode},
		Preset: model.PresetFullModism,
	}
	plan := planner.ResolvedPlan{Agents: selection.Agents}

	out := RenderDependencyTree(plan, selection, 0)

	if strings.Contains(out, "No components selected yet.") {
		t.Fatalf("RenderDependencyTree() showed generic empty copy for mixed Pi plan; output:\n%s", out)
	}
	for _, want := range []string{
		"Pi agent support will be installed.",
		"pi install npm:specai-pi",
		"pi install npm:sdd-memory-kevg1t",
		"pi install npm:pi-mcp-adapter",
		fmt.Sprintf("npm exec --yes --package sdd-memory-kevg1t@%s -- pi-sdd-memory init", versions.SDDMemory),
		"pi install npm:pi-subagents",
		"pi install npm:pi-intercom",
		"pi install npm:@juicesharp/rpiv-ask-user-question",
		"pi install npm:pi-web-access",
		"pi install npm:pi-lens",
		"pi install npm:@juicesharp/rpiv-todo",
		"pi install npm:pi-btw",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("RenderDependencyTree() missing %q for mixed Pi plan; output:\n%s", want, out)
		}
	}
}
