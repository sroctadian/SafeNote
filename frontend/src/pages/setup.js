import { api } from "../api.js";
import { navigate } from "../router.js";
import { toast } from "../components/layout.js";

export async function setupPage() {
  return {
    html: `
      <div class="min-h-screen flex items-center justify-center p-4">
        <div class="card bg-base-200 w-full max-w-md shadow-xl">
          <div class="card-body">
            <h2 class="card-title">Welcome to SafeNote</h2>
            <p class="text-sm opacity-70">
              Define your Secret Key. It is combined with each note's PIN to
              derive encryption keys — minimum 32 characters, any printable
              character. It is never stored in plaintext and never fully
              displayed again.
            </p>
            <textarea id="sk-input" class="textarea textarea-bordered w-full mt-2"
              rows="3" placeholder="Enter a strong Secret Key (min. 32 characters)"></textarea>
            <div id="sk-strength" class="text-xs mt-1 opacity-70">0 / 32 characters</div>
            <p id="sk-error" class="text-error text-sm hidden"></p>
            <div class="card-actions justify-end mt-4">
              <button id="sk-submit" class="btn btn-primary w-full">Set Secret Key</button>
            </div>
          </div>
        </div>
      </div>
    `,
    mount: (root) => {
      const input = root.querySelector("#sk-input");
      const strength = root.querySelector("#sk-strength");
      const error = root.querySelector("#sk-error");

      input.addEventListener("input", () => {
        strength.textContent = `${input.value.length} / 32 characters`;
        strength.classList.toggle("text-success", input.value.length >= 32);
      });

      root.querySelector("#sk-submit").addEventListener("click", async () => {
        const value = input.value;
        if (value.length < 32) {
          error.textContent = "Secret Key must be at least 32 characters.";
          error.classList.remove("hidden");
          return;
        }
        try {
          await api.setupSecretKey(value);
          toast("Secret Key configured.", "success");
          navigate("/home");
        } catch (err) {
          error.textContent = err.message;
          error.classList.remove("hidden");
        }
      });
    },
  };
}
