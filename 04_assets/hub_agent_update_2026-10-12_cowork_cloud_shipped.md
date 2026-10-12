# Hub Agent Update - ZeroNonsense.dev

**Date:** 2026-10-12 (evening) - **Owner:** Wilco de Tree
**Project:** BurnMon
**Purpose:** Cowork (Cloud) shipped on main and is live; a second park exception was granted.
**Read order:** this file, `C:\ZND\50_projects\burnmon\02_roadmap\2026-10-12_cowork_cloud_bucket_spec.md`, `C:\ZND\50_projects\burnmon\04_assets\hub_agent_update_2026-10-12_cloud_cowork_and_audit_fix.md`
**Supersedes:** nothing (adds to `hub_agent_update_2026-10-12_cloud_cowork_and_audit_fix.md`)

## 1. Headline
BurnMon and BurnMon Dev now count cloud Cowork sessions as their own bucket, "Cowork (Cloud)", fed by the `burnmon-cloud` plugin. Committed and pushed as `a249bc5`; seen live in BurnMon Dev on 2026-10-12.

## 2. What changed on disk
- **Committed** in `C:\ZND\50_projects\burnmon` (origin `wilcodetree/burnmon`): `c8d85d9` audit.jsonl double-count fix, `be652f9` docs and probe plugin, `b933994` probe results, `a249bc5` Cowork (Cloud) bucket and plugin. Pushed; `git ls-remote origin HEAD` returned `a249bc5` (checked live).
- **Uncommitted:** this brief, a `SESSION_LOG.md` line and roadmap row 10d marked done.
- **Untracked by design:** `C:\ZND\50_projects\burnmon\.burnmon\` (local usage drops) and `C:\ZND\50_projects\burnmon\tools\burnmon-cloud\dist\` (built plugin), both in `.gitignore`.

## 3. What did NOT happen (and why)
- No release, no tag, no version bump: Wilco's call.
- The "N cloud sessions not measured" counter is not built (needs the key names of `remote-session-spaces.json`).
- Test 4 of the spec (gap size against the runner's own totals) has not run.
- The plugin has not been given to anyone else. `tools\burnmon-cloud\build.ps1 -Target` makes a copy for another computer; handing it over is Wilco's step.
- `uicheck w1` still fails on its 1.5 s first-paint budget (known flake, not traced to this change).

## 4. Findings worth propagating
- [RESULT] Live, 2026-10-12 21:52 local: vendor strip shows Cowork (Cloud) 64.0M today; one cloud session delivered 242 requests with only the 19 whitelisted fields.
- [RESULT] The audit.jsonl fix removed at least 47.6M tokens of double count from today's Cowork total and 131.3M from the month (BurnMon Dev screenshots 16:57 and 17:26).
- [RESULT] Delivery needs a folder connected to the cloud session that contains the drop path; sessions with only other folders stay in the cloud uncounted.
- [STATE] Cloud numbers are a lower bound: the transcript misses side calls (about 6 percent of cache reads and all Haiku calls at one measured turn).

## 5. Hub-level decision
  **Decision:** Second BurnMon park exception: build and ship Cowork (Cloud) now instead of after 2026-11-01.
  **Context:** Wilco: "lets build now, because i need to see the tokens, also to learn and to give this to" a client. New Cowork sessions on his account run in the cloud and were invisible to BurnMon.
  **Alternatives considered:** build after the park, as decided earlier the same day.
  **Consequences:** Shipped the same day; the park otherwise stands until 2026-11-01.
  **Links:** `C:\ZND\50_projects\burnmon\02_roadmap\roadmap.md` row 10d, commit `a249bc5`.

## 6. What the next hub read should update
`C:\ZND\10_holding\03_logs\decisions.md` (the second park exception), `C:\ZND\10_holding\01_projects\burnmon.md` (Cowork (Cloud) live, park note), BurnMon `STATUS.md` (still names an older HEAD).

Tracker rows moved: none.

## 7. Open flags for next session
- Build the "not measured" counter; run spec test 4.
- Decide whether `burnmon-cloud` gets a release and how it is offered to other users.
- BurnMon `STATUS.md` lags the repo.

## 8. Related files
`C:\ZND\50_projects\burnmon\tools\burnmon-cloud\`, `C:\ZND\50_projects\burnmon\04_assets\2026-10-12_cloud_cowork_sessions_proof.md`, `C:\ZND\50_projects\burnmon\04_assets\2026-10-12_cloud_cowork_capture_research.md`.
