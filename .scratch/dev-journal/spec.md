# Dev Journal — first-version specification

Status: ready-for-agent

## Problem Statement

The user regularly works in the terminal and needs to remember progress without interrupting the workday. They want one accessible place for completed work, intended actions, and topics to discuss with their tech lead. Retrieving this information should simplify the daily and preparation for the monthly O2O without requiring a complex system to maintain.

## Solution

A visually clear terminal application operated by keyboard for opening, capturing, and closing. The first version provides Today, Daily, O2O, and Tasks in a single space, with local data and offline operation.

Today combines the daily log and today's plan. Daily helps prepare what to share using the last recorded workday, today's plan, and blockers. O2O collects topics, supports meeting notes, and preserves agreements and actions. Data can be exported as readable Markdown or recoverable JSON.

## User Stories

1. As a user, I want to open the application from the terminal, so that it fits my regular workday.
2. As a user, I want to land directly in Today, so that I can capture an entry immediately.
3. As a user, I want to add an entry through a quick command, so that I can record progress without opening the full TUI.
4. As a user, I want to record progress and outcomes in free text, so that I can avoid unnecessary forms.
5. As a user, I want to save without recording hours, so that maintaining the journal takes little effort.
6. As a user, I want my entries to survive closing and reopening, so that I can trust the journal.
7. As a user, I want to see the daily log at a glance, so that I can remember what happened.
8. As a user, I want to browse previous workdays, so that I can retrieve recent work.
9. As a user, I want to correct entries, so that my record remains accurate.
10. As a user, I want to distinguish completed work from today's plan, so that I do not present intentions as outcomes.
11. As a user, I want to record blockers, so that I remember matters requiring help.
12. As a user, I want to open a Daily view, so that I can quickly prepare my update.
13. As a user, I want to consult the latest previous workday with entries, so that I can prepare the daily after weekends or absences.
14. As a user, I want to see that workday's date, so that I know when the displayed work happened.
15. As a user, I want to select the progress I will share, so that I can keep my update brief.
16. As a user, I want to edit my daily proposal, so that I can express it in my own words.
17. As a user, I want recent work, today's plan, and blockers to be clearly separated, so that I can follow my update during the meeting.
18. As a user, I want to preserve my daily preparation, so that I can retrieve it after reopening.
19. As a user, I want to save tasks without assigning a workday, so that I can remember actions for later.
20. As a user, I want to select tasks for today's plan, so that I can choose my priorities deliberately.
21. As a user, I want unfinished actions to remain open without automatically moving to the next day, so that plans do not accumulate.
22. As a user, I want to mark tasks as completed, so that I can distinguish them from remaining work.
23. As a user, I want to collect O2O topics throughout the month, so that I do not forget them at the meeting.
24. As a user, I want to review open topics when preparing the O2O, so that I can organize what to discuss.
25. As a user, I want to take notes during an O2O, so that I can preserve the conversation.
26. As a user, I want to mark topics as addressed, so that I can distinguish them from open topics.
27. As a user, I want to close an O2O while preserving notes and agreements, so that I can consult them later.
28. As a user, I want unaddressed topics to remain available, so that I can bring them to another meeting.
29. As a user, I want O2O actions to become tasks, so that I can follow up on them.
30. As a user, I want to consult past meetings, so that I can recall agreements and proposals.
31. As a user, I want keyboard navigation with visible actions, so that I can use the application without guessing shortcuts.
32. As a user, I want to identify the current view and distinguish selection, content, and actions, so that I can orient myself during brief sessions.
33. As a user, I want to store data locally and work offline, so that I do not depend on external services.
34. As a user, I want to export the journal as Markdown, so that I can read or use it outside the application.
35. As a user, I want to export and restore a complete JSON backup, so that I can preserve my data when changing computers or installations.

## Implementation Decisions

- This is a new project: only design and glossary documents exist; there is no implementation or prior test infrastructure.
- The first version has four views: Today, Daily, O2O, and Tasks. Quick capture complements the TUI.
- A single space can contain both work and personal matters.
- Distinguish entries, tasks, today's plan, blockers, daily proposals, O2O topics, and O2O agreements.
- The daily's recent-work reference is the latest workday with entries before today, with its date visible. When no previous workday exists, show a clear empty state.
- Preparing the daily allows selecting and editing content without changing the original entries.
- Selecting a task for today does not complete it. An unfinished task remains available and is not automatically added to the following day's plan.
- Closing an O2O preserves its history and leaves unaddressed topics open. Resulting actions are managed as tasks.
- Data persists locally. Provide Markdown export and a complete, recoverable JSON backup that preserves relationships between journal concepts.
- For future automation, preserve the conceptual separation between a generated draft and personal preparation. Regenerating a draft must not overwrite the user's edits. The integration and its contracts are outside this version's scope.
- Language, terminal library, storage technology, commands, and shortcuts are implementation choices. No technology has been agreed upon.
- All repository content and application UI text use English.
- Visual design and developer experience are first-class requirements. The first implementation ticket includes intensive primary-source TUI research and validation of a representative Today screen and capture flow before the design is extended to other views.
- Use a coherent visual system for spacing, semantic colors, terminal text attributes, focus, selection, key hints, empty states, and save feedback. Keep the default screen uncluttered and capture directly accessible.
- Validate a compact layout at 80×24 and a richer layout at 120×40, resizing, long entries, light and dark backgrounds, low-color output, and decorative Unicode fallback. State and actions must remain understandable without color alone.
- Measure startup and interaction behavior in the development environment; target a usable initial screen within one second and immediate save feedback without visible flicker.

## Testing Decisions

Approach validated by the user:

- Test observable behavior through a public application interface shared by the TUI and quick capture, using temporary local storage. Prefer this single boundary for verifying workflows over tests of internal functions.
- Verify capture, persistence after reopening, and entry correction.
- Verify daily preparation, including Mondays, absences, and missing previous workdays; ensure proposal edits do not change the original log.
- Verify that selecting an action for today does not complete it and that open tasks do not carry over automatically.
- Verify the O2O cycle: open topics, notes, closing, agreements, actions, and history; ensure unaddressed topics remain open.
- Verify that JSON exports can be restored while preserving data and relationships, and that Markdown represents the content readably.
- Add a real terminal interaction check for keyboard navigation, capture, and exit. Review visual states in a real terminal as well.
- There is no prior code or comparable test suite to use as a reference. Tests should not assert internal structures or details of the chosen library.
- Real-terminal review includes the open–capture–save–exit flow, visible keyboard help and focus, supported sizes, resizing, color fallbacks, long content, and representative accumulated data. Record measured launch behavior and any supported-terminal limitations.

## Out of Scope

- The Linear skill, Linear connection, and automatic proposal generation.
- Scheduled execution each morning.
- Resources, training, and a dedicated learning catalog.
- Multiple spaces, synchronization, accounts, and collaboration.
- Timers and mandatory time tracking.
- Automatic carryover of open tasks into today's plan.

## Further Notes

The priority is the daily log and daily, followed by O2O, tasks, and, in future versions, training and resources. Visual comfort and fast capture are product requirements.

The user delegated the initial capture design to be refined through use. Agreed terminology is recorded in the project glossary. There are no existing ADRs.

Published in the project's local Markdown tracker with status `ready-for-agent`. The user validated the testing approach.
