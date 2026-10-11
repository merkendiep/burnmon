# Design session tag (2026-10-11)

Status: spec, decided by Wilco 2026-10-11. Build after the park (2026-11-01, STATUS.md top
entry of 2026-10-03). Never use em dashes.

## Why

Claude Design work now also runs inside the Claude Code tab of the desktop app. That session
is a normal local Claude Code session that publishes a Design canvas through the Artifact tool.
BurnMon already counts its tokens as `claude-code`, but nothing shows that it was design work.
Design work in the Chat sidebar (claude.ai) keeps no local transcript and stays out of scope
(README, "Coverage").

## Evidence (measured 2026-10-11)

Transcript `C:\Users\WilcoDeTree\.claude\projects\C--ZND-50-projects-siteoffice\b273352c-...-ee219c141cb0.jsonl`,
shown in BurnMon Dev as `claude-code siteoffice #1cb0`. Counts are lines, not matches
(`Select-String`, one result per line):

| Marker | Lines |
|---|---|
| `<command-name>/design</command-name>` | 0 |
| `"skill": "...design..."` | 0 |
| `"name": "Artifact"` (tool call) | 49 |
| `"type_url"` | 3 |
| `project/canvas.json` | 46 |
| `*.dc.html` | 351 |
| `project/design-system.json` | 0 |

The two `type_url` values resolve (Artifact read, 2026-10-11) to the Design type and the Design
System type. The Design type answered under a different canonical address than the one in the
transcript, so a type URL is NOT a stable identifier. The `/design` skill (Claude Code
v2.1.234, week 34) was not used in this session, so the slash command is not a reliable signal
either.

## Decisions

1. A tag, not a new agent. The session keeps agent `claude-code`; the vendor strip does not
   change. The tag is `design`.
2. Session level. A session gets the tag when at least one `tool_use` block with
   `name == "Artifact"` has an input that mentions `project/canvas.json` or a path ending in
   `.dc.html` (in `file_path`, `files` keys or values, or `path`).
3. Build after the park.

OPEN: turn-level tagging, so the filtered sum counts only design turns and not the coding turns
of a mixed session (waits on a rule for where a design turn ends, and on usage of the
session-level tag).

## Scope

- Adapter: `internal\adapter\claude\claude.go`, in the `tool_use` loop next to `skillTag`
  (around line 320). Set a per-session flag from the rule above. Local file data only, no
  network, no vendor endpoint.
- Store: one new session column (or tag table) through a migration in `internal\store\migrations`.
  Existing sessions get the tag on the next full re-parse.
- Output: `Surface` and the agent label do not change (`classifySurface`, `agentFor`).
- UI: a `design` badge on the session card (Now, `burnmon-dev`) and on the Sessions row; a
  filter on the tag in the Sessions tab of `burnmon.exe` and the saved report.
- Out of scope: Design System sessions (`project/design-system.json`), a `claude-design` vendor
  row, claude.ai Chat sidebar Design.

## Tests first

- Fixture transcript with one Artifact call that publishes `project/canvas.json` and
  `project/Main.dc.html`: tag set.
- Fixture with an Artifact call that publishes a plain `page.html`: tag not set.
- Fixture with a Design System publish (`project/design-system.json` only): tag not set.
- Fixture where `canvas.json` appears only in assistant text, not in a tool input: tag not set.
- Agent stays `claude-code` in every case.

## Done when

`go vet ./...`, `go test ./... -count=1`, `.\build.ps1` and `.\scripts\uicheck.ps1` pass, and
`#1cb0` shows the `design` badge in BurnMon Dev and in the Sessions tab after a re-parse.

## Risk

The canvas file layout is versioned by Anthropic (`"v":3` in `canvas.json`). If the paths
change, the tag stops matching. The fixtures make that visible; re-check the markers on a fresh
Design session when the tag goes quiet.
