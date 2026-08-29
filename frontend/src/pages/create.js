import { api } from "../api.js";
import { navigate, onCleanup } from "../router.js";
import { layout, toast } from "../components/layout.js";
import { createEditor, serializeContent, isContentEmpty } from "../components/richEditor.js";
import { promptPin, showPinError } from "../components/pinModal.js";

export async function createPage() {
  return {
    html: layout(
      "/create",
      `
      <div class="max-w-2xl mx-auto flex flex-col gap-4">
        <h1 class="text-2xl font-bold">New Note</h1>
        <input id="title-input" type="text" placeholder="Title"
          class="input input-bordered w-full" maxlength="75" />
        <div class="text-xs opacity-60 -mt-2"><span id="title-count">0</span> / 75 characters</div>
        <div class="editor-shell bg-base-100 border border-base-300 rounded-lg overflow-hidden h-[45vh] min-h-[280px] max-h-[520px]">
          <div id="content-editor"></div>
        </div>
        <input id="tags-input" type="text" placeholder="Tags (comma separated, max 25 chars each, optional)"
          class="input input-bordered w-full" />
        <p id="form-error" class="text-error text-sm hidden"></p>
        <div class="flex justify-end gap-2">
          <button class="btn btn-ghost" onclick="history.back()">Cancel</button>
          <button id="save-btn" class="btn btn-primary">Save (Ctrl+S)</button>
        </div>
      </div>
      `
    ),
    mount: (root) => {
      const quill = createEditor(root.querySelector("#content-editor"), {
        placeholder: "Write your note...",
      });

      const titleInput = root.querySelector("#title-input");
      const titleCount = root.querySelector("#title-count");
      titleInput.addEventListener("input", () => {
        titleCount.textContent = titleInput.value.length;
      });

      const save = async () => {
        const title = titleInput.value.trim();
        const content = serializeContent(quill);
        const tags = root
          .querySelector("#tags-input")
          .value.split(",")
          .map((t) => t.trim())
          .filter(Boolean);
        const errorEl = root.querySelector("#form-error");

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
        errorEl.classList.add("hidden");

        // REV3: PIN is no longer a field inside the form — it's asked
        // via the same modal used to open an existing note, right when
        // the user commits to saving. Keeps the writing form itself
        // free of security chrome until it's actually needed.
        await promptForPinAndSave(title, content, tags, errorEl);
      };

      root.querySelector("#save-btn").addEventListener("click", save);
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

async function promptForPinAndSave(title, content, tags, errorEl) {
  const pin = await promptPin("Set a 6-digit PIN to encrypt this note");
  if (pin === null) return; // user cancelled — stay on the form, nothing lost

  try {
    await api.createNote(title, content, pin, tags);
    toast("Note created and encrypted.", "success");
    navigate("/home"); // REV6: straight back to the list
  } catch (err) {
    showPinError(err.message);
    errorEl.textContent = err.message;
    errorEl.classList.remove("hidden");
  }
}
