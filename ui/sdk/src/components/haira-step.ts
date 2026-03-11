import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles, keyframes, animateInStyles } from "../core/styles";
import { iconStrings } from "../core/icons";
import type { StepStatus, StepLogEntry } from "../core/types";
import type { StepConfirmPayload } from "@haira/arp";

@customElement("haira-step")
export class HairaStep extends LitElement {
  static styles = [
    baseStyles,
    css`
      :host {
        display: block;
      }

      .step {
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius-sm);
        overflow: hidden;
        transition: border-color 0.2s;
      }
      .step.running {
        border-color: var(--haira-border-light);
      }
      .step.done {
        border-color: rgba(34, 197, 94, 0.15);
      }
      .step.failed {
        border-color: rgba(239, 68, 68, 0.15);
      }
      .step.retrying {
        border-color: rgba(234, 179, 8, 0.2);
      }
      .step.awaiting_confirm {
        border-color: rgba(59, 130, 246, 0.3);
      }

      .step-header {
        display: flex;
        align-items: center;
        gap: 0.55rem;
        padding: 0.55rem 0.75rem;
        background: var(--haira-bg-card);
        cursor: pointer;
        user-select: none;
        transition: background 0.12s;
      }
      .step-header:hover {
        background: var(--haira-bg-card-hover);
      }

      .step-icon {
        display: flex;
        align-items: center;
        flex-shrink: 0;
      }
      .step-icon.pending {
        color: var(--haira-muted);
      }
      .step-icon.running {
        color: var(--haira-accent);
      }
      .step-icon.done {
        color: var(--haira-success);
      }
      .step-icon.failed {
        color: var(--haira-error);
      }
      .step-icon.retrying {
        color: var(--haira-warn);
      }
      .step-icon.skipped {
        color: var(--haira-muted);
      }
      .step-icon.awaiting_confirm {
        color: var(--haira-info);
      }

      .step-name {
        flex: 1;
        font-size: 0.82rem;
        font-weight: 600;
        color: var(--haira-text);
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .step-meta {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        flex-shrink: 0;
      }
      .step-duration {
        font-size: 0.72rem;
        font-family: var(--haira-mono);
        color: var(--haira-muted);
      }
      .step-status-label {
        font-size: 0.7rem;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.03em;
      }
      .step-status-label.pending {
        color: var(--haira-muted);
      }
      .step-status-label.running {
        color: var(--haira-accent);
      }
      .step-status-label.done {
        color: var(--haira-success);
      }
      .step-status-label.failed {
        color: var(--haira-error);
      }
      .step-status-label.retrying {
        color: var(--haira-warn);
      }
      .step-status-label.skipped {
        color: var(--haira-muted);
      }
      .step-status-label.awaiting_confirm {
        color: var(--haira-info);
      }

      .step-expand {
        display: flex;
        align-items: center;
        color: var(--haira-muted);
        transition: transform 0.15s;
      }
      .step-expand.open {
        transform: rotate(90deg);
      }

      /* Error message */
      .step-error {
        padding: 0.45rem 0.75rem;
        font-size: 0.78rem;
        color: var(--haira-error);
        background: rgba(239, 68, 68, 0.06);
        border-top: 1px solid var(--haira-border);
      }

      /* Log panel */
      .logs {
        border-top: 1px solid var(--haira-border);
        background: var(--haira-bg);
        max-height: 240px;
        overflow-y: auto;
        animation: expandDown 0.2s ease-out;
      }
      .logs::-webkit-scrollbar {
        width: 4px;
      }
      .logs::-webkit-scrollbar-thumb {
        background: var(--haira-muted);
        border-radius: 2px;
      }

      .log-entry {
        display: flex;
        align-items: flex-start;
        gap: 0.5rem;
        padding: 0.35rem 0.75rem;
        border-bottom: 1px solid var(--haira-border);
        font-size: 0.76rem;
        line-height: 1.45;
      }
      .log-entry:last-child {
        border-bottom: none;
      }

      .log-badge {
        font-size: 0.62rem;
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.04em;
        padding: 0.1rem 0.35rem;
        border-radius: 3px;
        flex-shrink: 0;
        margin-top: 0.05rem;
      }
      .log-badge.info {
        color: var(--haira-info);
        background: rgba(59, 130, 246, 0.1);
      }
      .log-badge.warn {
        color: var(--haira-warn);
        background: rgba(234, 179, 8, 0.1);
      }
      .log-badge.error {
        color: var(--haira-error);
        background: rgba(239, 68, 68, 0.1);
      }
      .log-message {
        color: var(--haira-text-dim);
        word-break: break-word;
        flex: 1;
      }

      .log-empty {
        padding: 0.6rem 0.75rem;
        font-size: 0.76rem;
        color: var(--haira-muted);
        text-align: center;
      }

      /* Confirm panel */
      .confirm-panel {
        border-top: 1px solid var(--haira-border);
        padding: 0.65rem 0.75rem;
        background: rgba(59, 130, 246, 0.04);
        animation: expandDown 0.2s ease-out;
      }
      .confirm-title {
        font-size: 0.82rem;
        font-weight: 600;
        color: var(--haira-text);
        margin-bottom: 0.25rem;
      }
      .confirm-message {
        font-size: 0.78rem;
        color: var(--haira-text-dim);
        margin-bottom: 0.5rem;
      }
      .confirm-actions {
        display: flex;
        gap: 0.5rem;
        justify-content: flex-end;
      }
      .confirm-btn {
        padding: 0.4rem 0.9rem;
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius-sm);
        font-family: var(--haira-font);
        font-size: 0.78rem;
        font-weight: 600;
        cursor: pointer;
        transition: all 0.15s;
      }
      .confirm-btn:disabled {
        opacity: 0.5;
        cursor: not-allowed;
      }
      .confirm-btn.approve {
        background: var(--haira-accent);
        color: #fff;
        border-color: var(--haira-accent);
      }
      .confirm-btn.approve:hover:not(:disabled) {
        opacity: 0.9;
        transform: translateY(-1px);
      }
      .confirm-btn.deny {
        background: transparent;
        color: var(--haira-text-dim);
      }
      .confirm-btn.deny:hover:not(:disabled) {
        background: var(--haira-bg-elevated);
        color: var(--haira-text);
      }
      .confirm-resolved {
        font-size: 0.76rem;
        color: var(--haira-muted);
        font-style: italic;
        padding: 0.55rem 0.75rem;
        border-top: 1px solid var(--haira-border);
      }
    `,
  ];

