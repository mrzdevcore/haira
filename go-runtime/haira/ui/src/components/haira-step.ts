import {
  baseStyles,
  sharedKeyframes,
  iconPending,
  iconSpinner,
  iconCheck,
  iconX,
  iconRetry,
} from "../theme";
import type { StepStatus } from "../types";

export class HairaStep extends HTMLElement {
  private _status: StepStatus = "pending";
  private _duration: number | undefined;
  private _timerInterval: ReturnType<typeof setInterval> | null = null;
  private _timerStart = 0;

  connectedCallback() {
    this.render();
  }

  disconnectedCallback() {
    this.clearTimer();
  }

  private render() {
    const name = this.getAttribute("name") || "";
    const idx = this.getAttribute("index") || "0";
    const shadow = this.attachShadow({ mode: "open" });
    shadow.innerHTML = `
      <style>
        ${baseStyles}
        ${sharedKeyframes}
        :host {
          display: block;
          animation: fadeSlideUp 0.3s ease-out both;
        }
        .row {
          display: flex;
          align-items: center;
          gap: 0.75rem;
          padding: 0.6rem 0.85rem;
          border-radius: var(--haira-radius);
          border: 1px solid transparent;
          transition: all 0.3s ease;
        }
        .row.pending {
          opacity: 0.5;
        }
        .row.running {
          background: var(--haira-gold-dim);
          border-color: var(--haira-gold);
          border-left: 3px solid var(--haira-gold);
          animation: pulse 2s ease-in-out infinite;
        }
        .row.done {
          background: rgba(74, 222, 128, 0.04);
          border-color: rgba(74, 222, 128, 0.15);
          border-left: 3px solid var(--haira-success);
          animation: pop 0.3s ease-out;
        }
        .row.failed {
          background: rgba(248, 113, 113, 0.04);
          border-color: rgba(248, 113, 113, 0.15);
          border-left: 3px solid var(--haira-error);
          animation: pop 0.3s ease-out;
        }
        .row.retrying {
          background: var(--haira-gold-dim);
          border-color: var(--haira-gold);
          border-left: 3px solid var(--haira-gold);
          animation: blink 0.8s ease-in-out infinite;
        }
        .row.skipped {
          opacity: 0.4;
          border-left: 3px solid var(--haira-muted);
        }
        .icon-wrap {
          flex-shrink: 0;
          width: 28px;
          height: 28px;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          background: var(--haira-bg);
          border: 1.5px solid var(--haira-border);
          transition: all 0.3s ease;
        }
        .row.pending .icon-wrap {
          border-style: dashed;
          border-color: var(--haira-muted);
          color: var(--haira-muted);
        }
        .row.running .icon-wrap {
          border-color: var(--haira-gold);
          color: var(--haira-gold);
          background: rgba(232, 163, 23, 0.1);
        }
        .row.done .icon-wrap {
          border-color: var(--haira-success);
          color: #fff;
          background: var(--haira-success);
        }
        .row.failed .icon-wrap {
          border-color: var(--haira-error);
          color: #fff;
          background: var(--haira-error);
        }
        .row.retrying .icon-wrap {
          border-color: var(--haira-gold);
          color: var(--haira-gold);
          background: rgba(232, 163, 23, 0.1);
        }
        .row.skipped .icon-wrap {
          border-style: dashed;
          border-color: var(--haira-muted);
          color: var(--haira-muted);
        }
        .step-num {
          font-size: 0.7rem;
          font-weight: 700;
        }
        .info {
          flex: 1;
          min-width: 0;
        }
        .name {
          font-size: 0.88rem;
          font-weight: 500;
          color: var(--haira-text-dim);
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
        .row.pending .name { color: var(--haira-muted); }
        .row.running .name,
        .row.done .name,
        .row.failed .name { color: var(--haira-text); }
        .timer {
          flex-shrink: 0;
          font-size: 0.78rem;
          font-family: var(--haira-mono);
          color: var(--haira-muted);
          min-width: 40px;
          text-align: right;
        }
        .row.running .timer { color: var(--haira-gold-light); }
        .row.done .timer { color: var(--haira-success); }
        .row.failed .timer { color: var(--haira-error); }

        /* Logs area */
        .logs {
          display: none;
          flex-direction: column;
          gap: 2px;
          padding: 0.35rem 0.85rem 0.35rem 3.1rem;
        }
        .logs.visible {
          display: flex;
          animation: fadeIn 0.2s ease-out;
        }
        .log-entry {
          display: flex;
          align-items: flex-start;
          gap: 0.5rem;
          font-size: 0.78rem;
          font-family: var(--haira-mono);
          line-height: 1.4;
          padding: 0.2rem 0;
          animation: fadeSlideUp 0.15s ease-out both;
        }
        .log-badge {
          flex-shrink: 0;
          font-size: 0.65rem;
          font-weight: 700;
          text-transform: uppercase;
          padding: 0.05rem 0.35rem;
          border-radius: 3px;
          letter-spacing: 0.03em;
          margin-top: 0.1rem;
        }
        .log-badge.info {
          background: rgba(96, 165, 250, 0.15);
          color: #60a5fa;
        }
        .log-badge.warn {
          background: rgba(251, 191, 36, 0.15);
          color: #fbbf24;
        }
        .log-badge.error {
          background: rgba(248, 113, 113, 0.15);
          color: #f87171;
        }
        .log-msg {
          flex: 1;
          word-break: break-word;
          white-space: pre-wrap;
        }
        .log-msg.info { color: var(--haira-text-dim); }
        .log-msg.warn { color: #fbbf24; }
        .log-msg.error { color: #f87171; }

        /* Error detail shown on failed status */
        .error-detail {
          display: none;
          padding: 0.4rem 0.85rem 0.3rem 3.1rem;
          animation: fadeSlideUp 0.2s ease-out both;
        }
        .error-detail.visible {
          display: block;
        }
        .error-text {
          font-size: 0.78rem;
          font-family: var(--haira-mono);
          color: #f87171;
          background: rgba(248, 113, 113, 0.08);
          border: 1px solid rgba(248, 113, 113, 0.15);
          border-radius: 4px;
          padding: 0.4rem 0.6rem;
          line-height: 1.4;
          word-break: break-word;
          white-space: pre-wrap;
        }

        @keyframes spin {
          to { transform: rotate(360deg); }
        }
      </style>
      <div class="row pending" id="row">
        <div class="icon-wrap" id="icon-wrap">
          <span id="icon"><span class="step-num">${Number(idx) + 1}</span></span>
        </div>
        <div class="info">
          <div class="name">${this.esc(name)}</div>
        </div>
        <span class="timer" id="timer"></span>
      </div>
      <div class="logs" id="logs"></div>
      <div class="error-detail" id="error-detail">
        <div class="error-text" id="error-text"></div>
      </div>
    `;
  }

