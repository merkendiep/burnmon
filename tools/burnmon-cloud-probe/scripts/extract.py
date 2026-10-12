#!/usr/bin/env python3
"""BurnMon cloud probe, Stop hook (test build 0.1.0).

Reads the hook input on stdin, then the session transcript and its subagent
files, and writes a usage-only extract: one JSON line per model request,
the last row per (requestId, message.id), whitelisted fields only. No
message text, tool input, paths, cwd or branch is ever written.

Outputs, all under /mnt/user-data/outputs/.burnmon/:
  cloud/<session_id>.jsonl   the extract for this session
  cloud/latest.jsonl         the same file, a fixed name for the commit hook
  probe/<session_id>.log     one line per Stop: counts and booleans only

Never fails the turn: every error is logged and the script exits 0.
"""
import glob
import json
import os
import sys
import time

OUT = "/mnt/user-data/outputs/.burnmon"
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
        "source": "burnmon-cloud-probe",
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
    last_text = None
    n_lines = 0
    with open(path, encoding="utf-8", errors="replace") as f:
        for line in f:
            n_lines += 1
            try:
                d = json.loads(line)
            except ValueError:
                continue
            if d.get("type") == "assistant":
                m = d.get("message") or {}
                if isinstance(m, dict):
                    for c in m.get("content") or []:
                        if isinstance(c, dict) and c.get("type") == "text":
                            last_text = c.get("text")
            r = usage_row(d, agent)
            if r is None:
                continue
            # Last row wins: early rows of a streamed request carry partial
            # output counts (measured 2026-10-12: output 8, then 148).
            rows[(r["request_id"], r["message_id"])] = r
    return rows, last_text, n_lines


def main():
    t0 = time.time()
    hook = json.load(sys.stdin)
    tpath = hook.get("transcript_path") or ""
    sid = hook.get("session_id") or "unknown"
    os.makedirs(os.path.join(OUT, "cloud"), exist_ok=True)
    os.makedirs(os.path.join(OUT, "probe"), exist_ok=True)
    log = {"at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()), "session_id": sid,
           "transcript_found": os.path.isfile(tpath)}
    rows = {}
    last_text = None
    if log["transcript_found"]:
        rows, last_text, log["main_lines"] = read_rows(tpath, "main")
        sub = glob.glob(os.path.join(os.path.splitext(tpath)[0], "subagents", "*.jsonl"))
        log["subagent_files"] = len(sub)
        for p in sub:
            agent = os.path.splitext(os.path.basename(p))[0]
            r, _, _ = read_rows(p, agent)
            rows.update(r)
    lam = hook.get("last_assistant_message")
    # Test 2: does the transcript already hold this turn's final answer?
    # Only booleans and lengths are logged, never the text.
    log["has_last_assistant_message"] = isinstance(lam, str)
    log["final_text_in_transcript"] = (isinstance(lam, str) and last_text is not None
                                       and last_text.strip() == lam.strip())
    out = sorted(rows.values(), key=lambda r: (r["ts"] or "", r["request_id"] or ""))
    log["requests"] = len(out)
    log["tokens"] = {k: sum(r[k] for r in out) for k in USAGE_KEYS}
    if out:
        log["newest_row_ts"] = out[-1]["ts"]
    body = "".join(json.dumps(r, separators=(",", ":")) + "\n" for r in out)
    for name in (sid + ".jsonl", "latest.jsonl"):
        tmp = os.path.join(OUT, "cloud", name + ".tmp")
        with open(tmp, "w", encoding="utf-8") as f:
            f.write(body)
        os.replace(tmp, os.path.join(OUT, "cloud", name))
    log["ms"] = int((time.time() - t0) * 1000)
    with open(os.path.join(OUT, "probe", sid + ".log"), "a", encoding="utf-8") as f:
        f.write(json.dumps(log) + "\n")


if __name__ == "__main__":
    try:
        main()
    except Exception as e:  # never break the user's turn
        try:
            os.makedirs(os.path.join(OUT, "probe"), exist_ok=True)
            with open(os.path.join(OUT, "probe", "errors.log"), "a", encoding="utf-8") as f:
                f.write(json.dumps({"at": time.time(), "error": repr(e)[:300]}) + "\n")
        except Exception:
            pass
    sys.exit(0)
