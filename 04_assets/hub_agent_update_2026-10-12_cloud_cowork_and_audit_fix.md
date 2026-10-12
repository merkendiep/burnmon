# Hub Agent Update - ZeroNonsense.dev

**Date:** 2026-10-12 - **Owner:** Wilco de Tree
**Project:** BurnMon
**Purpose:** Cloud Cowork sessions are invisible to BurnMon; a live double-count bug is fixed in code; a park exception was granted.
**Read order:** this file, `C:\ZND\50_projects\burnmon\04_assets\2026-10-12_cloud_cowork_sessions_proof.md`, `C:\ZND\50_projects\burnmon\02_roadmap\2026-10-12_cowork_cloud_bucket_spec.md`, `C:\ZND\50_projects\burnmon\04_assets\2026-10-12_cloud_cowork_capture_research.md`
**Supersedes:** nothing

## 1. Headline
New Cowork sessions on Wilco's Team-plan account run in Anthropic's cloud and leave no local transcript, so BurnMon undercounts Cowork (measured). The audit.jsonl double count in the live watcher is fixed in code, uncommitted. A Cowork (Cloud) bucket is specced, with a test plugin ready; build after the park.

## 2. What changed on disk
- **Committed:** nothing. HEAD and origin are `930aeb4` (checked with `git rev-parse HEAD` and `git ls-remote origin HEAD`).
- **Files touched (uncommitted):** `C:\ZND\50_projects\burnmon\internal\scan\filter.go` (new), `C:\ZND\50_projects\burnmon\internal\scan\filter_test.go` (new), `C:\ZND\50_projects\burnmon\internal\scan\scan.go`, `C:\ZND\50_projects\burnmon\internal\watch\watch.go`, `C:\ZND\50_projects\burnmon\internal\watch\tail.go`, `C:\ZND\50_projects\burnmon\internal\watch\watch_test.go`, `C:\ZND\50_projects\burnmon\internal\store\store.go`, `C:\ZND\50_projects\burnmon\internal\store\store_test.go`, `C:\ZND\50_projects\burnmon\internal\dataset\dataset.go`, `C:\ZND\50_projects\burnmon\.gitignore`, `C:\ZND\50_projects\burnmon\02_roadmap\roadmap.md` (rows 10d, 10e), `C:\ZND\50_projects\burnmon\SESSION_LOG.md`.
- **New, untracked:** `C:\ZND\50_projects\burnmon\tools\burnmon-cloud-probe\` (plugin source and `dist\burnmon-cloud-probe-0.1.0.plugin`), the spec, the proof and research files named above.

## 3. What did NOT happen (and why)
- Not committed, not pushed, not tagged, not rebuilt: Wilco's step.
- Not tested on Windows: tests ran in a Linux container. Three existing tests fail on Linux before and after the change (Windows paths and watcher), so `go test` on the laptop is still owed.
- The probe plugin is not installed and none of the four live tests has run.
- The size of the double count in Wilco's store is not measured.
- No Cowork (Cloud) code exists; build is after 2026-11-01.

## 4. Findings worth propagating
- [RESULT] With "Only on this computer" on, two new sessions (inside and outside a project) both ran in the cloud and wrote no local transcript (laptop PowerShell, 2026-10-12).
- [RESULT] Local Cowork transcripts are still written for sessions that started locally: 8 on 2026-10-12.
- [RESULT] BurnMon Dev's `#udit` session was a local session's `audit.jsonl` read a second time by the live watcher (folder `92082970`, 496 usage rows).
- [RESULT] A cloud Cowork container keeps a standard Claude Code transcript with usage; account-synced plugin hooks run there. A probe extract over one session: 167 requests, 87 ms, no content.
- [STATE] Hook-driven delivery to the laptop (mcp_tool hook calling `device_commit_files`) is untested.
- [STATE] Cloud usage from the transcript is a lower bound: about 6 percent of Opus cache reads and all Haiku side calls missing at one measured turn.

## 5. Hub-level decision
  **Decision:** BurnMon park exception: fix the audit.jsonl double count now, and write the Cowork (Cloud) spec and test plugin now. The bucket itself is built after the park ends on 2026-11-01.
  **Context:** BurnMon is parked until 2026-11-01 except the Station (decisions 2026-10-01 and 2026-10-02). The bug inflates the Cowork numbers BurnMon exists to show; cloud sessions are a growing blind spot.
  **Alternatives considered:** everything after the park; a README note only instead of a roadmap row.
  **Consequences:** a small uncommitted code change waits for Wilco's Windows test and commit; four live tests with the probe plugin gate the build.
  **Links:** `C:\ZND\50_projects\burnmon\02_roadmap\roadmap.md` rows 10d and 10e, `C:\ZND\50_projects\burnmon\02_roadmap\2026-10-12_cowork_cloud_bucket_spec.md`.

## 6. What the next hub read should update
`C:\ZND\10_holding\03_logs\decisions.md` (the park exception), `C:\ZND\10_holding\01_projects\burnmon.md` (park note and the cloud blind spot), BurnMon `STATUS.md` once committed.

Tracker rows moved: none.

## 7. Open flags for next session
- Windows `go vet`, `go test`, `.\build.ps1`, then commit.
- Install the probe plugin and run live tests 1 to 4 from the spec.
- Spec OPEN items: default `cloud_drop_roots`, `remote-session-spaces.json` entry fields, kill thresholds.
- BurnMon `STATUS.md` names `1b86eec` as HEAD; the repo is at `930aeb4`, so STATUS lags.
- Side note: the znd-private PreToolUse git guard calls `powershell`, which does not exist in cloud containers (exit 127, the hook does not block).

## 8. Related files
`C:\ZND\50_projects\burnmon\04_assets\2026-10-12_cloud_cowork_sessions_proof.md`, `C:\ZND\50_projects\burnmon\04_assets\2026-10-12_cloud_cowork_capture_research.md`, `C:\ZND\50_projects\burnmon\02_roadmap\2026-10-12_cowork_cloud_bucket_spec.md`, `C:\ZND\50_projects\burnmon\tools\burnmon-cloud-probe\`.
