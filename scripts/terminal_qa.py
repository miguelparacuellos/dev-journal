#!/usr/bin/env python3
"""Exercise the built CLI in a real OS PTY, without terminal-library mocks."""
import argparse
import fcntl
import json
import os
import pathlib
import select
import statistics
import struct
import subprocess
import tempfile
import termios
import time

parser = argparse.ArgumentParser()
parser.add_argument("binary")
args = parser.parse_args()
binary = str(pathlib.Path(args.binary).resolve())

def launch(path, width=80, height=24, theme="mono", ascii_mode=True):
    master, slave = os.openpty()
    fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", height, width, 0, 0))
    command = [binary, "--data", str(path), "--theme", theme]
    if ascii_mode:
        command.append("--ascii")
    started = time.perf_counter()
    process = subprocess.Popen(command, stdin=slave, stdout=slave, stderr=slave,
                               env={**os.environ, "TERM": "xterm-256color"})
    os.close(slave)
    return master, process, started

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

def close(master, process):
    os.write(master, b"\x1b")
    time.sleep(0.05)
    os.write(master, b"q")
    process.wait(timeout=3)
    assert process.returncode == 0
    os.close(master)

with tempfile.TemporaryDirectory() as directory:
    path = pathlib.Path(directory) / "journal.json"
    launches = []
    for _ in range(30):
        master, process, started = launch(path)
        until(master, b"Ctrl+S")
        launches.append((time.perf_counter() - started) * 1000)
        close(master, process)
    master, process, _ = launch(path)
    until(master, b"Ctrl+S")
    # Bracketed paste includes newline and command-like letters; it must stay text.
    os.write(master, "\x1b[200~Fixed login\nEspaña 日本語 q?\x1b[201~".encode())
    until(master, "Espa".encode())
    started = time.perf_counter()
    os.write(master, b"\x13")
    until(master, b"Saved")
    save_ms = (time.perf_counter() - started) * 1000
    close(master, process)
    log = subprocess.check_output([binary, "--data", str(path), "log"]).decode()
    assert "Fixed login\nEspaña 日本語 q?" in log
    # Correct the selected entry and retain the original workday.
    master, process, _ = launch(path, 120, 40, "light", False)
    until(master, b"SELECTED ENTRY")
    os.write(master, b"\x1b")
    time.sleep(0.08)
    os.write(master, b"e")
    until(master, b"Correct entry")
    os.write(master, b" corrected\x13")
    until(master, b"Saved")
    # Resize down while a draft is pending and restore; navigation never discards it.
    os.write(master, b"Keep draft")
    until(master, b"Keep draft")
    fcntl.ioctl(master, termios.TIOCSWINSZ, struct.pack("HHHH", 18, 60, 0, 0))
    import signal
    process.send_signal(signal.SIGWINCH)
    until(master, b"Resize to at least")
    fcntl.ioctl(master, termios.TIOCSWINSZ, struct.pack("HHHH", 24, 80, 0, 0))
    process.send_signal(signal.SIGWINCH)
    until(master, b"Keep draft")
    os.write(master, b"\x1b")
    time.sleep(0.08)
    os.write(master, b"q")
    until(master, b"Unsaved draft")
    os.write(master, b"xq")
    process.wait(timeout=3)
    os.close(master)
    failure_parent = pathlib.Path(directory) / "unavailable"
    failure_path = failure_parent / "journal.json"
    master, process, _ = launch(failure_path)
    until(master, b"Ctrl+S")
    failure_parent.write_text("Unavailable destination")
    os.write(master, b"Do not lose this")
    until(master, b"Do not lose this")
    os.write(master, b"\x13")
    until(master, b"Not saved")
    failure_parent.unlink()
    os.write(master, b"\x13")
    until(master, b"Saved")
    close(master, process)
    assert "Do not lose this" in subprocess.check_output([binary, "--data", str(failure_path), "log"]).decode()
    # Representative accumulated journal fixture for render/load performance.
    fixture = {"version": 1, "entries": [
        {"id": f"fixture-{i}", "workday": "2026-09-26", "text": "A long historical outcome with Unicode 日本語 and a follow-up\nSecond line"}
        for i in range(10000)]}
    path.write_text(json.dumps(fixture))
    master, process, started = launch(path, 120, 40, "dark", False)
    until(master, b"Ctrl+S")
    large_ms = (time.perf_counter() - started) * 1000
    os.write(master, b"Fast capture")
    until(master, b"Fast capture")
    started = time.perf_counter()
    os.write(master, b"\x13")
    until(master, b"Saved")
    large_save_ms = (time.perf_counter() - started) * 1000
    close(master, process)
    p95 = sorted(launches)[28]
    assert p95 < 1000, f"Opening exceeded budget: {p95}ms"
    print(json.dumps({"samples": 30, "launch_median_ms": round(statistics.median(launches), 2),
                      "launch_p95_ms": round(p95, 2), "save_ms": round(save_ms, 2),
                      "large_launch_ms": round(large_ms, 2), "large_save_ms": round(large_save_ms, 2)}, indent=2))
