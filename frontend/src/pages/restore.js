import { api } from "../api.js";
import { layout, toast } from "../components/layout.js";
import { icon } from "../components/icons.js";

export async function restorePage() {
  return {
    html: layout(
      "/restore",
      `
      <div class="max-w-2xl mx-auto flex flex-col gap-4">
        <h1 class="text-2xl font-bold">Restore</h1>
        <p class="opacity-70 text-sm">
          Select a SafeNote backup file. It will be validated (format
          version + checksum) before anything is imported. Duplicate notes
          are never overwritten automatically.
        </p>
        <button id="select-btn" class="btn btn-primary self-start gap-1">${icon("arrowDownTray", "w-4 h-4")} Select Backup File</button>
        <div id="preview-container"></div>
      </div>
      `
    ),
    mount: (root) => {
      let selectedPath = null;

      root.querySelector("#select-btn").addEventListener("click", async () => {
        const container = root.querySelector("#preview-container");
        try {
          const path = await api.openFileDialog();
          if (!path) return;
          selectedPath = path;

          const preview = await api.previewRestore(path);
          container.innerHTML = `
            <div class="card bg-base-200 mt-2">
              <div class="card-body">
                <h2 class="card-title text-base">Backup Preview</h2>
                <p>Format version: ${preview.formatVersion}</p>
                <p>Notes in file: ${preview.noteCount}</p>
                <p>Duplicates found: ${preview.duplicateIds?.length || 0}</p>
                ${
                  preview.duplicateIds?.length
                    ? `<label class="label cursor-pointer justify-start gap-2">
                        <input id="overwrite-check" type="checkbox" class="checkbox checkbox-sm" />
                        <span class="label-text text-sm">Overwrite ${preview.duplicateIds.length} duplicate note(s)</span>
                       </label>`
                    : ""
                }
                <button id="confirm-restore-btn" class="btn btn-primary btn-sm self-end mt-2">Confirm Restore</button>
              </div>
            </div>
          `;

          container.querySelector("#confirm-restore-btn").addEventListener("click", async () => {
            try {
              const overwrite = container.querySelector("#overwrite-check")?.checked;
              const ids = overwrite ? preview.duplicateIds : [];
              const imported = await api.restoreBackup(selectedPath, ids);
              toast(`Restored ${imported} note(s).`, "success");
            } catch (err) {
              toast(err.message, "error");
            }
          });
        } catch (err) {
          container.innerHTML = `<div class="alert alert-error mt-2">${err.message}</div>`;
        }
      });
    },
  };
}
