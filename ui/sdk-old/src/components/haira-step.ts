import { baseCSS, scrollbarCSS, icons, esc } from "../core";
import type { StepStatus } from "../core/types";

export class HairaStep extends HTMLElement {
  private _status: StepStatus = "pending";
  private _duration: number | undefined;
  private _timerInterval: ReturnType<typeof setInterval> | null = null;
  private _timerStart = 0;
  private _expanded = false;
  private _logCount = 0;
  private _hasError = false;

  connectedCallback() {
    this.renderStep();
  }

  disconnectedCallback() {
    this.clearTimer();
  }

  private renderStep() {
    const name = this.getAttribute("name") || "";
    const idx = this.getAttribute("index") || "0";
    const shadow = this.attachShadow({ mode: "open" });
    shadow.innerHTML = `
      <style>
        ${baseCSS}
        :host { display: block; position: relative; }
        .step-header {
          display: flex; align-items: center; gap: 0.6rem;
          padding: 0.5rem 0.65rem; border-radius: var(--haira-radius-sm);
          cursor: pointer; user-select: none; transition: background 0.15s; position: relative;
        }
        .step-header:hover { background: rgba(255, 255, 255, 0.03); }
        .chevron {
          flex-shrink: 0; width: 16px; height: 16px; display: flex; align-items: center;
          justify-content: center; color: var(--haira-muted); transition: transform 0.2s ease, color 0.2s; opacity: 0;
        }
        .has-logs .chevron { opacity: 1; }
        .expanded .chevron { transform: rotate(90deg); }
        .status-icon {
          flex-shrink: 0; width: 22px; height: 22px; border-radius: 50%;
          display: flex; align-items: center; justify-content: center; transition: all 0.25s ease;
        }
        .pending .status-icon { border: 1.5px dashed var(--haira-muted); color: var(--haira-muted); }
        .running .status-icon { border: 1.5px solid var(--haira-accent); color: var(--haira-accent); background: rgba(232, 163, 23, 0.1); }
        .done .status-icon { background: var(--haira-success); color: #fff; }
        .failed .status-icon { background: var(--haira-error); color: #fff; }
        .retrying .status-icon {
          border: 1.5px solid var(--haira-accent); color: var(--haira-accent);
          background: rgba(232, 163, 23, 0.1); animation: pulse 1.5s ease-in-out infinite;
        }
        .skipped .status-icon { border: 1.5px dashed var(--haira-muted); color: var(--haira-muted); opacity: 0.5; }
        .step-num { font-size: 0.65rem; font-weight: 600; }
        .step-name {
          flex: 1; font-size: 0.85rem; font-weight: 500; color: var(--haira-muted);
          overflow: hidden; text-overflow: ellipsis; white-space: nowrap; transition: color 0.2s;
        }
        .running .step-name { color: var(--haira-text); font-weight: 600; }
        .done .step-name { color: var(--haira-text-dim); }
        .failed .step-name { color: var(--haira-text); }
        .retrying .step-name { color: var(--haira-text); }
        .log-count {
          font-size: 0.7rem; color: var(--haira-muted); padding: 0.1rem 0.4rem;
          border-radius: 10px; background: rgba(255, 255, 255, 0.04);
          font-family: var(--haira-mono); display: none;
        }
        .has-logs .log-count { display: inline-block; }
        .has-error .log-count { color: var(--haira-error); background: rgba(239, 68, 68, 0.1); }
        .timer {
          flex-shrink: 0; font-size: 0.75rem; font-family: var(--haira-mono);
          color: var(--haira-muted); min-width: 36px; text-align: right; transition: color 0.2s;
        }
        .running .timer { color: var(--haira-accent-light); }
        .done .timer { color: var(--haira-success); }
        .failed .timer { color: var(--haira-error); }
        .logs-wrapper {
          overflow: hidden; max-height: 0; opacity: 0;
          transition: max-height 0.25s ease, opacity 0.2s ease; margin-left: 2.55rem;
        }
        .logs-wrapper.open { max-height: 600px; opacity: 1; overflow-y: auto; ${scrollbarCSS} }
        .logs-inner { padding: 0.25rem 0 0.5rem 0; border-left: 1px solid rgba(63, 63, 70, 0.3); margin-left: 0.15rem; }
        .log-entry {
          display: flex; align-items: flex-start; gap: 0.5rem;
          font-size: 0.78rem; font-family: var(--haira-mono); line-height: 1.5;
          padding: 0.15rem 0 0.15rem 0.85rem; animation: fadeIn 0.15s ease-out both;
        }
        .log-badge {
          flex-shrink: 0; font-size: 0.6rem; font-weight: 700; text-transform: uppercase;
          padding: 0.08rem 0.35rem; border-radius: 3px; letter-spacing: 0.04em; margin-top: 0.12rem;
        }
        .log-badge.info { background: rgba(59, 130, 246, 0.12); color: var(--haira-info); }
        .log-badge.warn { background: rgba(234, 179, 8, 0.12); color: var(--haira-warn); }
        .log-badge.error { background: rgba(239, 68, 68, 0.12); color: var(--haira-error); }
        .log-msg { flex: 1; word-break: break-word; white-space: pre-wrap; color: var(--haira-text-dim); }
        .log-msg.warn { color: var(--haira-warn); }
        .log-msg.error { color: var(--haira-error); }
        .error-detail {
          margin: 0.25rem 0 0.5rem 0; padding: 0.5rem 0.75rem;
          font-size: 0.78rem; font-family: var(--haira-mono); color: var(--haira-error);
          background: rgba(239, 68, 68, 0.06); border: 1px solid rgba(239, 68, 68, 0.12);
          border-radius: var(--haira-radius-sm); margin-left: 2.55rem; line-height: 1.5;
          word-break: break-word; white-space: pre-wrap; display: none;
        }
        .error-detail.visible { display: block; animation: fadeIn 0.2s ease-out; }
      </style>
      <div class="step-header pending" id="header">
        <span class="chevron" id="chevron">${icons.chevron}</span>
        <span class="status-icon" id="status-icon">
          <span class="step-num" id="step-num">${Number(idx) + 1}</span>
        </span>
        <span class="step-name">${esc(name)}</span>
        <span class="log-count" id="log-count"></span>
        <span class="timer" id="timer"></span>
      </div>
      <div class="logs-wrapper" id="logs-wrapper">
        <div class="logs-inner" id="logs"></div>
      </div>
      <div class="error-detail" id="error-detail"></div>
    `;

    shadow.getElementById("header")!.addEventListener("click", () => {
      if (this._logCount === 0) return;
      this.toggleLogs();
    });
  }

