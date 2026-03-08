/// <reference types="vitest" />
import { defineConfig } from "vitest/config";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import tailwindcss from "@tailwindcss/vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte(), tailwindcss()],
  resolve: {
    conditions: ["browser", "svelte"],
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./src/vitest-setup.ts"],
    css: true,
    server: {
      deps: {
        inline: ["flowbite-svelte"],
      },
    },
  },
});
