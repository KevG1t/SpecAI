---
name: pytest
description: "Trigger: pytest, fixtures, parametrize, conftest, async tests, coverage. Apply pytest testing patterns."
license: MIT
metadata:
  author: kevg1t
  version: "1.0"
---

## Activation Contract

Load this skill when writing or reviewing pytest tests, defining fixtures, using parametrize, writing async tests, or configuring coverage.

## Hard Rules

- Define fixtures in `conftest.py` at the appropriate scope level; never define shared fixtures in test files.
- Use `@pytest.mark.parametrize` for data-driven tests; include a clear `ids=` list for readable output.
- Prefer function-scoped fixtures by default; use `scope="session"` only for expensive setups (DB connections, etc.).
- For async code, use `pytest-asyncio` with `@pytest.mark.asyncio` or set `asyncio_mode = "auto"` in `pyproject.toml`.
- Mock external I/O with `pytest-mock` (`mocker.patch`) — never hit real network or disk in unit tests.
- Use `tmp_path` fixture for filesystem tests; never use absolute paths or rely on current working directory.
- Fail fast on missing coverage: configure `--cov-fail-under=80` in CI.

## Decision Gates

| Scenario | Pattern |
|---|---|
| Multiple input cases | `@pytest.mark.parametrize` |
| Shared expensive setup | `scope="session"` fixture in `conftest.py` |
| External HTTP call | `mocker.patch("module.requests.get")` |
| Async function test | `pytest-asyncio` + `async def test_...` |
| File system interaction | `tmp_path` fixture |
| Expected exception | `pytest.raises(ExceptionType)` as context manager |
| Slow integration test | `@pytest.mark.slow` + `pytest -m "not slow"` in CI |

## Execution Steps

1. Install: `pip install pytest pytest-cov pytest-mock pytest-asyncio`.
2. Create `conftest.py` at project root and feature level for fixtures.
3. Name tests `test_<behavior>_<condition>` — not `test_function_name`.
4. Parametrize happy path and edge cases; add IDs for each case.
5. Run `pytest --cov=src --cov-report=term-missing` and address uncovered lines.
6. Configure `pyproject.toml` with `[tool.pytest.ini_options]` for markers and asyncio mode.

## Output Contract

Report: fixtures defined, parametrize cases added, async tests written, mocks introduced, coverage delta.

## References

- pytest docs: https://docs.pytest.org/
- pytest-asyncio: https://pytest-asyncio.readthedocs.io/
