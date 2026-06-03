---
name: ai-sdk-5
description: "Trigger: Vercel AI SDK, useChat, streamText, generateText, tool use, providers. Apply AI SDK 5 coding patterns."
license: MIT
metadata:
  author: kevg1t
  version: "1.0"
---

## Activation Contract

Load this skill when using the Vercel AI SDK (v5), implementing `useChat`, `streamText`, `generateText`, defining tools, or integrating AI providers.

## Hard Rules

- Always specify the `model` via the provider factory (e.g., `openai('gpt-4o')`, `anthropic('claude-opus-4-5')`); never use string model IDs directly.
- Use `streamText` for conversational flows; use `generateText` for one-shot completions.
- Define tools with `tool()` helper; include `description` and `parameters` (Zod schema) for every tool.
- Handle `finishReason` explicitly: `stop`, `tool-calls`, `length`, `content-filter` require different branching.
- Do not block on the full stream — consume via `result.textStream` or `result.fullStream` iteratively.
- Always handle errors from `result.text` or stream consumption — network and model errors are normal.
- For multi-step agents, use `maxSteps` in `generateText` to limit recursion depth.

## Decision Gates

| Scenario | Pattern |
|---|---|
| Chat UI with streaming | `useChat` hook (client) + `/api/chat` route (server) |
| Server-only completion | `generateText` with provider model |
| Real-time streaming response | `streamText` + `result.toDataStreamResponse()` |
| Agent with tool use | `generateText` with `tools` + `maxSteps` |
| Structured output | `generateObject` with a Zod schema |
| Switch provider at runtime | Provider factory from `ai` package |

## Execution Steps

1. Install: `npm install ai @ai-sdk/openai` (or desired provider package).
2. Create provider instance: `const openai = createOpenAI({ apiKey: process.env.OPENAI_API_KEY })`.
3. Define tools with `tool({ description, parameters: z.object({...}), execute: async (...) })`.
4. Call `streamText` / `generateText` with `model`, `messages`, and optional `tools`.
5. On the client, wire `useChat` to the `/api/chat` route handler.
6. Test tool calls by mocking the provider with `MockLanguageModelV1` from `ai/test`.

## Output Contract

Report: provider configured, tools defined, streaming or batch path chosen, error handling added, test mocks used.

## References

- AI SDK docs: https://sdk.vercel.ai/docs
- AI SDK 5 migration: https://sdk.vercel.ai/docs/migration-guides/migration-guide-4-5
