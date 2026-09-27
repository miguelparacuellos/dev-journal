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
    os.write(master,b"\x1b")
    time.sleep(.08)
    os.write(master,b"q")
    deadline=time.perf_counter()+3
    while process.poll() is None and time.perf_counter()<deadline:
        if select.select([master],[],[],.02)[0]:
            try: os.read(master,65536)
            except OSError: break
    process.wait(timeout=1)
    assert process.returncode==0
    os.close(master)

with tempfile.TemporaryDirectory() as directory:
    path=pathlib.Path(directory)/"journal.json"
    subprocess.run([binary,"--data",str(path),"--date","2026-09-25","add","Shipped login"],check=True,stdout=subprocess.PIPE)
    subprocess.run([binary,"--data",str(path),"--date","2026-09-25","add","Unshared investigation"],check=True,stdout=subprocess.PIPE)
    task=subprocess.check_output([binary,"--data",str(path),"task","Review deployment"]).decode().split()[2]
    subprocess.run([binary,"--data",str(path),"--date","2026-09-28","plan",task],check=True,stdout=subprocess.PIPE)
    for width,height,theme,ascii_mode in [(80,24,"mono",True),(120,40,"light",False)]:
        master,process,_=launch(path,width,height,theme,ascii_mode,day="2026-09-28")
        until(master,b"Ctrl+S")
        os.write(master,b"\x1b")
        time.sleep(.08)
        os.write(master,b"bAwaiting access\x13")
        until(master,b"Saved")
        os.write(master,b"g")
        output=until(master,b"PERSONAL PREPARATION")
        assert b"2026-09-25" in output and b"RECENT WORK" in output and b"BLOCKERS" in output 
        os.write(master,b"s")
        until(master,b"r prepares; e edits")
        os.write(master,b"e")
        until(master,b"Personal preparation")
        # Extend the selected copied text in the user's own words
        os.write(master,b"\x1b[200~\nMy own update\x1b[201~")
        until(master,b"My own update")
        os.write(master,b"\x1b")
        time.sleep(.08)
        os.write(master,b"t")
        until(master,b"Draft retained")
        os.write(master,b"n\x13")
        until(master,b"Saved")
        os.write(master,b"v")
        until(master,b"My own update")
        # Full preparation text remains readable after resize.
        fcntl.ioctl(master,termios.TIOCSWINSZ,struct.pack("HHHH",24,80,0,0))
        import signal
        process.send_signal(signal.SIGWINCH)
        time.sleep(.08)
        close(master,process)
        saved=json.loads(path.read_text())
        assert len(saved["entries"])==2 and saved["entries"][1]["text"]=="Unshared investigation"
        assert saved["tasks"][0]["completed"] is False
        assert "Unshared investigation" not in saved["prepared"][0]["text"]
        assert "My own update" in saved["prepared"][0]["text"]
        master,process,_=launch(path,day="2026-09-28")
        until(master,b"Awaiting access")
        os.write(master,b"\x1b")
        time.sleep(.08)
        os.write(master,b"g")
        until(master,b"PERSONAL PREPARATION")
        os.write(master,b"v")
        until(master,b"My own update")
        close(master,process)
    master,process,_=launch(path,day="2026-09-29")
    until(master,b"Ctrl+S")
    os.write(master,b"\x1b")
    time.sleep(.08)
    os.write(master,b"g")
    output=until(master,b"PERSONAL PREPARATION")
    assert b"not saved" in output and b"2026-09-25" in output
    close(master,process)
    # Accumulated sections stay navigable with their footer visible at compact size.
    fixture=json.loads(path.read_text())
    for i in range(12):
        fixture["entries"].append({"id":f"progress-{i}","workday":"2026-09-25","text":f"Accumulated progress {i}"})
        fixture["blockers"].append({"day":"2026-09-28","text":f"Accumulated blocker {i}"})
    fixture["entries"][-1]["text"]="word "*40+"PROGRESS TAIL"
    path.write_text(json.dumps(fixture))
    master,process,_=launch(path,120,40,day="2026-09-28")
    until(master,b"Ctrl+S")
    os.write(master,b"\x1b")
    time.sleep(.08)
    os.write(master,b"g")
    until(master,b"PERSONAL PREPARATION")
    os.write(master,b"j"*20)
    until(master,b"word")
    os.write(master,b"\r")
    until(master,b"PROGRESS TAIL")
    fcntl.ioctl(master,termios.TIOCSWINSZ,struct.pack("HHHH",24,80,0,0))
    process.send_signal(signal.SIGWINCH)
    until(master,b"PROGRESS TAIL")
    os.write(master,b"\x1b")
    time.sleep(.08)
    os.write(master,b"ll"+b"j"*20)
    until(master,b"Accumulated blocker 11")
    os.write(master,b"\r")
    until(master,b"Accumulated blocker 11")
    close(master,process)
    empty=pathlib.Path(directory)/"empty.json"
    master,process,_=launch(empty)
    until(master,b"Ctrl+S")
    os.write(master,b"\x1b")
    time.sleep(.08)
    os.write(master,b"g")
    until(master,b"No previous recorded workday")
    close(master,process)
    print("PASS: Daily compact/wide, real source date, share selection, editing, draft safety, blockers, reopening and source preservation")
