# Dev Journal design

## Agreements

- Priority: daily log and daily, followed by O2O, tasks, and training or resources.
- A single space for work and personal activities in the first version.
- Local, exportable storage; synchronization is not required.
- Brief terminal sessions: open, capture, and close.
- The initial capture design is delegated to implementation and will evolve through use.
- The interface must be visually clear and comfortable to use.
- Record progress and outcomes without a timer or mandatory hours.
- The daily shows the last recorded workday, today, and blockers; users can select and edit what to share, with the source date visible.
- Today's plan is distinct from completed work. Open tasks do not carry over automatically; users can select them again.
- O2O: collect topics, take meeting notes, and preserve agreements and actions on closing. Unaddressed topics remain available.
- Future goal: a predefined skill checks Linear each morning for tickets worked on during the previous workday and updates an automatically generated daily proposal.
- The first version includes neither the skill nor the Linear connection.
- The generated draft and the user's prepared version remain separate. Regeneration does not overwrite personal preparation; the user chooses what to incorporate.
- First version: Today, Daily, O2O, and a simple Tasks list. Resources and training are deferred to a later iteration.
- Markdown export for reading and external use; JSON for a complete, recoverable backup.
- All repository content and application UI text use English.
- Visual design and developer experience are central product priorities. Ticket 01 includes intensive TUI research, a coherent visual and interaction system, and real-terminal validation of the capture-first Today screen before expanding to other views.
- Evaluate compact and spacious layouts, keyboard discoverability, focus, safe editing, color and Unicode fallbacks, startup time, and rendering responsiveness. Research findings will inform the framework selection.

## Initial capture design

Primary-source findings and provisional recommendations are recorded in [TUI design research](research/tui-design.md). Framework choice and detailed interaction proposals require real-terminal validation; no stack has been selected yet.

The capture design was delegated and will evolve through use: open the TUI in Today and also provide a quick-capture command. Navigation uses the keyboard, with visible actions and saving that allows an immediate exit.

## Status

The scope interview is complete. The specification is published in the local tracker with status `ready-for-agent`. Technical decisions and interaction details will be resolved during implementation within these agreements.
