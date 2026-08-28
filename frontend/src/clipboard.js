import { api } from "./api.js";
import { toast } from "./components/layout.js";

/**
 * Write plain text to the native OS clipboard and schedule an
 * auto-clear based on the user's configured timeout (Settings ->
 * Clipboard Auto-Clear). Shared by the Home card "Copy" action and the
 * View page's copy affordances so both behave identically.
 */
export async function copyToClipboardWithAutoClear(text) {
  await api.setClipboardText(text);
  toast("Copied to clipboard.", "success");

  try {
    const settings = await api.getSettings();
    const seconds = settings.clipboardClearSeconds || 20;
    if (seconds > 0) {
      setTimeout(() => api.clearClipboard(), seconds * 1000);
    }
  } catch {
    // A failed settings fetch shouldn't undo the copy that just happened.
  }
}
