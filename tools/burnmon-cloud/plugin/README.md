# BurnMon cloud

Counts the tokens of Cowork sessions that run in Anthropic's cloud, for
[BurnMon](https://github.com/wilcodetree/burnmon), the local token and cost monitor.

A cloud session writes no transcript to your computer, so BurnMon cannot see it. This
plugin runs inside each cloud session, reads that session's own transcript and keeps only
the usage numbers: time, session and request ids, model and token counts. No prompts, no
answers, no tool input, no paths. The Claude desktop app then copies that file to
`~/.burnmon/cloud` on your computer, and BurnMon shows it as its own row, Cowork (Cloud).
The copy is a plugin hook, so it costs no model tokens.

Status: still under investigation. It only counts a cloud session that has `~/.burnmon`
connected, and Claude does not connect that folder by itself. If you know a better way, please
open an issue at https://github.com/wilcodetree/burnmon/issues.

## Set up

1. Install BurnMon and start it once. It creates `~/.burnmon/cloud` in your home folder
   (on Windows `C:\Users\<you>\.burnmon\cloud`).
2. Install this plugin from the ZND Skills Public marketplace.
3. Connect the folder `~/.burnmon` to the cloud sessions you want counted. In a Cowork
   project, add it once to the project's folders.

## Use it in Cowork only

This plugin is for Cowork cloud sessions. Its copy hook calls the `remote-devices` tool
server, which exists only in a cloud session linked to your computer. Claude Code on your
computer has no such server, and a plugin you install or sync in claude.ai can also load
there. Versions before 3.4.0 then showed this at every stop:

    Stop hook error: MCP server 'remote-devices' not connected

From 3.4.0 the copy runs only when the cloud session's file watcher fires, which never
happens in Claude Code, so the error is gone. On an older version, turn the plugin off for
Claude Code:

    claude plugin disable burnmon-cloud@synced

If you installed it from the marketplace instead, use `burnmon-cloud@znd-skills-public`.

## What it does not cover

- A session without `~/.burnmon` connected is not counted. Neither is a session started on
  the web or on your phone: those cannot reach your computer.
- The numbers are a lower bound: the transcript misses some side calls (0 to 7 percent of
  the main model's tokens on one measured session).
- The end of a turn reaches BurnMon about 10 to 15 seconds after the turn, and a subagent's
  work about as long after the subagent finishes.
- It reads Claude's transcript format, which Anthropic documents as internal and liable to
  change between versions.

Source, tests and the build script: `tools/burnmon-cloud` in the BurnMon repository.