  private toggleLogs(forceOpen?: boolean) {
    const wrapper = this.shadowRoot?.getElementById("logs-wrapper");
    const header = this.shadowRoot?.getElementById("header");
    if (!wrapper || !header) return;

    if (forceOpen === true) {
      this._expanded = true;
    } else if (forceOpen === false) {
      this._expanded = false;
    } else {
      this._expanded = !this._expanded;
    }

    if (this._expanded) {
      wrapper.classList.add("open");
      header.classList.add("expanded");
    } else {
      wrapper.classList.remove("open");
      header.classList.remove("expanded");
    }
  }

  setStatus(status: StepStatus, durationMs?: number, error?: string) {
    this._status = status;
    this._duration = durationMs;
    const header = this.shadowRoot?.getElementById("header");
    const icon = this.shadowRoot?.getElementById("status-icon");
    const timer = this.shadowRoot?.getElementById("timer");
    const errorDetail = this.shadowRoot?.getElementById("error-detail");
    if (!header || !icon || !timer) return;

    const extraClasses: string[] = [];
    if (this._logCount > 0) extraClasses.push("has-logs");
    if (this._hasError) extraClasses.push("has-error");
    if (this._expanded) extraClasses.push("expanded");
    header.className = `step-header ${status} ${extraClasses.join(" ")}`;

    switch (status) {
      case "pending":
        icon.innerHTML = `<span class="step-num">${this.getIndex()}</span>`;
        this.clearTimer();
        timer.textContent = "";
        errorDetail?.classList.remove("visible");
        break;
      case "running":
        icon.innerHTML = icons.spinner;
        this.startTimer(timer);
        errorDetail?.classList.remove("visible");
        break;
      case "done":
        icon.innerHTML = icons.check;
        this.clearTimer();
        timer.textContent = this.formatDuration(durationMs);
        errorDetail?.classList.remove("visible");
        if (!this._hasError) this.toggleLogs(false);
        break;
      case "failed":
        icon.innerHTML = icons.x;
        this.clearTimer();
        timer.textContent = this.formatDuration(durationMs);
        if (error && errorDetail) {
          errorDetail.textContent = error;
          errorDetail.classList.add("visible");
        }
        if (this._logCount > 0) this.toggleLogs(true);
        break;
      case "retrying":
        icon.innerHTML = icons.retry;
        errorDetail?.classList.remove("visible");
        break;
      case "skipped":
        icon.innerHTML = `<span class="step-num">${this.getIndex()}</span>`;
        this.clearTimer();
        timer.textContent = "skipped";
        errorDetail?.classList.remove("visible");
        break;
    }
  }

