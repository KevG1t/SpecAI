---
name: typescript
description: "Trigger: TypeScript files, strict mode, type safety, utility types, generics. Apply TypeScript 5.x coding patterns."
license: MIT
metadata:
  author: kevg1t
  version: "1.0"
---

## Activation Contract

Load this skill when writing or reviewing TypeScript code, enforcing strict mode, using utility types, defining generics, or auditing type safety.

## Hard Rules

- Enable `strict: true` in tsconfig; never disable individual strict checks.
- Prefer explicit return types on exported functions and class methods.
- Use `unknown` instead of `any`; narrow with type guards before use.
- Prefer `type` aliases for union/intersection shapes; use `interface` for object contracts that may be extended.
- Avoid non-null assertions (`!`) — handle `null | undefined` explicitly via narrowing or optional chaining.
- Use `satisfies` operator to validate literal shapes without widening.
- Prefer `const` enums only when bundle size is critical; use string union types otherwise.

## Decision Gates

| Scenario | Pattern |
|---|---|
| Shared object shape | `interface` |
| Union / tuple / mapped type | `type` alias |
| Runtime-safe narrowing | type guard (`is`) or assertion function |
| Conditional behavior by type | overloaded function signatures |
| Reusable transformation | generic utility type (`Pick`, `Omit`, `Partial`, custom) |
| API boundary input | `z.object` (Zod) or `io-ts` codec |

## Execution Steps

1. Set `strict: true` and resolve all pre-existing errors before adding new code.
2. Type every function parameter and return; let inference handle local variables.
3. For complex generics, write the constraint first, then the implementation.
4. Validate external data at entry points; keep internal types as narrow as possible.
5. Run `tsc --noEmit` before committing; treat type errors as test failures.

## Output Contract

Report: files changed, type errors resolved, new types introduced, any `any` or `!` usage with justification.

## References

- TypeScript Handbook: https://www.typescriptlang.org/docs/handbook/
- TypeScript 5.x release notes: https://devblogs.microsoft.com/typescript/
