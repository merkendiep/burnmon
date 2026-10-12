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

`tools\burnmon-cloud-probe` (removed 2026-10-12, in git history at `be652f9`) is the test build (0.1.0). The real plugin is the same shape:

- A Stop hook (`command`, python3) reads the hook input, the transcript and its subagent files, and
  writes `/mnt/user-data/outputs/.burnmon/cloud/<session_id>.jsonl`: one line per request, last row
  wins, whitelisted fields only (`ts`, `session_id`, `request_id`, `message_id`, `model`, `agent`,
  `is_sidechain`, `entrypoint`, `service_tier`, `speed`, the four token counts, `cache_5m`, `cache_1h`,
  `thinking_tokens`, `v`, `source`). Never message text, tool input, paths, `cwd` or `gitBranch`.
- Delivery copies that file to a connected laptop folder with `device_commit_files`. Preferred: an
  `mcp_tool` hook, so no model tokens are spent. Fallback if live test 3 fails: one model call per
  turn, budget 300 tokens.
- The file is rewritten in full each turn, so a lost delivery heals on the next turn.
- Plugin 1.1.0 (2026-10-12) moves the copy from "on Stop" to "after each write", to remove the
  one-turn lag of 1.0.0. A SessionStart command hook (`extract.py --watch`) returns `watchPaths`
  for `/mnt/user-data/outputs/.burnmon/cloud/latest.jsonl`; a FileChanged group (matcher
  `latest.jsonl`) runs the `mcp_tool` copy after each write. The extract runs on Stop, on
  SubagentStop and on PostToolUse (async, `--throttle`: skipped when `latest.jsonl` is under
  20 s old), and writes `latest.jsonl` last, through a temp file and a rename. The 1.0.0 copies on
  Stop, UserPromptSubmit and SessionEnd stay as the fallback. Each part is documented
  (code.claude.com/docs/en/hooks, FileChanged and SessionStart); the combination in a cloud
  session is not tested yet (test 5).

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
5. Copy after each write (plugin 1.1.0): install 1.1.0 over 1.0.0, start a new cloud session with
   `C:\ZND` connected, run one plain turn, then one turn that starts a subagent. Pass: the laptop's
   `latest.jsonl` holds each turn's rows within seconds of that turn ending, before the next
   message, and the subagent's rows land when the subagent finishes, before its turn ends. The
   subagent turn is the clean proof: every other copy runs at Stop or later, so subagent rows on
   the laptop while the main turn still runs can only come from the FileChanged copy. Fail: rows
   still arrive one turn late (FileChanged does not fire in cloud sessions); 1.0.0 behaviour
   then remains.

## Live test results

- 2026-10-12 17:32, probe 0.1.0, cloud session `551731f0`, folder `C:\ZND\50_projects\burnmon` connected.
  Test 1 (hook fires): PASS. Test 3 (hook delivery without the model): PASS. `probe-latest.jsonl` landed
  on the laptop at 17:32:20, 1,006 bytes, 2 request rows (17:32:10 and 17:32:13, `claude-sonnet-5-5`,
  `entrypoint` `remote_cowork`), 7 seconds after the last request. Every row carries exactly the 19
  whitelisted fields.
- Test 2 (final row seen): PASS in both test sessions. The probe log said `final_text_in_transcript: true`
  (session `1afdc570`, Opus 5.5, 1 request, 323 ms; session `551731f0`, Sonnet 5.5, 2 requests, 403 ms).
- Test 3 caveat 1, delivery lags one turn: the second delivery (17:33:59) still held the 2 rows of turn 1.
  The extract hook and the commit hook run in parallel on Stop, so the commit copies the previous file.
  The final turn of a session is therefore not delivered unless a later event (UserPromptSubmit,
  SessionEnd) commits again. To test in the build: extract also on PostToolUse, commit also on
  UserPromptSubmit and SessionEnd.
- Test 3 caveat 2, folder: session `1afdc570` had only a non-ZND folder connected, and nothing was
  delivered. Delivery needs a connected folder that contains the hook's `devicePath`. Session
  `551731f0` had `C:\ZND` connected and delivered into `C:\ZND\50_projects\burnmon`.
- Test 4 (gap size), measured 2026-10-12 on cloud session `46e3aa09` (this build session), at seven turn
  boundaries between 14:21 and 14:47 UTC. Reference: the runner's own cumulative `modelUsage` from its
  `result` lines (an undocumented runner file in `/tmp`, used here only as a yardstick, never in the
  product). Opus 5.5 (main model): the transcript matched exactly at the first three boundaries, then
  missed 7.1, 5.9, 2.7 and 4.0 percent of tokens (at 14:47: 18,692,590 against 17,950,524), almost all of
  it cache reads that arrive in steps of about 190,000 tokens. Inferred, not proven: the post-turn
  memory pass, a second Opus call on the same cache that the transcript does not record. Haiku 5.5:
  471,071 tokens, none of them in the transcript (side calls), costing 0.08 of 10.82 USD list price.
  Verdict against the kill criteria: under the 10 percent line, so the numbers stay unlabelled as
  estimates, with the lower-bound tooltip. Limit: one session, one day, seven boundaries.
- Test 5 (copy after each write): not run yet. Plugin 1.1.0 built 2026-10-12 for target
  `C:\ZND\50_projects\burnmon\.burnmon\cloud\latest.jsonl`; `claude plugin validate --strict`
  passes on the unpacked package (2.1.296), and 15 extract tests pass on Python 3.10 and 3.13.
  Measured on this build session's real transcript: 25 requests, 19 fields each, 15 ms.

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
- Answered 2026-10-12 (masked key dump on Wilco's laptop): `remote-session-spaces.json` holds
  `entries`, a list of 17, each `{sessionId: "session_" + 24 chars, spaceId, folders[], memoryEnabled}`.
  No timestamps. Its name and fields say it maps cloud sessions to project spaces; whether a cloud
  session outside any project is listed is not tested (inferred: not). The cloud runner's
  `CLAUDE_CODE_REMOTE_SESSION_ID` is `cse_` plus the same 24 characters, so the two can be matched.
- OPEN: build the "not measured" counter on this partial list, or close the item (waits on Wilco).
- OPEN: kill thresholds above (waits on Wilco).

## Known limits

- Sessions with no connected folder, and sessions started on web or mobile, cannot deliver.
- Plugin 1.0.0: the final turn arrives late or, if the session ends, only through a SessionEnd
  pass. Plugin 1.1.0 copies after each write instead; until test 5 passes, assume 1.0.0 timing.
- The probe's mcp_tool hook hard-codes the laptop path; the real plugin must take it from config.
