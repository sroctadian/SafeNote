import { layout } from "../components/layout.js";

export async function aboutPage() {
  return layout(
    "/about",
    `
    <div class="max-w-2xl mx-auto flex flex-col gap-4">
      <h1 class="text-2xl font-bold">About SafeNote</h1>
      <div class="card bg-base-200">
        <div class="card-body gap-2 text-sm">
          <p><strong>SafeNote</strong> is an offline-first, encrypted note application.</p>
          <p>Every note is protected with AES-256-GCM. Encryption keys are derived
            per-note via Argon2id from your Secret Key combined with a note-specific PIN.
            Plaintext content, PINs, and derived keys are never persisted.</p>
          <p>Stack: Go 1.24+, Wails v2, Vanilla JS, TailwindCSS + DaisyUI, SQLite.</p>
          <p class="opacity-60">No internet connection required. No telemetry.</p>
        </div>
      </div>
      <div class="text-xs opacity-50">Keyboard shortcuts: Ctrl+N new note · Ctrl+F search ·
        Ctrl+C copy note · Ctrl+S save · Esc close dialog</div>
    </div>
    `
  );
}
