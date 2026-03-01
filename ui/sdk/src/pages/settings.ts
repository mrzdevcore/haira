import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles } from "../core/styles";
import { iconStrings } from "../core/icons";
import { applyTheme } from "../services/theme-manager";
import type { WorkflowMeta } from "../core/types";

const STORAGE_KEY_THEME = "haira-ui-theme";
const STORAGE_KEY_ACCENT = "haira-ui-accent";

@customElement("haira-page-settings")
export class PageSettings extends LitElement {
  static styles = [
    baseStyles,
    css`
      :host {
        display: block;
        padding: 2rem 1.5rem;
      }
      .container {
        max-width: 800px;
        margin: 0 auto;
      }
      .page-desc {
        color: var(--haira-muted);
        font-size: 0.85rem;
        margin-bottom: 1.5rem;
      }

      .section {
        margin-bottom: 2rem;
      }
      .section-title {
        font-size: 0.95rem;
        font-weight: 700;
        color: var(--haira-text);
        margin-bottom: 0.75rem;
        display: flex;
        align-items: center;
        gap: 0.5rem;
      }
      .card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        padding: 1rem 1.25rem;
      }
      .kv-row {
        display: flex;
        align-items: center;
        gap: 1rem;
        padding: 0.5rem 0;
        border-bottom: 1px solid var(--haira-border);
        font-size: 0.85rem;
      }
      .kv-row:last-child {
        border-bottom: none;
      }
      .kv-key {
        color: var(--haira-muted);
        min-width: 140px;
        flex-shrink: 0;
        font-size: 0.78rem;
        text-transform: uppercase;
        letter-spacing: 0.03em;
      }
      .kv-value {
        color: var(--haira-text);
        font-family: var(--haira-mono);
        font-size: 0.82rem;
        word-break: break-all;
        flex: 1;
      }

      /* Theme toggle */
      .theme-toggle {
        display: flex;
        gap: 0;
        border-radius: 6px;
        overflow: hidden;
        border: 1px solid var(--haira-border);
      }
      .theme-btn {
        padding: 0.35rem 0.75rem;
        font-size: 0.78rem;
        font-weight: 600;
        cursor: pointer;
        border: none;
        background: transparent;
        color: var(--haira-muted);
        transition: all 0.15s;
        font-family: inherit;
      }
      .theme-btn:hover {
        color: var(--haira-text);
        background: var(--haira-bg-card-hover);
      }
      .theme-btn.active {
        background: var(--haira-accent);
        color: #000;
      }

      /* Accent color picker */
      .accent-row {
        display: flex;
        align-items: center;
        gap: 0.6rem;
      }
      .color-input {
        width: 32px;
        height: 32px;
        border: 2px solid var(--haira-border);
        border-radius: 6px;
        cursor: pointer;
        padding: 0;
        background: none;
        -webkit-appearance: none;
        appearance: none;
      }
      .color-input::-webkit-color-swatch-wrapper {
        padding: 2px;
      }
      .color-input::-webkit-color-swatch {
        border: none;
        border-radius: 3px;
      }
      .color-input::-moz-color-swatch {
        border: none;
        border-radius: 3px;
      }
      .accent-presets {
        display: flex;
        gap: 0.35rem;
        margin-left: 0.5rem;
      }
      .preset-dot {
        width: 20px;
        height: 20px;
        border-radius: 50%;
        cursor: pointer;
        border: 2px solid transparent;
        transition: border-color 0.15s, transform 0.1s;
      }
      .preset-dot:hover {
        transform: scale(1.15);
      }
      .preset-dot.active {
        border-color: var(--haira-text);
      }
      .accent-hex {
        font-family: var(--haira-mono);
        font-size: 0.8rem;
        color: var(--haira-text-dim);
        margin-left: 0.4rem;
      }

      /* Reset button */
      .reset-btn {
        padding: 0.35rem 0.75rem;
        font-size: 0.75rem;
        font-weight: 500;
        cursor: pointer;
        border: 1px solid var(--haira-border);
        border-radius: 6px;
        background: transparent;
        color: var(--haira-muted);
        transition: all 0.15s;
        font-family: inherit;
        margin-left: auto;
      }
      .reset-btn:hover {
        color: var(--haira-text);
        background: var(--haira-bg-card-hover);
        border-color: var(--haira-border-focus);
      }

      .endpoints-list {
        display: flex;
        flex-direction: column;
        gap: 0.35rem;
      }
      .endpoint {
        display: flex;
        align-items: center;
        gap: 0.6rem;
        padding: 0.5rem 0.65rem;
        background: var(--haira-bg);
        border-radius: var(--haira-radius-sm);
        font-size: 0.82rem;
      }
      .endpoint-method {
        font-size: 0.6rem;
        font-weight: 700;
        padding: 0.12rem 0.4rem;
        border-radius: 3px;
        color: #fff;
        flex-shrink: 0;
      }
      .endpoint-path {
        font-family: var(--haira-mono);
        color: var(--haira-text-dim);
      }
      .endpoint-desc {
        color: var(--haira-muted);
        font-size: 0.75rem;
        margin-left: auto;
      }
    `,
  ];

  @property({ type: Object }) meta: WorkflowMeta | null = null;

  @state() private _theme: string = "dark";
  @state() private _accent: string = "#e8a317";

  private _accentPresets = [
    "#e8a317", // Haira Gold
    "#F6821F", // Cloudflare Orange
    "#3b82f6", // Blue
    "#22c55e", // Green
    "#ef4444", // Red
    "#a855f7", // Purple
    "#ec4899", // Pink
    "#06b6d4", // Cyan
  ];

