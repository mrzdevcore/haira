import { defineConfig } from "vite";

export default defineConfig({
  build: {
    outDir: "dist",
    emptyOutDir: true,
    rollupOptions: {
      input: "src/index.ts",
      output: {
        entryFileNames: "haira-ui.js",
        format: "es",
      },
    },
    target: "es2021",
    minify: true,
  },
  server: {
    proxy: {
      "/_api": "http://localhost:8080",
      "/_observe": "http://localhost:8080",
      "/_arp": {
        target: "http://localhost:8080",
        ws: true,
      },
    },
  },
});
