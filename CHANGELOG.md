# Changelog

## [0.1.7] — 2026-08-13

### Added
- **Custom data directory** (Settings → Data Location): point SafeNote's
  database + vault key at any folder — including one already synced by
  Dropbox/OneDrive/Syncthing — as a zero-server way to access notes from
  multiple devices. Two flows, chosen automatically: **move** (target
  folder empty → current data copied there) or **adopt** (target folder
  already has a SafeNote database → point at it as-is, local copy backed
  up first as a safety net). Requires an app restart to take effect
  (deliberately not hot-swapped mid-session). New
  `internal/app/datalocation.go` + `ResolveDataDir`/`SetDataDirectory`
  + test suite. See `docs/ADR.md` (ADR-007) for the security trade-off
  this involves — the vault key now travels with the database when
  moving to a new folder, which narrows the defense-in-depth ADR-003
  established for the default (non-customized) location.

### Accessibility
- All icon-only buttons (favorite, pin, copy, open, edit, delete, swipe
  reveal, mobile nav toggle) now have matching `aria-label` in addition
  to `title`, with favorite/pin using dynamic labels + `aria-pressed`
  reflecting their current state (e.g. "Add to favorites" vs "Remove
  from favorites").

## [0.1.6] — 2026-08-13

### Changed (7 UX revisions)
- **REV1**: View page no longer shows a "This note is encrypted" holding
  screen — the PIN dialog now opens immediately on navigation, with
  automatic re-prompt on a wrong PIN (still rate-limited by the existing
  backend lockout tracker).
- **REV2**: Notes can now be deleted straight from Home without entering
  a PIN (deleting doesn't decrypt anything, so no PIN was ever needed).
  Swipe a card right to reveal a Delete action; a confirmation dialog
  still guards the actual delete. New `components/swipeToDelete.js`.
- **REV3**: Create/Edit forms no longer have an inline PIN field. PIN is
  requested via the same modal used to open a note, triggered at Save
  time for Create (Edit already prompted before showing the form, to
  decrypt existing content — unchanged).
- **REV4**: Replaced emoji icons throughout (sidebar nav, note cards,
  View/Backup/Restore actions) with inline Heroicons v2 (MIT licensed).
  New `components/icons.js`.
- **REV5**: Replaced the CSS-drawn logo mark and 🔒 emoji with the
  provided SafeNote logo — used in the sidebar, mobile navbar, splash
  screen, and browser/window favicon.
- **REV6**: Saving from Create or Edit now returns straight to the note
  list instead of routing back through the (now PIN-gated) View page.
- **REV7**: Home card actions are icon-only now (clipboard icon for
  Copy, eye icon for Open) instead of labeled buttons, matching REV4's
  icon set.

## [0.1.5] — 2026-08-13

### Fixed
- Router now has a proper cleanup lifecycle: `onCleanup(fn)` in
  `router.js` lets a page register teardown work (e.g.
  `removeEventListener`) that automatically runs the moment the user
  navigates to a different route. Previously, per-page `keydown`
  listeners (Ctrl+S on Create/Edit, Ctrl+C on View, Ctrl+N/Ctrl+F on
  Home) were attached to `document` and never removed, so revisiting
  those pages repeatedly without a full reload would stack up stale
  listeners holding references to old notes/containers.
- This uses an explicit registration function rather than a
  return-value-based cleanup pattern, because several pages have an
  `async mount()` that the router intentionally does not `await`
  (so slow pages don't block rendering) — a returned cleanup function
  would often resolve after the route had already changed.

## [0.1.4] — 2026-08-13

### Fixed
- View page's Ctrl+C previously always copied the *entire* note as plain
  text, discarding any text the user had actually selected. It's now
  selection-aware: an active text selection is copied natively by the
  browser (preserving formatting when pasted elsewhere); Ctrl+C only
  falls back to copying the whole note when nothing is selected.
- Added an explicit "📋 Copy" button on the View page for copying the
  whole note without relying on the keyboard shortcut.
- Extracted the copy-with-auto-clear logic (used by Home's card Copy
  button and View's new Copy button/Ctrl+C fallback) into a shared
  `clipboard.js` helper to avoid duplicated timeout logic.

## [0.1.3] — 2026-08-13

### Fixed
- Home: clicking anywhere on a note card now opens it (previously only
  the small "Open" button worked). Icon buttons (favorite/pin/copy)
  still work independently via event propagation stopping.
- View → Edit no longer re-prompts for a PIN that was just entered.
  A short-lived, in-memory-only PIN cache (`noteSession.js`) remembers
  the unlock for the *current* note only, and is cleared whenever the
  user returns to Home or deletes the note — so a note never stays
  "unlocked" beyond the immediate view/edit flow.

## [0.1.2] — 2026-08-13

### Fixed
- Rich text editor toolbar no longer scrolls out of view on long notes.
  The editor is now a fixed-height panel (`.editor-shell`): the Quill
  toolbar stays pinned, and only the note body (`.ql-editor`) scrolls
  internally. Applies to Create and Edit; the read-only View page keeps
  its natural (page-scrolling) height since it has no toolbar.

## [0.1.1] — 2026-08-09

### Added
- Rich text note editor using Quill.js (toolbar: headers, bold/italic/
  underline/strike, colors, lists, indent, blockquote, code block, link)
- Note content now stored as Quill Delta JSON (format-agnostic to the
  backend — no crypto/service/repository changes required)
- Read-only Quill rendering on the View page (safer than raw HTML
  injection: Quill re-renders from structured ops, not markup strings)
- `SetClipboardText` backend method — Copy Note now converts rich content
  to plain text on the frontend before writing to the OS clipboard
- Legacy plain-text notes (pre-0.1.1) still open correctly via automatic
  fallback parsing

## [0.1.0] — 2026-08-09

Initial scaffold.

### Added
- Core encryption: AES-256-GCM + Argon2id per-note key derivation (Secret Key + PIN)
- Note CRUD (create, open, edit, delete), favorite, pin, search, sort, pagination
- Settings: first-run Secret Key setup, change Secret Key, theme, clipboard auto-clear, export/import config
- Backup/Restore: encrypted export, checksum validation, duplicate detection, no-auto-overwrite
- Input validation: PIN must be exactly 6 digits, title max 75 chars, tags max 25 chars each
- Frontend: vanilla JS SPA (hash router) + TailwindCSS/DaisyUI, dark/light theme
- Unit tests for crypto, note/settings/backup services, and input validation

### Fixed
- Project structure: moved `main.go` to project root (Wails v2 requires this to generate bindings)
- `ListNotes` binding: Wails v2 does not support 3+ Go return values as an iterable JS array;
  changed to return a single `ListNotesResult{Notes, Total}` struct

### Known limitations (see docs/ADR.md)
- Changing the Secret Key does not retroactively re-encrypt existing notes
- Vault key is a local file (0600), not OS-keychain-backed
- Failed-attempt lockout is in-memory only, does not survive app restart
- No rich text editor (plain text notes only)
- No secure-delete (overwrite-before-delete) on the SQLite row
