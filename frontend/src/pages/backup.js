import { api } from "../api.js";
import { layout, toast } from "../components/layout.js";

export async function backupPage() {
  return {
    html: layout(
      "/backup",
      `
      <div class="max-w-2xl mx-auto flex flex-col gap-4">
        <h1 class="text-2xl font-bold">Backup</h1>
        <p class="opacity-70 text-sm">
          Export all notes as a single encrypted backup file. SafeNote never
          decrypts note content during backup — the exported file is exactly
          as protected as your live database.
        </p>
        <button id="backup-btn" class="btn btn-primary self-start">📤 Export Backup</button>
        <div id="backup-result"></div>
      </div>
      `
    ),
    mount: (root) => {
      root.querySelector("#backup-btn").addEventListener("click", async () => {
        const resultEl = root.querySelector("#backup-result");
        try {
          const path = await api.saveFileDialog(`safenote-backup-${dateStamp()}.json`);
          if (!path) return;
          const bf = await api.exportBackup(path);
          toast(`Backup exported: ${bf.notes.length} notes.`, "success");
          resultEl.innerHTML = `
            <div class="alert alert-success mt-2">
              Exported ${bf.notes.length} notes. Checksum: <code class="text-xs">${bf.checksum.slice(0, 16)}…</code>
            </div>`;
        } catch (err) {
          toast(err.message, "error");
        }
      });
    },
  };
}

function dateStamp() {
  return new Date().toISOString().slice(0, 10);
}
