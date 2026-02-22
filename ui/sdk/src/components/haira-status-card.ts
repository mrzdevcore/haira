import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles, keyframes, animateInStyles } from "../core/styles";
import { iconStrings } from "../core/icons";
import type { StatusCardProps } from "../core/types";

const STATUS_CONFIG: Record<
  string,
  { icon: string; color: string; bg: string }
> = {
  success: {
    icon: iconStrings.statusSuccess,
    color: "var(--haira-success)",
    bg: "rgba(34,197,94,0.08)",
  },
  error: {
    icon: iconStrings.statusError,
    color: "var(--haira-error)",
    bg: "rgba(239,68,68,0.08)",
  },
  warning: {
    icon: iconStrings.statusWarning,
    color: "var(--haira-warn)",
    bg: "rgba(234,179,8,0.08)",
  },
  info: {
    icon: iconStrings.statusInfo,
    color: "var(--haira-info)",
    bg: "rgba(59,130,246,0.08)",
  },
};

@customElement("haira-ui-status-card")
export class HairaStatusCard extends LitElement {
  static styles = [
    baseStyles,
    animateInStyles,
    css`
      .card {
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        overflow: hidden;
      }

      /* Inline variant */
      .inline {
        display: flex;
        align-items: center;
        gap: 0.55rem;
        padding: 0.55rem 0.85rem;
        background: var(--haira-bg-card);
      }
      .inline .icon {
        display: flex;
        align-items: center;
        flex-shrink: 0;
      }
      .inline .text {
        font-size: 0.84rem;
        font-weight: 600;
        color: var(--haira-text);
      }

      /* Full variant */
      .full-header {
        display: flex;
        align-items: center;
        gap: 0.55rem;
        padding: 0.6rem 0.85rem;
        border-bottom: 1px solid var(--haira-border);
      }
      .full-header .icon {
        display: flex;
        align-items: center;
        flex-shrink: 0;
      }
      .full-header .title {
        font-size: 0.88rem;
        font-weight: 700;
        color: var(--haira-text);
      }

      .full-body {
        padding: 0.65rem 0.85rem;
        background: var(--haira-bg-card);
      }
      .full-message {
        font-size: 0.84rem;
        line-height: 1.55;
        color: var(--haira-text-dim);
        white-space: pre-wrap;
        word-break: break-word;
      }

      /* Sections */
      .sections {
        border-top: 1px solid var(--haira-border);
        background: var(--haira-bg);
      }
      .section-item {
        padding: 0.5rem 0.85rem;
        border-bottom: 1px solid var(--haira-border);
      }
      .section-item:last-child {
        border-bottom: none;
      }
      .section-label {
        font-size: 0.72rem;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.03em;
        color: var(--haira-muted);
        margin-bottom: 0.2rem;
      }
      .section-content {
        font-size: 0.82rem;
        line-height: 1.5;
        color: var(--haira-text);
        white-space: pre-wrap;
        word-break: break-word;
      }
      .section-content.success {
        color: var(--haira-success);
      }
      .section-content.error {
        color: var(--haira-error);
      }
      .section-content.warning {
        color: var(--haira-warn);
      }
      .section-content.info {
        color: var(--haira-info);
      }
      .section-content.code {
        font-family: var(--haira-mono);
        font-size: 0.78rem;
        background: var(--haira-bg-input);
        padding: 0.35rem 0.55rem;
        border-radius: var(--haira-radius-sm);
      }
    `,
  ];

  @property({ type: String }) status: StatusCardProps["status"] = "info";
  @property({ type: String }) title: string = "";
  @property({ type: String }) message: string = "";
  @property({ type: Array }) sections: StatusCardProps["sections"] = [];
  @property({ type: Boolean }) _restored: boolean = false;

  /** Set all props at once */
  public setProps(props: StatusCardProps): void {
    this.status = props.status;
    this.title = props.title;
    this.message = props.message || "";
    this.sections = props.sections || [];
    if (props._restored) this._restored = true;
  }

  private _getConfig() {
    return STATUS_CONFIG[this.status] || STATUS_CONFIG.info;
  }

  private _isInline(): boolean {
    // Inline: no message and no sections
    return !this.message && (!this.sections || this.sections.length === 0);
  }

  render() {
    const cfg = this._getConfig();

    if (this._isInline()) {
      return html`
        <div class="card inline" style="border-left: 3px solid ${cfg.color}">
          <span class="icon" style="color:${cfg.color}">
            ${unsafeHTML(cfg.icon)}
          </span>
          <span class="text">${this.title}</span>
        </div>
      `;
    }

    return html`
      <div class="card">
        <div class="full-header" style="background:${cfg.bg}">
          <span class="icon" style="color:${cfg.color}">
            ${unsafeHTML(cfg.icon)}
          </span>
          <span class="title">${this.title}</span>
        </div>

        ${this.message
          ? html`
              <div class="full-body">
                <div class="full-message">${this.message}</div>
              </div>
            `
          : nothing}
        ${this.sections && this.sections.length > 0
          ? html`
              <div class="sections">
                ${this.sections.map(
                  (sec) => html`
                    <div class="section-item">
                      <div class="section-label">${sec.label}</div>
                      <div class="section-content ${sec.style || ""}">
                        ${sec.content}
                      </div>
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
    "haira-ui-status-card": HairaStatusCard;
  }
}
