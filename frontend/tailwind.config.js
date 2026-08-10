/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,html}"],
  darkMode: ["class", '[data-theme="safenoteDark"]'],
  theme: {
    extend: {
      fontFamily: {
        sans: ["Inter", "ui-sans-serif", "system-ui", "sans-serif"],
        mono: ["JetBrains Mono", "ui-monospace", "monospace"],
      },
    },
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: [
      {
        safenoteLight: {
          primary: "#2563eb",
          secondary: "#7c3aed",
          accent: "#0ea5e9",
          neutral: "#1f2937",
          "base-100": "#ffffff",
          "base-200": "#f3f4f6",
          "base-300": "#e5e7eb",
          info: "#0ea5e9",
          success: "#16a34a",
          warning: "#d97706",
          error: "#dc2626",
        },
      },
      {
        safenoteDark: {
          primary: "#3b82f6",
          secondary: "#a78bfa",
          accent: "#22d3ee",
          neutral: "#0b0f19",
          "base-100": "#0f1420",
          "base-200": "#161c2b",
          "base-300": "#1f273a",
          info: "#38bdf8",
          success: "#22c55e",
          warning: "#f59e0b",
          error: "#ef4444",
        },
      },
    ],
  },
};
