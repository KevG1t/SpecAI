---
name: nextjs-15
description: "Trigger: Next.js 15, App Router, Server Actions, caching, metadata, route handlers. Apply Next.js 15 coding patterns."
license: MIT
metadata:
  author: kevg1t
  version: "1.0"
---

## Activation Contract

Load this skill when building Next.js 15 applications, working with the App Router, Server Actions, route handlers, caching APIs, or metadata generation.

## Hard Rules

- Use the App Router (`app/` directory) exclusively — do not mix `pages/` and `app/` patterns.
- Colocate Server Components with their data fetching; pass data down via props, not global state.
- Define Server Actions with `"use server"` directive; validate all inputs before use.
- Use `cache()` for deduplicating repeated fetches within a request; use `unstable_cache` for cross-request caching with revalidation tags.
- Always export `metadata` or `generateMetadata` from page files — never set meta tags in `<head>` manually.
- Use `<Image>` from `next/image` for all images; set explicit `width` and `height` or `fill` with a sized container.
- Never `fetch()` inside a client component without React `use()` — use Server Components or route handlers for data.

## Decision Gates

| Scenario | Pattern |
|---|---|
| Page data fetch | Server Component + `async/await fetch` |
| Mutation (form submit) | Server Action + `useActionState` |
| REST/JSON endpoint | Route Handler (`app/api/.../route.ts`) |
| Cross-request cached data | `unstable_cache` with `revalidateTag` |
| Dynamic metadata | `generateMetadata()` export |
| Static params | `generateStaticParams()` export |
| Streaming large page | `<Suspense>` with `loading.tsx` |

## Execution Steps

1. Scaffold route: `app/(group)/feature/page.tsx` and optional `layout.tsx`.
2. Fetch data in the Server Component using `fetch` with appropriate `cache` / `next.revalidate` options.
3. Define Server Actions in a separate `actions.ts` file with `"use server"` at top.
4. Export `metadata` or `generateMetadata` from every public `page.tsx`.
5. Add `loading.tsx` for Suspense fallback and `error.tsx` for error boundaries.
6. Run `next build` and check output for any dynamic/static classification changes.

## Output Contract

Report: routes created, Server Actions defined, caching strategy used, metadata exported, any `pages/` patterns migrated, build output notes.

## References

- Next.js 15 docs: https://nextjs.org/docs
- App Router migration guide: https://nextjs.org/docs/app/building-your-application/upgrading
