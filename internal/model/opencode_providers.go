package model

// ModelEntry describes a single model within a provider.
type ModelEntry struct {
	ID             string
	Name           string
	SupportsEffort bool // true for models where effort level is configurable (e.g. o3)
}

// ProviderEntry describes a provider and its available models.
type ProviderEntry struct {
	ID     string
	Name   string
	Models []ModelEntry
}

// OpenCodeProviders is the static provider/model catalog used by the OpenCode model picker.
var OpenCodeProviders = []ProviderEntry{
	{
		ID:   "anthropic",
		Name: "Anthropic",
		Models: []ModelEntry{
			{ID: "claude-opus-4-8", Name: "Claude Opus 4.8"},
			{ID: "claude-sonnet-4-6", Name: "Claude Sonnet 4.6"},
			{ID: "claude-haiku-4-5", Name: "Claude Haiku 4.5"},
		},
	},
	{
		ID:   "openai",
		Name: "OpenAI",
		Models: []ModelEntry{
			{ID: "gpt-4o", Name: "GPT-4o"},
			{ID: "gpt-4.1", Name: "GPT-4.1"},
			{ID: "o3", Name: "o3", SupportsEffort: true},
		},
	},
	{
		ID:   "google",
		Name: "Google",
		Models: []ModelEntry{
			{ID: "gemini-2.5-pro", Name: "Gemini 2.5 Pro"},
			{ID: "gemini-2.5-flash", Name: "Gemini 2.5 Flash"},
		},
	},
	{
		ID:   "xai",
		Name: "xAI",
		Models: []ModelEntry{
			{ID: "grok-3", Name: "Grok 3"},
		},
	},
	{
		ID:   "groq",
		Name: "Groq",
		Models: []ModelEntry{
			{ID: "llama-3.3-70b", Name: "LLaMA 3.3 70B"},
		},
	},
}
