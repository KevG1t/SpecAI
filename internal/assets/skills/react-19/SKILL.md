---
name: react-19
description: "Trigger: React 19, Server Components, Actions, use(), Suspense, React hooks. Apply React 19 coding patterns."
license: MIT
metadata:
  author: kevg1t
  version: "1.0"
---

## Activation Contract

Load this skill when writing React 19 components, using Server Components, Actions, the `use()` hook, Suspense, or migrating from React 18 patterns.

## Hard Rules

- Default to Server Components; add `"use client"` only when you need browser APIs, event listeners, or React hooks.
- Never call hooks conditionally or inside loops; hooks must be at the top level of a function component.
- Use Actions for form mutations; wrap with `useActionState` to track pending and error states.
- Wrap async boundaries with `<Suspense>` and provide meaningful `fallback` UIs.
- Use `use(promise)` or `use(context)` in client components to unwrap values; do not manually `.then()` in render.
- Do not use `useEffect` for data fetching that can be done in a Server Component.
- Mark components that throw on error with `<ErrorBoundary>` — Suspense does not catch errors.

## Decision Gates

| Scenario | Pattern |
|---|---|
| Data fetch with no user interaction | Server Component + `async/await` |
| Interactive form submission | Action + `useActionState` |
| Shared data across components | React Context (client) or prop drilling (server) |
| Async value in client component | `use(promise)` inside `<Suspense>` |
| Optimistic UI update | `useOptimistic` hook |
| Ref forwarding | `ref` prop directly (React 19 removed `forwardRef`) |

## Execution Steps

1. Identify whether the component needs browser state or user events → determines server vs. client.
2. Co-locate data fetching in the nearest Server Component ancestor.
3. Define Actions as `async` server functions (`"use server"` directive or server module export).
4. Wrap children that suspend in `<Suspense fallback={...}>`.
5. Test server components with `react-dom/server` render utilities; test client components with Testing Library.

## Output Contract

Report: components created/modified, server vs. client split, Actions defined, Suspense boundaries added, any deprecated patterns removed.

## References

- React 19 blog post: https://react.dev/blog/2024/12/05/react-19
- React 19 API reference: https://react.dev/reference/react
