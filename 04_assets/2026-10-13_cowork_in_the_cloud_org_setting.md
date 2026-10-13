# "Cowork in the cloud" turned off for the organization: what it means for BurnMon

Date: 2026-10-13. Trigger: Wilco saw that an admin switched "Cowork in the cloud (Beta)" off this
morning (Settings > Cowork, "Set by admin"). The same page says "Chat and Cowork unified is on for
your organization. Members can run code and work with files from chat regardless of this setting."

## Documented (support.claude.com, fetched 2026-10-13)

- Team and Enterprise: "Whether sessions can run in the cloud is a separate control", the toggle
  "Run Cowork in the cloud". Team default: "on by default. An owner can turn it off any time".
  Enterprise default: off. Local sessions: "Claude's work runs on the user's computer, with code in
  an isolated virtual machine", and "For local sessions, Cowork stores conversation history on users'
  computers." The page does not say what happens to running cloud sessions when the toggle goes off.
  (Use Claude Cowork on Team and Enterprise plans, 13455879.)
- Pro and Max: new Cowork tasks run in the cloud since 2026-10-06, no setting; Team cloud Cowork is
  in beta on web and mobile. Cloud tasks reach only folders connected through the desktop app, and
  only while it is open. (Use Claude Cowork on web, desktop, and mobile, 15520349.)
- Unified experience: "Every conversation starts as a chat"; when a task needs code or files,
  "Claude continues the same conversation in an isolated cloud sandbox." It requires "Cloud code
  execution and file creation" and "doesn't require Cowork or Cowork in the cloud to be on. It is
  meant to replace both over time." The article is written for Enterprise; it does not mention Team,
  plugins, hooks or transcripts. (Set up and manage the new unified Claude experience, 17444462.)

## Measured today

- Test session `35082878` (started about 10:34 local) ran in the cloud: its extract wrote 7 rows,
  and the extract keeps only rows whose `entrypoint` is `remote_cowork`. Plugin 3.3.0 was synced
  into it, and its `diag.log` shows SessionStart, PostToolUse, SubagentStop, Stop, FileChanged and
  the settle rewrite all running. Its copies did not reach the laptop: only the `burnmon` folder was
  connected, not `~\.burnmon`.
- This build session (`137a77b1`, cloud since 2026-10-12) picked up plugin 3.3.0; its hook copies
  to `~/.burnmon/cloud/latest.jsonl` land on the laptop, and BurnMon Dev (built from `9607b0c`) shows
  them as cowork-cloud `#2a37` with live turns (screenshot 10:43). So `~` works from a hook, and the
  home drop folder works end to end.
- BurnMon Dev shows Cowork (local) 1.3M tokens today. Inferred, not proven: some Cowork work ran on
  the laptop again after the toggle went off.

## What it means for BurnMon (inferred)

- Cloud runs do not stop: with the unified experience on, a chat that needs code or files continues
  in a cloud sandbox whatever "Cowork in the cloud" says. Today's test session is consistent with
  that, if the toggle was already off at 10:34 (time of the change not yet known).
- So the BurnMon cloud plugin stays needed, and local Cowork sessions are counted from local
  transcripts as before. The plugin does nothing harmful in a local session: the extract only acts
  when `CLAUDE_CODE_ENTRYPOINT` is `remote_cowork`, and a copy hook without the remote-devices server
  fails quietly.

## Open

- OPEN: the exact time of the change (waits on the admin activity feed), to tie test session
  `35082878` to "toggle off".
- OPEN: whether the admin keeps it off (waits on the admin).

## Do other token monitors solve cloud sessions? (searched 2026-10-13)

- ccusage, Claude-Code-Usage-Monitor, claude-usage and similar read local transcripts only; none
  reads a cloud session.
- tokscale-desktop (PR #1, merged 2026-10-10): a Stop hook in each cloud session runs the tokscale
  engine and uploads daily totals to the user's own sync Worker over the internet. Same idea as
  BurnMon cloud (hook in the session), but delivery needs an online service and an account; it covers
  Claude Code on the web, not Cowork. Not usable under BurnMon's no-network rule.
- txcript issue #61: its Cowork harness misses cloud `cse_` sessions; the reporter's workaround reads
  Claude's authenticated session data (an undocumented route, ruled out for BurnMon).
- Documented org route: Cowork OpenTelemetry export (Team and Enterprise, admin only, Admin settings >
  Cowork). Per request it carries model, token counts, `cost_usd` and `session.id`; the exporter runs
  inside the Cowork VM, so the collector must be reachable from Anthropic's cloud; by default events
  also carry prompt and response text unless `otlpContentCapture` is set to `[]` on each device.
  Fits an organization with a collector, not a portable local monitor.
- Conclusion: no tool found that gets cloud usage to a local, offline monitor without a manual step.
  BurnMon's hook-plus-folder-copy is the only local route found; its weak point is the connected folder.

## Next tests (planned)

- Test 10: can the plugin ask for the folder itself? A probe plugin (`reports\burnmon-folder-probe-0.1.0.plugin`,
  not in git) runs `device_request_folder_access` for `~/.burnmon` on UserPromptSubmit. Measured from
  this session: asking again for a granted folder returns at once with `alreadyGranted: true`.
  Pass: one approval prompt in a new session, copies land, no prompt on the next message.
- Test 11: do folders added to a project reach every new session in that project?
- Test 12: does a session started on the web (desktop app open) deliver?
