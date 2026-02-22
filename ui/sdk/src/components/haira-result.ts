import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles, keyframes, animateInStyles } from "../core/styles";
import { iconStrings } from "../core/icons";

@customElement("haira-result")
export class HairaResult extends LitElement {
  static styles = [
    baseStyles,
    css`
      :host {
        display: block;
        animation: fadeSlideUp 0.25s ease-out;
      }
      :host([hidden]) {
        display: none;
      }

      .result {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        overflow: hidden;
      }

      .result-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.6rem 0.85rem;
        border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg);
      }
      .result-header-left {
        display: flex;
        align-items: center;
        gap: 0.45rem;
      }
      .status-dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        flex-shrink: 0;
      }
      .status-dot.success {
        background: var(--haira-success);
      }
      .status-dot.error {
        background: var(--haira-error);
      }
      .result-title {
        font-size: 0.82rem;
        font-weight: 600;
        color: var(--haira-text);
      }

      .copy-btn {
        display: flex;
        align-items: center;
        gap: 0.25rem;
        background: none;
        border: none;
        color: var(--haira-muted);
        cursor: pointer;
        font-size: 0.7rem;
        font-family: var(--haira-font);
        padding: 0.2rem 0.4rem;
        border-radius: 4px;
        transition: all 0.15s;
      }
      .copy-btn:hover {
        background: var(--haira-bg-elevated);
        color: var(--haira-text);
      }
      .copy-btn.copied {
        color: var(--haira-success);
      }

      .result-body {
        padding: 0.75rem 0.85rem;
      }

      /* Rich message mode */
      .rich-message {
        font-size: 0.88rem;
        line-height: 1.6;
        color: var(--haira-text);
        white-space: pre-wrap;
        word-break: break-word;
      }

      /* Key-value mode */
      .kv-grid {
        display: grid;
        grid-template-columns: auto 1fr;
        gap: 0.3rem 0.85rem;
      }
      .kv-key {
        font-size: 0.78rem;
        font-weight: 600;
        color: var(--haira-text-dim);
        white-space: nowrap;
      }
      .kv-value {
        font-size: 0.82rem;
        color: var(--haira-text);
        word-break: break-word;
      }

      /* Raw JSON mode */
      .raw-json {
        font-family: var(--haira-mono);
        font-size: 0.8rem;
        line-height: 1.5;
        color: var(--haira-text);
        white-space: pre-wrap;
        word-break: break-word;
        max-height: 400px;
        overflow-y: auto;
      }
      .raw-json::-webkit-scrollbar {
        width: 5px;
      }
      .raw-json::-webkit-scrollbar-thumb {
        background: var(--haira-muted);
        border-radius: 3px;
      }
    `,
  ];

  @state() private _data: unknown = null;
  @state() private _isError: boolean = false;
  @state() private _visible: boolean = false;
  @state() private _copied: boolean = false;

  /** Show the result panel */
  public show(data: unknown, isError: boolean = false): void {
    this._data = data;
    this._isError = isError;
    this._visible = true;
    this.removeAttribute("hidden");
  }

  /** Hide the result panel */
  public hide(): void {
    this._visible = false;
    this._data = null;
    this.setAttribute("hidden", "");
  }

  private _isRichMessage(data: unknown): data is { message: string } {
    return (
      typeof data === "object" &&
      data !== null &&
      "message" in data &&
      typeof (data as any).message === "string"
    );
  }

  private _isKeyValueObject(data: unknown): data is Record<string, unknown> {
    if (typeof data !== "object" || data === null || Array.isArray(data))
      return false;
    const keys = Object.keys(data);
    return keys.length > 0 && keys.length <= 10;
  }

  private async _copyToClipboard() {
    const text =
      typeof this._data === "string"
        ? this._data
        : JSON.stringify(this._data, null, 2);
    try {
      await navigator.clipboard.writeText(text);
      this._copied = true;
      setTimeout(() => {
        this._copied = false;
      }, 1500);
    } catch {
      // Clipboard write failed silently
    }
  }

  render() {
    if (!this._visible || this._data == null) return nothing;

    return html`
      <div class="result">
        <div class="result-header">
          <div class="result-header-left">
            <span class="status-dot ${this._isError ? "error" : "success"}">
            </span>
            <span class="result-title">
              ${this._isError ? "Error" : "Result"}
            </span>
          </div>
          <button
            class="copy-btn ${this._copied ? "copied" : ""}"
            @click=${this._copyToClipboard}
          >
            ${this._copied
              ? html`${unsafeHTML(iconStrings.copyDone)} Copied`
              : html`${unsafeHTML(iconStrings.copy)} Copy`}
          </button>
        </div>
        <div class="result-body">${this._renderContent()}</div>
      </div>
    `;
  }

  private _renderContent() {
    const data = this._data;

    // Mode 1: Rich message (has message field)
    if (this._isRichMessage(data)) {
      return html`<div class="rich-message">${data.message}</div>`;
    }

    // Mode 2: Key-value (flat object with <= 10 keys)
    if (this._isKeyValueObject(data)) {
      const entries = Object.entries(data);
      // Check if all values are primitives (flat object)
      const isFlat = entries.every(
        ([, v]) => typeof v !== "object" || v === null
      );
      if (isFlat) {
        return html`
          <div class="kv-grid">
            ${entries.map(
              ([key, value]) => html`
                <span class="kv-key">${key}</span>
                <span class="kv-value">${String(value ?? "\u2014")}</span>
              `
            )}
          </div>
        `;
      }
    }

    // Mode 3: Raw JSON
    const jsonStr =
      typeof data === "string"
        ? data
        : JSON.stringify(data, null, 2);
    return html`<div class="raw-json">${jsonStr}</div>`;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-result": HairaResult;
  }
}