  connectedCallback() {
    super.connectedCallback();
    // Load from localStorage, fall back to meta, fall back to defaults
    this._theme =
      localStorage.getItem(STORAGE_KEY_THEME) ||
      this.meta?.theme ||
      "dark";
    this._accent =
      localStorage.getItem(STORAGE_KEY_ACCENT) ||
      this.meta?.accent ||
      "#e8a317";
  }

  updated(changed: Map<string, unknown>) {
    if (changed.has("meta") && this.meta) {
      // Only apply meta defaults if no localStorage override exists
      if (!localStorage.getItem(STORAGE_KEY_THEME)) {
        this._theme = this.meta.theme || "dark";
      }
      if (!localStorage.getItem(STORAGE_KEY_ACCENT)) {
        this._accent = this.meta.accent || "#e8a317";
      }
    }
  }

  private _setTheme(theme: string) {
    this._theme = theme;
    localStorage.setItem(STORAGE_KEY_THEME, theme);
    this._applyToApp();
  }

  private _setAccent(accent: string) {
    this._accent = accent;
    localStorage.setItem(STORAGE_KEY_ACCENT, accent);
    this._applyToApp();
  }

  private _resetPreferences() {
    localStorage.removeItem(STORAGE_KEY_THEME);
    localStorage.removeItem(STORAGE_KEY_ACCENT);
    this._theme = this.meta?.theme || "dark";
    this._accent = this.meta?.accent || "#e8a317";
    this._applyToApp();
  }

  private _applyToApp() {
    const app = document.querySelector("haira-app");
    if (app) {
      applyTheme(app as HTMLElement, {
        theme: this._theme,
        accent: this._accent,
      });
    }
  }

  private _methodColor(m: string): string {
    switch (m) {
      case "GET":
        return "#3b82f6";
      case "POST":
        return "#22c55e";
      case "PUT":
        return "#f59e0b";
      case "DELETE":
        return "#ef4444";
      default:
        return "#71717a";
    }
  }

  render() {
    const workflows = this.meta?.workflows || [];

    return html`
      <div class="container">
        <p class="page-desc">Appearance preferences and API endpoints.</p>

        <!-- Appearance -->
        <div class="section">
          <h2 class="section-title">
            ${unsafeHTML(iconStrings.settings)} Appearance
          </h2>
          <div class="card">
            <div class="kv-row">
              <span class="kv-key">Theme</span>
              <div class="kv-value">
                <div class="theme-toggle">
                  <button
                    class="theme-btn ${this._theme === "dark" ? "active" : ""}"
                    @click=${() => this._setTheme("dark")}
                  >
                    Dark
                  </button>
                  <button
                    class="theme-btn ${this._theme === "light" ? "active" : ""}"
                    @click=${() => this._setTheme("light")}
                  >
                    Light
                  </button>
                </div>
              </div>
            </div>
            <div class="kv-row">
              <span class="kv-key">Accent</span>
              <div class="kv-value">
                <div class="accent-row">
                  <input
                    type="color"
                    class="color-input"
                    .value=${this._accent}
                    @input=${(e: Event) =>
                      this._setAccent(
                        (e.target as HTMLInputElement).value
                      )}
                  />
                  <span class="accent-hex">${this._accent}</span>
                  <div class="accent-presets">
                    ${this._accentPresets.map(
                      (c) => html`
                        <span
                          class="preset-dot ${this._accent === c
                            ? "active"
                            : ""}"
                          style="background:${c}"
                          @click=${() => this._setAccent(c)}
                        ></span>
                      `
                    )}
                  </div>
                </div>
              </div>
            </div>
            <div class="kv-row">
              <span class="kv-key">Workflows</span>
              <span class="kv-value">${workflows.length}</span>
              <button class="reset-btn" @click=${() => this._resetPreferences()}>
                Reset to defaults
              </button>
            </div>
          </div>
        </div>

        <!-- API Endpoints -->
        <div class="section">
          <h2 class="section-title">
            ${unsafeHTML(iconStrings.workflow)} API Endpoints
          </h2>
          <div class="card">
            <div class="endpoints-list">
              <!-- System endpoints -->
              <div class="endpoint">
                <span
                  class="endpoint-method"
                  style="background:${this._methodColor("GET")}"
                  >GET</span
                >
                <span class="endpoint-path">/_ui/</span>
                <span class="endpoint-desc">UI Console</span>
              </div>
              <div class="endpoint">
                <span
                  class="endpoint-method"
                  style="background:${this._methodColor("GET")}"
                  >GET</span
                >
                <span class="endpoint-path">/_api/runs</span>
                <span class="endpoint-desc">List runs</span>
              </div>
              <div class="endpoint">
                <span
                  class="endpoint-method"
                  style="background:${this._methodColor("GET")}"
                  >GET</span
                >
                <span class="endpoint-path">/_api/chats</span>
                <span class="endpoint-desc">List chat sessions</span>
              </div>
              <div class="endpoint">
                <span
                  class="endpoint-method"
                  style="background:${this._methodColor("GET")}"
                  >GET</span
                >
                <span class="endpoint-path">/_observe</span>
                <span class="endpoint-desc">Observability</span>
              </div>

              <!-- Workflow endpoints -->
              ${workflows.map(
                (wf) => html`
                  <div class="endpoint">
                    <span
                      class="endpoint-method"
                      style="background:${this._methodColor(wf.method)}"
                      >${wf.method}</span
                    >
                    <span class="endpoint-path">${wf.path}</span>
                    <span class="endpoint-desc"
                      >${wf.title || wf.name}</span
                    >
                  </div>
                `
              )}
            </div>
          </div>
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-page-settings": PageSettings;
  }
}
