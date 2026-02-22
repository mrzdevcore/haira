import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { styleMap } from "lit/directives/style-map.js";
import { baseStyles, scrollbarStyles } from "../core/styles";
import { iconStrings } from "../core/icons";
import { relativeTime } from "../core/utils";
import type { WorkflowMeta, DeploymentItem } from "../core/types";

function statusColor(status: string): string {
  switch (status) {
    case "running":
      return "var(--haira-success)";
    case "stopped":
      return "var(--haira-muted)";
    case "crashed":
      return "var(--haira-error)";
    case "deploying":
      return "var(--haira-accent)";
    default:
      return "var(--haira-muted)";
  }
}

@customElement("haira-page-deployments")
export class HairaPageDeployments extends LitElement {
  static styles = [
    baseStyles,
    scrollbarStyles,
    css`
      :host {
        display: block;
        padding: 1.5rem;
        animation: fadeIn 0.2s ease-out;
      }
      .container {
        max-width: 1200px;
        margin: 0 auto;
      }
      .page-header {
        margin-bottom: 1.75rem;
      }
      .page-header h1 {
        font-size: 1.35rem;
        font-weight: 700;
        color: var(--haira-text);
        margin-bottom: 0.25rem;
      }
      .page-header p {
        font-size: 0.82rem;
        color: var(--haira-muted);
      }

      /* Stats */
      .stats-row {
        display: grid;
        grid-template-columns: repeat(4, 1fr);
        gap: 0.85rem;
        margin-bottom: 2rem;
      }
      .stat-card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        padding: 1.1rem 1.15rem;
        display: flex;
        flex-direction: column;
        gap: 0.35rem;
        animation: fadeSlideUp 0.3s ease-out both;
      }
      .stat-card:nth-child(1) { animation-delay: 0ms; }
      .stat-card:nth-child(2) { animation-delay: 50ms; }
      .stat-card:nth-child(3) { animation-delay: 100ms; }
      .stat-card:nth-child(4) { animation-delay: 150ms; }
      .stat-label {
        font-size: 0.72rem;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.04em;
        color: var(--haira-muted);
        display: flex;
        align-items: center;
        gap: 0.4rem;
      }
      .stat-icon { display: flex; align-items: center; opacity: 0.6; }
      .stat-value {
        font-size: 1.65rem;
        font-weight: 700;
        color: var(--haira-accent);
        letter-spacing: -0.02em;
      }

      /* Deployment cards */
      .section {
        margin-bottom: 2rem;
      }
      .section-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 0.85rem;
      }
      .section-title {
        font-size: 0.95rem;
        font-weight: 650;
        color: var(--haira-text);
        display: flex;
        align-items: center;
        gap: 0.45rem;
      }
      .section-icon {
        display: flex;
        align-items: center;
        color: var(--haira-muted);
      }

      .deploy-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
        gap: 0.85rem;
      }
      .deploy-card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        padding: 1.1rem 1.2rem;
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
        animation: fadeSlideUp 0.3s ease-out both;
        transition: border-color 0.15s;
      }
      .deploy-card:hover {
        border-color: var(--haira-border-focus);
      }
      .deploy-card-header {
        display: flex;
        align-items: center;
        gap: 0.6rem;
      }
      .status-dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        flex-shrink: 0;
      }
      .status-dot.running {
        animation: pulse 1.5s ease-in-out infinite;
      }
      .deploy-name {
        font-size: 0.95rem;
        font-weight: 650;
        color: var(--haira-text);
        flex: 1;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }
      .deploy-name a {
        color: inherit;
        text-decoration: none;
      }
      .deploy-name a:hover {
        color: var(--haira-accent);
      }
      .status-badge {
        font-size: 0.65rem;
        font-weight: 600;
        padding: 0.15rem 0.5rem;
        border-radius: 20px;
        text-transform: uppercase;
        letter-spacing: 0.03em;
      }
      .deploy-meta {
        display: flex;
        gap: 1rem;
        flex-wrap: wrap;
      }
      .meta-item {
        font-size: 0.72rem;
        color: var(--haira-muted);
        display: flex;
        align-items: center;
        gap: 0.3rem;
      }
      .meta-item .label {
        color: var(--haira-text-dim);
      }
      .meta-item .value {
        font-family: var(--haira-mono);
        color: var(--haira-text);
      }
      .deploy-actions {
        display: flex;
        gap: 0.4rem;
        border-top: 1px solid var(--haira-border);
        padding-top: 0.7rem;
        flex-wrap: wrap;
      }
      .action-btn {
        font-size: 0.72rem;
        font-weight: 600;
        padding: 0.3rem 0.65rem;
        border-radius: 5px;
        border: 1px solid var(--haira-border);
        background: var(--haira-bg);
        color: var(--haira-text-dim);
        cursor: pointer;
        transition: all 0.12s;
        display: flex;
        align-items: center;
        gap: 0.3rem;
      }
      .action-btn:hover {
        background: var(--haira-bg-card-hover);
        color: var(--haira-text);
        border-color: var(--haira-border-focus);
      }
      .action-btn.danger:hover {
        color: var(--haira-error);
        border-color: var(--haira-error);
      }
      .action-btn:disabled {
        opacity: 0.4;
        cursor: not-allowed;
      }
      .action-btn .btn-icon {
        display: flex;
        align-items: center;
      }

      /* Logs panel */
      .logs-panel {
        margin-top: 0.5rem;
        background: var(--haira-bg);
        border: 1px solid var(--haira-border);
        border-radius: 6px;
        max-height: 200px;
        overflow-y: auto;
        padding: 0.5rem 0.75rem;
        font-family: var(--haira-mono);
        font-size: 0.7rem;
        line-height: 1.6;
        color: var(--haira-text-dim);
        white-space: pre-wrap;
        word-break: break-all;
      }

      /* Empty */
      .empty {
        text-align: center;
        padding: 3rem 1rem;
        color: var(--haira-muted);
      }
      .empty-icon {
        font-size: 2rem;
        margin-bottom: 0.75rem;
        opacity: 0.4;
      }
      .empty h2 {
        font-size: 1.1rem;
        font-weight: 600;
        color: var(--haira-text-dim);
        margin-bottom: 0.35rem;
      }
      .empty p {
        font-size: 0.82rem;
      }
      .empty code {
        background: var(--haira-bg-card);
        padding: 0.2rem 0.45rem;
        border-radius: 4px;
        font-family: var(--haira-mono);
        font-size: 0.78rem;
      }

      @media (max-width: 900px) {
        .stats-row { grid-template-columns: repeat(2, 1fr); }
      }
      @media (max-width: 540px) {
        :host { padding: 1rem; }
        .stats-row { grid-template-columns: 1fr; }
        .deploy-grid { grid-template-columns: 1fr; }
      }
    `,
  ];

