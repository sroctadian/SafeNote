import Quill from "quill";
import "quill/dist/quill.snow.css";

// SafeNote stores note content as Quill's "Delta" format, serialized to
// JSON. This is safer than storing rendered HTML (Quill re-renders from
// structured ops rather than injecting raw markup, so a tampered backup
// file can't smuggle in a script via innerHTML) and preserves formatting
// losslessly.
//
// Notes created before rich text support was added contain plain text
// content instead of Delta JSON. parseStoredContent() detects that case
// and falls back to treating the raw string as a single plain-text
// insert, so old notes keep opening correctly.

const FULL_TOOLBAR = [
  [{ header: [1, 2, 3, false] }],
  ["bold", "italic", "underline", "strike"],
  [{ color: [] }, { background: [] }],
  [{ list: "ordered" }, { list: "bullet" }],
  [{ indent: "-1" }, { indent: "+1" }],
  ["blockquote", "code-block"],
  ["link"],
  ["clean"],
];

/**
 * Create a Quill editor instance inside containerEl.
 * @param {HTMLElement} containerEl
 * @param {{ readOnly?: boolean, placeholder?: string }} options
 */
export function createEditor(containerEl, options = {}) {
  return new Quill(containerEl, {
    theme: "snow",
    readOnly: !!options.readOnly,
    placeholder: options.placeholder || "",
    modules: {
      toolbar: options.readOnly ? false : FULL_TOOLBAR,
    },
  });
}

/**
 * Parse a note's stored content string into a Quill Delta object.
 * Falls back to plain text for legacy (pre-rich-text) notes.
 */
export function parseStoredContent(raw) {
  if (!raw) return { ops: [{ insert: "\n" }] };
  try {
    const parsed = JSON.parse(raw);
    if (parsed && Array.isArray(parsed.ops)) {
      return parsed;
    }
  } catch {
    // Not JSON — legacy plain-text note.
  }
  return { ops: [{ insert: raw }] };
}

/** Serialize a Quill editor's current content to the stored format. */
export function serializeContent(quill) {
  return JSON.stringify(quill.getContents());
}

/**
 * Convert stored content (Delta JSON or legacy plain text) to plain
 * text, used for clipboard copy and search-safe previews.
 */
export function deltaToPlainText(raw) {
  const delta = parseStoredContent(raw);
  return delta.ops
    .map((op) => (typeof op.insert === "string" ? op.insert : ""))
    .join("")
    .replace(/\n+$/, "");
}

/** True if the note is non-empty (Quill's empty state is a lone "\n"). */
export function isContentEmpty(raw) {
  return deltaToPlainText(raw).trim().length === 0;
}
