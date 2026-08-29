// Thin wrapper around the Wails-generated Go bindings
// (window.go.app.App.*). Centralizing calls here keeps pages free of
// binding-path details and gives one place to add error normalization.

function backend() {
  if (!window.go || !window.go.app || !window.go.app.App) {
    throw new Error(
      "Wails backend not available. Run via `wails dev` / the packaged app, not a plain browser."
    );
  }
  return window.go.app.App;
}

function normalizeError(err) {
  const message = typeof err === "string" ? err : err?.message || "Unexpected error";
  return new Error(message);
}

async function call(method, ...args) {
  try {
    return await backend()[method](...args);
  } catch (err) {
    throw normalizeError(err);
  }
}

export const api = {
  // Settings / first run
  isSecretKeyConfigured: () => call("IsSecretKeyConfigured"),
  setupSecretKey: (secretKey) => call("SetupSecretKey", secretKey),
  changeSecretKey: (newSecretKey) => call("ChangeSecretKey", newSecretKey),
  getMaskedSecretKey: () => call("GetMaskedSecretKey"),
  getSettings: () => call("GetSettings"),
  updateTheme: (theme) => call("UpdateTheme", theme),
  updateClipboardTimeout: (seconds) => call("UpdateClipboardTimeout", seconds),
  exportConfig: (path) => call("ExportConfig", path),
  importConfig: (path) => call("ImportConfig", path),
  getDataDirectory: () => call("GetDataDirectory"),
  selectDirectoryDialog: () => call("SelectDirectoryDialog"),
  setDataDirectory: (path) => call("SetDataDirectory", path),

  // Notes
  createNote: (title, content, pin, tags) => call("CreateNote", title, content, pin, tags || []),
  openNote: (id, pin) => call("OpenNote", id, pin),
  copyNoteToClipboard: (id, pin) => call("CopyNoteToClipboard", id, pin),
  setClipboardText: (text) => call("SetClipboardText", text),
  clearClipboard: () => call("ClearClipboard"),
  editNote: (id, pin, title, content, tags) => call("EditNote", id, pin, title, content, tags || []),
  deleteNote: (id) => call("DeleteNote", id),
  setFavorite: (id, favorite) => call("SetFavorite", id, favorite),
  setPinned: (id, pinned) => call("SetPinned", id, pinned),
  listNotes: (search, sort, page, pageSize, onlyFavorite) =>
    call("ListNotes", search || "", sort || "newest", page || 1, pageSize || 20, !!onlyFavorite),

  // Backup / restore
  exportBackup: (path) => call("ExportBackup", path),
  previewRestore: (path) => call("PreviewRestore", path),
  restoreBackup: (path, overwriteIds) => call("RestoreBackup", path, overwriteIds || []),

  // Dialogs
  saveFileDialog: (defaultFilename) => call("SaveFileDialog", defaultFilename),
  openFileDialog: () => call("OpenFileDialog"),
};
