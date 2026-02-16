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
          padding: 2rem 1rem;
        }
        .container { max-width: 600px; width: 100%; }
        h1 {
          font-size: 1.35rem;
          font-weight: 600;
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
          padding: 1rem 1.25rem;
          margin-bottom: 0.6rem;
          text-decoration: none;
          color: var(--haira-text);
          transition: all 0.2s;
          animation: fadeSlideUp 0.3s ease-out both;
        }
        .wf:hover {
          border-color: var(--haira-gold);
          transform: translateY(-1px);
          box-shadow: 0 4px 20px rgba(232, 163, 23, 0.12);
        }
        .wf-name {
          font-weight: 600;
          font-size: 0.92rem;
        }
        .wf-path {
          font-family: var(--haira-mono);
          font-size: 0.8rem;
          color: var(--haira-muted);
          margin-top: 0.15rem;
        }
        .badge {
          font-size: 0.65rem;
          font-weight: 700;
          padding: 0.12rem 0.5rem;
          border-radius: 3px;
          color: #fff;
          margin-right: 0.5rem;
        }
        .wf-right {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          flex-shrink: 0;
        }
        .type-pill {
          font-size: 0.68rem;
          font-weight: 600;
          padding: 0.15rem 0.55rem;
          border-radius: 12px;
          border: 1px solid;
          text-transform: lowercase;
        }
        .empty {
          text-align: center;
          padding: 3rem 1rem;
          animation: fadeIn 0.4s ease-out;
        }
        .empty-icon {
          font-size: 2.5rem;
          color: var(--haira-gold);
          opacity: 0.3;
          margin-bottom: 0.75rem;
        }
        .empty-title {
          color: var(--haira-text-dim);
          font-size: 1rem;
          font-weight: 500;
          margin-bottom: 0.3rem;
        }
        .empty-sub {
          color: var(--haira-muted);
          font-size: 0.85rem;
        }
      </style>
      <div class="container">
        <h1>Workflows</h1>
        ${
          workflows.length === 0
            ? `<div class="empty">
              <div class="empty-icon">\u2699</div>
              <div class="empty-title">No workflows registered yet</div>
              <div class="empty-sub">Define a workflow in your .haira file to get started</div>
            </div>`
            : workflows
                .map(
                  (wf, i) => `
            <a class="wf" href="/_ui${wf.path}" style="animation-delay:${i * 60}ms">
              <div>
                <span class="badge" style="background:${methodColor(wf.method)}">${wf.method}</span>
                <span class="wf-name">${this.esc(wf.title || wf.name)}</span>
                <div class="wf-path">${this.esc(wf.path)}</div>
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
