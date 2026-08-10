import { registerRoute, setNotFound, startRouter } from "./router.js";
import { splashPage } from "./pages/splash.js";
import { setupPage } from "./pages/setup.js";
import { homePage } from "./pages/home.js";
import { createPage } from "./pages/create.js";
import { viewPage } from "./pages/view.js";
import { editPage } from "./pages/edit.js";
import { settingsPage } from "./pages/settings.js";
import { backupPage } from "./pages/backup.js";
import { restorePage } from "./pages/restore.js";
import { aboutPage } from "./pages/about.js";
import { api } from "./api.js";

registerRoute("/splash", splashPage);
registerRoute("/setup", setupPage);
registerRoute("/home", homePage);
registerRoute("/create", createPage);
registerRoute("/view", viewPage);
registerRoute("/edit", editPage);
registerRoute("/settings", settingsPage);
registerRoute("/backup", backupPage);
registerRoute("/restore", restorePage);
registerRoute("/about", aboutPage);

setNotFound(() => `
  <div class="min-h-screen flex flex-col items-center justify-center gap-4">
    <div class="text-xl">Page not found.</div>
    <a href="#/home" class="btn btn-primary btn-sm">Go Home</a>
  </div>
`);

async function applyStoredTheme() {
  try {
    const settings = await api.getSettings();
    document.documentElement.setAttribute(
      "data-theme",
      settings.theme === "light" ? "safenoteLight" : "safenoteDark"
    );
  } catch {
    // Not configured yet (first run) — keep default dark theme.
  }
}

document.addEventListener("keydown", (e) => {
  if (e.key === "Escape") {
    document.querySelectorAll(".modal-open").forEach((m) => m.classList.remove("modal-open"));
  }
});

applyStoredTheme();
startRouter();
