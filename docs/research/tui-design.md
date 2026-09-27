# Dev Journal TUI design research

Research date: 2026-09-27. Scope: an exceptional visual and keyboard experience for brief visits to a local journal. This document records research and recommendations, not an agreed technology decision or implementation. The repository currently has a specification and glossary but no application. The priority remains daily log and Daily, followed by O2O and Tasks.

## Findings and recommendation

Recommend prototyping **Go with Bubble Tea v2, Lip Gloss v2, and Bubbles v2**, then deciding from a real-terminal capture workflow. Recommend **Textual** as the strongest alternative if CSS-based composition and Python development prove more valuable. Neither recommendation is a performance benchmark: startup, typing latency, terminal compatibility, and visual comfort must be measured in the actual application.

The visual direction should be a calm writing instrument with strong hierarchy: generous content space, a clear date, a prominent capture surface, readable entries, restrained accent color, unmistakable focus, and contextual keyboard hints. The biggest design investment should go into editing, state transitions, responsiveness, and trust in saving. Decorative panels and animation cannot compensate for friction in recording an entry.

## What exemplary applications establish

These are verified capabilities of the referenced applications. The final column is a project-specific inference, not a claim that those applications use the proposed journal design.

| Reference | Verified pattern | Application to Dev Journal |
| --- | --- | --- |
| [Lazygit keybindings](https://github.com/jesseduffield/lazygit/blob/master/docs/keybindings/Keybindings_en.md) and [configuration](https://github.com/jesseduffield/lazygit/blob/master/docs/Config.md) | Contextual keybindings, searchable/filterable views, keybinding menu, bottom action line, panel-jump hints; icons can be disabled. | Always show the few actions relevant to the focused surface. Offer arrow navigation alongside optional Vim navigation. Avoid copying the dense multi-panel Git workspace. |
| [Yazi quick start](https://yazi-rs.github.io/docs/quick-start/) and [theme reference](https://yazi-rs.github.io/docs/configuration/theme/) | Keyboard-first browsing; theme roles cover selection, input, help, notification severity, and a which-key surface. | Distinguish selected row, focused surface, editing, and save errors through explicit visual roles. Keep help discoverable rather than requiring memorization. |
| [K9s commands](https://k9scli.io/topics/commands/) | Help, resource navigation, filtering, editing, and commands are accessible by keyboard. | Borrow consistent escape/back behavior and a discoverable action vocabulary. A command language should complement visible navigation, not become a prerequisite for capture. |
| [Television upstream README](https://raw.githubusercontent.com/alexpasmantier/television/main/README.md) | Fuzzy finding with live previews, shell integration, customizable channels and themes. | Use list-plus-preview when reviewing history or selecting Daily material. Search-first retrieval may be valuable later; do not add a channel system to this small journal. |

The shared lesson is orientation plus immediate action. These sources do not establish that any particular palette or number of panels is universally optimal. Dev Journal has a different interaction frequency and much more free-text writing than a file manager or infrastructure dashboard.

## Framework comparison

| Candidate | Verified building blocks | Project tradeoff and assessment |
| --- | --- | --- |
| [Bubble Tea v2](https://github.com/charmbracelet/bubbletea), [Lip Gloss v2](https://github.com/charmbracelet/lipgloss), [Bubbles](https://github.com/charmbracelet/bubbles) | Go model/update/view architecture; inline or full-window rendering; cell-based renderer and color downsampling. Styling includes borders, spacing, alignment and physical text measurement. Reusable inputs, multiline text areas, lists, help and viewports. | Recommended initial candidate: explicit state and extensive visual control with reusable editing components. Application must own focus, layout breakpoints, routing and coherent interactions; component defaults are not a finished design system. |
| [Ratatui layout](https://ratatui.rs/concepts/layout/) and [backends](https://ratatui.rs/concepts/backends/) | Rust widgets in rectangular regions; constraint-based layout; terminal backends including Crossterm and TestBackend. [Snapshot recipe](https://ratatui.rs/recipes/testing/snapshots/) documents rendered-buffer testing. | Excellent choice for precise cell-level composition and a Rust application. More interaction and editor assembly is likely for this product; choose if Rust familiarity or control outweighs that work. |
| [Textual styles](https://textual.textualize.io/guide/styles/), [themes](https://textual.textualize.io/guide/design/), [input](https://textual.textualize.io/guide/input/), [testing](https://textual.textualize.io/guide/testing/) | Python framework with CSS-style rules, theme variables, focus and bindings. Pilot simulates keys/clicks; snapshot plugin generates SVG comparisons. | Strongest alternative: efficient visual iteration and rich widget testing. Packaging and runtime startup need measuring. SVG previews are helpful but cannot prove font, emulator or multiplexer behavior. |
| [Ink](https://github.com/vadimdemedes/ink) | React components, Yoga Flexbox, input/paste/focus hooks, testing guidance and React Devtools. Basic screen-reader mode uses a subset of ARIA roles/states. | Attractive if React/TypeScript experience dominates. A journal still needs coherent full-screen composition and editing behavior; useful components have varying ownership. Basic screen-reader support is valuable but is not proof of universal accessibility. |

Charm explicitly released the v2 generations as stable in February 2026. Use current examples and compatible versions across the three libraries; old v1 imports and view APIs are not interchangeable with v2. [Official release](https://charm.land/blog/v2/), [Bubbles migration guide](https://github.com/charmbracelet/bubbles/blob/main/UPGRADE_GUIDE_V2.md).

The recommendation follows product fit, not screenshots, repository stars, or an assumed language preference. Validate startup and text editing before adopting a stack. A terminal renderer is an implementation detail behind the already agreed public application behavior boundary.

## Proposed capture-first experience

The following is a design proposal to validate in the first ticket.

1. Launch directly into Today with the capture input focused, the local date visible, and the recent daily log in view. Typing requires no navigation key or modal launch.
2. Provide a one-line capture surface with Enter to save. Multiline pasted content must remain one entry and must not trigger commands or submission. Offer an explicit expanded editor for longer writing; Enter inserts a newline there and a visible Save action completes editing.
3. After successful durable persistence, clear the composer, append the entry, show a readable saved confirmation, and keep focus ready for the next entry. On failure, retain text and expose retry. Never imply a write succeeded before it has.
4. Escape from an empty composer enters browsing. Escape from an editor returns safely while preserving a recoverable draft; do not silently discard it. Q exits only outside text editing. Ordinary letters, digits and question marks in text must never invoke global shortcuts.
5. Show Today / Daily / O2O / Tasks in a compact navigation row. Number shortcuts may operate in browsing mode; Tab and Shift+Tab move through interactive surfaces. Display the active view and focus independently.
6. In history, use a readable list with a full-text detail view for long entries; selection is distinct from editing. Editing should return to the same entry and workday without losing position.
7. Keep the quick CLI capture path available for the fastest shell workflow. It should share application behavior with the TUI and return an unambiguous success or error.

At wide sizes, Today can place the daily log beside today's plan and blockers. At ordinary sizes, prefer stacked sections with the log receiving most space. Daily should present recent work with its actual date, today's plan, and blockers as distinct reading sections. An editable daily proposal must be visibly separate from its source entries. The O2O view should prioritize open topics before meeting history; Tasks should make completion and selection for today visually different.

## Visual system proposal

- Define named roles for primary text, secondary text, separators, accent, selected row, focused control, success, warning, and error. Use the same roles across every view.
- Use the terminal background where possible, with tested light and dark variants. Choose a single cool accent, warm warnings and distinct error text; exact colors require terminal review. Avoid low-contrast gray body text.
- Use weight, spacing and short labels to establish hierarchy. Borders should explain focus or grouping rather than surround every row. Avoid nested frames that consume width and create visual noise.
- Keep free text left-aligned, wrap naturally, and avoid mandatory horizontal scrolling. Preserve content rather than truncating it beyond recovery; previews may truncate only when full text is easy to open.
- Always combine color with a text label or shape: `Saved`, `Not saved`, `Blocked`, `[x]`, and a selection marker remain understandable in monochrome. Do not require Nerd Fonts, emoji, images or custom terminal protocols.
- Show a short footer appropriate to focus, such as `Enter Save · Esc Browse` or `Enter Open · e Edit · ? Help · q Quit`. Expand full help on demand; do not display every possible action at once.
- Avoid launch splashes, typing animation and idle refresh loops. Motion is justified only to explain an operation or state; local saves normally need no decorative spinner.

Lip Gloss supplies measurement and color-profile adaptation, but application layout must still allocate content after borders and padding. Its background-color query can support theme choice; use a usable default and an override so terminal query responses never gate capture. [Styling and adaptive colors](https://github.com/charmbracelet/lipgloss).

## Terminal constraints and portability

Terminals do not deliver all keyboard combinations distinctly, and operating systems may intercept keys. Textual documents this limitation; the kitty keyboard protocol defines richer handling but requires support and negotiation. Core capture, save, browse, cancel and exit must work without that protocol. Do not make Shift+Enter, Command shortcuts or exotic Ctrl combinations the sole route to an action. [Textual input limitations](https://textual.textualize.io/guide/input/), [kitty keyboard protocol](https://sw.kovidgoyal.net/kitty/keyboard-protocol/).

Required layout targets in ticket 01 are 120×40 for a spacious view and 80×24 for a complete compact experience. A single focused surface at 60×18 is an optional stretch target, not an additional committed requirement. Below the supported size, preserve the draft and offer a clear resize message or quick-command alternative. Never let resize discard text or strand Save/Cancel outside the viewport. These dimensions are project targets, not industry standards.

Use display-cell width rather than byte length for wrapping, clipping and cursor placement. Test Spanish accents, combining marks, CJK, emoji, pasted paragraphs and long URLs even though UI labels are English. Text editing must preserve the original text. Provide simple ASCII alternatives for decorative box drawing; choose ordinary glyphs for core actions. Bubbles documents Unicode and paste support; Ink explicitly calls out wide-character cursor measurement. [Bubbles editing components](https://github.com/charmbracelet/bubbles), [Ink cursor guidance](https://github.com/vadimdemedes/ink#usecursor).

Honor nonempty `NO_COLOR` and verify a monochrome experience. The convention disables added color, not necessarily bold or underline. An explicit per-application color setting may override it. [NO_COLOR specification](https://no-color.org/).

Do not equate keyboard access with screen-reader accessibility. Repeated full-screen redraws can remain hard to consume. Provide plain CLI output and Markdown export as usable alternatives, and test with a real reader before claiming accessible TUI behavior. Ink explicitly describes its support as basic. [Ink accessibility](https://github.com/vadimdemedes/ink#screen-reader-support).

Restore terminal cursor, input mode and screen state on normal exit and handled failures. Keep diagnostic logs away from the interactive screen. Respect piping and shell exit codes in quick commands. These are consistent with the [Command Line Interface Guidelines](https://clig.dev/); the exact product behaviors should be tested.

## Measurable design acceptance targets

These are proposed budgets and workflow criteria, not measured results. Ticket 01 requires measuring a usable opening screen against a one-second target; the 200 ms figure below is an aspirational research target, not a replacement requirement. Report the machine, terminal, build type, dataset and sample size when measuring.

| Concern | Initial target | Verification |
| --- | --- | --- |
| Ready to write | First interactive frame and focused composer within 200 ms at p95 over 30 release-build launches on the developer machine | Measure launch-to-readiness with a PTY harness and real-terminal review; no network dependency. |
| Single entry | Launch, type, Enter to save; no category form or mandatory metadata | Exercise from a fresh installation and a populated journal. |
| Save and close | After writing, one Save action and one Exit action; success is visible | Terminal workflow with reopen and persisted-content check. |
| Input response | No visible lag; target key-to-frame latency under 50 ms at p95 | PTY timing plus typing review with a large history. |
| Keyboard completeness | Every first-version action reachable without mouse; current action hints always available | Walk capture, correction, Daily, Tasks, O2O and exports with keyboard only. |
| Focus integrity | Exactly one focus target; modal close restores prior focus and selection | Resize and open/close help/editor repeatedly while preserving text. |
| Responsive clarity | No overlap or unreachable primary action at required sizes | Review 120×40, 80×24 and transitions between them; explore 60×18 only as a stretch target. |
| Color and fonts | All statuses and selections readable in dark, light, 16-color and monochrome modes; ordinary monospace font sufficient | Check fallback palette and ASCII mode in real terminals. |
| Trust | No lost draft on save error or safe navigation; no false saved message | Inject storage failure through the public behavior boundary and retry. |
| Scale | Capture remains responsive with 10,000 historical entries | Seed representative wrapped text; avoid loading/rendering all history per keystroke. |

Review on macOS Terminal and the developer's preferred emulator, the VS Code terminal, and one tmux session. Treat other operating systems as additional compatibility targets if distribution expands. Record any terminal-specific limitations rather than claiming universal support from library documentation.

## First-ticket execution gate

Before implementing the rest of the product, build the real capture-and-browse slice with its final persistence behavior. Compare at least two layout treatments using the same content: a spacious split layout and a compact writing layout. Review empty, populated, long-entry, editing, help, save-failure and resized states. Choose the treatment from keyboard comfort and readability, not a static screenshot alone.

Keep behavioral workflow tests at the agreed public application seam. Add a small terminal interaction test for focus, paste, save and exit, plus representative rendering checks for clipping and hints. Do not snapshot every internal component. A visual snapshot is evidence of layout stability, not evidence that the application is comfortable or terminal-compatible.

Open questions to resolve through that slice: exact composer behavior for multiline text; the best history/detail arrangement; default light/dark palette; preferred launch performance budget after real measurement; whether an external editor merits a later shortcut. None blocks recording these design priorities now. Linear integration, learning resources and automatic generation remain outside this version.
