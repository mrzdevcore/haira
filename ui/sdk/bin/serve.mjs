#!/usr/bin/env node

/**
 * @haira/ui standalone server
 *
 * Serves the Haira UI SDK and proxies ARP/API requests to a Haira backend.
 *
 * Usage:
 *   npx @haira/ui --connect localhost:8080
 *   npx @haira/ui --connect localhost:8080 --port 3000
 */

import { createServer } from "node:http";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

const __dirname = dirname(fileURLToPath(import.meta.url));
const distDir = join(__dirname, "..", "dist");

// Parse CLI args
const args = process.argv.slice(2);
let backend = "localhost:8080";
let port = 3000;

for (let i = 0; i < args.length; i++) {
  if ((args[i] === "--connect" || args[i] === "-c") && args[i + 1]) {
    backend = args[i + 1];
    i++;
  } else if ((args[i] === "--port" || args[i] === "-p") && args[i + 1]) {
    port = parseInt(args[i + 1], 10);
    i++;
  } else if (args[i] === "--help" || args[i] === "-h") {
    console.log(`
@haira/ui — Standalone ARP renderer

Usage:
  haira-ui --connect <host:port> [--port <port>]

Options:
  --connect, -c  Haira backend address (default: localhost:8080)
  --port, -p     Local server port (default: 3000)
  --help, -h     Show this help
`);
    process.exit(0);
  }
}

if (!backend.startsWith("http")) {
  backend = `http://${backend}`;
}

// Load the UI bundle
let uiBundle;
try {
  uiBundle = readFileSync(join(distDir, "haira-ui.js"), "utf-8");
} catch {
  console.error("Error: dist/haira-ui.js not found. Run `npm run build` first.");
  process.exit(1);
}

// Loader HTML template
const loaderHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Haira</title>
<style>body{margin:0;background:#09090b}</style>
</head>
<body>
<script type="application/json" id="haira-meta">{{META}}</script>
<haira-app></haira-app>
<script type="module" src="/_ui/assets/haira-ui.js"></script>
</body>
</html>`;

const server = createServer(async (req, res) => {
  const url = new URL(req.url, `http://localhost:${port}`);

  // Serve UI bundle
  if (url.pathname === "/_ui/assets/haira-ui.js") {
    res.writeHead(200, {
      "Content-Type": "application/javascript; charset=utf-8",
      "Cache-Control": "public, max-age=3600",
    });
    res.end(uiBundle);
    return;
  }

  // UI config — point to local bundle
  if (url.pathname === "/_ui/config") {
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ src: "/_ui/assets/haira-ui.js" }));
    return;
  }

  // Serve loader HTML for UI routes
  if (url.pathname === "/_ui/" || url.pathname.startsWith("/_ui/")) {
    // Fetch metadata from backend
    try {
      const metaRes = await fetch(`${backend}/_api/workflows`);
      const workflows = await metaRes.json();
      const meta = JSON.stringify({ mode: "index", workflows });
      const html = loaderHTML.replace("{{META}}", meta);
      res.writeHead(200, { "Content-Type": "text/html; charset=utf-8" });
      res.end(html);
    } catch {
      const html = loaderHTML.replace("{{META}}", JSON.stringify({ mode: "index", workflows: [] }));
      res.writeHead(200, { "Content-Type": "text/html; charset=utf-8" });
      res.end(html);
    }
    return;
  }

  // Proxy API and ARP requests to backend
  if (url.pathname.startsWith("/_api/") || url.pathname.startsWith("/_arp/")) {
    try {
      const proxyUrl = `${backend}${req.url}`;
      const headers = { ...req.headers };
      delete headers.host;

      const proxyRes = await fetch(proxyUrl, {
        method: req.method,
        headers,
        body: req.method !== "GET" && req.method !== "HEAD" ? req : undefined,
        duplex: "half",
      });

      res.writeHead(proxyRes.status, Object.fromEntries(proxyRes.headers));
      const body = await proxyRes.arrayBuffer();
      res.end(Buffer.from(body));
    } catch (err) {
      res.writeHead(502, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ error: `Backend unreachable: ${err.message}` }));
    }
    return;
  }

  // Redirect root to /_ui/
  if (url.pathname === "/") {
    res.writeHead(302, { Location: "/_ui/" });
    res.end();
    return;
  }

  res.writeHead(404);
  res.end("Not found");
});

server.listen(port, () => {
  console.log(`\n  @haira/ui renderer`);
  console.log(`  Local:   http://localhost:${port}/_ui/`);
  console.log(`  Backend: ${backend}`);
  console.log(`  ARP:     ws://${backend.replace(/^https?:\/\//, "")}/_arp/v1\n`);
});
