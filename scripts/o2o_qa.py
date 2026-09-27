#!/usr/bin/env python3
"""Exercise O2O topic capture and review in a real OS PTY, without terminal-library mocks."""
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
    until(master, b"1-2 of 2")
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
    print("PASS: O2O empty state, keyboard and quick topic capture, validation, draft safety, review, full text, resize and reopening")
