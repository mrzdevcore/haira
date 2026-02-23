import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import dts from "vite-plugin-dts";
import { resolve } from "path";

export default defineConfig({
  plugins: [
    vue(),
    dts({ rollupTypes: true, tsconfigPath: "./tsconfig.json" }),
  ],
  build: {
    lib: {
      entry: {
        "arp-vue": resolve(__dirname, "src/index.ts"),
        ui: resolve(__dirname, "src/ui/index.ts"),
      },
      formats: ["es"],
    },
    rollupOptions: {
      external: ["vue", "@haira/arp"],
      output: {
        globals: {
          vue: "Vue",
        },
      },
    },
  },
});
