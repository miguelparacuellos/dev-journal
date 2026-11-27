# Records demo/dj.gif with `make demo`.
# Uses the fictional journal in demo/journal (copied to a temp dir) and demo/fake-claude,
# so no real data, Linear or Claude call is involved.

Output demo/dj.gif

Require dj

Set Shell "bash"
Set FontSize 15
Set Width 1400
Set Height 720
Set Padding 18
Set Theme "TokyoNight"
Set TypingSpeed 55ms
Set CursorBlink false

Hide
Type "unset HISTFILE; export PS1='$ ' DJ_HOME=$(mktemp -d)/journal DJ_DEMO_DATE=2026-11-27 DJ_CLAUDE_BIN=$PWD/demo/fake-claude; cp -R demo/journal $DJ_HOME; clear"
Enter
Sleep 500ms
Show

Type "dj"
Sleep 400ms
Enter
Sleep 2s

# Daily: draft from Linear (fake), pick suggestions, insert
Type "g"
Sleep 3s
Type "j"
Sleep 500ms
Type "j"
Sleep 600ms
Type "j"
Sleep 600ms
Enter
Sleep 1.5s

# Add a manual item to Today and tick it
Tab
Sleep 500ms
Type "a"
Sleep 300ms
Type "Pair with QA on ABC-433"
Enter
Sleep 600ms
Type "x"
Sleep 1.5s

# 1:1s
Type "2"
Sleep 2.5s

# Backlog: quick add, then cycle status
Type "3"
Sleep 1.2s
Type "a"
Sleep 300ms
Type "idea: Ticket-to-branch CLI #tooling !S"
Enter
Sleep 1s
Type "s"
Sleep 1.5s

# Wins
Type "4"
Sleep 2s

Type "1"
Sleep 2s
Type "q"
Sleep 300ms