  setStatus(status: StepStatus, durationMs?: number, error?: string) {
    this._status = status;
    this._duration = durationMs;
    const row = this.shadowRoot?.getElementById("row");
    const icon = this.shadowRoot?.getElementById("icon");
    const timer = this.shadowRoot?.getElementById("timer");
    const errorDetail = this.shadowRoot?.getElementById("error-detail");
    const errorText = this.shadowRoot?.getElementById("error-text");
    if (!row || !icon || !timer) return;

    row.className = `row ${status}`;

    switch (status) {
      case "pending":
        icon.innerHTML = `<span class="step-num">${this.getIndex()}</span>`;
        this.clearTimer();
        timer.textContent = "";
        errorDetail?.classList.remove("visible");
        break;
      case "running":
        icon.innerHTML = iconSpinner;
        this.startTimer(timer);
        errorDetail?.classList.remove("visible");
        break;
      case "done":
        icon.innerHTML = iconCheck;
        this.clearTimer();
        timer.textContent = this.formatDuration(durationMs);
        errorDetail?.classList.remove("visible");
        break;
      case "failed":
        icon.innerHTML = iconX;
        this.clearTimer();
        timer.textContent = this.formatDuration(durationMs);
        if (error && errorDetail && errorText) {
          errorText.textContent = error;
          errorDetail.classList.add("visible");
        }
        break;
      case "retrying":
        icon.innerHTML = iconRetry;
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
    if (!logsContainer) return;

    logsContainer.classList.add("visible");

    const entry = document.createElement("div");
    entry.className = "log-entry";
    entry.innerHTML = `<span class="log-badge ${level}">${level}</span><span class="log-msg ${level}">${this.esc(message)}</span>`;
    logsContainer.appendChild(entry);
  }

  clearLogs() {
    const logsContainer = this.shadowRoot?.getElementById("logs");
    if (!logsContainer) return;
    logsContainer.innerHTML = "";
    logsContainer.classList.remove("visible");

    const errorDetail = this.shadowRoot?.getElementById("error-detail");
    errorDetail?.classList.remove("visible");
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

  private esc(s: string): string {
    return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }
}
