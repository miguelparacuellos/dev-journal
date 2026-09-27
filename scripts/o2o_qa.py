#!/usr/bin/env python3
"""Exercise O2O topics and the meeting cycle in a real OS PTY, without terminal-library mocks."""
import argparse
import fcntl
import json
import os
import pathlib
import select
import signal
import struct
import subprocess
import tempfile
import termios
import time

parser = argparse.ArgumentParser()
parser.add_argument("binary")
args = parser.parse_args()
binary = str(pathlib.Path(args.binary).resolve())


def launch(path, width=80, height=24, theme="mono", ascii_mode=True, day="2026-09-27"):
    master, slave = os.openpty()
    fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", height, width, 0, 0))
    command = [binary, "--data", str(path), "--theme", theme, "--date", day]
    if ascii_mode:
        command.append("--ascii")
    process = subprocess.Popen(command, stdin=slave, stdout=slave, stderr=slave,
                               env={**os.environ, "TERM": "xterm-256color"})
    os.close(slave)
    return master, process


def until(master, token, timeout=3):
    output = b""
    deadline = time.perf_counter() + timeout
    while time.perf_counter() < deadline:
        if select.select([master], [], [], 0.02)[0]:
            try:
                output += os.read(master, 65536)
            except OSError:
                break
            if token in output:
                return output
    raise AssertionError(f"Expected {token!r}, received {output[-1800:]!r}")


def key(master, data):
    os.write(master, data)
    time.sleep(.08)


def resize(master, process, width, height):
    fcntl.ioctl(master, termios.TIOCSWINSZ, struct.pack("HHHH", height, width, 0, 0))
    process.send_signal(signal.SIGWINCH)


def close(master, process):
    key(master, b"\x1b")
    os.write(master, b"q")
    deadline = time.perf_counter() + 3
    while process.poll() is None and time.perf_counter() < deadline:
        if select.select([master], [], [], .02)[0]:
            try:
                os.read(master, 65536)
            except OSError:
                break
    process.wait(timeout=1)
    assert process.returncode == 0
    os.close(master)


def saved_until(path, check, timeout=3):
    """Poll the saved journal until check passes; saves are durable before feedback."""
    deadline = time.perf_counter() + timeout
    while True:
        saved = json.loads(path.read_text())
        if check(saved):
            return saved
        if time.perf_counter() > deadline:
            raise AssertionError(f"Saved journal never matched: {saved!r}")
        time.sleep(.05)


def cli(path, *command, day="2026-09-27"):
    return subprocess.run([binary, "--data", str(path), "--date", day, *command],
                          check=True, stdout=subprocess.PIPE).stdout.decode()


