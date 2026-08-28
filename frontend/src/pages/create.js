import { api } from "../api.js";
import { navigate, onCleanup } from "../router.js";
import { layout, toast } from "../components/layout.js";
import { createEditor, serializeContent, isContentEmpty } from "../components/richEditor.js";

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
        <input id="pin-input" type="password" inputmode="numeric" pattern="\d{6}"
          placeholder="6-digit PIN to encrypt this note"
          class="input input-bordered w-full" maxlength="6" />
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

      const pinInput = root.querySelector("#pin-input");
      pinInput.addEventListener("input", () => {
        pinInput.value = pinInput.value.replace(/\D/g, "").slice(0, 6);
      });

      const save = async () => {
        const title = titleInput.value.trim();
        const content = serializeContent(quill);
        const tags = root
          .querySelector("#tags-input")
          .value.split(",")
          .map((t) => t.trim())
          .filter(Boolean);
        const pin = pinInput.value;
        const errorEl = root.querySelector("#form-error");

        if (!title || isContentEmpty(content) || !pin) {
          errorEl.textContent = "Title, content, and PIN are required.";
          errorEl.classList.remove("hidden");
          return;
        }
        if (title.length > 75) {
          errorEl.textContent = "Title must be at most 75 characters.";
          errorEl.classList.remove("hidden");
          return;
        }
        if (!/^\d{6}$/.test(pin)) {
          errorEl.textContent = "PIN must be exactly 6 digits (0-9).";
          errorEl.classList.remove("hidden");
          return;
        }
        if (tags.some((t) => t.length > 25)) {
          errorEl.textContent = "Each tag must be at most 25 characters.";
          errorEl.classList.remove("hidden");
          return;
        }

        try {
          await api.createNote(title, content, pin, tags);
          toast("Note created and encrypted.", "success");
          navigate("/home");
        } catch (err) {
          errorEl.textContent = err.message;
          errorEl.classList.remove("hidden");
        }
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