  @property({ type: Object })
  meta: WorkflowMeta | null = null;

  @state() private _deployments: DeploymentItem[] = [];
  @state() private _logsOpen: Record<string, boolean> = {};
  @state() private _logsData: Record<string, string> = {};
  @state() private _actionLoading: Record<string, boolean> = {};

  private _pollTimer: ReturnType<typeof setInterval> | null = null;

  connectedCallback() {
    super.connectedCallback();
    this._loadDeployments();
    this._pollTimer = setInterval(() => this._loadDeployments(), 5000);
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    if (this._pollTimer) {
      clearInterval(this._pollTimer);
      this._pollTimer = null;
    }
  }

  private async _loadDeployments() {
    try {
      const res = await fetch("/_api/deployments");
      if (!res.ok) return;
      const data = await res.json();
      this._deployments = Array.isArray(data) ? data : [];
    } catch {
      // keep existing data on error
    }
  }

  private get _running() {
    return this._deployments.filter((d) => d.status === "running").length;
  }
  private get _stopped() {
    return this._deployments.filter((d) => d.status === "stopped").length;
  }
  private get _crashed() {
    return this._deployments.filter((d) => d.status === "crashed").length;
  }

  private async _doAction(name: string, action: "stop" | "restart") {
    this._actionLoading = { ...this._actionLoading, [name]: true };
    try {
      await fetch(`/_api/deployments/${name}/${action}`, { method: "POST" });
      await this._loadDeployments();
    } catch {
      // ignore
    }
    this._actionLoading = { ...this._actionLoading, [name]: false };
  }

  private async _undeploy(name: string) {
    if (!confirm(`Remove deployment "${name}"? This deletes the binary and logs.`)) return;
    this._actionLoading = { ...this._actionLoading, [name]: true };
    try {
      await fetch(`/_api/deployments/${name}`, { method: "DELETE" });
      await this._loadDeployments();
    } catch {
      // ignore
    }
    this._actionLoading = { ...this._actionLoading, [name]: false };
  }

  private async _toggleLogs(name: string) {
    const open = !this._logsOpen[name];
    this._logsOpen = { ...this._logsOpen, [name]: open };
    if (open && !this._logsData[name]) {
      try {
        const res = await fetch(`/_api/deployments/${name}/logs`);
        if (res.ok) {
          const text = await res.text();
          this._logsData = { ...this._logsData, [name]: text || "(empty)" };
        } else {
          this._logsData = { ...this._logsData, [name]: "(no logs available)" };
        }
      } catch {
        this._logsData = { ...this._logsData, [name]: "(failed to load logs)" };
      }
    }
  }

