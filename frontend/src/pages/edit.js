import { api } from "../api.js";
import { navigate, onCleanup } from "../router.js";
import { layout, toast, escapeHtml } from "../components/layout.js";
import { promptPin } from "../components/pinModal.js";
import { createEditor, parseStoredContent, serializeContent, isContentEmpty } from "../components/richEditor.js";
import { rememberUnlock, getRememberedPin } from "../noteSession.js";

export async function editPage(params) {
  const id = params.id;

  return {
    html: layout(
      "/home",
      `
      <div class="max-w-2xl mx-auto flex flex-col gap-4">
        <div id="edit-container" class="flex flex-col items-center justify-center py-24 gap-4">
          <div class="loading loading-spinner"></div>
        </div>
      </div>
      `
    ),
    mount: async (root) => {
      if (!id) {
        navigate("/home");
        return;
      }
      const container = root.querySelector("#edit-container");

      let pin = getRememberedPin(id);
      if (!pin) {
        pin = await promptPin("Enter PIN to edit this note");
        if (pin === null) {
          navigate("/home");
          return;
        }
      }

      let note;
      try {
        note = await api.openNote(id, pin);
        rememberUnlock(id, pin); // keep/refresh cache for this note
      } catch (err) {
        toast(err.message, "error");
        navigate("/home");
        return;
      }

      container.className = "flex flex-col gap-4";
      container.innerHTML = `
        <h1 class="text-2xl font-bold">Edit Note</h1>
        <input id="title-input" type="text" class="input input-bordered w-full"
          value="${escapeHtml(note.title)}" maxlength="75" />
        <div class="editor-shell bg-base-100 border border-base-300 rounded-lg overflow-hidden h-[45vh] min-h-[280px] max-h-[520px]">
          <div id="content-editor"></div>
        </div>
        <input id="tags-input" type="text" class="input input-bordered w-full"
          value="${escapeHtml((note.tags || []).join(", "))}" placeholder="Tags (comma separated)" />
        <p id="form-error" class="text-error text-sm hidden"></p>
        <div class="flex justify-end gap-2">
          <button class="btn btn-ghost" onclick="location.hash='#/view?id=${id}'">Cancel</button>
          <button id="save-btn" class="btn btn-primary">Save (Ctrl+S)</button>
        </div>
      `;

      const quill = createEditor(container.querySelector("#content-editor"));
      quill.setContents(parseStoredContent(note.content));

      const save = async () => {
        const title = container.querySelector("#title-input").value.trim();
        const content = serializeContent(quill);
        const tags = container
          .querySelector("#tags-input")
          .value.split(",")
          .map((t) => t.trim())
          .filter(Boolean);
        const errorEl = container.querySelector("#form-error");

        if (!title || isContentEmpty(content)) {
          errorEl.textContent = "Title and content are required.";
          errorEl.classList.remove("hidden");
          return;
        }
        if (title.length > 75) {
          errorEl.textContent = "Title must be at most 75 characters.";
          errorEl.classList.remove("hidden");
          return;
        }
        if (tags.some((t) => t.length > 25)) {
          errorEl.textContent = "Each tag must be at most 25 characters.";
          errorEl.classList.remove("hidden");
          return;
        }

        try {
          await api.editNote(id, pin, title, content, tags);
          toast("Note updated and re-encrypted.", "success");
          navigate("/home"); // REV6: straight back to the list
        } catch (err) {
          errorEl.textContent = err.message;
          errorEl.classList.remove("hidden");
        }
      };

      container.querySelector("#save-btn").addEventListener("click", save);
      const keyHandler = (e) => {
        if (e.ctrlKey && e.key.toLowerCase() === "s") {
          e.preventDefault();
          save();
        }
      };
      document.addEventListener("keydown", keyHandler);
      onCleanup(() => document.removeEventListener("keydown", keyHandler));
    },
  };
}
