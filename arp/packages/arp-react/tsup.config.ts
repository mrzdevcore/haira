import { defineConfig } from "tsup";

export default defineConfig([
  {
    entry: ["src/index.ts"],
    format: ["cjs", "esm"],
    dts: true,
    external: ["react", "react-dom", "@haira/arp"],
  },
  {
    entry: ["src/ui/index.ts"],
    outDir: "dist/ui",
    format: ["cjs", "esm"],
    dts: true,
    external: ["react", "react-dom", "@haira/arp"],
  },
]);
