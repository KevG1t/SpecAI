---
name: playwright
description: "Trigger: Playwright, e2e tests, page objects, fixtures, browser automation. Apply Playwright testing patterns."
license: MIT
metadata:
  author: kevg1t
  version: "1.0"
---

## Activation Contract

Load this skill when writing Playwright end-to-end tests, building page objects, defining fixtures, configuring CI, or debugging flaky tests.

## Hard Rules

- Use Page Object Model (POM) — never write raw `page.locator()` calls in test bodies.
- Prefer user-visible locators: `getByRole()`, `getByLabel()`, `getByText()` over CSS selectors or XPath.
- Never use `page.waitForTimeout()` — use `expect(locator).toBeVisible()` or `page.waitForResponse()` instead.
- Scope fixtures to the smallest granularity needed: `"test"` for UI state, `"worker"` for expensive setup.
- Use `test.describe.parallel()` only for truly independent tests; share state with worker-scoped fixtures.
- Run a11y checks with `@axe-core/playwright` on critical pages.
- Tag tests with `@smoke` and `@regression`; run smoke only in CI pre-merge.

## Decision Gates

| Scenario | Pattern |
|---|---|
| User interaction sequence | Page Object method with chained assertions |
| Shared authentication state | Worker-scoped `storageState` fixture |
| Mocked API response | `page.route()` intercept |
| Visual regression | `expect(page).toHaveScreenshot()` |
| Flaky network | `page.waitForResponse()` + `APIRequestContext` |
| Cross-browser run | `playwright.config.ts` `projects` array |

## Execution Steps

1. Install: `npm install -D @playwright/test && npx playwright install`.
2. Define Page Objects in `tests/pages/` — each method returns `this` for chaining.
3. Create fixtures in `tests/fixtures.ts`; extend `test` from `@playwright/test`.
4. Write tests that use fixtures and page objects; no raw locators in test bodies.
5. Configure `playwright.config.ts`: `baseURL`, `retries`, `reporter`, and `projects`.
6. Add CI step: `npx playwright test --reporter=github` with artifact upload for traces.

## Output Contract

Report: page objects created/modified, fixtures defined, test cases written, CI config updated, flaky patterns addressed.

## References

- Playwright docs: https://playwright.dev/docs/intro
- Best practices: https://playwright.dev/docs/best-practices
