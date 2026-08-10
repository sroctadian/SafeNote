import { api } from "../api.js";
import { navigate } from "../router.js";
import { layout, toast, escapeHtml } from "../components/layout.js";
import { promptPin, showPinError } from "../components/pinModal.js";
import { createEditor, parseStoredContent, deltaToPlainText } from "../components/richEditor.js";

export async function viewPage(params) {
  const id = params.id;

  return {
    html: layout(
      "/home",
      `
      <div class="max-w-2xl mx-auto flex flex-col gap-4">
        <div id="view-container" class="flex flex-col items-center justify-center py-24 gap-4">
          <div class="text-lg opacity-70">This note is encrypted.</div>
          <button id="unlock-btn" class="btn btn-primary">🔓 Enter PIN to Open</button>
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

      const unlock = async () => {
        const pin = await promptPin("Enter PIN to open this note");
        if (pin === null) return;
        try {
          const note = await api.openNote(id, pin);
          renderNote(container, note, id);
        } catch (err) {
          showPinError(err.message);
          toast(err.message, "error");
        }
      };

      root.querySelector("#unlock-btn").addEventListener("click", unlock);
    },
  };
}

function renderNote(container, note, id) {
  container.dataset.unlocked = "true";
  container.className = "flex flex-col gap-4";
  container.innerHTML = `
    <div class="flex items-start justify-between">
      <h1 class="text-2xl font-bold break-words pr-4">${escapeHtml(note.title)}</h1>
      <div class="flex gap-2 shrink-0">
        <button id="edit-btn" class="btn btn-sm">✏️ Edit</button>
        <button id="delete-btn" class="btn btn-sm btn-error btn-outline">🗑️ Delete</button>
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
    </div>
  `;

  // Read-only Quill instance: safer than injecting rendered HTML, since
  // Quill re-renders from structured Delta ops rather than raw markup.
  const quill = createEditor(container.querySelector("#content-viewer"), { readOnly: true });
  quill.setContents(parseStoredContent(note.content));

  container.querySelector("#edit-btn").addEventListener("click", () => navigate(`/edit?id=${id}`));
  container.querySelector("#delete-btn").addEventListener("click", () => deleteFlow(id));

  const keyHandler = (e) => {
    if (e.ctrlKey && e.key.toLowerCase() === "c") {
      e.preventDefault();
      const plainText = deltaToPlainText(note.content);
      api.setClipboardText(plainText).then(() => toast("Copied to clipboard.", "success"));
    }
  };
  document.addEventListener("keydown", keyHandler);
}

async function deleteFlow(id) {
  if (!confirm("Delete this note permanently? This cannot be undone.")) return;
  try {
    await api.deleteNote(id);
    toast("Note deleted.", "success");
    navigate("/home");
  } catch (err) {
    toast(err.message, "error");
  }
}
