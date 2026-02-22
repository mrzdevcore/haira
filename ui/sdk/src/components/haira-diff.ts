import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { baseStyles, keyframes, animateInStyles } from "../core/styles";
import type { DiffProps } from "../core/types";

@customElement("haira-ui-diff")
export class HairaDiff extends LitElement {
  static styles = [
    baseStyles,
    animateInStyles,
    css`
      .diff-card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        overflow: hidden;
      }

      .diff-header {
        padding: 0.55rem 0.85rem;
        border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg);
      }
      .diff-title {
        font-size: 0.85rem;
        font-weight: 700;
        color: var(--haira-text);
      }

      .diff-panels {
        display: grid;
        grid-template-columns: 1fr 1fr;
        min-height: 0;
      }

      .diff-panel {
        overflow: auto;
        max-height: 420px;
      }
      .diff-panel::-webkit-scrollbar {
        width: 4px;
        height: 4px;
      }
      .diff-panel::-webkit-scrollbar-thumb {
        background: var(--haira-muted);
        border-radius: 2px;
      }

      .diff-panel:first-child {
        border-right: 1px solid var(--haira-border);
      }

      .panel-label {
        padding: 0.35rem 0.65rem;
        font-size: 0.7rem;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.04em;
        border-bottom: 1px solid var(--haira-border);
        position: sticky;
        top: 0;
        z-index: 1;
      }
      .panel-label.before {
        color: var(--haira-error);
        background: rgba(239, 68, 68, 0.06);
      }
      .panel-label.after {
        color: var(--haira-success);
        background: rgba(34, 197, 94, 0.06);
      }

      .diff-content {
        padding: 0.55rem 0.75rem;
        font-family: var(--haira-mono);
        font-size: 0.8rem;
        line-height: 1.55;
        white-space: pre-wrap;
        word-break: break-word;
      }
      .diff-content.before {
        background: rgba(239, 68, 68, 0.03);
        color: var(--haira-text);
      }
      .diff-content.after {
        background: rgba(34, 197, 94, 0.03);
        color: var(--haira-text);
      }

      @media (max-width: 600px) {
        .diff-panels {
          grid-template-columns: 1fr;
        }
        .diff-panel:first-child {
          border-right: none;
          border-bottom: 1px solid var(--haira-border);
        }
      }
    `,
  ];

  @property({ type: String }) title: string = "";
  @property({ type: String }) beforeText: string = "";
  @property({ type: String }) afterText: string = "";
  @property({ type: String }) before_label: string = "Before";
  @property({ type: String }) after_label: string = "After";
  @property({ type: String }) language: string = "";

  /** Set all props at once */
  public setProps(props: DiffProps): void {
    this.title = props.title || "";
    this.beforeText = props.before || "";
    this.afterText = props.after || "";
    this.before_label = props.before_label || "Before";
    this.after_label = props.after_label || "After";
    this.language = props.language || "";
  }

  render() {
    return html`
      <div class="diff-card">
        ${this.title
          ? html`
              <div class="diff-header">
                <span class="diff-title">${this.title}</span>
              </div>
            `
          : nothing}

        <div class="diff-panels">
          <div class="diff-panel">
            <div class="panel-label before">${this.before_label}</div>
            <div class="diff-content before">${this.beforeText}</div>
          </div>
          <div class="diff-panel">
            <div class="panel-label after">${this.after_label}</div>
            <div class="diff-content after">${this.afterText}</div>
          </div>
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-ui-diff": HairaDiff;
  }
}
