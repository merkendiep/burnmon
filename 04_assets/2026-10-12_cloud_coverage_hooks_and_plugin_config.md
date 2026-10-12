# Claude Code hook and plugin capabilities for cloud-to-device file delivery and per-user config (Oct 2026, v2.1.296)

Sources were fetched live on 2026-10-12 as Markdown from code.claude.com (hooks.md is about 269 KB, read in full by section) and the raw CHANGELOG.md from anthropics/claude-code (top entry is 2.1.296). GitHub issue search through the API was blocked from this session, so I read issues individually through WebFetch. Anything not stated in a source is marked "undocumented".

## 1. mcp_tool hook `input` substitution: which placeholders work, and can devicePath include the session id?

### Takeaway
The only documented substitution in `mcp_tool` `input` is `${path}`, a lookup into the hook's own JSON input (for example `${tool_input.file_path}`). `session_id`, `cwd`, `transcript_path`, `scratchpad_dir` and `hook_event_name` are common input fields, so `${session_id}` fits the documented rule. Three things are undocumented: substitution in the middle of a string (`C:\\...\\${session_id}.jsonl`), substitution inside nested arrays such as `files[0].devicePath`, and env vars or `${user_config.*}` in `input`.

### Cited Findings
- MCP tool hook fields are `server`, `tool` and `input`. For `input`: "String values support `${path}` substitution from the hook's JSON input, such as `"${tool_input.file_path}"`". For a plugin-bundled server, `server` must be the scoped name `plugin:<plugin-name>:<server-name>`. [Hooks reference](https://code.claude.com/docs/en/hooks#mcp-tool-hook-fields)
- The common input fields on every event are `session_id`, `prompt_id`, `transcript_path`, `cwd`, `scratchpad_dir` (v2.1.257+), `permission_mode`, `effort` and `hook_event_name`. Event-specific fields come on top, for example `last_assistant_message` on Stop and `file_path` on FileChanged. [Hooks reference, common input fields](https://code.claude.com/docs/en/hooks#common-input-fields)
- The tool's text output is parsed like command-hook stdout. A result with `isError: true` is a non-blocking error and execution continues. [Hooks reference](https://code.claude.com/docs/en/hooks#how-the-tools-result-is-read)
- On blocking events such as `PreToolUse` or `Stop`, Claude Code waits for a server that is still connecting, up to `MCP_TIMEOUT`. On observational events such as `Notification` or `SessionEnd` it does not wait. If the server is not connected, the hook gives a non-blocking error. [Hooks reference](https://code.claude.com/docs/en/hooks#when-the-server-is-still-connecting)
- Changelog entries: `mcp_tool` hooks were introduced in 2.1.118 ("Hooks can now invoke MCP tools directly via `type: "mcp_tool"`"). In 2.1.281, mcp_tool hooks on blocking events stopped being skipped while their server was connecting. [CHANGELOG](https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md)
- `${user_config.KEY}` is documented as substituted only in "MCP server config, LSP server config, exec-form hook `args`, and skill and agent content". `mcp_tool` `input` is not on that list. [Plugin manifest reference](https://code.claude.com/docs/en/plugins-reference#reference-a-saved-value)
- The table of where plugin `${...}` variables resolve lists hook commands as "Anywhere in `command` and `args`" and has no row for `mcp_tool` input. [Plugin manifest reference, where each variable resolves](https://code.claude.com/docs/en/plugins-reference#where-each-variable-resolves)
- HTTP hooks have their own env-var interpolation in `headers`, limited to the variables listed in `allowedEnvVars`. The docs describe no equivalent for `mcp_tool` `input`. [Hooks reference, HTTP hook fields](https://code.claude.com/docs/en/hooks#http-hook-fields)
- In cloud sessions the hook-input `session_id` comes from `CLAUDE_CODE_SESSION_ID`, which is separate from `CLAUDE_CODE_REMOTE_SESSION_ID` (the `session_`-form id in the claude.ai/code URL). `CLAUDE_CODE_SESSION_ID` "matches the `session_id` field in the hook JSON input and is updated on `/clear`". [Env vars](https://code.claude.com/docs/en/env-vars)

### Inferences
- A Stop command hook can read `session_id` from stdin and write `/mnt/user-data/outputs/.burnmon/cloud/<session_id>.jsonl`. The mcp_tool hook could then reference the same name through `${session_id}` in `stagedPath` and `devicePath`. That stops one session's file from overwriting another's on the device, but it does not fix the race described in Q2.
- A safer pattern on documented ground uses whole-value substitution: on FileChanged (see Q2), `"stagedPath": "${file_path}"` passes the exact staged file. Partial-string substitution such as `...\\${session_id}.jsonl` needs a live test.
- `${session_id}` changes on `/clear` and differs between a session and its forks. If BurnMon wants one file per conversation it should expect several ids per conversation.

### Gaps
- The docs do not say whether `${path}` substitution applies inside a longer string, or inside objects nested in arrays (`input.files[0].devicePath`). The one example is a top-level whole-value string. The caller's measured hook with static strings in a nested array shows that nesting is accepted, not that it is substituted.
- The behaviour for a `${path}` that does not resolve (empty string, literal left in place, or error) is undocumented.
- Whether `${CLAUDE_PLUGIN_ROOT}`, `${CLAUDE_PLUGIN_DATA}`, `${user_config.*}` or `$ENV` are expanded in `mcp_tool` `input` is undocumented, and none of them is on the documented lists.

## 2. Ordering: can hooks within one event run sequentially, and can a command hook make an MCP tool call happen?

### Takeaway
There is no documented setting to run hooks sequentially. All matching hooks for one event run in parallel, which matches what was measured. Ordering is only available across events. The cleanest documented chain is: Stop command hook writes the file, then a FileChanged `mcp_tool` hook on that path (registered through `watchPaths`) commits it. A command hook cannot call an MCP tool itself. It can only make Claude continue the turn (Stop `decision: "block"` or `additionalContext`) so that Claude makes the call.

### Cited Findings
- "All matching hooks run in parallel. If you define the same handler in more than one settings file, it runs once. A plugin's or skill's copy of the same handler stays separate." [Hooks reference, hook handler fields](https://code.claude.com/docs/en/hooks#hook-handler-fields)
- The hooks guide says the same ("Claude Code runs all matching hooks in parallel"). It also warns that when several PreToolUse hooks return `updatedInput`, "the last one to finish takes effect. Since hooks run in parallel, the order is non-deterministic." [Hooks guide](https://code.claude.com/docs/en/hooks-guide)
- Feature request #21533, "[FEATURE] Sequential Hook Execution Option", proposes `"sequential": true` on a matcher group and states "Claude Code currently executes all matching hooks in parallel." I could not read the issue's status or any maintainer reply, and the current docs have no such field. [GitHub issue #21533](https://github.com/anthropics/claude-code/issues/21533)
- Issue #57800 reports a contradiction between "parallel" in the Agent SDK docs and an older hooks-guide passage saying "Hooks run in order, and the first failure blocks the rest". The current hooks guide says parallel. I could not see a resolution. [GitHub issue #57800](https://github.com/anthropics/claude-code/issues/57800)
- `async: true` is "only available on `type: "command"` hooks". It runs the hook in the background, `timeout` is not enforced, and output arrives on the next turn. `asyncRewake: true` wakes Claude on exit code 2 and shows stderr as a system reminder. Neither option orders hooks. [Hooks reference, run hooks in the background](https://code.claude.com/docs/en/hooks#run-hooks-in-the-background)
- FileChanged "Runs when a watched file changes on disk ... it runs the hook no matter what changed the file: ... a script Claude runs with `Bash`, or a process outside Claude Code entirely." The input includes `file_path` (absolute) and `event` (`change`, `add`, `unlink`). [Hooks reference, FileChanged](https://code.claude.com/docs/en/hooks#filechanged)
- The FileChanged matcher registers literal filenames in the working directory and also filters hook groups by basename. Absolute paths elsewhere are added through `watchPaths`, which can be returned from SessionStart ("Array of absolute paths to watch for FileChanged events during this session") or from FileChanged itself. [Hooks reference, SessionStart decision control](https://code.claude.com/docs/en/hooks#sessionstart-decision-control) and [FileChanged output](https://code.claude.com/docs/en/hooks#filechanged-output)
- FileChanged supports `command`, `http` and `mcp_tool` hooks. [Hooks reference, prompt-based hooks](https://code.claude.com/docs/en/hooks#prompt-based-hooks)
- The docs' own example notes that a hook rewriting a watched file triggers FileChanged again ("Claude Code runs the hook again after every rewrite"), so writes made by hook processes are seen. [Hooks reference, FileChanged](https://code.claude.com/docs/en/hooks#filechanged)
- Stop decision control: `decision: "block"` with `reason`, or `hookSpecificOutput.additionalContext`, keeps the conversation going "so Claude can act on it". Stop fires again with `stop_hook_active: true`, and there is an 8-consecutive-continuation cap (`CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`). [Hooks reference, Stop](https://code.claude.com/docs/en/hooks#stop-decision-control)
- `PostToolBatch` fires once "after every tool call in a batch has resolved, before Claude Code sends the next request to the model" and supports `mcp_tool`. [Hooks reference, PostToolBatch](https://code.claude.com/docs/en/hooks#posttoolbatch)
- The transcript "is written asynchronously and may lag the in-memory conversation". Stop hooks should use `last_assistant_message` instead. [Hooks reference, common input fields](https://code.claude.com/docs/en/hooks#common-input-fields)
- Agent hooks (experimental) spawn a subagent that "can use tools like Read, Grep, and Glob". Whether that subagent can call MCP tools such as `device_commit_files` is not stated. [Hooks reference, agent-based hooks](https://code.claude.com/docs/en/hooks#agent-based-hooks)

### Inferences
- Option A, a FileChanged chain (documented building blocks, untested combination). SessionStart command hook returns `{"hookSpecificOutput":{"hookEventName":"SessionStart","watchPaths":["/mnt/user-data/outputs/.burnmon/cloud/latest.jsonl"]}}`. The Stop command hook writes to a temp file and renames it into place so the watcher sees one complete write. A FileChanged group with matcher `latest.jsonl` runs the `mcp_tool` hook with `"stagedPath": "${file_path}"`. Because the commit is triggered by the write, it carries the current turn's data rather than the previous turn's. Risks: FileChanged is observational, so it probably does not wait for a connecting `remote-devices` server. An `unlink` event would also fire the commit and fail harmlessly. Behaviour in cloud Cowork is unverified.
- Option B, a two-pass Stop. On the first Stop (`stop_hook_active` false), the command hook writes the file and returns `additionalContext`. Claude makes one more short response, Stop fires again, and that run's parallel `mcp_tool` hook copies the file written on the first pass. This is documented behaviour but costs an extra model call every turn, which is a poor fit for a token monitor.
- Option C, split events. Extraction runs on `PostToolBatch` or `UserPromptSubmit` and the commit runs on `Stop`. That orders the two, but the transcript lag and the missing final message mean the file still trails the turn slightly.
- Option D, have Claude call `device_commit_files` itself (prompted by Stop `additionalContext` or a skill). Ordering is reliable, but it uses the model and depends on Claude complying.
- A command hook has no documented way to invoke an MCP tool. Hook processes are ordinary subprocesses, and `remote-devices` is a host-provided connector with no local endpoint the script could call.

### Gaps
- No documented `sequential` flag, priority, or ordering between matcher groups or between plugin and settings hooks.
- Whether FileChanged watchers and `watchPaths` work in cloud and Cowork sessions, and whether they fire while a session is idle between turns, is undocumented.
- Whether agent hooks can call MCP tools is undocumented.

## 3. Plugin configuration: does plugin.json support per-user configuration that hooks can read?

### Takeaway
Yes, locally: `userConfig` in plugin.json prompts each user when the plugin is enabled. Exec-form hook `args` receive `${user_config.KEY}`, and every hook process receives `CLAUDE_PLUGIN_OPTION_<KEY>`. Two gaps remain. `mcp_tool` `input` is not a documented substitution target. And values are stored in the user's local `~/.claude/settings.json` (`pluginConfigs`, user or managed scope only), which cloud sessions do not read. Cowork org-plugin docs do not mention userConfig at all.

### Cited Findings
- "`userConfig` declares values Claude Code prompts the user for when the plugin is enabled, so users don't edit `settings.json` themselves." Field types are `string`, `number`, `boolean`, `directory` and `file`. Fields: `title`, `description` (both required), `required`, `default`, `options` (v2.1.271+), `multiple`, `sensitive`, `min`/`max`. Unknown keys fail validation. [Plugin manifest reference, user configuration](https://code.claude.com/docs/en/plugins-reference#user-configuration)
- Each option appears as a row in `/config` (v2.1.269+), except `sensitive` options and `multiple` lists. [Plugin manifest reference](https://code.claude.com/docs/en/plugins-reference#user-configuration)
- Non-sensitive values go under `pluginConfigs` in the user's `settings.json`, and sensitive values go to the OS credential store. [Plugin manifest reference, where values are stored](https://code.claude.com/docs/en/plugins-reference#where-values-are-stored)
- `pluginConfigs` has scope "User or managed". "Claude Code ignores project and local entries because it substitutes these values into plugin hook, MCP, and LSP configurations, and a cloned repository must not be able to supply them. Before v2.1.207, project and local settings were also read." [Settings reference, pluginConfigs](https://code.claude.com/docs/en/settings-reference#pluginconfigs)
- Plugin hooks substitute `${user_config.*}` "in exec form only". A shell-form hook that references it "fails with an error instead of running". Shell-form hooks should read `$CLAUDE_PLUGIN_OPTION_<KEY>` instead. [Hooks reference, exec and shell form](https://code.claude.com/docs/en/hooks#exec-form-and-shell-form); changelog 2.1.207: "`${user_config.*}` in shell-form commands is now rejected (shell-injection fix)". [CHANGELOG](https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md)
- userConfig history: "Plugin options (`manifest.userConfig`) now available externally ... with `sensitive: true` values stored in keychain" (2.1.83). `claude plugin validate` reports undeclared `${user_config.*}` references (2.1.281). [CHANGELOG](https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md)
- Cloud sessions: "Plugins enabled only in your user settings: No". Plugins declared in a repo's `.claude/settings.json` are not installed either, and server-managed settings do carry over. [Cloud environments, what carries over](https://code.claude.com/docs/en/cloud-environments#what-carries-over-from-your-setup)
- In Anthropic-hosted cloud environments, a setup script can write files under `~/.claude/`. The docs show it writing `~/.claude/CLAUDE.md` for personal preferences. [Cloud environments](https://code.claude.com/docs/en/cloud-environments#add-personal-preferences-without-committing-to-the-repo)
- The Cowork org marketplace help article covers install preferences (Installed by default, Available, Not available, Required) and says members "can't edit organization-managed plugins". It does not describe any per-user configuration or userConfig mechanism. [Claude Help Center 13837433](https://support.claude.com/en/articles/13837433)

### Inferences
- For local Claude Code, a `userConfig` option such as `device_target` (type `string`, since `directory` would be validated against the cloud filesystem, not the Windows device) lets one published plugin serve any user. A command hook reads `CLAUDE_PLUGIN_OPTION_DEVICE_TARGET`. The `mcp_tool` hook, though, has no documented way to receive that value.
- Workaround within documented features: have the command hook copy the value into the hook JSON path space. That is not possible, because hook input is generated by Claude Code. The only documented carrier of `${user_config.*}` to an MCP call is the plugin's own MCP server config, and a plugin MCP server runs in the cloud container, not on the device.
- For cloud Cowork the realistic carriers of a per-user target path are: (a) skill content with `${user_config.KEY}`, which is substituted only if the value exists in that environment; (b) server-managed `pluginConfigs` (managed scope), which is org-wide and not per-user; (c) a convention-based path relative to a connected folder; or (d) Claude reading `get_device_info` and choosing the path. Whether the Cowork cloud runtime prompts for or stores userConfig per member is undocumented.

### Gaps
- No source says whether Cowork (desktop or cloud) shows the userConfig dialog for org marketplace plugins, or where it stores the answers.
- No source says whether `${user_config.*}` is substituted in `mcp_tool` `input`. It is not on the documented list, so treat it as unsupported until tested.

## 4. Plugin data dir, SessionStart hooks, and SessionEnd behaviour and timeouts

### Takeaway
`${CLAUDE_PLUGIN_DATA}` is `~/.claude/plugins/data/<id>/`. It persists across plugin updates and is exported to hook processes, but in a cloud VM it lives only as long as that container. SessionStart supports only command and mcp_tool hooks, and its mcp_tool hooks are skipped at launch and resume because MCP is not ready yet. SessionEnd has a 1.5-second shared budget that plugin-provided timeouts cannot raise, which makes it a poor delivery point.

### Cited Findings
- `${CLAUDE_PLUGIN_DATA}` is "`~/.claude/plugins/data/<id>/`, created on first reference and kept across plugin updates", where `<id>` is the plugin id with non `[A-Za-z0-9_-]` characters replaced by `-`. It is deleted on uninstall from the last place the plugin is installed, unless `--keep-data` is used. [Plugin manifest reference, environment variables](https://code.claude.com/docs/en/plugins-reference#environment-variables); added in 2.1.78. [CHANGELOG](https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md)
- Hook processes receive `CLAUDE_PLUGIN_ROOT`, `CLAUDE_PLUGIN_DATA`, `CLAUDE_PROJECT_DIR` and `CLAUDE_PLUGIN_OPTION_<KEY>`. These variables are not present for Bash tool commands. [Plugin manifest reference](https://code.claude.com/docs/en/plugins-reference#where-each-variable-resolves)
- SessionStart matchers are `startup`, `resume` (`--resume`, `--continue`, `/resume`), `clear`, `compact` and `fork`. Only `command` and `mcp_tool` hooks are supported. [Hooks reference, SessionStart](https://code.claude.com/docs/en/hooks#sessionstart)
- "`SessionStart` at launch, including with `--continue` or `--resume`, and every `Setup` event fire before the session's MCP servers are available to hooks. Claude Code skips their `mcp_tool` hooks." They do run when SessionStart fires again after `/clear` or compaction. [Hooks reference](https://code.claude.com/docs/en/hooks#events-that-fire-before-mcp-servers-are-available)
- SessionStart output can include `additionalContext`, `initialUserMessage` (`-p` only), `sessionTitle`, `watchPaths` and `reloadSkills`. SessionStart, Setup, CwdChanged and FileChanged hooks get `CLAUDE_ENV_FILE`. [Hooks reference, SessionStart decision control](https://code.claude.com/docs/en/hooks#sessionstart-decision-control)
- In cloud sessions, SessionStart hooks "run each time a session starts or resumes". A command hook is cancelled after 600 s unless `timeout` is set. Remote sessions keep sending keep-alives during long SessionStart hooks (changelog). [Cloud environments](https://code.claude.com/docs/en/cloud-environments); [CHANGELOG](https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md)
- SessionEnd reasons are `clear`, `resume`, `logout`, `prompt_input_exit` and `other`. It has no decision control and its JSON output is discarded. Default budget 1.5 s. It is raised by per-hook `timeout` in settings files up to 60 s, but "Timeouts set on plugin-provided hooks don't raise the budget". `CLAUDE_CODE_SESSIONEND_HOOKS_TIMEOUT_MS` overrides the budget. [Hooks reference, SessionEnd](https://code.claude.com/docs/en/hooks#sessionend)
- SessionEnd is classed as observational for mcp_tool, so Claude Code does not wait for a connecting MCP server. [Hooks reference](https://code.claude.com/docs/en/hooks#when-the-server-is-still-connecting)
- In `-p` mode, async hooks still running at teardown are killed. [Hooks reference, async](https://code.claude.com/docs/en/hooks#run-hooks-in-the-background)

### Inferences
- For a Cowork cloud session, `CLAUDE_PLUGIN_DATA` is container-local and should be treated as per-session scratch, not durable state. The `/mnt/user-data/outputs` staging path remains the handoff point.
- A plugin's SessionEnd `mcp_tool` delivery runs under 1.5 s with no wait for the server. It cannot be extended from the plugin, because the env var must be set in the environment that launches Claude Code. Do not rely on it as the final flush.
- A SessionStart command hook is the right place to register `watchPaths` for the Option A chain in Q2.

### Gaps
- Whether a cloud Cowork session ever emits SessionEnd (for example when the container is reaped or the user closes the chat) is undocumented.

## 5. Any hook event or setting fired when a session is resumed on another device

### Takeaway
No device-specific hook event exists. The closest signals are SessionStart with `source: "resume"` (or `fork`), the env vars `CLAUDE_CODE_REMOTE`, `CLAUDE_CODE_REMOTE_SESSION_ID` and `CLAUDE_CODE_BRIDGE_SESSION_ID`, and the fact that `--teleport` makes a separate local copy. Viewing or steering a cloud session from another device (web, phone, Remote Control) is not documented to fire any hook.

### Cited Findings
- SessionStart `resume` covers `--resume`, `--continue` and `/resume`. Resumed sessions also get `seconds_since_last_response`, `context_tokens`, `prompt_cache_likely_expired` and `estimated_cache_write_usd` (v2.1.251+). [Hooks reference, SessionStart input](https://code.claude.com/docs/en/hooks#sessionstart-input)
- `--teleport` "pulls a cloud session and its branch" into the terminal. "The terminal gets its own copy of the session: new work there stays local and doesn't appear in the cloud session." `--teleport` is distinct from `--resume`. [Claude Code on the web](https://code.claude.com/docs/en/claude-code-on-the-web)
- `CLAUDE_CODE_REMOTE` is `true` in cloud sessions. `CLAUDE_CODE_REMOTE_SESSION_ID` holds the cloud session id. `CLAUDE_CODE_BRIDGE_SESSION_ID` is set in hook subprocesses while a local session has an active Remote Control connection (v2.1.199+) and removed when the connection ends. [Env vars](https://code.claude.com/docs/en/env-vars); [Hooks reference](https://code.claude.com/docs/en/hooks#hook-handler-fields)
- `PreModelSwitch` input reports `source: "sdk"` for a model change requested from an Agent SDK host or Remote Control. That is the only hook input value I found that relates to Remote Control. [Hooks reference, PreModelSwitch](https://code.claude.com/docs/en/hooks#premodelswitch-input)

### Inferences
- A teleported session runs local hooks on the user's machine, where BurnMon can read the local transcript directly, so cloud delivery is not needed for the teleported part. The cloud part is a separate session id, so session-id-keyed files from Q1 avoid collisions.
- When the user switches device while a cloud session keeps running, nothing fires. BurnMon has to rely on per-turn Stop or FileChanged delivery, not on a "device changed" signal.
- Which device receives `device_commit_files` writes depends on the `remote-devices` connector's linked device at call time. That belongs to the connector, not to Claude Code hooks, and is not covered by these docs.

### Gaps
- No documentation on whether SessionStart `resume` fires when a cloud Cowork session is reopened from a different client or device.
- No documented hook input field that identifies the user's device or the connected folders. Only the `remote-devices` tools (`get_device_info`) expose that, and they are outside the hook system.
