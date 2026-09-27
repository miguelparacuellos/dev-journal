# 02: Manage tasks and today's plan

**What to build:** Keep a persistent Tasks list and deliberately select actions for today. Clearly separate intended actions from completed work, without automatically carrying unfinished actions into the next day.

**Blocked by:** 01: Capture and browse the daily log.

**Status:** ready-for-agent

- [ ] Users can create tasks without assigning a workday and view open tasks in Tasks.
- [ ] Users can select an open task for today and see it in today’s plan, separately from daily log entries.
- [ ] Selecting a task for today does not mark it completed.
- [ ] Users can mark a task completed and distinguish completed tasks from open tasks.
- [ ] An unfinished task remains open after the day changes but is not automatically added to the new day’s plan; the user can select it again.
- [ ] Tasks, completion states, and dated plan selections survive reopening; the workflows are usable by keyboard with visible actions.
- [ ] Behavior tests through the public application interface verify creation, planning, completion, persistence, and absence of automatic carryover across a date change.