with tempfile.TemporaryDirectory() as directory:
    path = pathlib.Path(directory) / "journal.json"

    # Empty state, keyboard capture without a meeting, validation, and draft safety at 80x24.
    master, process = launch(path, day="2026-09-03")
    until(master, b"Ctrl+S")
    key(master, b"\x1b")
    os.write(master, b"o")
    until(master, b"Nothing to raise yet")
    os.write(master, b"n\x13")
    until(master, b"Not saved: topic cannot be empty")
    os.write(master, b"Feedback on the incident review\x13")
    until(master, b"Topic saved")
    os.write(master, b"\x1b[200~Proposal: rotate on-call\nwith a written handoff\x1b[201~\x13")
    # Cell-diff redraws can rewrite only part of a range footer; the new row is emitted whole.
    until(master, b"rotate on-call / with a written handoff")
    os.write(master, b"Unfinished thought")
    until(master, b"Unfinished thought")
    key(master, b"\x1b")
    os.write(master, b"t")
    until(master, b"Draft retained")
    os.write(master, b"x")
    until(master, b"discarded")
    # Blockers captured from O2O stay blockers.
    os.write(master, b"bAwaiting access\x13")
    until(master, b"Saved")
    os.write(master, b"\r")
    until(master, b"with a written handoff")
    close(master, process)

    # Quick capture on another day, then review after restarting later in the month.
    assert "stays open" in cli(path, "topic", "Ask about the promotion path " + "word " * 40 + "TOPIC TAIL", day="2026-09-15")
    listing = cli(path, "topics")
    assert "2026-09-03  Feedback on the incident review" in listing and "2026-09-15" in listing
    master, process = launch(path, 120, 40, "light", False, day="2026-09-29")
    until(master, b"Ctrl+S")
    key(master, b"\x1b")
    os.write(master, b"o")
    until(master, b"SELECTED TOPIC")
    os.write(master, b"jj")
    until(master, b"word TOPIC TAIL")
    os.write(master, b"\r")
    until(master, b"TOPIC TAIL")
    resize(master, process, 80, 24)
    until(master, b"TOPIC TAIL")
    key(master, b"\x1b")
    os.write(master, b"\t")
    until(master, b"No tasks here")
    os.write(master, b"o")
    until(master, b"1-3 of 3")
    close(master, process)

    saved = json.loads(path.read_text())
    assert [topic["day"] for topic in saved["topics"]] == ["2026-09-03", "2026-09-03", "2026-09-15"]
    assert saved["topics"][1]["text"] == "Proposal: rotate on-call\nwith a written handoff"
    assert saved["entries"] == [] and "tasks" not in saved
    assert [blocker["text"] for blocker in saved["blockers"]] == ["Awaiting access"]
    # Abandoning an emptied correction must not turn a later topic into an entry correction.
    cli(path, "add", "Original entry", day="2026-09-29")
    master, process = launch(path, day="2026-09-29")
    until(master, b"Ctrl+S")
    key(master, b"\x1b")
    os.write(master, b"e")
    until(master, b"Correct entry")
    os.write(master, b"\x7f" * len("Original entry"))
    until(master, b"hat moved forward?")
    key(master, b"\x1b")
    os.write(master, b"on")
    until(master, b"New O2O topic")
    os.write(master, b"Raised after an abandoned correction\x13")
    until(master, b"Topic saved")
    close(master, process)
    corrected = json.loads(path.read_text())
    assert [entry["text"] for entry in corrected["entries"]] == ["Original entry"]
    assert corrected["topics"][-1]["text"] == "Raised after an abandoned correction"

    # Complete meeting cycle: start, consult, address, notes, agreements, follow-ups, close.
    path = pathlib.Path(directory) / "meeting.json"
    for text in ["Feedback on the incident review", "Promotion path", "Rotate on-call"]:
        cli(path, "topic", text, day="2026-09-03")
    master, process = launch(path, day="2026-09-30")
    until(master, b"Ctrl+S")
    key(master, b"\x1b")
    os.write(master, b"o")
    until(master, b"No meetings recorded yet")
    os.write(master, b"a")
    until(master, b"No meeting in progress")
    os.write(master, b"s")
    until(master, b"Meeting started")
    os.write(master, b"a")
    until(master, b"Addressed in this meeting")
    saved_until(path, lambda d: d["topics"][0].get("addressed_in") == d["meetings"][0]["id"])
    # A topic marked by mistake is reopened with the same key.
    os.write(master, b"ja")
    saved_until(path, lambda d: d["topics"][1].get("addressed_in") == d["meetings"][0]["id"])
    os.write(master, b"a")
    until(master, b"Open again")
    saved_until(path, lambda d: "addressed_in" not in d["topics"][1])
    os.write(master, b"w")
    until(master, b"Meeting notes")
    os.write(master, b"Talked about on-call\x1b[200~\nand the promotion path\x1b[201~\x13")
    until(master, b"Notes saved")
    os.write(master, b"rRotate on-call monthly\x13")
    until(master, b"Agreement recorded")
    os.write(master, b"fDraft the on-call handoff\x13")
    until(master, b"also listed in Tasks")
    # A retained draft blocks closing so nothing written during the meeting is lost.
    os.write(master, b"rUnfinished agreement")
    until(master, b"O2O agreement")
    key(master, b"\x1b")
    until(master, b"Draft retained")
    os.write(master, b"c")
    until(master, b"before closing the meeting")
    os.write(master, b"x")
    until(master, b"discarded")
    os.write(master, b"v")
    until(master, b"2026-09-30 | in progress")
    key(master, b"\x1b")
    os.write(master, b"c")
    until(master, b"Close this meeting?")
    os.write(master, b"j")
    until(master, b"still in progress")
    os.write(master, b"cc")
    until(master, b"kept for the next one")
    saved = saved_until(path, lambda d: d["meetings"][0].get("closed_on") == "2026-09-30")
    # Follow-ups behave like any task: plan and complete them from Tasks.
    os.write(master, b"\t")
    until(master, b"O2O: Draft the on-call handoff")
    os.write(master, b"p")
    until(master, b"Selected for today")
    os.write(master, b"d")
    until(master, b"Completed")
    close(master, process)

    meeting = saved["meetings"][0]
    assert meeting["day"] == "2026-09-30" and meeting["notes"] == "Talked about on-call\nand the promotion path"
    assert [t.get("addressed_in") for t in saved["topics"]] == [meeting["id"], None, None]
    assert [(a["meeting_id"], a["text"]) for a in saved["agreements"]] == [(meeting["id"], "Rotate on-call monthly")]
    final = json.loads(path.read_text())
    assert [(t["meeting_id"], t["text"], t["completed"]) for t in final["tasks"]] == [(meeting["id"], "Draft the on-call handoff", True)]
    assert final["plan"] == [{"day": "2026-09-30", "task_id": final["tasks"][0]["id"]}]
    assert "2 open topics" not in cli(path, "topics") and len(cli(path, "topics").splitlines()) == 2
    record = cli(path, "meetings")
    assert "closed 2026-09-30" in record and "Rotate on-call monthly" in record and "[done] Draft the on-call handoff" in record

    # Reopen later: browse past meetings, read the record across a resize, then meet again.
    master, process = launch(path, 120, 40, "light", False, day="2026-10-28")
    until(master, b"Ctrl+S")
    key(master, b"\x1b")
    os.write(master, b"o")
    until(master, b"last on 2026-09-30")
    os.write(master, b"m")
    until(master, b"Rotate on-call monthly")
    os.write(master, b"\r")
    until(master, b"and the promotion path")
    resize(master, process, 80, 24)
    until(master, b"and the promotion path")
    key(master, b"\x1b")
    os.write(master, b"s")
    until(master, b"Meeting started")
    os.write(master, b"ac")
    until(master, b"Close this meeting?")
    os.write(master, b"c")
    until(master, b"kept for the next one")
    close(master, process)
    final = json.loads(path.read_text())
    assert [m["day"] for m in final["meetings"]] == ["2026-09-30", "2026-10-28"]
    assert [t.get("addressed_in") for t in final["topics"]] == [m["id"] for m in final["meetings"]] + [None]
    assert cli(path, "topics").count("\n") == 1

    print("PASS: O2O empty state, keyboard and quick topic capture, validation, draft safety, review, full text, resize, reopening, correction routing, and the complete meeting cycle with follow-up tasks and history")
