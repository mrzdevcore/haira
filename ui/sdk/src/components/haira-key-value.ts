import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { baseStyles, keyframes, animateInStyles } from "../core/styles";
import type { KeyValueProps } from "../core/types";

@customElement("haira-ui-key-value")
export class HairaKeyValue extends LitElement {
  static styles = [
    baseStyles,
    animateInStyles,
    css`
      .kv-card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        overflow: hidden;
      }

      .kv-header {
        padding: 0.55rem 0.85rem;
        border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg);
      }
      .kv-title {
        font-size: 0.85rem;
        font-weight: 700;
        color: var(--haira-text);
      }

      .kv-list {
        padding: 0.45rem 0;
      }
      .kv-row {
        display: flex;
        align-items: flex-start;
        gap: 0.75rem;
        padding: 0.4rem 0.85rem;
        transition: background 0.12s;
      }
      .kv-row:hover {
        background: var(--haira-bg-card-hover);
      }

      .kv-key {
        font-size: 0.78rem;
        font-weight: 600;
        color: var(--haira-text-dim);
        min-width: 100px;
        flex-shrink: 0;
        padding-top: 0.05rem;
      }
      .kv-value {
        font-size: 0.84rem;
        color: var(--haira-text);
        word-break: break-word;
        flex: 1;
      }

      /* Style coloring */
      .kv-value.success {
        color: var(--haira-success);
      }
      .kv-value.error {
        color: var(--haira-error);
      }
      .kv-value.warning {
        color: var(--haira-warn);
      }
      .kv-value.info {
        color: var(--haira-info);
      }
      .kv-value.code {
        font-family: var(--haira-mono);
        font-size: 0.8rem;
        background: var(--haira-bg-input);
        padding: 0.2rem 0.45rem;
        border-radius: 4px;
      }
    `,
  ];

  @property({ type: String }) title: string = "";
  @property({ type: Array }) items: KeyValueProps["items"] = [];

  /** Set all props at once */
  public setProps(props: KeyValueProps): void {
    this.title = props.title || "";
    this.items = props.items || [];
  }

  render() {
    return html`
      <div class="kv-card">
        ${this.title
          ? html`
              <div class="kv-header">
                <span class="kv-title">${this.title}</span>
              </div>
            `
          : nothing}
        <div class="kv-list">
          ${this.items.map(
            (item) => html`
              <div class="kv-row">
                <span class="kv-key">${item.key}</span>
                <span class="kv-value ${item.style || ""}">${item.value}</span>
              </div>
            `
          )}
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-ui-key-value": HairaKeyValue;
  }
}
