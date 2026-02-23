import { defineConfig } from "vite";
import path from "path";

export default defineConfig({
  resolve: {
    alias: {
      "@haira/arp": path.resolve(__dirname, "../../arp/packages/arp/src/index.ts"),
    },
  },
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
