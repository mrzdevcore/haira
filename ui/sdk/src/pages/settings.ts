import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles } from "../core/styles";
import { iconStrings } from "../core/icons";
import type { WorkflowMeta } from "../core/types";

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
        align-items: baseline;
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
        <p class="page-desc">Server configuration and API endpoints.</p>

        <!-- Server Info -->
        <div class="section">
          <h2 class="section-title">
            ${unsafeHTML(iconStrings.settings)} Server
          </h2>
          <div class="card">
            <div class="kv-row">
              <span class="kv-key">Theme</span>
              <span class="kv-value">${this.meta?.theme || "dark"}</span>
            </div>
            <div class="kv-row">
              <span class="kv-key">Accent</span>
              <span class="kv-value"
                >${this.meta?.accent || "#e8a317"}
                <span
                  style="display:inline-block;width:12px;height:12px;border-radius:3px;background:${this.meta?.accent || "#e8a317"};vertical-align:middle;margin-left:0.4rem"
                ></span>
              </span>
            </div>
            <div class="kv-row">
              <span class="kv-key">Workflows</span>
              <span class="kv-value">${workflows.length}</span>
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
