import { baseCSS, logoSvg, esc } from "../core";
import type { WorkflowMeta } from "../core/types";
import { applyTheme } from "../services/theme-manager";

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
    if (this.meta) {
      const name = this.meta.title || this.meta.name;
      document.title = name ? `${name} — Haira` : "Haira";
    }

    this.renderApp();
  }

  private renderApp() {
    const shadow = this.attachShadow({ mode: "open" });
    shadow.innerHTML = `
      <style>
        ${baseCSS}
        :host { display: block; height: 100vh; overflow: hidden; background: var(--haira-bg); }
        .shell { height: 100%; display: flex; flex-direction: column; overflow: hidden; }
        header {
          padding: 0.6rem 1.25rem; border-bottom: 1px solid var(--haira-border);
          display: flex; align-items: center; gap: 0.6rem;
          background: var(--haira-bg); position: sticky; top: 0; z-index: 100;
        }
        .logo { display: flex; align-items: center; gap: 0.4rem; text-decoration: none; }
        .logo-icon { display: flex; align-items: center; }
        .logo-icon img { width: 22px; height: 22px; object-fit: contain; }
        .logo-text { font-weight: 700; font-size: 0.92rem; color: var(--haira-text); letter-spacing: -0.01em; }
        .logo-text .ai { color: var(--haira-accent); }
        .sep { color: var(--haira-muted); font-size: 0.75rem; opacity: 0.5; }
        .title { color: var(--haira-text-dim); font-size: 0.85rem; font-weight: 500; }
        main { flex: 1; display: flex; flex-direction: column; overflow: hidden; min-height: 0; }
        main.scrollable { overflow-y: auto; }
      </style>
      <div class="shell">
        <header>
          <a class="logo" href="/_ui/">
            <span class="logo-icon">${this.meta?.logo ? `<img src="${esc(this.meta.logo)}" alt="logo">` : logoSvg}</span>
            <span class="logo-text">home</span>
          </a>
          ${
            this.meta && this.meta.mode !== "index"
              ? `
            <span class="sep">/</span>
            <span class="title">${esc(this.meta.title || this.meta.name || "")}</span>
          `
              : ""
          }
        </header>
        <main id="content" class="${this.meta?.mode !== "chat" ? "scrollable" : ""}"></main>
      </div>
    `;

    applyTheme(shadow.host as HTMLElement, {
      theme: this.meta?.theme,
      accent: this.meta?.accent,
    });

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
}
