# 01: Capture and browse the daily log

**What to build:** Open the application in Today, capture and correct entries, browse previous workdays, and retain the daily log across sessions. Provide both keyboard-driven TUI capture and a quick-capture command.

**Blocked by:** None (can start immediately).

**Status:** ready-for-agent

## Design research

The repository's TUI design research compares exemplary applications and framework options using primary sources. It recommends a calm, capture-first experience and provisionally favors Go with Charm v2, with Textual as the strongest alternative. Treat this as input to a real-terminal evaluation, not a settled stack decision. Visual quality and developer experience are core acceptance requirements for this ticket and must carry through later views.

## Acceptance criteria

- [ ] Before committing to a TUI framework or interaction model, complete an intensive primary-source investigation of exemplary TUIs, framework capabilities, and terminal constraints. Capture cited findings and project-specific recommendations in the repository.
- [ ] Use that research to compare realistic framework options and explain the selected option in terms of visual quality, keyboard interaction, rendering performance, testability, packaging, and maintenance. Technology remains open until this evaluation is complete.
- [ ] Treat visual design and developer experience as core deliverables: define typography through terminal attributes, spacing, semantic colors, focus states, navigation, visible key hints, empty states, and save feedback as a coherent system.
- [ ] Validate a representative Today screen and the complete open–capture–save–exit flow in a real terminal before expanding the implementation. Keep research, design validation, and implementation in this ticket rather than creating a disconnected infrastructure task.
- [ ] Support an uncluttered default layout at 80×24 and a richer layout at 120×40; preserve access to content and actions when resizing or using longer entries. Below the supported minimum, display a clear size message and preserve saved data.
- [ ] Make capture directly reachable from Today without traversing menus. Show how to save, cancel, and exit; preserve line breaks while editing and avoid losing unsubmitted text through accidental navigation.
- [ ] Clearly distinguish focus, selection, completed actions, and blockers using text or attributes as well as color. Check dark and light backgrounds, low-color output, and an ASCII fallback for decorative Unicode.
- [ ] Use concise, stable navigation and immediate save feedback. Verify that representative accumulated journal data does not cause visible flicker, stale content, or an unresponsive capture flow.
- [ ] Record real-terminal design checks, supported terminal behavior, known limitations, and launch/capture interaction timings. Target a usable opening screen within one second on the development machine and measure the result rather than claiming it without evidence.
- [ ] The application launches from the terminal into Today and supports a single local space without a network connection.
- [ ] Users can add free-text entries without recording hours and see today’s daily log at a glance.
- [ ] Users can browse previous workdays and correct entries without changing their workday association.
- [ ] A quick-capture command saves an entry that is visible when the TUI opens.
- [ ] Saved entries and corrections survive closing and reopening the application.
- [ ] The current view, selected item, available actions, and empty states are clear; capture, navigation, saving, and exit are usable by keyboard.
- [ ] All repository content and application UI text use English; document how to launch the application and use quick capture.
- [ ] Behavior tests exercise capture, correction, and persistence through the public application interface with temporary local storage; verify keyboard capture and exit in a real terminal.
