import { icon } from "./icons.js";
import logoUrl from "../assets/logo.png";

export function layout(activeRoute, content) {
  const navItem = (path, label, iconName) => `
    <a href="#${path}" class="btn btn-ghost btn-sm justify-start gap-2 ${
    activeRoute === path ? "btn-active" : ""
  }">
      ${icon(iconName, "w-5 h-5")} ${label}
    </a>`;

  return `
    <div class="drawer lg:drawer-open min-h-screen">
      <input id="nav-drawer" type="checkbox" class="drawer-toggle" />
      <div class="drawer-content flex flex-col">
        <div class="navbar bg-base-200 border-b border-base-300 lg:hidden">
          <label for="nav-drawer" class="btn btn-square btn-ghost">☰</label>
          <img src="${logoUrl}" alt="SafeNote" class="w-6 h-6 ml-2" />
          <span class="text-lg font-semibold ml-2">SafeNote</span>
        </div>
        <main class="flex-1 p-4 md:p-8 max-w-6xl w-full mx-auto">
          ${content}
        </main>
      </div>
      <div class="drawer-side">
        <label for="nav-drawer" class="drawer-overlay"></label>
        <aside class="w-64 bg-base-200 border-r border-base-300 min-h-full p-4 flex flex-col gap-1">
          <div class="text-xl font-bold px-2 mb-4 flex items-center gap-2">
            <img src="${logoUrl}" alt="SafeNote" class="w-7 h-7" /> SafeNote
          </div>
          ${navItem("/home", "Home", "home")}
          ${navItem("/create", "New Note", "plus")}
          ${navItem("/settings", "Settings", "cog")}
          ${navItem("/backup", "Backup", "arrowUpTray")}
          ${navItem("/restore", "Restore", "arrowDownTray")}
          ${navItem("/about", "About", "informationCircle")}
        </aside>
      </div>
    </div>
  `;
}

export function toast(message, kind = "info") {
  const el = document.createElement("div");
  el.className = "toast toast-top toast-end z-50";
  el.innerHTML = `<div class="alert alert-${kind}"><span>${escapeHtml(message)}</span></div>`;
  document.body.appendChild(el);
  setTimeout(() => el.remove(), 3500);
}

export function escapeHtml(str) {
  const div = document.createElement("div");
  div.textContent = str ?? "";
  return div.innerHTML;
}
