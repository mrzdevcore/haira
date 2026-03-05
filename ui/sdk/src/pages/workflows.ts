import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles, scrollbarStyles, methodColor, uiTypeColor } from "../core/styles";
import { iconStrings } from "../core/icons";
import { relativeTime, shortId, formatMs } from "../core/utils";
import { navigate } from "../core/router";
import type { WorkflowMeta, RunSummary } from "../core/types";

@customElement("haira-page-workflows")
export class PageWorkflows extends LitElement {
  static styles = [
    baseStyles,
    scrollbarStyles,
    css`
      :host {
        display: block;
        padding: 2rem 1.5rem;
      }
      .container {
        max-width: 1200px;
        margin: 0 auto;
      }
      .page-desc {
        color: var(--haira-muted);
        font-size: 0.85rem;
        margin-bottom: 1.5rem;
      }

      /* Workflow cards */
      .wf-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
        gap: 0.75rem;
        margin-bottom: 2rem;
      }
      .wf-card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        padding: 1rem;
        cursor: pointer;
        transition: all 0.15s;
        animation: fadeSlideUp 0.3s ease-out both;
      }
      .wf-card:hover {
        border-color: rgba(232, 163, 23, 0.3);
        background: var(--haira-bg-card-hover);
      }
      .wf-header {
        display: flex;
        align-items: center;
        gap: 0.6rem;
        margin-bottom: 0.5rem;
      }
      .badge {
        font-size: 0.6rem;
        font-weight: 700;
        padding: 0.12rem 0.4rem;
        border-radius: 3px;
        color: #fff;
        flex-shrink: 0;
        letter-spacing: 0.02em;
      }
      .wf-name {
        font-weight: 600;
        font-size: 0.9rem;
        flex: 1;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      .type-pill {
        font-size: 0.65rem;
        font-weight: 600;
        padding: 0.12rem 0.5rem;
        border-radius: 10px;
        border: 1px solid;
        text-transform: lowercase;
        flex-shrink: 0;
      }
      .wf-path {
        font-family: var(--haira-mono);
        font-size: 0.75rem;
        color: var(--haira-muted);
      }
      .wf-meta {
        display: flex;
        gap: 1rem;
        margin-top: 0.6rem;
        font-size: 0.72rem;
        color: var(--haira-text-dim);
      }
      .wf-meta-item {
        display: flex;
        align-items: center;
        gap: 0.3rem;
      }

      /* Runs section */
      .section-title {
        font-size: 1rem;
        font-weight: 700;
        color: var(--haira-text);
        margin-bottom: 0.75rem;
        display: flex;
        align-items: center;
        gap: 0.5rem;
      }
      .runs-table {
        width: 100%;
        border-collapse: collapse;
      }
      .runs-table th {
        text-align: left;
        font-size: 0.72rem;
        font-weight: 600;
        color: var(--haira-muted);
        text-transform: uppercase;
        letter-spacing: 0.04em;
        padding: 0.5rem 0.75rem;
        border-bottom: 1px solid var(--haira-border);
      }
      .runs-table td {
        padding: 0.6rem 0.75rem;
        font-size: 0.82rem;
        border-bottom: 1px solid var(--haira-border);
        color: var(--haira-text-dim);
      }
      .runs-table tr {
        cursor: pointer;
        transition: background 0.12s;
      }
      .runs-table tbody tr:hover {
        background: var(--haira-bg-card);
      }
      .status-dot {
        display: inline-block;
        width: 7px;
        height: 7px;
        border-radius: 50%;
        margin-right: 0.4rem;
        background: var(--haira-muted);
      }
      .status-dot.completed {
        background: var(--haira-success);
      }
      .status-dot.failed {
        background: var(--haira-error);
      }
      .status-dot.running {
        background: var(--haira-accent);
        animation: pulse 1.5s ease-in-out infinite;
      }
      .status-badge {
        font-size: 0.62rem;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.03em;
        padding: 0.08rem 0.35rem;
        border-radius: 3px;
      }
      .status-badge.completed {
        color: var(--haira-success);
        background: rgba(34, 197, 94, 0.1);
      }
      .status-badge.failed {
        color: var(--haira-error);
        background: rgba(239, 68, 68, 0.1);
      }
      .status-badge.running {
        color: var(--haira-accent);
        background: rgba(232, 163, 23, 0.1);
      }
      .mono {
        font-family: var(--haira-mono);
        font-size: 0.72rem;
        color: var(--haira-muted);
      }
      .empty {
        text-align: center;
        padding: 2rem;
        color: var(--haira-muted);
        font-size: 0.85rem;
      }
    `,
  ];

  @property({ type: Object }) meta: WorkflowMeta | null = null;
  @state() private _runs: RunSummary[] = [];

  connectedCallback() {
    super.connectedCallback();
    this._loadRuns();
  }

  private async _loadRuns() {
    try {
      const resp = await fetch("/_api/runs");
      if (resp.ok) {
        this._runs = (await resp.json()) || [];
      }
    } catch {
      // silently fail
    }
  }

  render() {
    const workflows = this.meta?.workflows || [];

    return html`
      <div class="container">
        <p class="page-desc">
          All registered workflows and their execution history.
        </p>

        <div class="wf-grid">
          ${workflows.length === 0
            ? html`<div class="empty">
                No workflows registered. Define a workflow in your .haira file.
              </div>`
            : workflows.map(
                (wf, i) => html`
                  <div
                    class="wf-card"
                    style="animation-delay:${i * 40}ms"
                    @click=${() =>
                      navigate({ page: "workbench", path: wf.path })}
                  >
                    <div class="wf-header">
                      <span
                        class="badge"
                        style="background:${methodColor(wf.method)}"
                        >${wf.method}</span
                      >
                      <span class="wf-name">${wf.title || wf.name}</span>
                      <span
                        class="type-pill"
                        style="color:${uiTypeColor(
                          wf.uiType
                        )};border-color:${uiTypeColor(wf.uiType)}30"
                        >${wf.uiType}</span
                      >
                    </div>
                    <div class="wf-path">${wf.path}</div>
                    <div class="wf-meta">
                      <span class="wf-meta-item">
                        ${unsafeHTML(iconStrings.workflow)}
                        ${wf.method}
                      </span>
                    </div>
                  </div>
                `
              )}
        </div>

        ${this._runs.length > 0
          ? html`
              <h2 class="section-title">
                ${unsafeHTML(iconStrings.activity)} Recent Runs
              </h2>
              <table class="runs-table">
                <thead>
                  <tr>
                    <th>Status</th>
                    <th>Workflow</th>
                    <th>Run ID</th>
                    <th>Steps</th>
                    <th>Started</th>
                  </tr>
                </thead>
                <tbody>
                  ${this._runs.map(
                    (run) => html`
                      <tr
                        @click=${() =>
                          navigate({
                            page: "workbench",
                            path: run.workflow_path,
                          })}
                      >
                        <td>
                          <span class="status-badge ${run.status}"
                            >${run.status}</span
                          >
                        </td>
                        <td>${run.workflow_name}</td>
                        <td class="mono">${shortId(run.id)}</td>
                        <td>${run.step_count}</td>
                        <td class="mono">${relativeTime(run.started_at)}</td>
                      </tr>
                    `
                  )}
                </tbody>
              </table>
            `
          : nothing}
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-page-workflows": PageWorkflows;
  }
}
