import { baseStyles, logoSvg } from "../theme";
import type { WorkflowMeta } from "../types";

const lightThemeVars = `
  --haira-bg: #ffffff;
  --haira-bg-card: #f7f7f8;
  --haira-bg-card-hover: #eeeff1;
  --haira-bg-elevated: #e8e8ec;
  --haira-bg-input: #f2f2f4;
  --haira-border: rgba(0, 0, 0, 0.1);
  --haira-border-light: rgba(0, 0, 0, 0.06);
  --haira-border-focus: rgba(0, 0, 0, 0.25);
  --haira-text: #1a1a1a;
  --haira-text-dim: #4a4a4a;
  --haira-muted: #8a8a8a;
`;

function hexToRgb(hex: string): { r: number; g: number; b: number } | null {
  const m = hex.match(/^#?([0-9a-f]{6})$/i);
  if (!m) return null;
  return {
    r: parseInt(m[1].substring(0, 2), 16),
    g: parseInt(m[1].substring(2, 4), 16),
    b: parseInt(m[1].substring(4, 6), 16),
  };
}

function lighten(hex: string, amount: number): string {
  const rgb = hexToRgb(hex);
  if (!rgb) return hex;
  const r = Math.min(255, rgb.r + Math.round((255 - rgb.r) * amount));
  const g = Math.min(255, rgb.g + Math.round((255 - rgb.g) * amount));
  const b = Math.min(255, rgb.b + Math.round((255 - rgb.b) * amount));
  return `#${r.toString(16).padStart(2, "0")}${g.toString(16).padStart(2, "0")}${b.toString(16).padStart(2, "0")}`;
}

export class HairaApp extends HTMLElement {
  private meta: WorkflowMeta | null = null;

  connectedCallback() {
    const metaEl = document.getElementById("haira-meta");
    if (metaEl) {
      try {
        this.meta = JSON.parse(metaEl.textContent || "{}");
      } catch {
        this.meta = null;
      }
    }
    this.render();
  }

  private applyTheme(host: HTMLElement) {
    if (!this.meta) return;

    // Light theme overrides
    if (this.meta.theme === "light") {
      for (const line of lightThemeVars.split("\n")) {
        const match = line.match(/(--[\w-]+):\s*(.+);/);
        if (match) host.style.setProperty(match[1], match[2].trim());
      }
    }

    // Accent color overrides
    const accent = this.meta.accent;
    if (accent) {
      host.style.setProperty("--haira-accent", accent);
      host.style.setProperty("--haira-accent-light", lighten(accent, 0.25));
      const rgb = hexToRgb(accent);
      if (rgb) {
        host.style.setProperty(
          "--haira-accent-dim",
          `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, 0.06)`,
        );
        host.style.setProperty(
          "--haira-border-light",
          `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, 0.12)`,
        );
        host.style.setProperty(
          "--haira-border-focus",
          `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, 0.4)`,
        );
      }
    }
  }

  private render() {
    const shadow = this.attachShadow({ mode: "open" });
    shadow.innerHTML = `
      <style>
        ${baseStyles}
        :host {
          display: block;
          height: 100vh;
          overflow: hidden;
          background: var(--haira-bg);
        }
        .shell {
          height: 100%;
          display: flex;
          flex-direction: column;
          overflow: hidden;
        }
        header {
          padding: 0.6rem 1.25rem;
          border-bottom: 1px solid var(--haira-border);
          display: flex;
          align-items: center;
          gap: 0.6rem;
          background: var(--haira-bg);
          position: sticky;
          top: 0;
          z-index: 100;
        }
        .logo {
          display: flex;
          align-items: center;
          gap: 0.4rem;
          text-decoration: none;
        }
        .logo-icon {
          display: flex;
          align-items: center;
        }
        .logo-icon img {
          width: 22px;
          height: 22px;
          object-fit: contain;
        }
        .logo-text {
          font-weight: 700;
          font-size: 0.92rem;
          color: var(--haira-text);
          letter-spacing: -0.01em;
        }
        .logo-text .ai {
          color: var(--haira-accent);
        }
        .sep {
          color: var(--haira-muted);
          font-size: 0.75rem;
          opacity: 0.5;
        }
        .title {
          color: var(--haira-text-dim);
          font-size: 0.85rem;
          font-weight: 500;
        }
        main {
          flex: 1;
          display: flex;
          flex-direction: column;
          overflow: hidden;
          min-height: 0;
        }
      </style>
      <div class="shell">
        <header>
          <a class="logo" href="/_ui/">
            <span class="logo-icon">${this.meta?.logo ? `<img src="${this.escapeHtml(this.meta.logo)}" alt="logo">` : logoSvg}</span>
            <span class="logo-text">h<span class="ai">ai</span>ra</span>
          </a>
          ${
            this.meta && this.meta.mode !== "index"
              ? `
            <span class="sep">/</span>
            <span class="title">${this.escapeHtml(this.meta.title || this.meta.name || "")}</span>
          `
              : ""
          }
        </header>
        <main id="content"></main>
      </div>
    `;

    // Apply theme overrides to the shadow host
    this.applyTheme(shadow.host as HTMLElement);

    const content = shadow.getElementById("content")!;

    if (!this.meta) {
      content.innerHTML = `<p style="padding:2rem;color:var(--haira-muted)">No workflow metadata found.</p>`;
      return;
    }

    switch (this.meta.mode) {
      case "index": {
        const el = document.createElement("haira-index");
        el.setAttribute("data-meta", JSON.stringify(this.meta));
        content.appendChild(el);
        break;
      }
      case "form": {
        const el = document.createElement("haira-form");
        el.setAttribute("data-meta", JSON.stringify(this.meta));
        content.appendChild(el);
        break;
      }
      case "chat": {
        const el = document.createElement("haira-chat");
        el.setAttribute("data-meta", JSON.stringify(this.meta));
        content.appendChild(el);
        break;
      }
    }
  }

  private escapeHtml(str: string): string {
    return str
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;");
  }
}
