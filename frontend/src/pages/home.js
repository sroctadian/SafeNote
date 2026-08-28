import { api } from "../api.js";
import { navigate, onCleanup } from "../router.js";
import { layout, toast, escapeHtml } from "../components/layout.js";
import { promptPin, showPinError } from "../components/pinModal.js";
import { deltaToPlainText } from "../components/richEditor.js";
import { forgetUnlock } from "../noteSession.js";
import { copyToClipboardWithAutoClear } from "../clipboard.js";

const state = {
  search: "",
  sort: "newest",
  page: 1,
  pageSize: 12,
  onlyFavorite: false,
  view: "grid", // "grid" | "list"
};

function noteCardHtml(note) {
  const updated = new Date(note.updatedAt).toLocaleString();
  const created = new Date(note.createdAt).toLocaleDateString();
  return `
    <div class="note-card" data-id="${note.id}">
      <div class="flex items-start justify-between">
        <h3 class="font-semibold truncate pr-2">${escapeHtml(note.title)}</h3>
        <div class="flex gap-1 shrink-0">
          <button class="btn-icon fav-btn" data-id="${note.id}" data-fav="${note.favorite}" title="Favorite">
            ${note.favorite ? "⭐" : "☆"}
          </button>
          <button class="btn-icon pin-btn" data-id="${note.id}" data-pinned="${note.pinned}" title="Pin">
            ${note.pinned ? "📌" : "📍"}
          </button>
        </div>
      </div>
      <div class="text-xs opacity-60 mt-2">Created ${created}</div>
      <div class="text-xs opacity-60">Updated ${updated}</div>
      <div class="card-actions justify-end mt-3 gap-1">
        <button class="btn btn-xs copy-btn" data-id="${note.id}">Copy</button>
        <button class="btn btn-xs btn-primary open-btn" data-id="${note.id}">Open</button>
      </div>
    </div>
  `;
}

export async function homePage() {
  return {
    html: layout(
      "/home",
      `
      <div class="flex flex-col gap-4">
        <div class="flex flex-wrap gap-2 items-center justify-between">
          <h1 class="text-2xl font-bold">Your Notes</h1>
          <button class="btn btn-primary btn-sm" onclick="location.hash='#/create'">➕ New Note</button>
        </div>

        <div class="flex flex-wrap gap-2 items-center">
          <input id="search-input" type="text" placeholder="Search by title (Ctrl+F)"
            class="input input-bordered input-sm w-full max-w-xs" value="${escapeHtml(state.search)}" />
          <select id="sort-select" class="select select-bordered select-sm">
            <option value="newest" ${state.sort === "newest" ? "selected" : ""}>Newest</option>
            <option value="oldest" ${state.sort === "oldest" ? "selected" : ""}>Oldest</option>
            <option value="alphabet" ${state.sort === "alphabet" ? "selected" : ""}>Alphabetical</option>
          </select>
          <label class="label cursor-pointer gap-2">
            <span class="label-text text-sm">Favorites only</span>
            <input id="fav-filter" type="checkbox" class="toggle toggle-sm" ${state.onlyFavorite ? "checked" : ""} />
          </label>
          <div class="join ml-auto">
            <button id="view-grid" class="btn btn-sm join-item ${state.view === "grid" ? "btn-active" : ""}">Grid</button>
            <button id="view-list" class="btn btn-sm join-item ${state.view === "list" ? "btn-active" : ""}">List</button>
          </div>
        </div>

        <div id="notes-container" class="min-h-[200px]">
          <div class="loading loading-spinner"></div>
        </div>

        <div id="pagination" class="flex justify-center gap-2 mt-2"></div>
      </div>
      `
    ),
    mount: async (root) => {
      forgetUnlock(); // arriving at Home always exits any unlocked-note context
      await loadAndRender(root);

      root.querySelector("#search-input").addEventListener(
        "input",
        debounce((e) => {
          state.search = e.target.value;
          state.page = 1;
          loadAndRender(root);
        }, 300)
      );

      root.querySelector("#sort-select").addEventListener("change", (e) => {
        state.sort = e.target.value;
        loadAndRender(root);
      });

      root.querySelector("#fav-filter").addEventListener("change", (e) => {
        state.onlyFavorite = e.target.checked;
        state.page = 1;
        loadAndRender(root);
      });

      root.querySelector("#view-grid").addEventListener("click", () => {
        state.view = "grid";
        loadAndRender(root);
      });
      root.querySelector("#view-list").addEventListener("click", () => {
        state.view = "list";
        loadAndRender(root);
      });

      document.addEventListener("keydown", globalShortcuts);
      onCleanup(() => document.removeEventListener("keydown", globalShortcuts));
    },
  };
}

