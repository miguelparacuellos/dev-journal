#!/usr/bin/env python3
"""Exercise the built CLI in a real OS PTY, without terminal-library mocks."""
import argparse
import fcntl
import json
import os
import pathlib
import select
import re
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

def launch(path, width=80, height=24, theme="mono", ascii_mode=True, day="2026-09-27"):
    master, slave = os.openpty()
    fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", height, width, 0, 0))
    command = [binary, "--data", str(path), "--theme", theme, "--date", day]
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
    master, process, started = launch(path)
    until(master, b"Ctrl+S")
    os.write(master,b"\x1b")
    time.sleep(0.08)
    os.write(master,b"\t")
    until(master,b"OPEN TASKS")
    os.write(master,b"nReview deployment\x13")
    until(master,b"Saved")
    os.write(master,b"\x1b")
    time.sleep(0.08)
    os.write(master,b"p")
    until(master,b"Selected for today")
    os.write(master,b"\t")
    until(master,b"TODAY'S PLAN")
    os.write(master,b"p")
    until(master,b"u Remove")
    os.write(master,b"u")
    until(master,b"Removed from")
    os.write(master,b"\t")
    until(master,b"OPEN TASKS")
    os.write(master,b"p")
    until(master,b"Selected for today")
    close(master,process)
    planned=subprocess.check_output([binary,"--data",str(path),"--date","2026-09-27","planned"]).decode()
    assert "[open] Review deployment" in planned
    assert "Review deployment" not in subprocess.check_output([binary,"--data",str(path),"--date","2026-09-27","log"]).decode()
    master,process,_=launch(path,120,40,"light",False,day="2026-09-28")
    until(master,b"No actions selected")
    os.write(master,b"\x1b")
    time.sleep(0.08)
    os.write(master,b"\t")
    until(master,b"Review deployment")
    os.write(master,b"p")
    until(master,b"Selected for today")
    os.write(master,b"d")
    until(master,b"Completed")
    os.write(master,b"c")
    until(master,b"COMPLETED TASKS")
    os.write(master,b"n")
    time.sleep(0.08)
    os.write(master,b"\x1b[200~Retained task draft\x1b[201~")
    until(master,b"Retained task draft")
    os.write(master,b"\x1b")
    time.sleep(0.08)
    os.write(master,b"\t")
    until(master,b"Draft retained")
    os.write(master,b"n\x13")
    until(master,b"Saved")
    close(master,process)
    tasks=subprocess.check_output([binary,"--data",str(path),"tasks"]).decode()
    assert "[done] Review deployment" in tasks and "[open] Retained task draft" in tasks
    assert "[done] Review deployment" in subprocess.check_output([binary,"--data",str(path),"--date","2026-09-27","planned"]).decode()
    # Accumulated compact Tasks keeps the action footer visible while navigating.
    for i in range(12):
        subprocess.run([binary,"--data",str(path),"task",f"Accumulated action {i}"],check=True,stdout=subprocess.PIPE)
    master,process,_=launch(path)
    until(master,b"Ctrl+S")
    os.write(master,b"\x1b")
    time.sleep(0.08)
    os.write(master,b"\t")
    until(master,b"Tab Today")
    os.write(master,b"j"*12)
    until(master,b"Accumulated action 11")
    close(master,process)
    for line in subprocess.check_output([binary,"--data",str(path),"tasks"]).decode().splitlines():
        if "[open]" in line:
            subprocess.run([binary,"--data",str(path),"--date","2026-09-27","plan",line.split()[0]],check=True,stdout=subprocess.PIPE)
    master,process,_=launch(path)
    until(master,b"Ctrl+S")
    os.write(master,b"\x1b")
    time.sleep(0.08)
    os.write(master,b"p")
    until(master,b"u Remove")
    os.write(master,b"j"*13)
    until(master,b"Accumulated action 11")
    close(master,process)
    # Full task text is rewrapped when shrinking the actual terminal.
    long_path=pathlib.Path(directory)/"long-task.json"
    subprocess.run([binary,"--data",str(long_path),"task","word "*40+"TAIL"],check=True,stdout=subprocess.PIPE)
    master,process,_=launch(long_path,120,40)
    until(master,b"Ctrl+S")
    os.write(master,b"\x1b")
    time.sleep(0.08)
    os.write(master,b"\t")
    until(master,b"OPEN TASKS")
    os.write(master,b"\r")
    until(master,b"TAIL")
    fcntl.ioctl(master,termios.TIOCSWINSZ,struct.pack("HHHH",24,80,0,0))
    import signal
    process.send_signal(signal.SIGWINCH)
    until(master,b"TAIL")
    close(master,process)
    print("PASS: 80x24 / 120x40 Tasks capture, planning, deselection, completion, date rollover, reopening, draft routing")
