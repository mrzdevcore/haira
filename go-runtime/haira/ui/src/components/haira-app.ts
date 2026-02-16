import { baseStyles, logoSvg } from "../theme";
import type { WorkflowMeta } from "../types";

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

  private render() {
    const shadow = this.attachShadow({ mode: "open" });
    shadow.innerHTML = `
      <style>
        ${baseStyles}
        :host {
          display: block;
          min-height: 100vh;
          background: var(--haira-bg);
        }
        .shell {
          min-height: 100vh;
          display: flex;
          flex-direction: column;
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
        .logo-text {
          font-weight: 700;
          font-size: 0.92rem;
          color: var(--haira-text);
          letter-spacing: -0.01em;
        }
        .logo-text .ai {
          color: var(--haira-gold);
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
        }
      </style>
      <div class="shell">
        <header>
          <a class="logo" href="/_ui/">
            <span class="logo-icon">${logoSvg}</span>
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
