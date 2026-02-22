import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles, keyframes, animateInStyles } from "../core/styles";
import { iconStrings } from "../core/icons";
import { formatToolName } from "../core/utils";

@customElement("haira-tool-card")
export class HairaToolCard extends LitElement {
  static styles = [
    baseStyles,
    css`
      :host {
        display: block;
        animation: fadeSlideUp 0.2s ease-out;
      }
      .tool-card {
        display: flex;
        align-items: center;
        gap: 0.55rem;
        padding: 0.45rem 0.75rem;
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius-sm);
        font-size: 0.8rem;
        color: var(--haira-text-dim);
        transition: border-color 0.2s;
      }
      .tool-card.running {
        border-color: var(--haira-border-light);
      }
      .tool-card.success {
        border-color: rgba(34, 197, 94, 0.2);
      }
      .tool-card.error {
        border-color: rgba(239, 68, 68, 0.2);
      }

      .icon {
        display: flex;
        align-items: center;
        flex-shrink: 0;
      }
      .icon.running {
        color: var(--haira-accent);
      }
      .icon.success {
        color: var(--haira-success);
      }
      .icon.error {
        color: var(--haira-error);
      }

      .name {
        font-weight: 600;
        color: var(--haira-text);
        flex: 1;
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      .status {
        font-size: 0.72rem;
        color: var(--haira-muted);
      }
      .duration {
        font-size: 0.72rem;
        font-family: var(--haira-mono);
        color: var(--haira-muted);
        flex-shrink: 0;
      }
    `,
  ];

  @state() private _name: string = "";
  @state() private _status: "running" | "success" | "error" = "running";
  @state() private _startTime: number = 0;
  @state() private _elapsed: string = "";
  private _timer: number | null = null;

  /** Set the tool name and begin timing */
  public setTool(name: string): void {
    this._name = name;
    this._status = "running";
    this._startTime = Date.now();
    this._elapsed = "";
    this._startTimer();
  }

  /** Mark the tool as complete */
  public complete(ok: boolean): void {
    this._status = ok ? "success" : "error";
    this._stopTimer();
    const elapsed = Date.now() - this._startTime;
    this._elapsed = this._formatDuration(elapsed);
  }

  private _startTimer(): void {
    this._stopTimer();
    this._timer = window.setInterval(() => {
      const ms = Date.now() - this._startTime;
      this._elapsed = this._formatDuration(ms);
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

  private _renderIcon() {
    if (this._status === "running") {
      return html`<span class="icon running"
        >${unsafeHTML(iconStrings.spinner)}</span
      >`;
    }
    if (this._status === "success") {
      return html`<span class="icon success"
        >${unsafeHTML(iconStrings.check)}</span
      >`;
    }
    return html`<span class="icon error"
      >${unsafeHTML(iconStrings.x)}</span
    >`;
  }

  private _statusLabel(): string {
    switch (this._status) {
      case "running":
        return "Running";
      case "success":
        return "Done";
      case "error":
        return "Failed";
    }
  }

  render() {
    if (!this._name) return nothing;

    return html`
      <div class="tool-card ${this._status}">
        ${this._renderIcon()}
        <span class="name">${formatToolName(this._name)}</span>
        <span class="status">${this._statusLabel()}</span>
        ${this._elapsed
          ? html`<span class="duration">${this._elapsed}</span>`
          : nothing}
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-tool-card": HairaToolCard;
  }
}
