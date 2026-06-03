---
name: zod-4
description: "Trigger: Zod, z.object, schema validation, transforms, coerce, parse. Apply Zod 4 coding patterns."
license: MIT
metadata:
  author: kevg1t
  version: "1.0"
---

## Activation Contract

Load this skill when defining Zod schemas, validating runtime data, building transforms, using `coerce`, or integrating Zod with form libraries or API boundaries.

## Hard Rules

- Always define schemas at module scope; do not recreate them inside functions.
- Use `.parse()` when invalid input must throw; use `.safeParse()` in error-handled flows.
- Use `z.coerce.number()` / `z.coerce.boolean()` for URL params and form fields, not manual casting.
- Add `.describe()` to every top-level field when generating JSON Schema or OpenAPI specs.
- Use `z.infer<typeof schema>` to derive TypeScript types — never duplicate them manually.
- Prefer `.transform()` for shaping valid data; use `.refine()` for cross-field validations.
- Chain `.default()` before `.optional()` so the default is applied when the key is absent.

## Decision Gates

| Scenario | Pattern |
|---|---|
| Validate API response body | `schema.parse(response)` with try/catch |
| Validate form input gracefully | `schema.safeParse(formData)` → inspect `error.flatten()` |
| Coerce string to number | `z.coerce.number()` |
| Cross-field validation | `.superRefine()` or `.refine()` with `ctx.addIssue()` |
| Discriminated union | `z.discriminatedUnion("type", [...])` |
| Recursive / circular schema | `z.lazy(() => schema)` |
| Partial update payload | `schema.partial()` |

## Execution Steps

1. Define the schema that matches the expected data shape exactly.
2. Add `.describe()` to every field if the schema will be used for documentation.
3. Export the schema and its inferred type: `export type MyType = z.infer<typeof mySchema>`.
4. Validate at the boundary (API handler, form submit, CLI arg parse).
5. Surface errors with `error.flatten()` for field-level messages or `error.format()` for nested shapes.

## Output Contract

Report: schemas defined, types inferred, validation call sites updated, any coercions introduced, error handling approach.

## References

- Zod docs: https://zod.dev
- Zod v4 migration: https://zod.dev/v4
