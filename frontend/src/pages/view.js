import { api } from "../api.js";
import { navigate, onCleanup } from "../router.js";
import { layout, toast, escapeHtml } from "../components/layout.js";
import { promptPin, showPinError } from "../components/pinModal.js";
import { createEditor, parseStoredContent, deltaToPlainText } from "../components/richEditor.js";
import { rememberUnlock, forgetUnlock, getRememberedPin } from "../noteSession.js";
import { copyToClipboardWithAutoClear } from "../clipboard.js";
import { icon } from "../components/icons.js";

export async function viewPage(params) {
  const id = params.id;

  return {
    html: layout(
      "/home",
      `
      <div class="max-w-2xl mx-auto flex flex-col gap-4">
        <div id="view-container" class="flex items-center justify-center py-24">
          <div class="loading loading-spinner loading-lg"></div>
        </div>
      </div>
      `
    ),
    mount: async (root) => {
      if (!id) {
        navigate("/home");
        return;
      }

      const container = root.querySelector("#view-container");

      // REV1: no intermediate "This note is encrypted" screen — the PIN
      // dialog opens immediately, same as before but without the extra
      // click. On a wrong PIN, re-prompt automatically (rate-limited by
      // the backend's existing lockout tracker) rather than dumping the
      // user back on a dead-end screen. If we already unlocked this
      // exact note moments ago (e.g. coming back from Edit), reuse that
      // PIN instead of asking again.
      const attemptUnlock = async (useCache = true) => {
        let pin = useCache ? getRememberedPin(id) : null;
        if (!pin) {
          pin = await promptPin("Enter PIN to open this note");
          if (pin === null) {
            navigate("/home");
            return;
          }
        }
        try {
          const note = await api.openNote(id, pin);
          rememberUnlock(id, pin);
          renderNote(container, note, id);
        } catch (err) {
          showPinError(err.message);
          toast(err.message, "error");
          attemptUnlock(false); // never retry with a cached PIN that just failed
        }
      };

      attemptUnlock();
    },
  };
}

function renderNote(container, note, id) {
  container.dataset.unlocked = "true";
  container.className = "flex flex-col gap-4";
  container.innerHTML = `
    <div class="flex items-start justify-between">
      <h1 class="text-2xl font-bold break-words pr-4">${escapeHtml(note.title)}</h1>
      <div class="flex gap-1 shrink-0">
        <button id="copy-btn" class="btn-icon" title="Copy note" aria-label="Copy note">${icon("clipboardDocument")}</button>
        <button id="edit-btn" class="btn-icon" title="Edit note" aria-label="Edit note">${icon("pencilSquare")}</button>
        <button id="delete-btn" class="btn-icon text-error" title="Delete note" aria-label="Delete note">${icon("trash")}</button>
      </div>
    </div>
    ${
      note.tags?.length
        ? `<div class="flex gap-1 flex-wrap">${note.tags
            .map((t) => `<span class="badge badge-outline">${escapeHtml(t)}</span>`)
            .join("")}</div>`
        : ""
    }
    <div class="bg-base-100 border border-base-300 rounded-lg overflow-hidden">
      <div id="content-viewer"></div>
    </div>
    <div class="text-xs opacity-50">
      Content is decrypted in memory only for this view. Navigating away clears it.
      Select any text and press Ctrl+C to copy just that selection.
    </div>
  `;

  // Read-only Quill instance: safer than injecting rendered HTML, since
  // Quill re-renders from structured Delta ops rather than raw markup.
  const quill = createEditor(container.querySelector("#content-viewer"), { readOnly: true });
  quill.setContents(parseStoredContent(note.content));

  container.querySelector("#edit-btn").addEventListener("click", () => navigate(`/edit?id=${id}`));
  container.querySelector("#delete-btn").addEventListener("click", () => deleteFlow(id));
  container.querySelector("#copy-btn").addEventListener("click", () => {
    copyToClipboardWithAutoClear(deltaToPlainText(note.content));
  });

  // Ctrl+C is selection-aware: if the user has highlighted part of the
  // note, let the browser's native copy handle it (this preserves
  // exactly what was selected, and even carries formatting when pasted
  // into another rich-text target). Only fall back to copying the
  // *entire* note when there's no active selection — previously this
  // shortcut always copied everything, silently discarding whatever
  // the user had actually selected.
  const keyHandler = (e) => {
    if (!(e.ctrlKey && e.key.toLowerCase() === "c")) return;

    const selectedText = window.getSelection()?.toString() ?? "";
    if (selectedText.trim().length > 0) {
      return; // let the native copy proceed untouched
    }

    e.preventDefault();
    copyToClipboardWithAutoClear(deltaToPlainText(note.content));
  };
  document.addEventListener("keydown", keyHandler);
  onCleanup(() => document.removeEventListener("keydown", keyHandler));
}

async function deleteFlow(id) {
  if (!confirm("Delete this note permanently? This cannot be undone.")) return;
  try {
    await api.deleteNote(id);
    forgetUnlock();
    toast("Note deleted.", "success");
    navigate("/home");
  } catch (err) {
    toast(err.message, "error");
  }
}
