---
name: tailwind-4
description: "Trigger: Tailwind CSS v4, @theme, CSS-first config, new utilities. Apply Tailwind 4 coding patterns."
license: MIT
metadata:
  author: kevg1t
  version: "1.0"
---

## Activation Contract

Load this skill when writing Tailwind CSS v4 styles, configuring `@theme`, using new utilities, or migrating from Tailwind v3.

## Hard Rules

- Configure Tailwind via `@theme` in CSS — not `tailwind.config.js` (deprecated in v4).
- Use CSS custom properties (`--color-*`, `--font-*`, `--spacing-*`) for design tokens; reference them via `theme()` or directly in CSS vars.
- Do not use `@apply` for complex multi-property compositions — use component classes or CSS layers instead.
- Prefer the new gradient syntax (`bg-gradient-to-r from-blue-500 to-purple-600`) over manual `background` declarations.
- Use `@layer base`, `@layer components`, `@layer utilities` to control cascade order.
- Avoid arbitrary values (`[value]`) unless no utility exists; prefer extending `@theme` with a design token.

## Decision Gates

| Scenario | Pattern |
|---|---|
| Brand color | Define in `@theme` as `--color-brand: #hex` |
| Custom spacing scale | Extend `@theme` with `--spacing-*` tokens |
| One-off responsive tweak | Responsive variant (`md:`, `lg:`) |
| Dark mode | `dark:` variant with `@media (prefers-color-scheme: dark)` or `class` strategy |
| Animation | `animate-*` utility or custom `@keyframes` in `@layer base` |
| Component-scoped styles | `@layer components` block |

## Execution Steps

1. Install Tailwind v4: `npm install tailwindcss@4 @tailwindcss/vite` (or PostCSS plugin).
2. Import in main CSS: `@import "tailwindcss"`.
3. Define design tokens in `@theme { }` block.
4. Apply utilities in markup; use responsive and state variants as needed.
5. Audit for removed v3 utilities; replace with v4 equivalents per upgrade guide.
6. Run `tailwindcss --input input.css --output output.css` to verify no build errors.

## Output Contract

Report: theme tokens added, utilities used, v3 patterns migrated, any arbitrary values introduced with justification.

## References

- Tailwind v4 docs: https://tailwindcss.com/docs
- v4 upgrade guide: https://tailwindcss.com/docs/upgrade-guide
