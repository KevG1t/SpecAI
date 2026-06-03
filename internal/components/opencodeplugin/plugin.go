package opencodeplugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/KevG1t/SpecAI/internal/components/filemerge"
	"github.com/KevG1t/SpecAI/internal/model"
)

type Definition struct {
	ID          model.OpenCodeCommunityPluginID
	Name        string
	PackageName string
	RepoURL     string
	Owner       string
	Repo        string
	Description string
}

type Result struct {
	Changed bool
	Files   []string
}

var definitions = []Definition{
	{
		ID:          model.OpenCodePluginSubAgentStatusline,
		Name:        "Sub-agent Statusline",
		PackageName: "opencode-subagent-statusline",
		RepoURL:     "https://github.com/Joaquinvesapa/sub-agent-statusline",
		Owner:       "Joaquinvesapa",
		Repo:        "sub-agent-statusline",
		Description: "OpenCode sidebar/statusline for sub-agent activity",
	},
	{
		ID:          model.OpenCodePluginSDDMemoryManage,
		Name:        "SDD Memory Manager",
		PackageName: "opencode-sdd-memory-manage",
		RepoURL:     "https://github.com/j0k3r-dev-rgl/sdd-memory-plugin",
		Owner:       "j0k3r-dev-rgl",
		Repo:        "sdd-memory-plugin",
		Description: "OpenCode TUI for SDD profiles and sdd-memory memories",
	},
}

const argentinaLogoPluginFile = "argentina-logo.tsx"

const argentinaLogoPluginSource = `// @ts-nocheck
/** @jsxImportSource @opentui/solid */
import type { TuiPlugin, TuiThemeCurrent } from "@opencode-ai/plugin/tui"
import { useTerminalDimensions } from "@opentui/solid"
import { createMemo } from "solid-js"

const id = "argentina-logo"

const roseArt = [
  "             ⣠⣾⣷⣶⣦⣤⣤⣄⣠⣄⣀  ⢀⣀⣀",
  "          ⢀⣴⣿⣿⠿⣋⣭⣭⣯⣭⣍⣭⣿⣟⠛⠛⠿⠿⣿⣷⣄",
  "      ⢀⣴⣾⡟⢻⣿⡟⠁⣼⣿⠏⣵⢻⣿⣻⣿⣿⢿⡻⣿⣿⣶⡌⢿⣿⣷⣦⣤⡄",
  "   ⣤⣶⣾⣿⣿⠏ ⠈⢿⣄ ⢹⣏⠠⠟⣾⣿⣿⣿⣿⣿⠷⣏⣼⠟⢡⣿⡟⠋⢻⣿⣿⡄",
  "   ⠈⣿⣿⣿⣿⡆   ⣽⢧⡘⠈⠳⣦⣍⠛⠛⢦⣉⣴⣛⣫⣭⣴⡟⠋  ⣾⣿⣿⡿",
  "   ⢀⠹⣿⣿⣿⣷⣤⡄ ⠋ ⠙⢆ ⣠⠴⠟⠛⣛⣛⣛⠟⠋⠁⠺⡇ ⣀⣴⣿⣿⡟⠁",
  "   ⠈⣀⠈⠛⠷⠿⣿⣿⣷⣤⣀ ⢠⠋   ⠈⠉⠉    ⣠⣴⣥⠾⠛⠉⣰⣿⣷",
  "          ⠹⣯⣝⠛⠛⠷⢶⣤⣤⣀   ⢀⡠⠖⠋⠉⢉⣀⣀⣴⣾⣿⠿⠟⠃ ⠠⠦",
  "⠁       ⠖  ⠘⠻⢿⣦⣄⡀  ⠉⠛⢦⠠⢊⠤⠴⢒⣛⣛⣩⣽⡿⠟⠁⢀⡀",
  "⠲⠶⣦⠴⠶⠶⠶⠶⡶⠶⢶⣤⣄⡀⠨⠭⠽⠟⣓⢦⣀⠈⢇⡥⠖⠛⠋⠉⠉⠉    ⠈  ⢠⡤",
  "  ⠈⢷ ⠐⠂⢤⣽⣄ ⠰⡎⠙⠳⣄⡀ ⠈⢣⠘⢦⠋⣀⡬⠟⠛⠛⠉⢀⣀⣀⣠⡤⠄⠃",
  "   ⠈⢳⣀⡒⠉⠉⣉⠙⡲⣽⣄ ⣏⠳⡄ ⠘⡇ ⡾⠁ ⢀⡤⠖⣻⣿⡏⢡⡎ ⠰⠄",
  "     ⠛⠻⢦⣄⣉⡁⣀⣀⣈⣙⣺⣌⡇⢠⢀⡇⡾  ⣴⣿⡷⠊ ⢲⣠⠟",
  "          ⠈⠉    ⠈⠳⡄⣸⢱⠇⢀⣰⣯⣭⣥⠭⠾⠛⠃",
  "                  ⡷⠡⡯⢖⠉   ⢠⠤",
  "                ⡠⢊⡴⠤⠂⠃ ⠒",
  "             ⢀⡴⢪⠔⣉⠔⠋",
  "               ⠐⠈",
]

const compactArt = ["✦ SpecAI ✦"]

const Logo = (props: { theme: TuiThemeCurrent }) => {
  const dim = useTerminalDimensions()
  const lines = createMemo(() => {
    const term = dim()
    return term.height >= roseArt.length + 6 && term.width >= 64 ? roseArt : compactArt
  })

  return (
    <box flexDirection="column" alignItems="center">
      {lines().map((line) => (
        <text fg={props.theme.accent}>{line}</text>
      ))}
    </box>
  )
}

const tui: TuiPlugin = async (api) => {
  api.slots.register({
    id,
    order: 100,
    slots: {
      home_logo(ctx) {
        return <Logo theme={ctx.theme.current} />
      },
    },
  })
}

const plugin = { id: "argentina-logo", tui }
export default plugin
`

