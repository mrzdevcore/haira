import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { styleMap } from "lit/directives/style-map.js";
import { baseStyles, scrollbarStyles, methodColor, uiTypeColor } from "../core/styles";
import { iconStrings } from "../core/icons";
import { relativeTime, shortId } from "../core/utils";
import { navigate } from "../core/router";
import type { WorkflowMeta, RunSummary, ChatSessionSummary, DeploymentItem } from "../core/types";

@customElement("haira-page-home")
export class HairaPageHome extends LitElement {
  static styles = [
    baseStyles,
    scrollbarStyles,
    css`
      :host {
        display: block;
        padding: 1.5rem;
        animation: fadeIn 0.2s ease-out;
      }

      .home-container {
        max-width: 1200px;
        margin: 0 auto;
      }

      /* Header */
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

      /* Stats row */
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
      .stat-label .stat-icon {
        display: flex;
        align-items: center;
        opacity: 0.6;
      }
      .stat-value {
        font-size: 1.65rem;
        font-weight: 700;
        color: var(--haira-accent);
        letter-spacing: -0.02em;
      }

      /* Section headers */
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
      .section-title .section-icon {
        display: flex;
        align-items: center;
        color: var(--haira-muted);
      }
      .section-link {
        font-size: 0.75rem;
        color: var(--haira-accent);
        cursor: pointer;
        text-decoration: none;
        font-weight: 500;
        transition: opacity 0.12s;
      }
      .section-link:hover {
        opacity: 0.8;
      }

      /* Workflow cards grid */
      .workflow-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
        gap: 0.75rem;
      }
      .workflow-card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        padding: 1rem 1.1rem;
        cursor: pointer;
        transition: all 0.15s ease;
        text-decoration: none;
        color: inherit;
        display: flex;
        flex-direction: column;
        gap: 0.55rem;
        animation: fadeSlideUp 0.3s ease-out both;
      }
      .workflow-card:hover {
        background: var(--haira-bg-card-hover);
        border-color: var(--haira-border-focus);
        transform: translateY(-1px);
      }
      .workflow-card-top {
        display: flex;
        align-items: center;
        gap: 0.5rem;
      }
      .method-badge {
        font-size: 0.62rem;
        font-weight: 700;
        font-family: var(--haira-mono);
        letter-spacing: 0.03em;
        padding: 0.15rem 0.4rem;
        border-radius: 4px;
        flex-shrink: 0;
      }
      .workflow-name {
        font-size: 0.88rem;
        font-weight: 600;
        color: var(--haira-text);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }
      .workflow-card-bottom {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 0.5rem;
      }
      .workflow-path {
        font-size: 0.72rem;
        font-family: var(--haira-mono);
        color: var(--haira-muted);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }
      .ui-type-pill {
        font-size: 0.62rem;
        font-weight: 600;
        padding: 0.12rem 0.45rem;
        border-radius: 20px;
        flex-shrink: 0;
        text-transform: lowercase;
      }

      /* List-style sections (chats / runs) */
      .two-col {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 1.25rem;
      }

      .list-card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        overflow: hidden;
      }
      .list-item {
        padding: 0.75rem 1rem;
        border-bottom: 1px solid var(--haira-border);
        display: flex;
        align-items: center;
        gap: 0.65rem;
        transition: background 0.12s;
        cursor: pointer;
        animation: fadeSlideUp 0.3s ease-out both;
      }
      .list-item:last-child {
        border-bottom: none;
      }
      .list-item:hover {
        background: var(--haira-bg-card-hover);
      }
      .list-item-body {
        flex: 1;
        min-width: 0;
      }
      .list-item-title {
        font-size: 0.82rem;
        font-weight: 600;
        color: var(--haira-text);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }
      .list-item-sub {
        font-size: 0.72rem;
        color: var(--haira-muted);
        display: flex;
        align-items: center;
        gap: 0.5rem;
        margin-top: 0.15rem;
      }
      .list-item-meta {
        font-size: 0.7rem;
        color: var(--haira-muted);
        flex-shrink: 0;
        text-align: right;
      }
      .list-item-meta .time {
        display: block;
      }
      .list-item-meta .count {
        font-size: 0.65rem;
        color: var(--haira-text-dim);
        margin-top: 0.1rem;
      }

      /* Status dot */
      .status-dot {
        width: 7px;
        height: 7px;
        border-radius: 50%;
        flex-shrink: 0;
      }
      .status-dot.running {
        background: var(--haira-accent);
        animation: pulse 1.5s ease-in-out infinite;
      }
      .status-dot.completed {
        background: var(--haira-success);
      }
      .status-dot.failed {
        background: var(--haira-error);
      }

      .run-id {
        font-family: var(--haira-mono);
        font-size: 0.68rem;
        color: var(--haira-text-dim);
      }

      /* Empty state */
      .empty {
        padding: 2rem 1rem;
        text-align: center;
        color: var(--haira-muted);
        font-size: 0.8rem;
      }

      /* Responsive */
      @media (max-width: 900px) {
        .stats-row {
          grid-template-columns: repeat(2, 1fr);
        }
        .two-col {
          grid-template-columns: 1fr;
        }
      }
      @media (max-width: 540px) {
        :host {
          padding: 1rem;
        }
        .stats-row {
          grid-template-columns: 1fr;
        }
        .workflow-grid {
          grid-template-columns: 1fr;
        }
      }
    `,
  ];