  addLog(level: "info" | "warn" | "error", message: string) {
    const logsContainer = this.shadowRoot?.getElementById("logs");
    const header = this.shadowRoot?.getElementById("header");
    const countEl = this.shadowRoot?.getElementById("log-count");
    if (!logsContainer || !header || !countEl) return;

    this._logCount++;
    header.classList.add("has-logs");
    countEl.textContent = String(this._logCount);

    if (level === "error") {
      this._hasError = true;
      header.classList.add("has-error");
      this.toggleLogs(true);
    }

    const maxLen = 200;
    const display = message.length > maxLen ? message.slice(0, maxLen) + "..." : message;

    const entry = document.createElement("div");
    entry.className = "log-entry";
    entry.innerHTML = `<span class="log-badge ${level}">${level}</span><span class="log-msg ${level}">${esc(display)}</span>`;
    logsContainer.appendChild(entry);

    if (this._status === "running" && !this._expanded) {
      this.toggleLogs(true);
    }

    const wrapper = this.shadowRoot?.getElementById("logs-wrapper");
    if (wrapper && this._expanded) {
      wrapper.scrollTop = wrapper.scrollHeight;
    }
  }

  clearLogs() {
    const logsContainer = this.shadowRoot?.getElementById("logs");
    const header = this.shadowRoot?.getElementById("header");
    const errorDetail = this.shadowRoot?.getElementById("error-detail");
    if (!logsContainer) return;
    logsContainer.innerHTML = "";
    this._logCount = 0;
    this._hasError = false;
    this._expanded = false;
    header?.classList.remove("has-logs", "has-error", "expanded");
    errorDetail?.classList.remove("visible");
    this.toggleLogs(false);

    const countEl = this.shadowRoot?.getElementById("log-count");
    if (countEl) countEl.textContent = "";
  }

  private getIndex(): string {
    const idx = this.getAttribute("index");
    return idx !== null ? String(Number(idx) + 1) : "";
  }

  private startTimer(el: HTMLElement) {
    this.clearTimer();
    this._timerStart = performance.now();
    el.textContent = "0.0s";
    this._timerInterval = setInterval(() => {
      const elapsed = (performance.now() - this._timerStart) / 1000;
      el.textContent = `${elapsed.toFixed(1)}s`;
    }, 100);
  }

  private clearTimer() {
    if (this._timerInterval !== null) {
      clearInterval(this._timerInterval);
      this._timerInterval = null;
    }
  }

  private formatDuration(ms?: number): string {
    if (ms === undefined) return "";
    return ms < 1000 ? `${ms}ms` : `${(ms / 1000).toFixed(1)}s`;
  }

  get status(): StepStatus {
    return this._status;
  }
  get duration(): number | undefined {
    return this._duration;
  }
}
