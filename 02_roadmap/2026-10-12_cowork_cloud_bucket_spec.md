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
- Plugin 1.1.1 (2026-10-12) is a diagnostic build for the 4-minute delay found in test 5. Every
  extract run appends one line to `diag.log` (UTC time, hook, result, row count, newest request
  time, short session id; no content). FileChanged also runs `extract.py --fired` next to the
  copy, so `diag.log` records the moment the copy started. UserPromptSubmit copies `diag.log` to
  `.burnmon\cloud\diag.log` as its own hook entry; BurnMon reads only `.jsonl` there.

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
6. Where the turn-end delay sits (plugin 1.1.1): install 1.1.1 over 1.1.0, new cloud session, one
   plain turn, then 5 minutes with no messages in any session, then a second plain turn, which
   copies `diag.log` to the laptop. Compare three times for turn 1: the Stop extract's "wrote"
   line, the FileChanged "fired" line after it (both in `diag.log`, container clock), and the
   laptop's write time of the file holding that turn. Fired about 1 s after the write but landed
   about 4 minutes later: the delay sits in the copy (the `mcp_tool` call or the device bridge).
   Fired about 4 minutes later: the watcher or the runner holds the event.

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
- Test 5 (copy after each write), 2026-10-12, cloud session `b2ca13ab`. CORRECTION: this session ran
  plugin 1.0.0. Plugin 1.1.0 was built but not installed (the plugin page still showed 1.0.0, 4 hooks).
  So it tested 1.0.0: subagent rows landed mid-turn (22:34:10 local) and turn ends landed about 4
  minutes after the turn (4 min 8 s and 4 min 9 s) without a next message. Inferred from 1.1.1's log
  below: the cloud runner stops Claude Code after a few idle minutes (SessionEnd copies the file) and
  restarts it at the next message (SessionStart `resume`). Also seen: a second cloud session with the
  same target (the build session) overwrote the laptop's `latest.jsonl`; BurnMon keeps both sessions'
  rows, the file holds one session at a time.
- Test 6, plugin 1.1.1 (`607b00e`), 2026-10-12, cloud session `eb7fbea0`, plus the build session after
  it picked up 1.1.1. Plain turn: Stop extract wrote at 21:07:21.189 UTC, FileChanged fired at
  21:07:21.837, laptop file written 23:07:28 local (about 6.5 s). Subagent turn: subagent rows landed
  at 23:14:42 and 23:15:15 local, before the turn ended. FAIL for the end of that turn: it landed only
  with the next message, 5 minutes later. Root cause, from `diag.log` and the runner's own debug log
  (`/tmp/claude-code.log`, a yardstick only): after each SessionStart registration, FileChanged fired
  exactly once (three registrations, two sessions); every later write fired nothing. Reproduced in
  headless Claude Code 2.1.296 on the same rclone mount: with `latest.jsonl` already present when the
  watch is registered, temp-file-plus-rename writes fire once and then never again (the watcher stays
  on the replaced inode); in-place writes fired on all four writes. Also measured: SubagentStop fires
  about 2 minutes after some turns with no new rows (a post-turn subagent, inferred).
- Plugin 1.1.2: `latest.jsonl` is rewritten in place (same inode; write first, cut the tail after, so
  it is never empty); the session file keeps temp-file-plus-rename. 24 extract tests pass on Python
  3.10 and 3.13 (new: the inode survives a rewrite and a shorter extract leaves no tail);
  `claude plugin validate --strict` passes.
- Test 7, 2026-10-12, cloud session `688bb225` (installed version not confirmed on the plugin page;
  the pattern fits 1.1.2). Plain turn: request 23:43:47 local, on the laptop 23:43:54. Subagent turn:
  three copies at 23:44:35 to 23:44:38 (FileChanged now fires on every write), but each carried the
  previous content; the subagent rows landed at 23:45:09 with the next write's copy, and the turn's
  last request did not land within the watch. Cause, measured from the build session with
  `device_commit_files` itself: a copy 1 s after an in-place write delivered the old content, a copy
  8 s after delivered the new. The outputs folder is an rclone mount with `vfs_cache_mode` full
  (read from its config), so a write reaches the store the copy reads from only after a short
  upload delay (rclone's default is 5 s: inferred, not read from this config).
- Plugin 1.1.3: every extract write starts a detached `extract.py --settle <sha256>` that waits 10 s
  and rewrites `latest.jsonl` in place with the same bytes if they are unchanged; that write's
  FileChanged copy finds the new content uploaded. The hook still returns in about 50 ms (measured).
  29 extract tests pass; `claude plugin validate --strict` passes.
- Test 8 (plugin 1.1.3, live): not run yet. Same steps as test 6, watcher open for 60 s after each
  turn. Pass: each turn's last request lands within about 20 s of the turn ending, and subagent rows
  within about 20 s of the subagent finishing.

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
- Plugin 1.0.0: the end of a turn arrives with the next message, or when the runner stops the
  idle session (about 4 minutes, test 5). Plugins 1.1.0 and 1.1.1: FileChanged fires once per
  session start (test 6). Plugin 1.1.2: fires on every write, but each copy carries the previous
  upload (test 7). Plugin 1.1.3 adds a settle rewrite 10 s after each write; live result pending
  (test 8).
- The extract has no tool calls, so the Station cannot tell a cloud session that works from one
  that waits. BurnMon Dev shows it as Thinking until 3 minutes after its newest request, then
  Waiting on you (the Copilot in VS Code rule).
- Two cloud sessions at once overwrite each other's `latest.jsonl` on the laptop. No rows are
  lost in BurnMon (upserts only), but the file is not a per-session record.
- The probe's mcp_tool hook hard-codes the laptop path; the real plugin must take it from config.
