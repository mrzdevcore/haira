import { BaseComponent, animateInCSS, cardCSS, icons } from "../core";
import type { ProgressProps } from "../core/types";

const stepIcons: Record<string, { icon: string; color: string }> = {
  done:    { icon: icons.stepDone,    color: "var(--haira-success)" },
  active:  { icon: icons.stepActive,  color: "var(--haira-accent)" },
  pending: { icon: icons.stepPending, color: "var(--haira-muted)" },
  failed:  { icon: icons.stepFailed,  color: "var(--haira-error)" },
};

export class HairaProgressView extends BaseComponent<ProgressProps> {
  protected render() {
    return `
      <div class="card">
        <div class="title-bar" id="title"></div>
        <div class="steps" id="steps"></div>
      </div>`;
  }

  protected styles() {
    return `
      ${animateInCSS}
      .card { ${cardCSS} }
      .title-bar {
        padding: 0.6rem 1rem;
        font-size: 0.8rem;
        font-weight: 600;
        color: var(--haira-text);
        border-bottom: 1px solid var(--haira-border);
        display: none;
      }
      .steps { padding: 0.6rem 1rem; }
      .step {
        display: flex; align-items: flex-start; gap: 0.6rem;
        position: relative; padding-bottom: 0.75rem;
      }
      .step:last-child { padding-bottom: 0; }
      .step::before {
        content: ""; position: absolute;
        left: 6.5px; top: 18px; bottom: 0;
        width: 1px; background: var(--haira-border);
      }
      .step:last-child::before { display: none; }
      .step-icon { display: flex; flex-shrink: 0; margin-top: 1px; }
      .step-content { flex: 1; min-width: 0; }
      .step-name { font-size: 0.8rem; font-weight: 500; line-height: 1.3; }
      .step-detail { font-size: 0.72rem; color: var(--haira-muted); margin-top: 0.15rem; }
      @keyframes spin { to { transform: rotate(360deg); } }`;
  }

  protected onUpdate() {
    const { title, steps = [] } = this.props;
    const titleEl = this.$("title");
    if (title) {
      titleEl.textContent = title;
      titleEl.style.display = "";
    }
    this.$("steps").innerHTML = steps
      .map((step) => {
        const si = stepIcons[step.status] || stepIcons.pending;
        return `<div class="step">
          <span class="step-icon" style="color:${si.color}">${si.icon}</span>
          <div class="step-content">
            <div class="step-name" style="color:${si.color}">${this.esc(step.name || "")}</div>
            ${step.detail ? `<div class="step-detail">${this.esc(step.detail)}</div>` : ""}
          </div>
        </div>`;
      })
      .join("");
  }
}
