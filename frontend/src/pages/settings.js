import { api } from "../api.js";
import { layout, toast, escapeHtml } from "../components/layout.js";
import { icon } from "../components/icons.js";

export async function settingsPage() {
  let settings, masked, dataDir;
  try {
    settings = await api.getSettings();
    masked = await api.getMaskedSecretKey();
  } catch (err) {
    settings = { theme: "dark", clipboardClearSeconds: 20 };
    masked = "unavailable";
  }
  try {
    dataDir = await api.getDataDirectory();
  } catch (err) {
    dataDir = "unavailable";
  }

  return {
    html: layout(
      "/settings",
      `
      <div class="max-w-2xl mx-auto flex flex-col gap-6">
        <h1 class="text-2xl font-bold">Settings</h1>

        <div class="card bg-base-200">
          <div class="card-body">
            <h2 class="card-title text-base">Secret Key</h2>
            <p class="text-sm opacity-70">Current: <code>${masked}</code></p>
            <textarea id="new-sk" class="textarea textarea-bordered w-full mt-2" rows="2"
              placeholder="New Secret Key (min. 32 characters)"></textarea>
            <p class="text-xs opacity-60 mt-1">
              Note: existing notes remain tied to the Secret Key active when they were
              created/edited. See docs/ADR.md for the re-encryption rollout strategy.
            </p>
            <button id="change-sk-btn" class="btn btn-sm btn-primary mt-2 self-end">Change Secret Key</button>
          </div>
        </div>

        <div class="card bg-base-200">
          <div class="card-body">
            <h2 class="card-title text-base">Theme</h2>
            <div class="join">
              <button id="theme-light" class="btn btn-sm join-item ${settings.theme === "light" ? "btn-active" : ""}">☀️ Light</button>
              <button id="theme-dark" class="btn btn-sm join-item ${settings.theme === "dark" ? "btn-active" : ""}">🌙 Dark</button>
            </div>
          </div>
        </div>

        <div class="card bg-base-200">
          <div class="card-body">
            <h2 class="card-title text-base">Clipboard Auto-Clear</h2>
            <input id="clipboard-timeout" type="number" min="0" max="300"
              class="input input-bordered w-32" value="${settings.clipboardClearSeconds ?? 20}" />
            <span class="text-xs opacity-60">seconds (0 = never)</span>
            <button id="clipboard-save-btn" class="btn btn-sm mt-2 self-end">Save</button>
          </div>
        </div>

        <div class="card bg-base-200">
          <div class="card-body">
            <h2 class="card-title text-base">Data Location</h2>
            <p class="text-sm opacity-70">Your notes database and vault key currently live at:</p>
            <code id="data-dir-display" class="text-xs bg-base-300 rounded px-2 py-1.5 break-all">${escapeHtml(dataDir)}</code>
            <p class="text-xs opacity-60 mt-2">
              Point this at a folder synced by Dropbox, OneDrive, Syncthing, etc. to access
              your notes from another device — no account or server involved. Moving to an
              <strong>empty</strong> folder copies your current data there; pointing at a folder
              that <strong>already has</strong> SafeNote data adopts it instead (your local
              copy is backed up first, just in case).
            </p>
            <p class="text-xs text-warning mt-1">
              ⚠ This also relocates your vault key, which trades away some of the
              defense-in-depth SafeNote normally keeps between your database and the key that
              protects your Secret Key — keep access to the synced folder as tightly controlled
              as your notes themselves. See docs/ADR.md (ADR-007).
            </p>
            <button id="change-data-dir-btn" class="btn btn-sm mt-3 self-end gap-1">
              ${icon("arrowUpTray", "w-4 h-4")} Choose Folder…
            </button>
          </div>
        </div>

        <div class="card bg-base-200">
          <div class="card-body">
            <h2 class="card-title text-base">Configuration</h2>
            <div class="flex gap-2">
              <button id="export-config-btn" class="btn btn-sm">Export Configuration</button>
              <button id="import-config-btn" class="btn btn-sm">Import Configuration</button>
            </div>
          </div>
        </div>
      </div>
      `
    ),
    mount: (root) => {
      root.querySelector("#change-sk-btn").addEventListener("click", async () => {
        const value = root.querySelector("#new-sk").value;
        if (value.length < 32) {
          toast("Secret Key must be at least 32 characters.", "error");
          return;
        }
        try {
          await api.changeSecretKey(value);
          toast("Secret Key changed.", "success");
        } catch (err) {
          toast(err.message, "error");
        }
      });

      root.querySelector("#theme-light").addEventListener("click", () => setTheme(root, "light"));
      root.querySelector("#theme-dark").addEventListener("click", () => setTheme(root, "dark"));

      root.querySelector("#clipboard-save-btn").addEventListener("click", async () => {
        const seconds = Number(root.querySelector("#clipboard-timeout").value);
        try {
          await api.updateClipboardTimeout(seconds);
          toast("Clipboard timeout saved.", "success");
        } catch (err) {
          toast(err.message, "error");
        }
      });

      root.querySelector("#change-data-dir-btn").addEventListener("click", async () => {
        try {
          const selected = await api.selectDirectoryDialog();
          if (!selected) return; // user cancelled the picker

          const proceed = confirm(
            "Change SafeNote's data location to:\n\n" +
              selected +
              "\n\nSafeNote must be restarted afterward for this to take effect. Continue?"
          );
          if (!proceed) return;

          const status = await api.setDataDirectory(selected);
          const display = root.querySelector("#data-dir-display");
          if (display) display.textContent = status.newPath;

          if (status.adopted) {
            toast(
              status.oldDataBackedUp
                ? "Adopted existing data at that folder. Your previous local data was backed up. Restart SafeNote to apply."
                : "Adopted existing data at that folder. Restart SafeNote to apply.",
              "success"
            );
          } else if (status.moved) {
            toast("Data moved to the new folder. Restart SafeNote to apply.", "success");
          } else {
            toast("That's already your current data folder.", "info");
          }
        } catch (err) {
          toast(err.message, "error");
        }
      });

      root.querySelector("#export-config-btn").addEventListener("click", async () => {
        try {
          const path = await api.saveFileDialog("safenote-config.json");
          if (!path) return;
          await api.exportConfig(path);
          toast("Configuration exported.", "success");
        } catch (err) {
          toast(err.message, "error");
        }
      });

      root.querySelector("#import-config-btn").addEventListener("click", async () => {
        try {
          const path = await api.openFileDialog();
          if (!path) return;
          await api.importConfig(path);
          toast("Configuration imported.", "success");
        } catch (err) {
          toast(err.message, "error");
        }
      });
    },
  };
}

async function setTheme(root, theme) {
  try {
    await api.updateTheme(theme);
    document.documentElement.setAttribute("data-theme", theme === "dark" ? "safenoteDark" : "safenoteLight");
    toast(`Theme set to ${theme}.`, "success");
  } catch (err) {
    toast(err.message, "error");
  }
}
