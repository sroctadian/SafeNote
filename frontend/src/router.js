const routes = new Map();
let notFoundHandler = () => `<div class="p-8">Page not found.</div>`;
let currentCleanups = [];

export function registerRoute(path, renderFn) {
  routes.set(path, renderFn);
}

export function setNotFound(fn) {
  notFoundHandler = fn;
}

/**
 * Register a cleanup callback to run the next time the route changes.
 * Pages should call this immediately after adding any document/window
 * -level event listener (e.g. a keydown shortcut), so the listener
 * doesn't outlive the page that created it. Safe to call from either a
 * synchronous or async mount(), and from callbacks that fire later
 * (e.g. inside a button click handler) as long as the user is still on
 * that route when it's called.
 *
 * Deliberately NOT relying on mount()'s return value: several pages
 * have an async mount(), and the router does not await it (fire-and-
 * forget, so slow pages don't block rendering) — by the time an async
 * mount's return value resolved, the route may have already changed,
 * so a returned cleanup function would frequently arrive too late to
 * be captured correctly.
 */
export function onCleanup(fn) {
  if (typeof fn === "function") {
    currentCleanups.push(fn);
  }
}

function runCleanups() {
  for (const fn of currentCleanups) {
    try {
      fn();
    } catch (err) {
      console.error("router: cleanup error", err);
    }
  }
  currentCleanups = [];
}

function currentPath() {
  const hash = window.location.hash.replace(/^#/, "");
  return hash || "/splash";
}

export function navigate(path) {
  window.location.hash = path;
}

export async function renderRoute() {
  runCleanups();

  const [path, query] = currentPath().split("?");
  const params = Object.fromEntries(new URLSearchParams(query || ""));

  const renderFn = routes.get(path) || notFoundHandler;
  const result = await renderFn(params);

  const root = document.getElementById("app");
  if (typeof result === "string") {
    root.innerHTML = result;
    return;
  }
  root.innerHTML = result.html;
  if (typeof result.mount === "function") {
    result.mount(root);
  }
}

export function startRouter() {
  window.addEventListener("hashchange", renderRoute);
  renderRoute();
}
