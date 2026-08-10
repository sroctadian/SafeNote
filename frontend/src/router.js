const routes = new Map();
let notFoundHandler = () => `<div class="p-8">Page not found.</div>`;

export function registerRoute(path, renderFn) {
  routes.set(path, renderFn);
}

export function setNotFound(fn) {
  notFoundHandler = fn;
}

function currentPath() {
  const hash = window.location.hash.replace(/^#/, "");
  return hash || "/splash";
}

export function navigate(path) {
  window.location.hash = path;
}

export async function renderRoute() {
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
