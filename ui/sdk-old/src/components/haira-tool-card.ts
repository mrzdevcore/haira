import { BaseComponent, animateInCSS, icons, formatToolName } from "../core";

export class HairaToolCard extends BaseComponent {
  private startTime = 0;

  protected render() {
    return `
      <div class="card" id="card">
        <div class="icon running" id="icon">${icons.spinner}</div>
        <div class="info">
          <div class="tool-name" id="name"></div>
          <div class="tool-status" id="status">Running...</div>
        </div>
        <span class="duration" id="duration"></span>
      </div>`;
  }

  protected styles() {
    return `
      ${animateInCSS}
      .card {
        display: flex; align-items: center; gap: 0.5rem;
        padding: 0.5rem 0.75rem;
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: 8px;
      }
      .icon {
        display: flex; align-items: center; justify-content: center;
        width: 24px; height: 24px; border-radius: 6px; flex-shrink: 0;
      }
      .icon.running { background: rgba(232, 163, 23, 0.1); color: var(--haira-accent); }
      .icon.done { background: rgba(34, 197, 94, 0.1); color: var(--haira-success); }
      .icon.failed { background: rgba(239, 68, 68, 0.1); color: var(--haira-error); }
      .info { flex: 1; min-width: 0; }
      .tool-name { font-size: 0.78rem; font-weight: 600; color: var(--haira-text); }
      .tool-status { font-size: 0.7rem; color: var(--haira-muted); display: flex; align-items: center; gap: 0.3rem; }
      .duration { font-family: var(--haira-mono); font-size: 0.68rem; color: var(--haira-muted); flex-shrink: 0; }
      @keyframes spin { to { transform: rotate(360deg); } }`;
  }

  protected onMount() {
    this.startTime = Date.now();
  }

  setTool(name: string) {
    const nameEl = this.root?.getElementById("name");
    if (nameEl) nameEl.textContent = formatToolName(name);
  }

  complete(ok: boolean) {
    const iconEl = this.root?.getElementById("icon");
    const status = this.root?.getElementById("status");
    const duration = this.root?.getElementById("duration");
    const elapsed = Date.now() - this.startTime;
    const seconds = (elapsed / 1000).toFixed(1);

    if (iconEl) {
      iconEl.className = `icon ${ok ? "done" : "failed"}`;
      iconEl.innerHTML = ok ? icons.check : icons.x;
    }
    if (status) status.textContent = ok ? "Completed" : "Failed";
    if (duration) duration.textContent = `${seconds}s`;
  }
}
