# 07: Back up and restore the complete journal

**What to build:** Export a complete JSON backup and restore it into an empty installation so the user can recover the journal with its history, states, and relationships intact.

**Blocked by:** 03: Prepare and save a daily proposal; 05: Record O2O meetings and follow-up actions.

**Status:** ready-for-agent

- [ ] Users can invoke documented local JSON backup and restore actions without a network connection.
- [ ] The backup includes all entries, dates, plans, blockers, daily proposals, tasks, O2O topics, meetings, notes, agreements, actions, states, and their relationships.
- [ ] Restoring into an empty installation reproduces the journal’s observable content and behavior, including historical daily preparation and O2O follow-up tasks.
- [ ] Restored data survives closing and reopening the application.
- [ ] Empty-journal backup and restore work; malformed or unsupported backup input produces a clear error without leaving partially restored data.
- [ ] The supported restore behavior and any limits for non-empty installations are documented; unrelated existing data must not be silently overwritten.
- [ ] A behavioral round-trip test exports representative data and restores it into fresh temporary storage, then verifies it through the public application interface.
