import { baseStyles, sharedKeyframes } from "../theme";

const stepIcons: Record<string, { icon: string; color: string }> = {
  done: {
    icon: `<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="7" fill="currentColor" opacity="0.15"/><path d="M5 8.5L7 10.5L11 5.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>`,
    color: "var(--haira-success)",
  },
  active: {
    icon: `<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-dasharray="28 10" style="animation:spin 0.7s linear infinite;transform-origin:center"/></svg>`,
    color: "var(--haira-gold)",
  },
  pending: {
    icon: `<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="6.5" stroke="currentColor" stroke-width="1.5" stroke-dasharray="3 2"/></svg>`,
    color: "var(--haira-muted)",
  },
  failed: {
    icon: `<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="7" fill="currentColor" opacity="0.15"/><path d="M5.5 5.5L10.5 10.5M10.5 5.5L5.5 10.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>`,
    color: "var(--haira-error)",
  },
};

export class HairaProgressView extends HTMLElement {
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
        .steps {
          padding: 0.6rem 1rem;
        }
        .step {
          display: flex;
          align-items: flex-start;
          gap: 0.6rem;
          position: relative;
          padding-bottom: 0.75rem;
        }
        .step:last-child { padding-bottom: 0; }
        .step::before {
          content: "";
          position: absolute;
          left: 6.5px;
          top: 18px;
          bottom: 0;
          width: 1px;
          background: var(--haira-border);
        }
        .step:last-child::before { display: none; }
        .step-icon { display: flex; flex-shrink: 0; margin-top: 1px; }
        .step-content { flex: 1; min-width: 0; }
        .step-name {
          font-size: 0.8rem;
          font-weight: 500;
          line-height: 1.3;
        }
        .step-detail {
          font-size: 0.72rem;
          color: var(--haira-muted);
          margin-top: 0.15rem;
        }
        @keyframes spin {
          to { transform: rotate(360deg); }
        }
      </style>
      <div class="card">
        <div class="title-bar" id="title"></div>
        <div class="steps" id="steps"></div>
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

      const steps = (props.steps as Array<Record<string, string>>) || [];
      const container = this.shadowRoot!.getElementById("steps")!;
      container.innerHTML = steps
        .map((step) => {
          const status = step.status || "pending";
          const si = stepIcons[status] || stepIcons.pending;
          return `<div class="step">
            <span class="step-icon" style="color:${si.color}">${si.icon}</span>
            <div class="step-content">
              <div class="step-name" style="color:${si.color}">${this.esc(step.name || "")}</div>
              ${step.detail ? `<div class="step-detail">${this.esc(step.detail)}</div>` : ""}
            </div>
          </div>`;
        })
        .join("");
    } catch {
      // Graceful fallback
    }
  }

  private esc(s: string): string {
    return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }
}
