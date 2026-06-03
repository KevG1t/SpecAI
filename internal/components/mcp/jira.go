package mcp

import "os/exec"

// Jira MCP server overlay variables.
// Primary transport: uvx mcp-atlassian (Python, preferred).
// Fallback: @modelcontextprotocol/server-atlassian (npm).
// If neither is available, injection is skipped silently.

var defaultJiraServerJSONUvx = []byte(`{
  "command": "uvx",
  "args": [
    "mcp-atlassian",
    "--jira-url", "<your-jira-url>",
    "--jira-username", "<your-email>",
    "--jira-api-token", "<your-token>"
  ]
}
`)

var defaultJiraServerJSONNpm = []byte(`{
  "command": "npx",
  "args": [
    "-y",
    "@modelcontextprotocol/server-atlassian"
  ],
  "env": {
    "JIRA_HOST": "<your-jira-url>",
    "JIRA_USERNAME": "<your-email>",
    "JIRA_API_TOKEN": "<your-token>"
  }
}
`)

var defaultJiraOverlayJSONUvx = []byte(`{
  "mcpServers": {
    "jira": {
      "command": "uvx",
      "args": [
        "mcp-atlassian",
        "--jira-url", "<your-jira-url>",
        "--jira-username", "<your-email>",
        "--jira-api-token", "<your-token>"
      ]
    }
  }
}
`)

var defaultJiraOverlayJSONNpm = []byte(`{
  "mcpServers": {
    "jira": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-atlassian"],
      "env": {
        "JIRA_HOST": "<your-jira-url>",
        "JIRA_USERNAME": "<your-email>",
        "JIRA_API_TOKEN": "<your-token>"
      }
    }
  }
}
`)

var openCodeJiraOverlayJSONUvx = []byte(`{
  "mcp": {
    "jira": {
      "__replace__": {
        "type": "stdio",
        "command": "uvx",
        "args": ["mcp-atlassian", "--jira-url", "<your-jira-url>", "--jira-username", "<your-email>", "--jira-api-token", "<your-token>"],
        "enabled": true
      }
    }
  }
}
`)

var openCodeJiraOverlayJSONNpm = []byte(`{
  "mcp": {
    "jira": {
      "__replace__": {
        "type": "stdio",
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-atlassian"],
        "env": {
          "JIRA_HOST": "<your-jira-url>",
          "JIRA_USERNAME": "<your-email>",
          "JIRA_API_TOKEN": "<your-token>"
        },
        "enabled": true
      }
    }
  }
}
`)

var openClawJiraOverlayJSONUvx = []byte(`{
  "mcp": {
    "servers": {
      "jira": {
        "command": "uvx",
        "args": ["mcp-atlassian", "--jira-url", "<your-jira-url>", "--jira-username", "<your-email>", "--jira-api-token", "<your-token>"]
      }
    }
  }
}
`)

var openClawJiraOverlayJSONNpm = []byte(`{
  "mcp": {
    "servers": {
      "jira": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-atlassian"],
        "env": {
          "JIRA_HOST": "<your-jira-url>",
          "JIRA_USERNAME": "<your-email>",
          "JIRA_API_TOKEN": "<your-token>"
        }
      }
    }
  }
}
`)

var vsCodeJiraOverlayJSONUvx = []byte(`{
  "servers": {
    "jira": {
      "type": "stdio",
      "command": "uvx",
      "args": ["mcp-atlassian", "--jira-url", "<your-jira-url>", "--jira-username", "<your-email>", "--jira-api-token", "<your-token>"]
    }
  }
}
`)

var vsCodeJiraOverlayJSONNpm = []byte(`{
  "servers": {
    "jira": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-atlassian"],
      "env": {
        "JIRA_HOST": "<your-jira-url>",
        "JIRA_USERNAME": "<your-email>",
        "JIRA_API_TOKEN": "<your-token>"
      }
    }
  }
}
`)

// uvxAvailableFn is injectable for testing.
var uvxAvailableFn = func() bool {
	_, err := exec.LookPath("uvx")
	return err == nil
}

// UvxAvailable reports whether the uvx Python runner is available on PATH.
func UvxAvailable() bool { return uvxAvailableFn() }

// JiraAuthGuidance returns the post-install guidance message for Jira MCP auth.
func JiraAuthGuidance(configPath string, usedUvx bool) string {
	transport := "npm (@modelcontextprotocol/server-atlassian)"
	if usedUvx {
		transport = "uvx (mcp-atlassian)"
	}
	return "Jira MCP server installed using " + transport + ".\n" +
		"  Config file: " + configPath + "\n" +
		"  Next step: replace placeholder values with your Jira URL, username, and API token.\n" +
		"  Docs: https://github.com/sooperset/mcp-atlassian"
}

func DefaultJiraServerJSON() []byte {
	if uvxAvailableFn() {
		content := make([]byte, len(defaultJiraServerJSONUvx))
		copy(content, defaultJiraServerJSONUvx)
		return content
	}
	content := make([]byte, len(defaultJiraServerJSONNpm))
	copy(content, defaultJiraServerJSONNpm)
	return content
}

func DefaultJiraOverlayJSON() []byte {
	if uvxAvailableFn() {
		content := make([]byte, len(defaultJiraOverlayJSONUvx))
		copy(content, defaultJiraOverlayJSONUvx)
		return content
	}
	content := make([]byte, len(defaultJiraOverlayJSONNpm))
	copy(content, defaultJiraOverlayJSONNpm)
	return content
}

func OpenCodeJiraOverlayJSON() []byte {
	if uvxAvailableFn() {
		content := make([]byte, len(openCodeJiraOverlayJSONUvx))
		copy(content, openCodeJiraOverlayJSONUvx)
		return content
	}
	content := make([]byte, len(openCodeJiraOverlayJSONNpm))
	copy(content, openCodeJiraOverlayJSONNpm)
	return content
}

func OpenClawJiraOverlayJSON() []byte {
	if uvxAvailableFn() {
		content := make([]byte, len(openClawJiraOverlayJSONUvx))
		copy(content, openClawJiraOverlayJSONUvx)
		return content
	}
	content := make([]byte, len(openClawJiraOverlayJSONNpm))
	copy(content, openClawJiraOverlayJSONNpm)
	return content
}

func VSCodeJiraOverlayJSON() []byte {
	if uvxAvailableFn() {
		content := make([]byte, len(vsCodeJiraOverlayJSONUvx))
		copy(content, vsCodeJiraOverlayJSONUvx)
		return content
	}
	content := make([]byte, len(vsCodeJiraOverlayJSONNpm))
	copy(content, vsCodeJiraOverlayJSONNpm)
	return content
}
