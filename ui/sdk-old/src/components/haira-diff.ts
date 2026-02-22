import { BaseComponent, animateInCSS, cardCSS, scrollbarCSS } from "../core";
import type { DiffProps } from "../core/types";

export class HairaDiff extends BaseComponent<DiffProps> {
  protected render() {
    return `
      <div class="card">
        <div class="title-bar" id="title"></div>
        <div class="diff-grid">
          <div class="pane">
            <div class="pane-header before" id="before-label">Before</div>
            <pre id="before"></pre>
          </div>
          <div class="pane">
            <div class="pane-header after" id="after-label">After</div>
            <pre id="after"></pre>
          </div>
        </div>
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
      .diff-grid { display: grid; grid-template-columns: 1fr 1fr; }
      .pane { overflow-x: auto; ${scrollbarCSS} }
      .pane + .pane { border-left: 1px solid var(--haira-border); }
      .pane-header {
        padding: 0.4rem 0.75rem;
        font-size: 0.72rem;
        font-weight: 600;
        color: var(--haira-muted);
        text-transform: uppercase;
        letter-spacing: 0.04em;
        background: var(--haira-bg);
        border-bottom: 1px solid var(--haira-border);
      }
      .pane-header.before { color: var(--haira-error); }
      .pane-header.after { color: var(--haira-success); }
      pre {
        margin: 0;
        padding: 0.6rem 0.75rem;
        font-family: var(--haira-mono);
        font-size: 0.75rem;
        color: var(--haira-text-dim);
        line-height: 1.6;
        white-space: pre;
        min-height: 3rem;
      }
      .pane:first-child pre { background: rgba(239, 68, 68, 0.03); }
      .pane:last-child pre { background: rgba(34, 197, 94, 0.03); }`;
  }

  protected onUpdate() {
    const p = this.props;
    const titleEl = this.$("title");
    if (p.title) {
      titleEl.textContent = p.title;
      titleEl.style.display = "";
    }
    this.$("before-label").textContent = p.before_label || "Before";
    this.$("after-label").textContent = p.after_label || "After";
    this.$("before").textContent = p.before || "";
    this.$("after").textContent = p.after || "";
  }
}
