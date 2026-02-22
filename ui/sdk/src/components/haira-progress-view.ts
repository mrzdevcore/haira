import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles, keyframes, animateInStyles } from "../core/styles";
import { iconStrings } from "../core/icons";
import type { ProgressProps } from "../core/types";

@customElement("haira-ui-progress-view")
export class HairaProgressView extends LitElement {
  static styles = [
    baseStyles,
    animateInStyles,
    css`
      .progress-card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        overflow: hidden;
      }

      .progress-header {
        padding: 0.55rem 0.85rem;
        border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg);
      }
      .progress-title {
        font-size: 0.85rem;
        font-weight: 700;
        color: var(--haira-text);
      }

      .step-list {
        padding: 0.65rem 0.85rem;
        position: relative;
      }

      /* Connecting line */
      .step-list::before {
        content: "";
        position: absolute;
        left: calc(0.85rem + 6px);
        top: 0.85rem;
        bottom: 0.85rem;
        width: 2px;
        background: var(--haira-border);
      }

      .step-item {
        display: flex;
        align-items: flex-start;
        gap: 0.65rem;
        position: relative;
        padding: 0.35rem 0;
      }

      .step-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 14px;
        height: 14px;
        flex-shrink: 0;
        margin-top: 0.15rem;
        position: relative;
        z-index: 1;
      }
      .step-icon.done {
        color: var(--haira-success);
      }
      .step-icon.running {
        color: var(--haira-accent);
      }
      .step-icon.pending {
        color: var(--haira-muted);
      }
      .step-icon.failed {
        color: var(--haira-error);
      }

      .step-body {
        flex: 1;
        min-width: 0;
      }
      .step-name {
        font-size: 0.84rem;
        font-weight: 600;
        color: var(--haira-text);
      }
      .step-detail {
        font-size: 0.76rem;
        color: var(--haira-muted);
        margin-top: 0.1rem;
      }
    `,
  ];

  @property({ type: String }) title: string = "";
  @property({ type: Array }) steps: ProgressProps["steps"] = [];

  /** Set all props at once */
  public setProps(props: ProgressProps): void {
    this.title = props.title || "";
    this.steps = props.steps || [];
  }

  private _getStepIcon(status: string) {
    switch (status) {
      case "done":
      case "complete":
      case "completed":
        return html`<span class="step-icon done"
          >${unsafeHTML(iconStrings.stepDone)}</span
        >`;
      case "running":
      case "active":
      case "in_progress":
        return html`<span class="step-icon running"
          >${unsafeHTML(iconStrings.stepActive)}</span
        >`;
      case "failed":
      case "error":
        return html`<span class="step-icon failed"
          >${unsafeHTML(iconStrings.stepFailed)}</span
        >`;
      default:
        return html`<span class="step-icon pending"
          >${unsafeHTML(iconStrings.stepPending)}</span
        >`;
    }
  }

  render() {
    return html`
      <div class="progress-card">
        ${this.title
          ? html`
              <div class="progress-header">
                <span class="progress-title">${this.title}</span>
              </div>
            `
          : nothing}
        <div class="step-list">
          ${this.steps.map(
            (step) => html`
              <div class="step-item">
                ${this._getStepIcon(step.status)}
                <div class="step-body">
                  <div class="step-name">${step.name}</div>
                  ${step.detail
                    ? html`<div class="step-detail">${step.detail}</div>`
                    : nothing}
                </div>
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
    "haira-ui-progress-view": HairaProgressView;
  }
}
