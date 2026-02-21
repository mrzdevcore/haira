import { baseCSS, keyframes, methodColor, uiTypeColor, esc } from "../core";
import type { WorkflowMeta, RunSummary, ChatSessionSummary } from "../core/types";

export class HairaIndex extends HTMLElement {
  connectedCallback() {
    const meta: WorkflowMeta = JSON.parse(this.getAttribute("data-meta") || "{}");
    const workflows = meta.workflows || [];

    const shadow = this.attachShadow({ mode: "open" });
    shadow.innerHTML = `
      <style>
        ${baseCSS}
        :host { display: flex; justify-content: center; padding: 2.5rem 1rem; }
        .container { max-width: 960px; width: 100%; }
        @media (min-width: 768px) { :host { padding: 2.5rem 2rem; } }
        h1 { font-size: 1.3rem; font-weight: 700; color: var(--haira-text); margin-bottom: 1.25rem; }
        .wf {
          display: flex; align-items: center; justify-content: space-between;
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius); padding: 0.85rem 1rem; margin-bottom: 0.5rem;
          text-decoration: none; color: var(--haira-text); transition: all 0.15s;
          animation: fadeSlideUp 0.3s ease-out both;
        }
        .wf:hover { border-color: rgba(232, 163, 23, 0.3); background: var(--haira-bg-card-hover); }
        .wf-left { display: flex; align-items: center; gap: 0.6rem; min-width: 0; }
        .badge {
          font-size: 0.6rem; font-weight: 700; padding: 0.12rem 0.4rem;
          border-radius: 3px; color: #fff; flex-shrink: 0; letter-spacing: 0.02em;
        }
        .wf-info { min-width: 0; }
        .wf-name {
          font-weight: 600; font-size: 0.88rem; white-space: nowrap;
          overflow: hidden; text-overflow: ellipsis;
        }
        .wf-path { font-family: var(--haira-mono); font-size: 0.75rem; color: var(--haira-muted); margin-top: 0.1rem; }
        .wf-right { display: flex; align-items: center; flex-shrink: 0; margin-left: 0.75rem; }
        .type-pill {
          font-size: 0.65rem; font-weight: 600; padding: 0.12rem 0.5rem;
          border-radius: 10px; border: 1px solid; text-transform: lowercase;
        }
        .empty { text-align: center; padding: 3rem 1rem; animation: fadeIn 0.4s ease-out; }
        .empty-title { color: var(--haira-text-dim); font-size: 0.95rem; font-weight: 500; margin-bottom: 0.25rem; }
        .empty-sub { color: var(--haira-muted); font-size: 0.82rem; }
        .section-title { font-size: 1rem; font-weight: 700; color: var(--haira-text); margin: 2rem 0 0.75rem; }
        .run {
          display: flex; align-items: center; gap: 0.6rem;
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius-sm); padding: 0.6rem 0.85rem; margin-bottom: 0.35rem;
          text-decoration: none; color: var(--haira-text); transition: all 0.15s;
          animation: fadeIn 0.25s ease-out both;
        }
        .run:hover { border-color: rgba(232, 163, 23, 0.3); background: var(--haira-bg-card-hover); }
        .run-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
        .run-dot.completed { background: var(--haira-success); }
        .run-dot.failed { background: var(--haira-error); }
        .run-dot.running { background: var(--haira-accent); animation: pulse 1.5s ease-in-out infinite; }
        .run-name {
          flex: 1; font-size: 0.82rem; font-weight: 500;
          overflow: hidden; text-overflow: ellipsis; white-space: nowrap; min-width: 0;
        }
        .run-name .run-id { font-family: var(--haira-mono); font-size: 0.7rem; color: var(--haira-muted); margin-left: 0.4rem; }
        .run-time { font-size: 0.72rem; font-family: var(--haira-mono); color: var(--haira-muted); flex-shrink: 0; }
        .run-status {
          font-size: 0.62rem; font-weight: 600; text-transform: uppercase;
          letter-spacing: 0.03em; flex-shrink: 0; padding: 0.08rem 0.35rem; border-radius: 3px;
        }
        .run-status.completed { color: var(--haira-success); background: rgba(34, 197, 94, 0.1); }
        .run-status.failed { color: var(--haira-error); background: rgba(239, 68, 68, 0.1); }
        .run-status.running { color: var(--haira-accent); background: rgba(232, 163, 23, 0.1); }
        .chat {
          display: flex; align-items: center; gap: 0.6rem;
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius-sm); padding: 0.6rem 0.85rem; margin-bottom: 0.35rem;
          text-decoration: none; color: var(--haira-text); transition: all 0.15s;
          animation: fadeIn 0.25s ease-out both;
        }
        .chat:hover { border-color: rgba(232, 163, 23, 0.3); background: var(--haira-bg-card-hover); }
        .chat-icon { color: var(--haira-accent); display: flex; flex-shrink: 0; opacity: 0.6; }
        .chat-title {
          flex: 1; font-size: 0.82rem; font-weight: 500;
          overflow: hidden; text-overflow: ellipsis; white-space: nowrap; min-width: 0;
        }
        .chat-wf { font-family: var(--haira-mono); font-size: 0.7rem; color: var(--haira-muted); margin-left: 0.4rem; }
        .chat-time { font-size: 0.72rem; font-family: var(--haira-mono); color: var(--haira-muted); flex-shrink: 0; }
        .chat-count {
          font-size: 0.62rem; color: var(--haira-muted); flex-shrink: 0;
          padding: 0.08rem 0.35rem; border-radius: 3px; background: var(--haira-bg-elevated);
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
                  <div class="wf-name">${esc(wf.title || wf.name)}</div>
                  <div class="wf-path">${esc(wf.path)}</div>
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
        <div id="chats-section"></div>
        <div id="runs-section"></div>
      </div>
    `;

    this.loadChats(shadow);
    this.loadRuns(shadow);
  }

