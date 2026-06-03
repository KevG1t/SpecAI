package opencode

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func DefaultCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".cache", "opencode", "models.json")
}

func DefaultSettingsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "opencode", "opencode.json")
}

func DefaultAuthPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".local", "share", "opencode", "auth.json")
}

type ModelCost struct {
	Input  float64 `json:"input"`
	Output float64 `json:"output"`
}

type ModelLimit struct {
	Context int `json:"context"`
	Output  int `json:"output"`
}

type Model struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Family    string     `json:"family"`
	ToolCall  bool       `json:"tool_call"`
	Reasoning bool       `json:"reasoning"`
	Cost      ModelCost  `json:"cost"`
	Limit     ModelLimit `json:"limit"`
	Variants  []string   `json:"-"`
}

type Provider struct {
	ID     string           `json:"id"`
	Name   string           `json:"name"`
	Env    []string         `json:"env"`
	Models map[string]Model `json:"models"`
}

func LoadModels(cachePath string) (map[string]Provider, error) {
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("read models cache %q: %w", cachePath, err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse models cache: %w", err)
	}

	providers := make(map[string]Provider, len(raw))
	for id, providerJSON := range raw {
		var p Provider
		if err := json.Unmarshal(providerJSON, &p); err != nil {
			continue
		}
		p.ID = id
		providers[id] = p
	}

	return providers, nil
}

func LoadModelsOrEmpty(cachePath string) (map[string]Provider, error) {
	providers, err := LoadModels(cachePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]Provider{}, nil
		}
		return nil, err
	}
	return providers, nil
}

func loadAuthProviders(authPath string) map[string]bool {
	data, err := os.ReadFile(authPath)
	if err != nil {
		return nil
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}

	result := make(map[string]bool, len(raw))
	for id := range raw {
		result[id] = true
	}
	return result
}

var envLookup = os.Getenv

var authPath = DefaultAuthPath

func DetectAvailableProviders(providers map[string]Provider, customProviderIDs ...string) []string {
	authProviders := loadAuthProviders(authPath())

	customSet := make(map[string]bool, len(customProviderIDs))
	for _, id := range customProviderIDs {
		customSet[id] = true
	}

	var available []string
	for id, provider := range providers {
		if !hasToolCallModel(provider) {
			continue
		}

		if customSet[id] {
			available = append(available, id)
			continue
		}

		if authProviders[id] {
			available = append(available, id)
			continue
		}

		if id == "opencode" {
			available = append(available, id)
			continue
		}

		if len(provider.Env) > 0 && allEnvVarsSet(provider.Env) {
			available = append(available, id)
			continue
		}
	}

	sort.Strings(available)
	return available
}

func hasToolCallModel(provider Provider) bool {
	for _, m := range provider.Models {
		if m.ToolCall {
			return true
		}
	}
	return false
}

func allEnvVarsSet(envVars []string) bool {
	for _, v := range envVars {
		if envLookup(v) == "" {
			return false
		}
	}
	return true
}

func FilterModelsForSDD(provider Provider) []Model {
	var models []Model
	for _, m := range provider.Models {
		if m.ToolCall {
			models = append(models, m)
		}
	}

	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})

	return models
}

func (m Model) EffortLevels() []string {
	if len(m.Variants) == 0 {
		return nil
	}
	return m.Variants
}

func DefaultVariantsCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".specai", "cache", "model-variants.json")
}

func LoadVariants(variantsPath string) (map[string]map[string][]string, error) {
	data, err := os.ReadFile(variantsPath)
	if err != nil {
		return nil, err
	}
	var variants map[string]map[string][]string
	if err := json.Unmarshal(data, &variants); err != nil {
		return nil, err
	}
	return variants, nil
}

func EnrichWithVariants(cached map[string]Provider, variantsPath string) {
	variants, err := LoadVariants(variantsPath)
	if err != nil {
		return
	}
	for provID, models := range variants {
		cachedProv, ok := cached[provID]
		if !ok {
			continue
		}
		for modelID, levels := range models {
			if cachedModel, ok := cachedProv.Models[modelID]; ok {
				cachedModel.Variants = levels
				cachedProv.Models[modelID] = cachedModel
			}
		}
		cached[provID] = cachedProv
	}
}

type ConfigModel struct {
	Name     string `json:"name"`
	ToolCall bool   `json:"tool_call"`
}

type ConfigProvider struct {
	Name   string                 `json:"name"`
	Models map[string]ConfigModel `json:"models"`
}

func LoadConfigProviders(path string) (map[string]ConfigProvider, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]ConfigProvider{}, nil
		}
		return map[string]ConfigProvider{}, err
	}

	var raw struct {
		Provider map[string]ConfigProvider `json:"provider"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return map[string]ConfigProvider{}, fmt.Errorf("parse opencode settings %q: %w", path, err)
	}
	if raw.Provider == nil {
		return map[string]ConfigProvider{}, nil
	}
	return raw.Provider, nil
}

func MergeCustomProviders(providers map[string]Provider, config map[string]ConfigProvider) map[string]Provider {
	if len(config) == 0 {
		return providers
	}

	merged := make(map[string]Provider, len(providers)+len(config))
	for id, p := range providers {
		clone := Provider{ID: p.ID, Name: p.Name, Env: append([]string(nil), p.Env...), Models: make(map[string]Model, len(p.Models))}
		for mid, m := range p.Models {
			clone.Models[mid] = m
		}
		merged[id] = clone
	}

	for id, cp := range config {
		existing, ok := merged[id]
		if !ok {
			existing = Provider{ID: id, Name: cp.Name, Models: make(map[string]Model, len(cp.Models))}
		}
		if existing.Models == nil {
			existing.Models = make(map[string]Model, len(cp.Models))
		}
		for mid, cm := range cp.Models {
			name := cm.Name
			if name == "" {
				name = mid
			}
			existing.Models[mid] = Model{ID: mid, Name: name, ToolCall: cm.ToolCall}
		}
		merged[id] = existing
	}

	return merged
}

func SDDPhases() []string {
	return []string{
		"sdd-init",
		"sdd-explore",
		"sdd-propose",
		"sdd-spec",
		"sdd-design",
		"sdd-tasks",
		"sdd-apply",
		"sdd-verify",
		"sdd-archive",
		"sdd-onboard",
	}
}

func JDPhases() []string {
	return []string{
		"jd-judge-a",
		"jd-judge-b",
		"jd-fix-agent",
	}
}

func ConfigurableAgentPhases() []string {
	phases := SDDPhases()
	phases = append(phases, JDPhases()...)
	return phases
}
