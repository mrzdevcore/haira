import { baseStyles, sharedKeyframes } from "../theme";

const iconSpinner = `<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-dasharray="28 10" style="animation:spin 0.7s linear infinite;transform-origin:center"/></svg>`;
const iconCheck = `<svg width="13" height="13" viewBox="0 0 16 16" fill="none"><path d="M4.5 8.5L7 11L11.5 5.5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>`;
const iconX = `<svg width="13" height="13" viewBox="0 0 16 16" fill="none"><path d="M5 5L11 11M11 5L5 11" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>`;
const iconTool = `<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.7 6.3a1 1 0 000 1.4l1.6 1.6a1 1 0 001.4 0l3.77-3.77a6 6 0 01-7.94 7.94l-6.91 6.91a2.12 2.12 0 01-3-3l6.91-6.91a6 6 0 017.94-7.94l-3.76 3.76z"/></svg>`;

// Human-readable tool names
function formatToolName(name: string): string {
  return name.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

export class HairaToolCard extends HTMLElement {
  private startTime = 0;

  connectedCallback() {
    this.startTime = Date.now();
    const shadow = this.attachShadow({ mode: "open" });
    shadow.innerHTML = `
      <style>
        ${baseStyles}
        ${sharedKeyframes}
        :host {
          display: block;
          animation: fadeSlideUp 0.2s ease-out;
        }
        .card {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          padding: 0.5rem 0.75rem;
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: 8px;
        }
        .icon {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 24px;
          height: 24px;
          border-radius: 6px;
          flex-shrink: 0;
        }
        .icon.running {
          background: rgba(232, 163, 23, 0.1);
          color: var(--haira-accent);
        }
        .icon.done {
          background: rgba(34, 197, 94, 0.1);
          color: var(--haira-success);
        }
        .icon.failed {
          background: rgba(239, 68, 68, 0.1);
          color: var(--haira-error);
        }
        .info {
          flex: 1;
          min-width: 0;
        }
        .tool-name {
          font-size: 0.78rem;
          font-weight: 600;
          color: var(--haira-text);
        }
        .tool-status {
          font-size: 0.7rem;
          color: var(--haira-muted);
          display: flex;
          align-items: center;
          gap: 0.3rem;
        }
        .duration {
          font-family: var(--haira-mono);
          font-size: 0.68rem;
          color: var(--haira-muted);
          flex-shrink: 0;
        }
        @keyframes spin {
          to { transform: rotate(360deg); }
        }
      </style>
      <div class="card" id="card">
        <div class="icon running" id="icon">${iconSpinner}</div>
        <div class="info">
          <div class="tool-name" id="name"></div>
          <div class="tool-status" id="status">Running...</div>
        </div>
        <span class="duration" id="duration"></span>
      </div>
    `;
  }

  setTool(name: string) {
    const nameEl = this.shadowRoot?.getElementById("name");
    if (nameEl) nameEl.textContent = formatToolName(name);
  }

  complete(ok: boolean) {
    const icon = this.shadowRoot?.getElementById("icon");
    const status = this.shadowRoot?.getElementById("status");
    const duration = this.shadowRoot?.getElementById("duration");

    const elapsed = Date.now() - this.startTime;
    const seconds = (elapsed / 1000).toFixed(1);

    if (icon) {
      icon.className = `icon ${ok ? "done" : "failed"}`;
      icon.innerHTML = ok ? iconCheck : iconX;
    }
    if (status) {
      status.textContent = ok ? "Completed" : "Failed";
    }
    if (duration) {
      duration.textContent = `${seconds}s`;
    }
  }
}