  private _renderCard(dep: DeploymentItem, i: number) {
    const loading = this._actionLoading[dep.name] || false;
    const logsOpen = this._logsOpen[dep.name] || false;
    const logsText = this._logsData[dep.name] || "";
    const isRunning = dep.status === "running";

    return html`
      <div
        class="deploy-card"
        style=${styleMap({ animationDelay: `${200 + i * 40}ms` })}
      >
        <div class="deploy-card-header">
          <span
            class="status-dot ${dep.status}"
            style=${styleMap({ background: statusColor(dep.status) })}
          ></span>
          <span class="deploy-name">
            <a href="${dep.url}_ui/" target="_blank" rel="noopener">${dep.name}</a>
          </span>
          <span
            class="status-badge"
            style=${styleMap({
              color: statusColor(dep.status),
              background: `${statusColor(dep.status)}18`,
            })}
          >${dep.status}</span>
        </div>

        <div class="deploy-meta">
          <span class="meta-item">
            <span class="label">Port</span>
            <span class="value">${dep.port}</span>
          </span>
          <span class="meta-item">
            <span class="label">PID</span>
            <span class="value">${dep.pid || "\u2014"}</span>
          </span>
          ${dep.restarts > 0
            ? html`<span class="meta-item">
                <span class="label">Restarts</span>
                <span class="value">${dep.restarts}</span>
              </span>`
            : nothing}
          <span class="meta-item">
            <span class="label">Created</span>
            <span class="value">${relativeTime(dep.created_at)}</span>
          </span>
        </div>

        <div class="deploy-actions">
          ${isRunning
            ? html`<button
                class="action-btn"
                ?disabled=${loading}
                @click=${() => this._doAction(dep.name, "stop")}
              >
                <span class="btn-icon">${unsafeHTML(iconStrings.x)}</span>
                Stop
              </button>`
            : html`<button
                class="action-btn"
                ?disabled=${loading}
                @click=${() => this._doAction(dep.name, "restart")}
              >
                <span class="btn-icon">${unsafeHTML(iconStrings.retry)}</span>
                Start
              </button>`}
          ${isRunning
            ? html`<button
                class="action-btn"
                ?disabled=${loading}
                @click=${() => this._doAction(dep.name, "restart")}
              >
                <span class="btn-icon">${unsafeHTML(iconStrings.retry)}</span>
                Restart
              </button>`
            : nothing}
          <button
            class="action-btn"
            @click=${() => this._toggleLogs(dep.name)}
          >
            <span class="btn-icon">${unsafeHTML(iconStrings.file)}</span>
            Logs
          </button>
          ${isRunning
            ? html`<a
                class="action-btn"
                href="${dep.url}_ui/"
                target="_blank"
                rel="noopener"
                style="text-decoration:none"
              >
                <span class="btn-icon">${unsafeHTML(iconStrings.chevronRight)}</span>
                Console
              </a>`
            : nothing}
          <button
            class="action-btn danger"
            ?disabled=${loading}
            @click=${() => this._undeploy(dep.name)}
          >
            <span class="btn-icon">${unsafeHTML(iconStrings.trash)}</span>
            Remove
          </button>
        </div>

        ${logsOpen
          ? html`<div class="logs-panel">${logsText}</div>`
          : nothing}
      </div>
    `;
  }

  render() {
    const deps = this._deployments;

    return html`
      <div class="container">
        <div class="page-header">
          <h1>Deployments</h1>
          <p>Manage your deployed Haira workflows and agents.</p>
        </div>

        <div class="stats-row">
          <div class="stat-card">
            <span class="stat-label">
              <span class="stat-icon">${unsafeHTML(iconStrings.workflow)}</span>
              Total
            </span>
            <span class="stat-value">${deps.length}</span>
          </div>
          <div class="stat-card">
            <span class="stat-label">
              <span class="stat-icon">${unsafeHTML(iconStrings.check)}</span>
              Running
            </span>
            <span class="stat-value" style="color: var(--haira-success)">${this._running}</span>
          </div>
          <div class="stat-card">
            <span class="stat-label">
              <span class="stat-icon">${unsafeHTML(iconStrings.pending)}</span>
              Stopped
            </span>
            <span class="stat-value" style="color: var(--haira-muted)">${this._stopped}</span>
          </div>
          <div class="stat-card">
            <span class="stat-label">
              <span class="stat-icon">${unsafeHTML(iconStrings.x)}</span>
              Crashed
            </span>
            <span class="stat-value" style="color: var(--haira-error)">${this._crashed}</span>
          </div>
        </div>

        ${deps.length > 0
          ? html`
              <div class="section">
                <div class="section-header">
                  <span class="section-title">
                    <span class="section-icon">${unsafeHTML(iconStrings.activity)}</span>
                    All Deployments
                  </span>
                </div>
                <div class="deploy-grid">
                  ${deps.map((dep, i) => this._renderCard(dep, i))}
                </div>
              </div>
            `
          : html`
              <div class="empty">
                <div class="empty-icon">${unsafeHTML(iconStrings.workflow)}</div>
                <h2>No deployments yet</h2>
                <p>Deploy your first workflow with <code>haira deploy &lt;file&gt;</code></p>
              </div>
            `}
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-page-deployments": HairaPageDeployments;
  }
}
