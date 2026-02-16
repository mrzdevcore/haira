import { baseStyles, sharedKeyframes, scrollbarStyles } from "../theme";

export class HairaDiff extends HTMLElement {
  connectedCallback() {
    this.attachShadow({ mode: "open" });
    this.shadowRoot!.innerHTML = `
      <style>
        ${baseStyles}
        ${sharedKeyframes}
        :host {
          display: block;
          animation: fadeSlideUp 0.25s ease-out;
        }
        .card {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          overflow: hidden;
        }
        .title-bar {
          padding: 0.6rem 1rem;
          font-size: 0.8rem;
          font-weight: 600;
          color: var(--haira-text);
          border-bottom: 1px solid var(--haira-border);
          display: none;
        }
        .diff-grid {
          display: grid;
          grid-template-columns: 1fr 1fr;
        }
        .pane {
          overflow-x: auto;
          ${scrollbarStyles}
        }
        .pane + .pane {
          border-left: 1px solid var(--haira-border);
        }
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
        .pane:first-child pre {
          background: rgba(239, 68, 68, 0.03);
        }
        .pane:last-child pre {
          background: rgba(34, 197, 94, 0.03);
        }
      </style>
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
      </div>
    `;
  }

  setProps(props: Record<string, unknown>) {
    try {
      const titleEl = this.shadowRoot!.getElementById("title")!;
      if (props.title) {
        titleEl.textContent = props.title as string;
        titleEl.style.display = "";
      }

      const beforeLabel = this.shadowRoot!.getElementById("before-label")!;
      beforeLabel.textContent = (props.before_label as string) || "Before";

      const afterLabel = this.shadowRoot!.getElementById("after-label")!;
      afterLabel.textContent = (props.after_label as string) || "After";

      this.shadowRoot!.getElementById("before")!.textContent =
        (props.before as string) || "";
      this.shadowRoot!.getElementById("after")!.textContent =
        (props.after as string) || "";
    } catch {
      // Graceful fallback
    }
  }
}
