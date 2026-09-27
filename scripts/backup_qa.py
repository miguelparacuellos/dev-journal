#!/usr/bin/env python3
"""Exercise JSON backup and restore through the built binary and its help in a real OS PTY."""
import argparse
import fcntl
import os
import pathlib
import select
import struct
import subprocess
import tempfile
import termios
import time

parser = argparse.ArgumentParser()
parser.add_argument("binary")
args = parser.parse_args()
binary = str(pathlib.Path(args.binary).resolve())


def cli(path, *command, day="2026-09-27", check=True, stdin=None):
    result = subprocess.run([binary, "--data", str(path), "--date", day, *command], input=stdin,
                            check=check, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    return result.stdout.decode() + result.stderr.decode()


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


with tempfile.TemporaryDirectory() as directory:
    directory = pathlib.Path(directory)
    path = directory / "journal.json"

    # An empty journal backs up and restores without creating the source journal.
    assert "JSON backup saved" in cli(path, "backup", str(directory / "empty.json"))
    assert not path.exists()
    assert "Journal restored into" in cli(directory / "empty-restored.json", "restore", str(directory / "empty.json"))
    assert "The journal is empty" in cli(directory / "empty-restored.json", "export", "-")
    assert "use backup FILE.json" in cli(path, "backup", check=False)
    assert "use restore FILE.json" in cli(path, "restore", check=False)

    # Representative data the CLI can record; journal tests cover meetings and proposals.
    cli(path, "add", "Reviewed the payments PR", day="2026-09-24")
    cli(path, "add", "Fixed login", day="2026-09-25")
    task = cli(path, "task", "Ship the login fix").split()[-1]
    cli(path, "plan", task, day="2026-09-25")
    cli(path, "complete", task)
    cli(path, "topic", "Team rotation")
    before = path.read_bytes()

    backup = directory / "backup.json"
    assert "the journal is unchanged" in cli(path, "backup", str(backup))
    assert '"format": "devjournal-backup"' in backup.read_text()
    assert cli(path, "backup", "-") == backup.read_text()
    assert "cannot replace the journal" in cli(path, "backup", str(path), check=False)
    assert path.read_bytes() == before, "backing up changed the stored journal"

    # Restoring into a fresh installation reproduces what every command shows.
    fresh = directory / "fresh" / "journal.json"
    fresh.parent.mkdir()
    assert f"Journal restored into {fresh}" in cli(fresh, "restore", str(backup))
    for command in [("export", "-"), ("tasks",), ("topics",), ("meetings",), ("planned",)]:
        assert cli(fresh, *command, day="2026-09-25") == cli(path, *command, day="2026-09-25"), command
    assert "Fixed login" in cli(fresh, "log", day="2026-09-25")

    # Standard input works too, and a restore never overwrites existing records.
    piped = directory / "piped.json"
    cli(piped, "restore", "-", stdin=backup.read_bytes())
    assert cli(piped, "export", "-") == cli(path, "export", "-")
    restored = fresh.read_bytes()
    assert "restore needs an empty journal" in cli(fresh, "restore", str(backup), check=False)
    assert fresh.read_bytes() == restored

    # Malformed and unsupported input fails clearly and restores nothing.
    rejected = directory / "rejected.json"
    (directory / "bad.json").write_text('{"format": "devjournal-backup", "version": 1, "backed_up_on": "2026-09-27", "tasks": [{"id": "t1", "text": "Ship", "meeting_id": "gone"}]}')
    assert "backup is inconsistent" in cli(rejected, "restore", str(directory / "bad.json"), check=False)
    assert "not a Dev Journal backup" in cli(rejected, "restore", str(path), check=False)
    cli(path, "export", str(directory / "journal.md"))
    assert "backup cannot be read" in cli(rejected, "restore", str(directory / "journal.md"), check=False)
    assert "backup cannot be read" in cli(rejected, "restore", str(directory / "missing.json"), check=False)
    assert "backup cannot be read" in cli(rejected, "restore", "-", stdin=b"{", check=False)
    assert not rejected.exists()

    # The restored journal opens in the TUI, and the help names both commands at 80x24.
    master, slave = os.openpty()
    fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", 24, 80, 0, 0))
    process = subprocess.Popen([binary, "--data", str(fresh), "--theme", "mono", "--ascii", "--date", "2026-09-25"],
                               stdin=slave, stdout=slave, stderr=slave, env={**os.environ, "TERM": "xterm-256color"})
    os.close(slave)
    until(master, b"Fixed login")
    os.write(master, b"\x1b")
    time.sleep(.08)
    os.write(master, b"?")
    until(master, b"backup/restore FILE.json")
    os.write(master, b"\x1b")
    time.sleep(.08)
    os.write(master, b"q")
    process.wait(timeout=3)
    assert process.returncode == 0
    os.close(master)
    assert fresh.read_bytes() == restored

    print("PASS: JSON backup and restore of empty and populated journals via files and pipes, matching CLI output after restore, non-empty and invalid restores refused without changes, TUI opens the restored journal, help names the commands at 80x24")
