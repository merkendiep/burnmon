# Cowork (Cloud) bucket, spec

Status: draft, 2026-10-12. Roadmap row 10d. Decided by Wilco on 2026-10-12 ("go"): fix row 10e now,
write this spec and the test plugin now, build the bucket after the park ends on 2026-11-01.
Research: `..\04_assets\2026-10-12_cloud_cowork_capture_research.md`.
Proof: `..\04_assets\2026-10-12_cloud_cowork_sessions_proof.md`.

## Problem

New Cowork sessions on Wilco's Team-plan account run in Anthropic's cloud, with or without a project,
and whatever the "Only on this computer" toggle says (measured 2026-10-12). They write no transcript
on the laptop. BurnMon reads only local files, so the Cowork row undercounts and nobody can see by how
much.

## Goal

The vendor strip, History, the Now chart and BurnMon Dev show two rows where there is one today:

| Agent key | Label | Source |
|---|---|---|
| `cowork` | Cowork | local `local-agent-mode-sessions` transcripts, unchanged |
| `cowork-cloud` | Cowork (Cloud) | usage-only extracts delivered by the BurnMon cloud plugin |

Plus one line that says what is missing: "N cloud sessions not measured".

## Facts this spec stands on

Measured, 2026-10-12:

1. A cloud Cowork session runs Claude Code in a container (`CLAUDE_CODE_ENTRYPOINT=remote_cowork`) and
   keeps a standard JSONL transcript with `message.usage` on every assistant row, plus one file per
   subagent under `<session>\subagents\`.
2. Plugins synced to the account run their hooks in that container (a PreToolUse hook from
   znd-private fired there).
3. The probe extract over this session: 167 requests, 87 ms, 19 whitelisted fields, no content.
4. Rows of one streamed request repeat with growing output counts. Keep the last row per
   (`requestId`, `message.id`); summing all rows overcounts.
5. The transcript is a lower bound: at one turn boundary it missed about 6 percent of Opus cache reads
   and all Haiku side calls compared with the runner's own totals.
6. The laptop's only trace of a cloud session is `remote-session-spaces.json` (key `entries`, no token
   fields).

Documented: the transcript may lag the in-memory conversation when a Stop hook fires
(code.claude.com/docs/en/hooks), and the transcript format "is internal to Claude Code and changes
between versions" (code.claude.com/docs/en/sessions).

## Design

### Part A, the plugin (runs in the cloud container)

`tools\burnmon-cloud-probe` is the test build (0.1.0). The real plugin is the same shape:

- A Stop hook (`command`, python3) reads the hook input, the transcript and its subagent files, and
  writes `/mnt/user-data/outputs/.burnmon/cloud/<session_id>.jsonl`: one line per request, last row
  wins, whitelisted fields only (`ts`, `session_id`, `request_id`, `message_id`, `model`, `agent`,
  `is_sidechain`, `entrypoint`, `service_tier`, `speed`, the four token counts, `cache_5m`, `cache_1h`,
  `thinking_tokens`, `v`, `source`). Never message text, tool input, paths, `cwd` or `gitBranch`.
- Delivery copies that file to a connected laptop folder with `device_commit_files`. Preferred: an
  `mcp_tool` hook, so no model tokens are spent. Fallback if live test 3 fails: one model call per
  turn, budget 300 tokens.
- The file is rewritten in full each turn, so a lost delivery heals on the next turn.

### Part B, BurnMon and BurnMon Dev

- Config key `cloud_drop_roots` (list). OPEN: default value (waits on Wilco, see Open questions).
- Full scan and live watcher pick up `<root>\**\.burnmon\cloud\*.jsonl` through a small adapter
  (`internal\adapter\claudecloud`) that maps one line to one `schema.Event`: vendor `anthropic`,
  agent `cowork-cloud`, surface `cloud`, session id from the line, request id from the line. The file
  is rewritten, not appended, so the adapter re-reads it whole when its size or mtime moves; upserts
  are idempotent on (vendor, session_id, request_id).
- Drop roots are not Claude transcript roots, so the Claude adapter never sees a drop file; only the
  new adapter reads them.
- Pricing: the same Anthropic price book as `cowork`.
- Labels: add `"cowork-cloud": "Cowork (Cloud)"` to `internal\vendorstrip`, `internal\history`,
  `internal\scan\types.go`, `cmd\burnmon-dev\station\station.js` and the colour maps in
  `cmd\burnmon-dev\page.html` (a lighter tint of the Cowork blue).
- Lower bound marker: the Cowork (Cloud) row carries a tooltip "lower bound: side calls not in the
  transcript".
- "Not measured" counter: entries in `remote-session-spaces.json` with no matching drop file.
  OPEN: the entry fields (waits on one look at the file's key names, no values).

## Live tests, in order (before any Part B code)

1. Hook fires: install the probe, start a cloud session, two turns. Pass: `probe\<session>.log` has
   two lines in the session's outputs.
2. Final row seen: pass if `final_text_in_transcript` is true on most turns; if false, the extract of
   turn N is completed by turn N+1 and the last turn needs a SessionEnd pass.
3. Hook delivery: pass if `C:\ZND\50_projects\burnmon\.burnmon\cloud\probe-latest.jsonl` appears on
   the laptop without the model calling any tool. Needs that folder connected to the session.
4. Gap size: compare the extract with the session's own totals at three turn boundaries.

## Done when

- Vendor strip shows Cowork and Cowork (Cloud) as separate rows, today, week and month.
- A cloud session's tokens appear within one turn of the session continuing, and match the extract.
- No drop file ever contains a field outside the whitelist (test on a fixture with content).
- The "not measured" count matches the cloud sessions without a drop file.
- go vet, go test, `.\build.ps1` and `.\scripts\uicheck.ps1` pass.

## Kill criteria (proposed, Wilco confirms)

- Kill the plugin if live test 1 fails.
- Fall back to the model-driven commit if test 3 fails; drop that fallback if it costs more than 500
  tokens per turn or fails on more than 1 turn in 10.
- Stop at once if any non-whitelisted field reaches a delivered file.
- Label numbers as estimates if they fall more than 10 percent under the runner totals; kill if more
  than 25 percent.

## Open questions

- OPEN: default `cloud_drop_roots`, and which folder every cloud session connects (waits on Wilco).
- OPEN: `remote-session-spaces.json` entry fields (waits on a key-name dump from the laptop).
- OPEN: kill thresholds above (waits on Wilco).

## Known limits

- Sessions with no connected folder, and sessions started on web or mobile, cannot deliver.
- The final turn arrives late or, if the session ends, only through a SessionEnd pass.
- The probe's mcp_tool hook hard-codes the laptop path; the real plugin must take it from config.
