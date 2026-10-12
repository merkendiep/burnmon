# Cloud Cowork sessions leave no local transcript (proof, 2026-10-12)

Question (Wilco): after the 9 October Claude mail ("Chat and Cowork are becoming one Claude"), does BurnMon or BurnMon Dev miss Cowork sessions?

## Answer

The merge itself does not break ingest. Local Cowork sessions are still written to the folder BurnMon scans. Cowork sessions that run in Anthropic's cloud write no transcript and no token data to the laptop, so BurnMon cannot count them. Measured on Wilco's laptop, on his Team account.

## Measured (PowerShell on the laptop, 2026-10-12)

1. Only one Desktop root exists: the Microsoft Store install, `%LOCALAPPDATA%\Packages\Claude_pzs8sxrjxfjjc\LocalCache\Roaming\Claude\local-agent-mode-sessions`. `internal\scan\scan.go` scans it. There is no `Claude-3p` folder.
2. Local Cowork transcripts per day, 2026-09-28 to 2026-10-12: 3 to 23 a day, 8 on 2026-10-12. No drop after 6 October or 9 October.
3. Cowork transcripts written in the last 90 minutes: two, `#daf8` (15:51) and `#436d` (16:00). These are the two blue Cowork sessions BurnMon Dev showed at 16:20.
4. This session (cloud, id `session_018WsGrhhfQU5FKkq7b8BeYG`, started about 16:18) wrote no `.jsonl` anywhere under the Claude roots or `~\.claude\projects`.
5. The only local file that mentions its id is `local-agent-mode-sessions\<account>\<org>\remote-session-spaces.json`: 3,174 bytes, one top-level key `entries`, 0 `input_tokens`, 0 `output_tokens`. It is an index of cloud sessions, not a transcript.

## Read online (not measured)

- Since 2026-10-06, new Cowork tasks on Pro and Max run in the cloud, and the "Only on this computer" option is removed (Claude Help Center). Team is not named; point 4 shows the Team account starts cloud sessions too.
- With unified Claude on, a Desktop home-screen message on a device that can run Cowork starts a Cowork session, and Chat conversations use the same `local-agent-mode-sessions` layout (Claude docs, Desktop 3P data storage and Unified Claude pages).
- Third-party (3P) Desktop installs store data under `%LOCALAPPDATA%\Claude-3p\`, which BurnMon does not scan. Unified Claude becomes mandatory on 3P on 2026-11-10.

## Inferred

- Cloud Cowork usage is invisible to BurnMon by design (no account, no network). The only local trace, `remote-session-spaces.json`, could at most show that a cloud session exists, not what it cost.
- `cmd\burnmon\main.go` line 577 already tells users that cloud Cowork sessions leave no local transcript.

Sources: https://support.claude.com/fr/articles/15520349-utilisez-claude-cowork-sur-le-web-le-bureau-et-le-mobile , https://claude.com/docs/third-party/claude-desktop/data-storage , https://claude.com/docs/third-party/claude-desktop/unified-claude , https://techcrunch.com/2026/09/16/anthropic-merges-claude-chat-and-cowork-in-one-interface/

## Update 2026-10-12 16:50: the "Only on this computer" toggle does not keep new sessions local

Measured: on this build, Settings > General > Tasks shows "Only on this computer", and it was ON. There is no Settings > Cowork page and no "Run new tasks in the cloud" toggle, so the help article's wording does not match this build. With the toggle ON, a new session started inside a project ("Hub update", around 16:45) wrote no local transcript. In the last 30 minutes the only transcript written was `#436d` (16:45:08, 7,774,254 bytes), an older local session that is still running. Inferred: new sessions on this account go to the cloud whatever the toggle says, and local sessions that already exist stay local until they end. Not yet tested: a new session started outside any project.

Update 16:57 (measured): a new session started outside any project, after one "say hi" turn, also wrote no local transcript (audit.jsonl excluded). Both test sessions ran in the cloud.

Side finding 16:57 (measured in code, cause of the screenshot not yet proven): BurnMon Dev showed a Cowork session tagged `#udit` at 16:45 (in 4, cache 1.1M, out 20). The tag is the last four characters of the session key, so the key is probably `audit`. `FindJSONL` skips `audit.jsonl` (`internal\scan\scan.go`), but the live watcher does not: `internal\watch\watch_windows.go` passes every `.jsonl` change, and `Track` in `internal\watch\tail.go` filters only on the suffix. Open: which session's `audit.jsonl` it was (a local session's, which means double counting, or the cloud session's, which means a local usage trace exists).

Update 17:05 (measured): only one `audit.jsonl` changed in the last hour. It sits in session folder `92082970`, the folder of the local session `#436d`, and was written at 16:45:08, the same second as that session's transcript (4,355,059 bytes, 496 `input_tokens` matches). So `#udit` in BurnMon Dev is `#436d`'s own audit log, counted a second time by the live watcher. Cloud sessions leave no `audit.jsonl` with usage on the laptop. This is a double-count bug in the live path, not a new data source. Size of the inflation in the store: not measured.