  @property({ type: Object })
  meta: WorkflowMeta | null = null;

  @state() private _agentCount = 0;
  @state() private _runs: RunSummary[] = [];
  @state() private _chats: ChatSessionSummary[] = [];
  @state() private _deployments: DeploymentItem[] = [];

  private get _isOrchestrator(): boolean {
    return this.meta?.mode === "orchestrator";
  }

  connectedCallback() {
    super.connectedCallback();
    if (this._isOrchestrator) {
      this._loadDeployments();
    } else {
      this._loadAgents();
      this._loadRuns();
      this._loadChats();
    }
  }

  private async _loadDeployments() {
    try {
      const res = await fetch("/_api/deployments");
      if (!res.ok) return;
      const data = await res.json();
      this._deployments = Array.isArray(data) ? data : [];
    } catch {
      this._deployments = [];
    }
  }

  private async _loadAgents() {
    try {
      const res = await fetch("/_observe/api/agents");
      if (!res.ok) return;
      const data: unknown[] = await res.json();
      this._agentCount = Array.isArray(data) ? data.length : 0;
    } catch {
      this._agentCount = 0;
    }
  }

  private async _loadRuns() {
    try {
      const res = await fetch("/_api/runs");
      if (!res.ok) return;
      const data = await res.json();
      this._runs = Array.isArray(data) ? data : [];
    } catch {
      this._runs = [];
    }
  }

  private async _loadChats() {
    try {
      const res = await fetch("/_api/chats");
      if (!res.ok) return;
      const data = await res.json();
      this._chats = Array.isArray(data) ? data : [];
    } catch {
      this._chats = [];
    }
  }

  private get _workflowCount(): number {
    return this.meta?.workflows?.length ?? 0;
  }

  private _navigateWorkflow(path: string, query?: Record<string, string>) {
    navigate({ page: "workbench", path, query });
  }

