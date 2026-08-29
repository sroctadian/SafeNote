import { api } from "../api.js";
import { navigate } from "../router.js";
import logoUrl from "../assets/logo.png";

export async function splashPage() {
  return {
    html: `
      <div class="min-h-screen flex flex-col items-center justify-center gap-4">
        <img src="${logoUrl}" alt="SafeNote" class="w-16 h-16" />
        <div class="text-2xl font-bold">SafeNote</div>
        <div class="loading loading-dots loading-md"></div>
      </div>
    `,
    mount: async () => {
      try {
        const configured = await api.isSecretKeyConfigured();
        navigate(configured ? "/home" : "/setup");
      } catch (err) {
        document.getElementById("app").innerHTML = `
          <div class="min-h-screen flex items-center justify-center p-8 text-center">
            <div class="alert alert-error max-w-md">${err.message}</div>
          </div>`;
      }
    },
  };
}
