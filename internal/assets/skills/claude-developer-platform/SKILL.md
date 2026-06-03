---
name: claude-developer-platform
description: "Trigger: Anthropic SDK, Claude API, tool use, prompt caching, streaming, claude_client. Apply Claude developer platform patterns."
license: MIT
metadata:
  author: kevg1t
  version: "1.0"
---

## Activation Contract

Load this skill when integrating with the Anthropic API, building tool use, implementing prompt caching, streaming responses, or using the Claude SDK in any language.

## Hard Rules

- Always pin the SDK version and model name — never use `latest` aliases in production.
- Pass `max_tokens` explicitly; do not rely on model defaults.
- Use prompt caching (`cache_control`) for large system prompts or repeated context to reduce cost and latency.
- Handle streaming chunks incrementally; never buffer the entire stream before processing.
- Validate tool input schemas with strict JSON Schema; reject unexpected properties.
- Never log raw API keys or bearer tokens; mask them in error messages.
- Implement exponential backoff on `429 rate_limit_error` and `529 overloaded_error`.

## Decision Gates

| Scenario | Pattern |
|---|---|
| Long system prompt reused across calls | `cache_control: {"type": "ephemeral"}` on the system block |
| Multi-step reasoning | multi-turn messages with assistant prefill |
| External tool integration | `tools` array + `tool_use` / `tool_result` message pair |
| Real-time output | `stream: true` with delta accumulation |
| Structured output | tool use with a single schema-defined tool |
| Token counting before send | `client.beta.messages.count_tokens()` |

## Execution Steps

1. Initialize the client with `ANTHROPIC_API_KEY` from env; never hardcode.
2. Define the system prompt and any tool schemas before the message loop.
3. Apply `cache_control` to static blocks (system, tool definitions) when reusing context.
4. Build the messages array; include `tool_result` blocks when tools were invoked.
5. Parse `stop_reason`; branch on `tool_use`, `end_turn`, or `max_tokens`.
6. Log token usage (`input_tokens`, `output_tokens`, `cache_read_input_tokens`) for cost tracking.

## Output Contract

Report: API calls made, cache hits, tools invoked, token usage summary, any errors with retry behavior.

## References

- Anthropic API docs: https://docs.anthropic.com/
- Prompt caching guide: https://docs.anthropic.com/en/docs/build-with-claude/prompt-caching
- Tool use guide: https://docs.anthropic.com/en/docs/build-with-claude/tool-use
