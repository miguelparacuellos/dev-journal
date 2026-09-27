# 01: Capture and browse the daily log

**What to build:** Open the application in Today, capture and correct entries, browse previous workdays, and retain the daily log across sessions. Provide both keyboard-driven TUI capture and a quick-capture command.

**Blocked by:** None (can start immediately).

**Status:** ready-for-agent

**Completion:** Implementation complete; native terminal visual QA pending.

## Design research

The repository's TUI design research compares exemplary applications and framework options using primary sources. It recommends a calm, capture-first experience and provisionally favors Go with Charm v2, with Textual as the strongest alternative. Treat this as input to a real-terminal evaluation, not a settled stack decision. Visual quality and developer experience are core acceptance requirements for this ticket and must carry through later views.

## Acceptance criteria

- [x] Before committing to a TUI framework or interaction model, complete an intensive primary-source investigation of exemplary TUIs, framework capabilities, and terminal constraints. Capture cited findings and project-specific recommendations in the repository.
- [x] Use that research to compare realistic framework options and explain the selected option in terms of visual quality, keyboard interaction, rendering performance, testability, packaging, and maintenance. Technology remains open until this evaluation is complete.
- [x] Treat visual design and developer experience as core deliverables: define typography through terminal attributes, spacing, semantic colors, focus states, navigation, visible key hints, empty states, and save feedback as a coherent system.
- [x] Validate a representative Today screen and the complete open–capture–save–exit flow in a real terminal before expanding the implementation. Keep research, design validation, and implementation in this ticket rather than creating a disconnected infrastructure task.
- [x] Support an uncluttered default layout at 80×24 and a richer layout at 120×40; preserve access to content and actions when resizing or using longer entries. Below the supported minimum, display a clear size message and preserve saved data.
- [x] Make capture directly reachable from Today without traversing menus. Show how to save, cancel, and exit; preserve line breaks while editing and avoid losing unsubmitted text through accidental navigation.
- [ ] Clearly distinguish focus, selection, completed actions, and blockers using text or attributes as well as color. Check dark and light backgrounds, low-color output, and an ASCII fallback for decorative Unicode.
- [x] Use concise, stable navigation and immediate save feedback. Verify that representative accumulated journal data does not cause visible flicker, stale content, or an unresponsive capture flow.
- [x] Record real-terminal design checks, supported terminal behavior, known limitations, and launch/capture interaction timings. Target a usable opening screen within one second on the development machine and measure the result rather than claiming it without evidence.
- [x] The application launches from the terminal into Today and supports a single local space without a network connection.
- [x] Users can add free-text entries without recording hours and see today’s daily log at a glance.
- [x] Users can browse previous workdays and correct entries without changing their workday association.
- [x] A quick-capture command saves an entry that is visible when the TUI opens.
- [x] Saved entries and corrections survive closing and reopening the application.
- [x] The current view, selected item, available actions, and empty states are clear; capture, navigation, saving, and exit are usable by keyboard.
- [x] All repository content and application UI text use English; document how to launch the application and use quick capture.
- [x] Behavior tests exercise capture, correction, and persistence through the public application interface with temporary local storage; verify keyboard capture and exit in a real terminal.

## Comments

Implementation committed with the complete daily-log capture/browse/correct slice.
Public behavior tests cover persistence, workday-preserving corrections, concurrent
sessions, invalid captures, and failed-save retry. Real PTY checks cover keyboard
capture, bracketed multiline paste, long-entry correction beyond 99 lines, draft
protection, resizing, failed-save retry, and exit. Launch p95 was 86.15 ms over
30 compiled-binary samples, below the one-second requirement. A 10,000-entry
fixture opened in 52.91 ms and saved in 33.92 ms in the final measured run.

Standards review: no actionable findings. Spec review: one P2 found and fixed
(editor default row limit silently dropped newlines in long-entry corrections).
Its regression is verified through the built binary in a real PTY. No scope
creep found. The two review axes remain separate in the validation document.

The unchecked visual-compatibility criterion is deliberate: PTY checks establish
layout and keyboard behavior but native emulator/font/contrast review remains
pending. The computer-use tool rejected root's native Terminal access for safety
reasons. No bypass was attempted. Dark/light roles, monochrome and ASCII modes
are implemented, with representative PTY output checks; native macOS Terminal,
preferred emulator, VS Code and tmux checks are unverified.

The triage status remains from the configured five-state vocabulary; completion
is recorded separately because this tracker defines no completed triage state.