  private _renderOrchestratorHome() {
    const deps = this._deployments;
    const running = deps.filter((d) => d.status === "running").length;
    const stopped = deps.filter((d) => d.status === "stopped").length;
    const crashed = deps.filter((d) => d.status === "crashed").length;

    return html`
      <div class="home-container">
        <div class="page-header">
          <h1>Haira Orchestrator</h1>
          <p>Overview of all deployed workflows and agents.</p>
        </div>

        <div class="stats-row">
          <div class="stat-card">
            <span class="stat-label">
              <span class="stat-icon">${unsafeHTML(iconStrings.workflow)}</span>
              Total Deployments
            </span>
            <span class="stat-value">${deps.length}</span>
          </div>
          <div class="stat-card">
            <span class="stat-label">
              <span class="stat-icon">${unsafeHTML(iconStrings.check)}</span>
              Running
            </span>
            <span class="stat-value" style="color: var(--haira-success)">${running}</span>
          </div>
          <div class="stat-card">
            <span class="stat-label">
              <span class="stat-icon">${unsafeHTML(iconStrings.pending)}</span>
              Stopped
            </span>
            <span class="stat-value">${stopped}</span>
          </div>
          <div class="stat-card">
            <span class="stat-label">
              <span class="stat-icon">${unsafeHTML(iconStrings.x)}</span>
              Crashed
            </span>
            <span class="stat-value" style="color: var(--haira-error)">${crashed}</span>
          </div>
        </div>

        ${deps.length > 0
          ? html`
              <div class="section">
                <div class="section-header">
                  <span class="section-title">
                    <span class="section-icon">${unsafeHTML(iconStrings.activity)}</span>
                    Deployments
                  </span>
                  <a
                    class="section-link"
                    href="/deployments"
                    @click=${(e: Event) => {
                      e.preventDefault();
                      navigate({ page: "deployments" });
                    }}
                  >Manage ${unsafeHTML(iconStrings.chevronRight)}</a>
                </div>
                <div class="workflow-grid">
                  ${deps.map(
                    (dep, i) => html`
                      <a
                        class="workflow-card"
                        href="${dep.url}_ui/"
                        target="_blank"
                        rel="noopener"
                        style=${styleMap({
                          animationDelay: `${200 + i * 40}ms`,
                        })}
                      >
                        <div class="workflow-card-top">
                          <span
                            class="method-badge"
                            style=${styleMap({
                              color:
                                dep.status === "running"
                                  ? "var(--haira-success)"
                                  : dep.status === "crashed"
                                    ? "var(--haira-error)"
                                    : "var(--haira-muted)",
                              background:
                                dep.status === "running"
                                  ? "var(--haira-success)18"
                                  : dep.status === "crashed"
                                    ? "var(--haira-error)18"
                                    : "var(--haira-muted)18",
                            })}
                          >${dep.status.toUpperCase()}</span>
                          <span class="workflow-name">${dep.name}</span>
                        </div>
                        <div class="workflow-card-bottom">
                          <span class="workflow-path">:${dep.port}</span>
                          <span class="workflow-path">${relativeTime(dep.created_at)}</span>
                        </div>
                      </a>
                    `
                  )}
                </div>
              </div>
            `
          : html`
              <div class="section" style="text-align:center;padding:3rem 1rem;color:var(--haira-muted)">
                <p style="font-size:0.9rem;margin-bottom:0.5rem">No deployments yet.</p>
                <p style="font-size:0.8rem">Run <code style="background:var(--haira-bg-card);padding:0.2rem 0.45rem;border-radius:4px;font-family:var(--haira-mono)">haira deploy &lt;file&gt;</code> to get started.</p>
              </div>
            `}
      </div>
    `;
  }