  @property({ type: String }) name: string = "";

  @state() private _status: StepStatus = "pending";
  @state() private _duration: string = "";
  @state() private _error: string = "";
  @state() private _logs: StepLogEntry[] = [];
  @state() private _expanded: boolean = false;
  @state() private _liveElapsed: string = "";
  @state() private _confirmData: StepConfirmPayload | null = null;
  @state() private _confirmResolved: string | null = null;

  private _startTime: number = 0;
  private _timer: number | null = null;

  /** Set confirm data for awaiting_confirm status */
  public setConfirmData(data: StepConfirmPayload): void {
    this._confirmData = data;
  }

  /** Update step status */
  public setStatus(
    status: StepStatus,
    duration?: number,
    error?: string
  ): void {
    this._status = status;

    if (status === "running") {
      this._startTime = Date.now();
      this._startTimer();
      this._expanded = true; // auto-expand the active step
    } else if (status === "awaiting_confirm") {
      this._stopTimer();
      this._expanded = true;
    } else {
      this._stopTimer();
    }

    if (duration != null) {
      this._duration = this._formatDuration(duration);
      this._liveElapsed = "";
    }

    if (error) {
      this._error = error;
    }
  }

  /** Add a log entry */
  public addLog(entry: StepLogEntry): void {
    this._logs = [...this._logs, entry];
    this._expanded = true; // auto-expand to make logs visible
  }

  /** Clear all logs */
  public clearLogs(): void {
    this._logs = [];
  }

  private _startTimer(): void {
    this._stopTimer();
    this._timer = window.setInterval(() => {
      const ms = Date.now() - this._startTime;
      this._liveElapsed = this._formatDuration(ms);
    }, 100);
  }

