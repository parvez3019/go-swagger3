---
title: "AI agent skills for go-swagger3"
description: "Claude Code and Cursor skills that let AI coding agents annotate Go APIs and generate OpenAPI 3 specs with go-swagger3."
---

# AI agent skills

go-swagger3 ships instructions for AI coding agents, so Claude Code or Cursor can add `@` annotations to an existing Go
API and generate an OpenAPI 3 spec for you.

| Agent | Source | Copy to |
|-------|--------|---------|
| Claude Code | [agent-skills/claude/SKILL.md](../agent-skills/claude/SKILL.md) | `.claude/skills/go-swagger3-docs/SKILL.md` |
| Cursor | [agent-skills/cursor/cursor_rules.mdc](../agent-skills/cursor/cursor_rules.mdc) | `.cursor/rules/go-swagger3.mdc` |

## Usage

1. Copy the file for your agent into your project at the path above.
2. Ask the agent, for example: *"Document the HTTP handlers in `internal/api` with go-swagger3 annotations and
   generate `oas.json`."*
3. Review the generated annotations and spec, then commit them.

The skills cover the same annotations as the [annotation reference](annotations.md).
