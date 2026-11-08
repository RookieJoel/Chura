import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import path from "path";
import { pathToFileURL } from "url";
import zlib from "zlib";

// Only on Node releases older than undici 8 supports; see vitest.node-polyfill.mjs.
const hasZstd = typeof (zlib as { createZstdDecompress?: unknown }).createZstdDecompress === "function";
const nodePolyfillArgs = hasZstd
  ? []
  : ["--import", pathToFileURL(path.resolve(__dirname, "vitest.node-polyfill.mjs")).href];

export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./vitest.setup.ts"],
    execArgv: nodePolyfillArgs,
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
});