  private _stopTimer(): void {
    if (this._timer !== null) {
      clearInterval(this._timer);
      this._timer = null;
    }
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this._stopTimer();
  }

  private _formatDuration(ms: number): string {
    if (ms < 1000) return `${Math.round(ms)}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  }

  private _toggleExpand(): void {
    this._expanded = !this._expanded;
  }

  private _renderIcon() {
    switch (this._status) {
      case "pending":
        return html`<span class="step-icon pending"
          >${unsafeHTML(iconStrings.stepPending)}</span
        >`;
      case "running":
        return html`<span class="step-icon running"
          >${unsafeHTML(iconStrings.stepActive)}</span
        >`;
      case "done":
        return html`<span class="step-icon done"
          >${unsafeHTML(iconStrings.stepDone)}</span
        >`;
      case "failed":
        return html`<span class="step-icon failed"
          >${unsafeHTML(iconStrings.stepFailed)}</span
        >`;
      case "retrying":
        return html`<span class="step-icon retrying"
          >${unsafeHTML(iconStrings.retry)}</span
        >`;
      case "skipped":
        return html`<span class="step-icon skipped"
          >${unsafeHTML(iconStrings.pending)}</span
        >`;
      case "awaiting_confirm":
        return html`<span class="step-icon awaiting_confirm"
          >${unsafeHTML(iconStrings.stepPending)}</span
        >`;
    }
  }

  private _onConfirm(): void {
    if (this._confirmResolved) return;
    this._confirmResolved = "confirmed";
    this.dispatchEvent(
      new CustomEvent("step-confirm", {
        bubbles: true,
        composed: true,
        detail: { confirmed: true },
      })
    );
  }

  private _onDeny(): void {
    if (this._confirmResolved) return;
    this._confirmResolved = "denied";
    this.dispatchEvent(
      new CustomEvent("step-confirm", {
        bubbles: true,
        composed: true,
        detail: { confirmed: false },
      })
    );
  }

  render() {
    const displayDuration = this._duration || this._liveElapsed;
    const hasLogs = this._logs.length > 0;

    return html`
      <div class="step ${this._status}">
        <div class="step-header" @click=${this._toggleExpand}>
          ${this._renderIcon()}
          <span class="step-name">${this.name}</span>
          <div class="step-meta">
            ${displayDuration
              ? html`<span class="step-duration">${displayDuration}</span>`
              : nothing}
            <span class="step-status-label ${this._status}">
              ${this._status}
            </span>
            ${hasLogs
              ? html`<span
                  class="step-expand ${this._expanded ? "open" : ""}"
                  >${unsafeHTML(iconStrings.chevron)}</span
                >`
              : nothing}
          </div>
        </div>
        ${this._error
          ? html`<div class="step-error">${this._error}</div>`
          : nothing}
        ${this._status === "awaiting_confirm" && this._confirmData && !this._confirmResolved
          ? html`
              <div class="confirm-panel">
                <div class="confirm-title">${this._confirmData.title}</div>
                ${this._confirmData.message
                  ? html`<div class="confirm-message">${this._confirmData.message}</div>`
                  : nothing}
                <div class="confirm-actions">
                  <button class="confirm-btn deny" @click=${this._onDeny}>
                    ${this._confirmData.deny_label || "Cancel"}
                  </button>
                  <button class="confirm-btn approve" @click=${this._onConfirm}>
                    ${this._confirmData.confirm_label || "Confirm"}
                  </button>
                </div>
              </div>
            `
          : nothing}
        ${this._confirmResolved
          ? html`<div class="confirm-resolved">
              ${this._confirmResolved === "confirmed" ? "Approved" : "Denied"}
            </div>`
          : nothing}
        ${this._expanded && hasLogs
          ? html`
              <div class="logs">
                ${this._logs.map(
                  (log) => html`
                    <div class="log-entry">
                      <span class="log-badge ${log.level}">${log.level}</span>
                      <span class="log-message">${log.message}</span>
                    </div>
                  `
                )}
              </div>
            `
          : nothing}
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-step": HairaStep;
  }
}
