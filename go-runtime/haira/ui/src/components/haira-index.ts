import {
  baseStyles,
  sharedKeyframes,
  methodColor,
  uiTypeColor,
} from "../theme";
import type { WorkflowMeta } from "../types";

export class HairaIndex extends HTMLElement {
  connectedCallback() {
    const meta: WorkflowMeta = JSON.parse(
      this.getAttribute("data-meta") || "{}",
    );
    const workflows = meta.workflows || [];

    const shadow = this.attachShadow({ mode: "open" });
    shadow.innerHTML = `
      <style>
        ${baseStyles}
        ${sharedKeyframes}
        :host {
          display: flex;
          justify-content: center;
          padding: 2.5rem 1rem;
        }
        .container { max-width: 520px; width: 100%; }
        h1 {
          font-size: 1.3rem;
          font-weight: 700;
          color: var(--haira-text);
          margin-bottom: 1.25rem;
        }
        .wf {
          display: flex;
          align-items: center;
          justify-content: space-between;
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          padding: 0.85rem 1rem;
          margin-bottom: 0.5rem;
          text-decoration: none;
          color: var(--haira-text);
          transition: all 0.15s;
          animation: fadeSlideUp 0.3s ease-out both;
        }
        .wf:hover {
          border-color: rgba(232, 163, 23, 0.3);
          background: var(--haira-bg-card-hover);
        }
        .wf-left {
          display: flex;
          align-items: center;
          gap: 0.6rem;
          min-width: 0;
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
        .wf-info { min-width: 0; }
        .wf-name {
          font-weight: 600;
          font-size: 0.88rem;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }
        .wf-path {
          font-family: var(--haira-mono);
          font-size: 0.75rem;
          color: var(--haira-muted);
          margin-top: 0.1rem;
        }
        .wf-right {
          display: flex;
          align-items: center;
          flex-shrink: 0;
          margin-left: 0.75rem;
        }
        .type-pill {
          font-size: 0.65rem;
          font-weight: 600;
          padding: 0.12rem 0.5rem;
          border-radius: 10px;
          border: 1px solid;
          text-transform: lowercase;
        }
        .empty {
          text-align: center;
          padding: 3rem 1rem;
          animation: fadeIn 0.4s ease-out;
        }
        .empty-title {
          color: var(--haira-text-dim);
          font-size: 0.95rem;
          font-weight: 500;
          margin-bottom: 0.25rem;
        }
        .empty-sub {
          color: var(--haira-muted);
          font-size: 0.82rem;
        }
      </style>
      <div class="container">
        <h1>Workflows</h1>
        ${
          workflows.length === 0
            ? `<div class="empty">
              <div class="empty-title">No workflows registered</div>
              <div class="empty-sub">Define a workflow in your .haira file to get started</div>
            </div>`
            : workflows
                .map(
                  (wf, i) => `
            <a class="wf" href="/_ui${wf.path}" style="animation-delay:${i * 50}ms">
              <div class="wf-left">
                <span class="badge" style="background:${methodColor(wf.method)}">${wf.method}</span>
                <div class="wf-info">
                  <div class="wf-name">${this.esc(wf.title || wf.name)}</div>
                  <div class="wf-path">${this.esc(wf.path)}</div>
                </div>
              </div>
              <div class="wf-right">
                <span class="type-pill" style="color:${uiTypeColor(wf.uiType)};border-color:${uiTypeColor(wf.uiType)}30">${wf.uiType}</span>
              </div>
            </a>
          `,
                )
                .join("")
        }
      </div>
    `;
  }

  private esc(s: string): string {
    return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }
}
