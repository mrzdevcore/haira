import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { baseStyles, keyframes, animateInStyles } from "../core/styles";
import type { ConfirmProps } from "../core/types";

@customElement("haira-ui-confirm")
export class HairaConfirm extends LitElement {
  static styles = [
    baseStyles,
    animateInStyles,
    css`
      :host {
        width: fit-content;
        max-width: 100%;
        margin-left: calc(28px + 0.65rem);
      }

      .confirm-card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        overflow: hidden;
      }

      .confirm-body {
        padding: 0.75rem 0.85rem;
      }
      .confirm-title {
        font-size: 0.88rem;
        font-weight: 700;
        color: var(--haira-text);
        margin-bottom: 0.3rem;
      }
      .confirm-message {
        font-size: 0.84rem;
        line-height: 1.55;
        color: var(--haira-text-dim);
        white-space: pre-wrap;
        word-break: break-word;
      }

      .confirm-actions {
        display: flex;
        gap: 0.5rem;
        padding: 0.65rem 0.85rem;
        border-top: 1px solid var(--haira-border);
        justify-content: flex-end;
      }

      .btn {
        padding: 0.45rem 1rem;
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius-sm);
        font-family: var(--haira-font);
        font-size: 0.82rem;
        font-weight: 600;
        cursor: pointer;
        transition: all 0.15s;
      }
      .btn:disabled {
        opacity: 0.5;
        cursor: not-allowed;
      }

      .btn-confirm {
        background: var(--haira-accent);
        color: #fff;
        border-color: var(--haira-accent);
      }
      .btn-confirm:hover:not(:disabled) {
        opacity: 0.9;
        transform: translateY(-1px);
      }

      .btn-deny {
        background: transparent;
        color: var(--haira-text-dim);
      }
      .btn-deny:hover:not(:disabled) {
        background: var(--haira-bg-elevated);
        color: var(--haira-text);
      }

      .selection-made {
        padding: 0.55rem 0.85rem;
        border-top: 1px solid var(--haira-border);
        font-size: 0.78rem;
        color: var(--haira-muted);
        font-style: italic;
      }
    `,
  ];

  @property({ type: String }) title: string = "";
  @property({ type: String }) message: string = "";
  @property({ type: String }) confirm_label: string = "Confirm";
  @property({ type: String }) deny_label: string = "Cancel";
  @property({ type: Boolean }) _restored: boolean = false;

  @state() private _selected: string | null = null;

  /** Set all props at once */
  public setProps(props: ConfirmProps): void {
    this.title = props.title;
    this.message = props.message || "";
    this.confirm_label = props.confirm_label || "Confirm";
    this.deny_label = props.deny_label || "Cancel";
    if (props._restored) this._restored = true;
  }

  private _isDisabled(): boolean {
    return this._selected !== null || this._restored;
  }

  private _onConfirm() {
    if (this._isDisabled()) return;
    this._selected = "confirm";
    this.dispatchEvent(
      new CustomEvent("haira-chat-input", {
        bubbles: true,
        composed: true,
        detail: { text: `[Confirmed] ${this.title}` },
      })
    );
  }

  private _onDeny() {
    if (this._isDisabled()) return;
    this._selected = "deny";
    this.dispatchEvent(
      new CustomEvent("haira-chat-input", {
        bubbles: true,
        composed: true,
        detail: { text: `[Denied] ${this.title}` },
      })
    );
  }

  render() {
    const disabled = this._isDisabled();

    return html`
      <div class="confirm-card">
        <div class="confirm-body">
          <div class="confirm-title">${this.title}</div>
          ${this.message
            ? html`<div class="confirm-message">${this.message}</div>`
            : nothing}
        </div>

        ${this._selected || this._restored
          ? html`
              <div class="selection-made">
                ${this._selected === "confirm"
                  ? `Selected: ${this.confirm_label}`
                  : this._selected === "deny"
                    ? `Selected: ${this.deny_label}`
                    : "Interaction completed"}
              </div>
            `
          : html`
              <div class="confirm-actions">
                <button
                  class="btn btn-deny"
                  ?disabled=${disabled}
                  @click=${this._onDeny}
                >
                  ${this.deny_label}
                </button>
                <button
                  class="btn btn-confirm"
                  ?disabled=${disabled}
                  @click=${this._onConfirm}
                >
                  ${this.confirm_label}
                </button>
              </div>
            `}
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-ui-confirm": HairaConfirm;
  }
}