  private async loadChats(shadow: ShadowRoot) {
    const section = shadow.getElementById("chats-section");
    if (!section) return;

    try {
      const resp = await fetch("/_api/chats");
      if (!resp.ok) return;
      const chats: ChatSessionSummary[] = await resp.json();
      if (!chats || chats.length === 0) return;

      const iconChat = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/></svg>`;

      section.innerHTML = `
        <h2 class="section-title">Recent Chats</h2>
        ${chats
          .map(
            (chat, i) => `
          <a class="chat" href="/_ui${esc(chat.workflow_path)}?session=${esc(chat.id)}" style="animation-delay:${i * 30}ms">
            <span class="chat-icon">${iconChat}</span>
            <span class="chat-title">${esc(chat.title || "New chat")}<span class="chat-wf">${esc(chat.workflow_name)}</span></span>
            <span class="chat-time">${this.relativeTime(chat.updated_at)}</span>
            <span class="chat-count">${chat.message_count} msg</span>
          </a>
        `,
          )
          .join("")}
      `;
    } catch {
      // Silently fail
    }
  }

  private async loadRuns(shadow: ShadowRoot) {
    const section = shadow.getElementById("runs-section");
    if (!section) return;

    try {
      const resp = await fetch("/_api/runs");
      if (!resp.ok) return;
      const runs: RunSummary[] = await resp.json();
      if (!runs || runs.length === 0) return;

      section.innerHTML = `
        <h2 class="section-title">Recent Runs</h2>
        ${runs
          .map(
            (run, i) => `
          <a class="run" href="/_ui${esc(run.workflow_path)}?run=${esc(run.id)}" style="animation-delay:${i * 30}ms">
            <span class="run-dot ${run.status}"></span>
            <span class="run-name">${esc(run.workflow_name)}<span class="run-id">${this.shortId(run.id)}</span></span>
            <span class="run-time">${this.relativeTime(run.started_at)}</span>
            <span class="run-status ${run.status}">${run.status}</span>
          </a>
        `,
          )
          .join("")}
      `;
    } catch {
      // Silently fail
    }
  }

  private relativeTime(iso: string): string {
    const now = Date.now();
    const then = new Date(iso).getTime();
    const diffMs = now - then;
    const diffSec = Math.floor(diffMs / 1000);
    if (diffSec < 60) return "just now";
    const diffMin = Math.floor(diffSec / 60);
    if (diffMin < 60) return `${diffMin}m ago`;
    const diffHr = Math.floor(diffMin / 60);
    if (diffHr < 24) return `${diffHr}h ago`;
    const diffDay = Math.floor(diffHr / 24);
    return `${diffDay}d ago`;
  }

  private shortId(id: string): string {
    const parts = id.split("_");
    if (parts.length >= 4) return parts.slice(2).join("_");
    return id;
  }
}