  render() {
    if (this._isOrchestrator) return this._renderOrchestratorHome();

    const workflows = this.meta?.workflows ?? [];
    const recentRuns = this._runs.slice(0, 8);
    const recentChats = this._chats.slice(0, 8);

    return html`
      <div class="home-container">
        <div class="page-header">
          <h1>${this.meta?.title || this.meta?.name || "Console"}</h1>
          <p>Dashboard overview for your Haira application.</p>
        </div>

        <!-- Stats -->
        <div class="stats-row">
          <div class="stat-card">
            <span class="stat-label">
              <span class="stat-icon">${unsafeHTML(iconStrings.workflow)}</span>
              Total Workflows
            </span>
            <span class="stat-value">${this._workflowCount}</span>
          </div>
          <div class="stat-card">
            <span class="stat-label">
              <span class="stat-icon">${unsafeHTML(iconStrings.agent)}</span>
              Active Agents
            </span>
            <span class="stat-value">${this._agentCount}</span>
          </div>
          <div class="stat-card">
            <span class="stat-label">
              <span class="stat-icon">${unsafeHTML(iconStrings.activity)}</span>
              Recent Runs
            </span>
            <span class="stat-value">${this._runs.length}</span>
          </div>
          <div class="stat-card">
            <span class="stat-label">
              <span class="stat-icon">${unsafeHTML(iconStrings.chat)}</span>
              Chat Sessions
            </span>
            <span class="stat-value">${this._chats.length}</span>
          </div>
        </div>

        <!-- Workflows -->
        ${workflows.length > 0
          ? html`
              <div class="section">
                <div class="section-header">
                  <span class="section-title">
                    <span class="section-icon">${unsafeHTML(iconStrings.workflow)}</span>
                    Workflows
                  </span>
                  <a
                    class="section-link"
                    href="/workflows"
                    @click=${(e: Event) => {
                      e.preventDefault();
                      navigate({ page: "workflows" });
                    }}
                  >View all ${unsafeHTML(iconStrings.chevronRight)}</a>
                </div>
                <div class="workflow-grid">
                  ${workflows.map(
                    (wf, i) => html`
                      <a
                        class="workflow-card"
                        href="/workbench${wf.path}"
                        style=${styleMap({
                          animationDelay: `${200 + i * 40}ms`,
                        })}
                        @click=${(e: Event) => {
                          e.preventDefault();
                          this._navigateWorkflow(wf.path);
                        }}
                      >
                        <div class="workflow-card-top">
                          <span
                            class="method-badge"
                            style=${styleMap({
                              color: methodColor(wf.method),
                              background: `${methodColor(wf.method)}18`,
                            })}
                          >${wf.method}</span>
                          <span class="workflow-name">${wf.title || wf.name}</span>
                        </div>
                        <div class="workflow-card-bottom">
                          <span class="workflow-path">${wf.path}</span>
                          <span
                            class="ui-type-pill"
                            style=${styleMap({
                              color: uiTypeColor(wf.uiType),
                              background: `${uiTypeColor(wf.uiType)}18`,
                            })}
                          >${wf.uiType}</span>
                        </div>
                      </a>
                    `
                  )}
                </div>
              </div>
            `
          : nothing}

        <!-- Recent Chats & Runs -->
        <div class="two-col">
          <!-- Recent Chats -->
          <div class="section">
            <div class="section-header">
              <span class="section-title">
                <span class="section-icon">${unsafeHTML(iconStrings.chat)}</span>
                Recent Chats
              </span>
            </div>
            <div class="list-card">
              ${recentChats.length > 0
                ? recentChats.map(
                    (chat, i) => html`
                      <div
                        class="list-item"
                        style=${styleMap({
                          animationDelay: `${300 + i * 40}ms`,
                        })}
                        @click=${() =>
                          this._navigateWorkflow(chat.workflow_path, { session: chat.id })}
                      >
                        <div class="list-item-body">
                          <div class="list-item-title">${chat.title || "Untitled"}</div>
                          <div class="list-item-sub">
                            <span>${chat.workflow_name}</span>
                          </div>
                        </div>
                        <div class="list-item-meta">
                          <span class="time">${relativeTime(chat.updated_at)}</span>
                          <span class="count">${chat.message_count} msg${chat.message_count !== 1 ? "s" : ""}</span>
                        </div>
                      </div>
                    `
                  )
                : html`<div class="empty">No chat sessions yet.</div>`}
            </div>
          </div>

          <!-- Recent Runs -->
          <div class="section">
            <div class="section-header">
              <span class="section-title">
                <span class="section-icon">${unsafeHTML(iconStrings.activity)}</span>
                Recent Runs
              </span>
            </div>
            <div class="list-card">
              ${recentRuns.length > 0
                ? recentRuns.map(
                    (run, i) => html`
                      <div
                        class="list-item"
                        style=${styleMap({
                          animationDelay: `${300 + i * 40}ms`,
                        })}
                        @click=${() =>
                          this._navigateWorkflow(run.workflow_path, { run: run.id })}
                      >
                        <span class="status-dot ${run.status}"></span>
                        <div class="list-item-body">
                          <div class="list-item-title">${run.workflow_name}</div>
                          <div class="list-item-sub">
                            <span class="run-id">${shortId(run.id)}</span>
                          </div>
                        </div>
                        <div class="list-item-meta">
                          <span class="time">${relativeTime(run.started_at)}</span>
                        </div>
                      </div>
                    `
                  )
                : html`<div class="empty">No runs recorded yet.</div>`}
            </div>
          </div>
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-page-home": HairaPageHome;
  }
}
