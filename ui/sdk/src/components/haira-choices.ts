import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { baseStyles, keyframes, animateInStyles } from "../core/styles";
import type { ChoicesProps } from "../core/types";

@customElement("haira-ui-choices")
export class HairaChoices extends LitElement {
  static styles = [
    baseStyles,
    animateInStyles,
    css`
      :host {
        width: fit-content;
        max-width: 100%;
        margin-left: calc(28px + 0.65rem);
      }

      .choices-card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        overflow: hidden;
      }

      .choices-header {
        padding: 0.6rem 0.85rem;
        border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg);
      }
      .choices-title {
        font-size: 0.88rem;
        font-weight: 700;
        color: var(--haira-text);
      }

      .choices-body {
        padding: 0.65rem 0.85rem;
      }

      /* Buttons style */
      .choices-buttons {
        display: flex;
        flex-wrap: wrap;
        gap: 0.45rem;
      }
      .choice-btn {
        padding: 0.45rem 0.95rem;
        background: var(--haira-bg-elevated);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius-sm);
        color: var(--haira-text);
        font-family: var(--haira-font);
        font-size: 0.82rem;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.15s;
      }
      .choice-btn:hover:not(:disabled) {
        background: var(--haira-accent-dim);
        border-color: var(--haira-accent);
        color: var(--haira-accent);
      }
      .choice-btn:disabled {
        opacity: 0.5;
        cursor: not-allowed;
      }
      .choice-btn.selected {
        background: var(--haira-accent);
        border-color: var(--haira-accent);
        color: #fff;
      }

      /* List style */
      .choices-list {
        display: flex;
        flex-direction: column;
        gap: 0.3rem;
      }
      .choice-list-item {
        display: flex;
        align-items: center;
        gap: 0.55rem;
        padding: 0.5rem 0.7rem;
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius-sm);
        cursor: pointer;
        transition: all 0.15s;
        background: transparent;
        color: var(--haira-text);
        font-family: var(--haira-font);
        font-size: 0.84rem;
        font-weight: 500;
        width: 100%;
        text-align: left;
      }
      .choice-list-item:hover:not(:disabled) {
        background: var(--haira-bg-elevated);
        border-color: var(--haira-border-focus);
      }
      .choice-list-item:disabled {
        opacity: 0.5;
        cursor: not-allowed;
      }
      .choice-list-item.selected {
        background: var(--haira-accent-dim);
        border-color: var(--haira-accent);
        color: var(--haira-accent);
      }

      .choice-radio {
        width: 14px;
        height: 14px;
        border-radius: 50%;
        border: 2px solid var(--haira-border);
        flex-shrink: 0;
        position: relative;
        transition: all 0.15s;
      }
      .choice-list-item.selected .choice-radio {
        border-color: var(--haira-accent);
      }
      .choice-list-item.selected .choice-radio::after {
        content: "";
        position: absolute;
        top: 2px;
        left: 2px;
        width: 6px;
        height: 6px;
        border-radius: 50%;
        background: var(--haira-accent);
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
  @property({ type: Array }) options: string[] = [];
  @property({ type: String }) variant: ChoicesProps["style"] = "buttons";
  @property({ type: Boolean }) _restored: boolean = false;

  @state() private _selected: string | null = null;

  /** Set all props at once */
  public setProps(props: ChoicesProps): void {
    this.title = props.title;
    this.options = props.options || [];
    this.variant = props.style || "buttons";
    if (props._restored) this._restored = true;
  }

  private _isDisabled(): boolean {
    return this._selected !== null || this._restored;
  }

  private _onSelect(option: string) {
    if (this._isDisabled()) return;
    this._selected = option;
    this.dispatchEvent(
      new CustomEvent("haira-chat-input", {
        bubbles: true,
        composed: true,
        detail: { text: option },
      })
    );
  }

  private _renderButtons() {
    const disabled = this._isDisabled();
    return html`
      <div class="choices-buttons">
        ${this.options.map(
          (opt) => html`
            <button
              class="choice-btn ${this._selected === opt ? "selected" : ""}"
              ?disabled=${disabled}
              @click=${() => this._onSelect(opt)}
            >
              ${opt}
            </button>
          `
        )}
      </div>
    `;
  }

  private _renderList() {
    const disabled = this._isDisabled();
    return html`
      <div class="choices-list">
        ${this.options.map(
          (opt) => html`
            <button
              class="choice-list-item ${this._selected === opt
                ? "selected"
                : ""}"
              ?disabled=${disabled}
              @click=${() => this._onSelect(opt)}
            >
              <span class="choice-radio"></span>
              ${opt}
            </button>
          `
        )}
      </div>
    `;
  }

  render() {
    return html`
      <div class="choices-card">
        <div class="choices-header">
          <span class="choices-title">${this.title}</span>
        </div>
        <div class="choices-body">
          ${this.variant === "list" ? this._renderList() : this._renderButtons()}
        </div>
        ${this._selected
          ? html`
              <div class="selection-made">
                Selected: ${this._selected}
              </div>
            `
          : nothing}
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-ui-choices": HairaChoices;
  }
}
