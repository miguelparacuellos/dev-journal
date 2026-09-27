#!/usr/bin/env python3
"""Exercise the Markdown export through the built binary and its help in a real OS PTY."""
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


def cli(path, *command, day="2026-09-27", check=True):
    result = subprocess.run([binary, "--data", str(path), "--date", day, *command],
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

    # An empty journal exports a short document that says so, without creating the journal.
    assert "Markdown export saved" in cli(path, "export", str(directory / "empty.md"))
    assert "The journal is empty" in (directory / "empty.md").read_text()
    assert not path.exists()
    assert "use export FILE.md" in cli(path, "export", check=False)

    cli(path, "add", "Reviewed the payments PR", day="2026-09-24")
    cli(path, "add", "Fixed login", day="2026-09-25")
    task = cli(path, "task", "Ship the login fix").split()[-1]
    cli(path, "plan", task, day="2026-09-25")
    cli(path, "complete", task)
    cli(path, "topic", "Team rotation")
    before = path.read_bytes()

    exported = directory / "journal.md"
    assert "the journal is unchanged" in cli(path, "export", str(exported))
    markdown = exported.read_text()
    for part in ["# Dev Journal", "Exported on 2026-09-27", "### 2026-09-25", "- Fixed login",
                 "- [x] Done: Ship the login fix", "### Completed tasks (1)",
                 "## Open O2O topics (1)", "- Team rotation (collected 2026-09-27)", "## O2O meetings (0)"]:
        assert part in markdown, f"export lacks {part!r}:\n{markdown}"
    assert markdown.index("### 2026-09-25") < markdown.index("### 2026-09-24")
    assert cli(path, "export", "-") == markdown
    assert "cannot replace the journal" in cli(path, "export", str(path), check=False)
    assert path.read_bytes() == before, "exporting changed the stored journal"

    # The help screen names the export command within the compact layout.
    master, slave = os.openpty()
    fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", 24, 80, 0, 0))
    process = subprocess.Popen([binary, "--data", str(path), "--theme", "mono", "--ascii", "--date", "2026-09-27"],
                               stdin=slave, stdout=slave, stderr=slave, env={**os.environ, "TERM": "xterm-256color"})
    os.close(slave)
    until(master, b"Ctrl+S")
    os.write(master, b"\x1b")
    time.sleep(.08)
    os.write(master, b"?")
    until(master, b"devjournal export FILE.md")
    os.write(master, b"\x1b")
    time.sleep(.08)
    os.write(master, b"q")
    process.wait(timeout=3)
    assert process.returncode == 0
    os.close(master)
    assert path.read_bytes() == before

    print("PASS: Markdown export of empty and populated journals to a file and stdout, journal unchanged, self-overwrite refused, help names the command at 80x24")
