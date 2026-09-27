# Issue tracker: Local Markdown

Issues and specifications live in `.scratch/`.

- One directory per feature: `.scratch/<feature-slug>/`.
- Specification: `.scratch/<feature-slug>/spec.md`.
- Tickets: `.scratch/<feature-slug>/issues/<NN>-<slug>.md`, numbered from 01, one file per ticket.
- State: a `Status:` line near the top of the file. Values are defined in `triage-labels.md`.
- Comments: append at the bottom under `## Comments`.

Publishing means creating the corresponding file.
Fetching a ticket means reading its file.