function globalShortcuts(e) {
  if (!location.hash.startsWith("#/home")) return;
  if (e.ctrlKey && e.key.toLowerCase() === "f") {
    e.preventDefault();
    document.getElementById("search-input")?.focus();
  }
  if (e.ctrlKey && e.key.toLowerCase() === "n") {
    e.preventDefault();
    navigate("/create");
  }
}

async function loadAndRender(root) {
  const container = root.querySelector("#notes-container");
  const paginationEl = root.querySelector("#pagination");
  container.innerHTML = `<div class="loading loading-spinner"></div>`;

  try {
    const result = await api.listNotes(
      state.search,
      state.sort,
      state.page,
      state.pageSize,
      state.onlyFavorite
    );
    const cards = result.notes;
    const total = result.total;

    if (!cards || cards.length === 0) {
      container.innerHTML = `<div class="text-center opacity-60 py-16">No notes found. Create your first note.</div>`;
      paginationEl.innerHTML = "";
      return;
    }

    container.className =
      state.view === "grid"
        ? "grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3"
        : "flex flex-col gap-2";
    container.innerHTML = cards.map(noteCardHtml).join("");

    attachCardHandlers(root);
    renderPagination(paginationEl, total);
  } catch (err) {
    container.innerHTML = `<div class="alert alert-error">${err.message}</div>`;
  }
}

function renderPagination(el, total) {
  const pages = Math.max(1, Math.ceil(total / state.pageSize));
  if (pages <= 1) {
    el.innerHTML = "";
    return;
  }
  let html = "";
  for (let p = 1; p <= pages; p++) {
    html += `<button class="btn btn-xs page-btn ${p === state.page ? "btn-active" : ""}" data-page="${p}">${p}</button>`;
  }
  el.innerHTML = html;
  el.querySelectorAll(".page-btn").forEach((btn) => {
    btn.addEventListener("click", () => {
      state.page = Number(btn.dataset.page);
      loadAndRender(document);
    });
  });
}

function attachCardHandlers(root) {
  // Clicking anywhere on the card opens the note (same as the "Open"
  // button); icon buttons inside stop propagation so they don't also
  // trigger a navigation.
  root.querySelectorAll(".note-card").forEach((card) =>
    card.addEventListener("click", () => navigate(`/view?id=${card.dataset.id}`))
  );

  root.querySelectorAll(".open-btn").forEach((btn) =>
    btn.addEventListener("click", (e) => {
      e.stopPropagation();
      navigate(`/view?id=${btn.dataset.id}`);
    })
  );

  root.querySelectorAll(".fav-btn").forEach((btn) =>
    btn.addEventListener("click", async (e) => {
      e.stopPropagation();
      const id = btn.dataset.id;
      const current = btn.dataset.fav === "true";
      try {
        await api.setFavorite(id, !current);
        loadAndRender(root);
      } catch (err) {
        toast(err.message, "error");
      }
    })
  );

  root.querySelectorAll(".pin-btn").forEach((btn) =>
    btn.addEventListener("click", async (e) => {
      e.stopPropagation();
      const id = btn.dataset.id;
      const current = btn.dataset.pinned === "true";
      try {
        await api.setPinned(id, !current);
        loadAndRender(root);
      } catch (err) {
        toast(err.message, "error");
      }
    })
  );

  root.querySelectorAll(".copy-btn").forEach((btn) =>
    btn.addEventListener("click", (e) => {
      e.stopPropagation();
      copyNoteFlow(btn.dataset.id);
    })
  );
}

async function copyNoteFlow(id) {
  const pin = await promptPin("Enter PIN to copy note content");
  if (pin === null) return;
  try {
    const note = await api.openNote(id, pin);
    const plainText = deltaToPlainText(note.content);
    await copyToClipboardWithAutoClear(plainText);
  } catch (err) {
    showPinError(err.message);
    toast(err.message, "error");
  }
}

function debounce(fn, delay) {
  let timer;
  return (...args) => {
    clearTimeout(timer);
    timer = setTimeout(() => fn(...args), delay);
  };
}
