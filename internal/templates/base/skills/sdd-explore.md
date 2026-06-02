# Skill: SDD Explore

This skill guides the AI agent through exploring a codebase before making changes.

## Steps
1. **Find entry points:** Locate the main file, routing, or core packages.
2. **Scan files:** Read up to 5 critical files related to the user request.
3. **Map dependencies:** Understand imports and side effects.
4. **Document findings:** Save your mapping in `openspec/exploration.md` or sdd-memory.
