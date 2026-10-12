#!/usr/bin/env python3
"""BurnMon cloud, extract hook (1.1.0).

Runs inside a cloud Cowork session. Reads the hook input on stdin, then the
session transcript and its subagent files, and writes a usage-only extract:
one JSON line per model request, the last row per (requestId, message.id),
whitelisted fields only. No message text, tool input, paths, cwd or branch is
ever written.

Modes:
  (no flag)    extract now (Stop, SubagentStop)
  --throttle   extract, unless latest.jsonl was written less than 20 s ago
               (PostToolUse, which fires after every tool call)
  --watch      SessionStart: print the watchPaths output that makes the
               plugin's FileChanged hook fire on every write of latest.jsonl,
               then exit; no extract. Prints nothing outside a cloud session.

Outputs, under /mnt/user-data/outputs/.burnmon/cloud/:
  <session_id>.jsonl   the extract for this session
  latest.jsonl         the same rows under a fixed name; the plugin's copy
                       hooks send this file to BurnMon's drop folder on the
                       user's computer
  errors.log           only when something failed

Never fails the turn: every error is logged and the script exits 0.
"""
import glob
import json
import os
import sys
import time

__version__ = "1.1.0"

OUT = "/mnt/user-data/outputs/.burnmon/cloud"
CLOUD_ENTRYPOINT = "remote_cowork"
THROTTLE_SECONDS = 20
USAGE_KEYS = (
    "input_tokens",
    "output_tokens",
    "cache_creation_input_tokens",
    "cache_read_input_tokens",
)


def usage_row(d, agent):
    m = d.get("message") or {}
    if not isinstance(m, dict):
        return None
    u = m.get("usage")
    if not isinstance(u, dict):
        return None
    cc = u.get("cache_creation") or {}
    otd = u.get("output_tokens_details") or {}
    row = {
        "v": 1,
        "source": "burnmon-cloud",
        "ts": d.get("timestamp"),
        "session_id": d.get("sessionId"),
        "request_id": d.get("requestId"),
        "message_id": m.get("id"),
        "model": m.get("model"),
        "agent": agent,
        "is_sidechain": bool(d.get("isSidechain")),
        "entrypoint": d.get("entrypoint"),
        "service_tier": u.get("service_tier"),
        "speed": u.get("speed"),
        "cache_5m": cc.get("ephemeral_5m_input_tokens"),
        "cache_1h": cc.get("ephemeral_1h_input_tokens"),
        "thinking_tokens": otd.get("thinking_tokens"),
    }
    for k in USAGE_KEYS:
        v = u.get(k)
        row[k] = v if isinstance(v, (int, float)) else 0
    return row


def read_rows(path, agent):
    rows = {}
    with open(path, encoding="utf-8", errors="replace") as f:
        for line in f:
            try:
                d = json.loads(line)
            except ValueError:
                continue
            r = usage_row(d, agent)
            if r is None:
                continue
            # Last row wins: early rows of a streamed request carry partial
            # output counts (measured 2026-10-12: output 8, then 148).
            rows[(r["request_id"], r["message_id"])] = r
    return rows


def main_transcript(tpath):
    # The docs give the main transcript as transcript_path, also on
    # SubagentStop and on tool calls inside a subagent. If a subagent file
    # (<session>/subagents/<agent>.jsonl) ever arrives instead, go up to the
    # session's own file, so latest.jsonl never shrinks to one subagent.
    sub_dir = os.path.dirname(tpath)
    if os.path.basename(sub_dir) == "subagents":
        main = os.path.dirname(sub_dir) + ".jsonl"
        if os.path.isfile(main):
            return main
    return tpath


def watch():
    if os.environ.get("CLAUDE_CODE_ENTRYPOINT") != CLOUD_ENTRYPOINT:
        return
    out = {"hookSpecificOutput": {
        "hookEventName": "SessionStart",
        "watchPaths": [os.path.join(OUT, "latest.jsonl")],
    }}
    sys.stdout.write(json.dumps(out, separators=(",", ":")) + "\n")


def recently_written():
    try:
        age = time.time() - os.path.getmtime(os.path.join(OUT, "latest.jsonl"))
    except OSError:
        return False
    return age < THROTTLE_SECONDS


def extract():
    if os.environ.get("CLAUDE_CODE_ENTRYPOINT", CLOUD_ENTRYPOINT) != CLOUD_ENTRYPOINT:
        return
    hook = json.load(sys.stdin)
    tpath = hook.get("transcript_path") or ""
    sid = hook.get("session_id") or "unknown"
    if not os.path.isfile(tpath):
        return
    tpath = main_transcript(tpath)
    rows = read_rows(tpath, "main")
    for p in glob.glob(os.path.join(os.path.splitext(tpath)[0], "subagents", "*.jsonl")):
        rows.update(read_rows(p, os.path.splitext(os.path.basename(p))[0]))
    # Cloud sessions only. A local Cowork or Claude Code session already
    # leaves its own transcript on the user's computer, and BurnMon reads
    # that; sending it again would count it twice under Cowork (Cloud).
    out = sorted((r for r in rows.values() if r["entrypoint"] == CLOUD_ENTRYPOINT),
                 key=lambda r: (r["ts"] or "", r["request_id"] or ""))
    if not out:
        return
    body = "".join(json.dumps(r, separators=(",", ":")) + "\n" for r in out)
    os.makedirs(OUT, exist_ok=True)
    # latest.jsonl last: its write is what fires the FileChanged copy, so the
    # session file is complete by then. Temp file plus rename, so the watcher
    # sees one finished write and never a half file.
    for name in (sid + ".jsonl", "latest.jsonl"):
        tmp = os.path.join(OUT, name + "." + str(os.getpid()) + ".tmp")
        with open(tmp, "w", encoding="utf-8") as f:
            f.write(body)
        os.replace(tmp, os.path.join(OUT, name))


def main(argv):
    flags = set(argv)
    unknown = flags - {"--watch", "--throttle"}
    if unknown:
        raise ValueError("unknown flag " + " ".join(sorted(unknown)))
    if "--watch" in flags:
        watch()
        return
    if "--throttle" in flags and recently_written():
        return
    extract()


def run(argv):
    try:
        main(argv)
    except Exception as e:  # never break the user's turn
        try:
            os.makedirs(OUT, exist_ok=True)
            with open(os.path.join(OUT, "errors.log"), "a", encoding="utf-8") as f:
                f.write(json.dumps({"at": time.time(), "v": __version__,
                                    "error": repr(e)[:300]}) + "\n")
        except Exception:
            pass


if __name__ == "__main__":
    run(sys.argv[1:])
    sys.exit(0)
