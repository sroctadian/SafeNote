// Reusable PIN-entry modal. Usage:
//   const pin = await promptPin("Enter PIN to open this note");
//   if (pin === null) { /* user cancelled */ }
export function promptPin(title = "Enter PIN") {
  return new Promise((resolve) => {
    const existing = document.getElementById("pin-modal");
    if (existing) existing.remove();

    const wrapper = document.createElement("div");
    wrapper.id = "pin-modal";
    wrapper.innerHTML = `
      <div class="modal modal-open">
        <div class="modal-box max-w-sm">
          <h3 class="font-bold text-lg mb-4">${title}</h3>
          <input id="pin-input" type="password" inputmode="numeric" autocomplete="off" pattern="\\d{6}"
            class="input input-bordered w-full text-center tracking-[0.5em] text-xl"
            maxlength="6" placeholder="••••••" />
          <p id="pin-error" class="text-error text-sm mt-2 hidden"></p>
          <div class="modal-action">
            <button id="pin-cancel" class="btn btn-ghost">Cancel</button>
            <button id="pin-submit" class="btn btn-primary">Unlock</button>
          </div>
        </div>
      </div>
    `;
    document.body.appendChild(wrapper);

    const input = wrapper.querySelector("#pin-input");
    input.focus();
    input.addEventListener("input", () => {
      input.value = input.value.replace(/\D/g, "").slice(0, 6);
    });

    const cleanup = () => wrapper.remove();

    wrapper.querySelector("#pin-cancel").addEventListener("click", () => {
      cleanup();
      resolve(null);
    });

    const submit = () => {
      const value = input.value.trim();
      if (!/^\d{6}$/.test(value)) {
        showPinError("PIN must be exactly 6 digits (0-9).");
        return;
      }
      cleanup();
      resolve(value);
    };

    wrapper.querySelector("#pin-submit").addEventListener("click", submit);
    input.addEventListener("keydown", (e) => {
      if (e.key === "Enter") submit();
      if (e.key === "Escape") {
        cleanup();
        resolve(null);
      }
    });
  });
}

export function showPinError(message) {
  const el = document.getElementById("pin-error");
  if (!el) return;
  el.textContent = message;
  el.classList.remove("hidden");
}