func Definitions() []Definition {
	out := make([]Definition, len(definitions))
	copy(out, definitions)
	return out
}

func DefinitionFor(id model.OpenCodeCommunityPluginID) (Definition, bool) {
	for _, def := range definitions {
		if def.ID == id {
			return def, true
		}
	}
	return Definition{}, false
}

func Install(homeDir string, id model.OpenCodeCommunityPluginID) (Result, error) {
	if id == model.OpenCodePluginArgentinaLogo {
		return installArgentinaLogo(homeDir)
	}

	def, ok := DefinitionFor(id)
	if !ok {
		return Result{}, fmt.Errorf("unknown OpenCode community plugin %q", id)
	}

	opencodeDir := filepath.Join(homeDir, ".config", "opencode")
	if err := os.MkdirAll(opencodeDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("create OpenCode config dir: %w", err)
	}

	tuiPath := filepath.Join(opencodeDir, "tui.json")
	written, err := ensureTUIPlugin(tuiPath, def.PackageName)
	if err != nil {
		return Result{}, err
	}

	return Result{Changed: written, Files: []string{tuiPath}}, nil
}

func installArgentinaLogo(homeDir string) (Result, error) {
	opencodeDir := filepath.Join(homeDir, ".config", "opencode")
	pluginDir := filepath.Join(opencodeDir, "tui-plugins")
	pluginPath := filepath.Join(pluginDir, argentinaLogoPluginFile)
	tuiPath := filepath.Join(opencodeDir, "tui.json")

	pluginWrite, err := filemerge.WriteFileAtomic(pluginPath, []byte(argentinaLogoPluginSource), 0o644)
	if err != nil {
		return Result{}, fmt.Errorf("write Argentina Logo TUI plugin: %w", err)
	}
	tuiChanged, err := ensureTUIPlugin(tuiPath, pluginPath)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Changed: pluginWrite.Changed || tuiChanged,
		Files:   []string{pluginPath, tuiPath},
	}, nil
}

func ensureTUIPlugin(path, pkg string) (bool, error) {
	root := map[string]any{"$schema": "https://opencode.ai/tui.json"}
	if data, err := os.ReadFile(path); err == nil && len(bytes.TrimSpace(data)) > 0 {
		if err := json.Unmarshal(data, &root); err != nil {
			return false, fmt.Errorf("parse OpenCode TUI config %q: %w", path, err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("read OpenCode TUI config %q: %w", path, err)
	}

	plugins := stringSlice(root["plugin"])
	for _, existing := range plugins {
		if existing == pkg {
			return false, nil
		}
	}
	plugins = append(plugins, pkg)
	root["plugin"] = plugins

	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return false, err
	}
	out = append(out, '\n')
	wr, err := filemerge.WriteFileAtomic(path, out, 0o644)
	if err != nil {
		return false, err
	}
	return wr.Changed, nil
}

func stringSlice(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}
