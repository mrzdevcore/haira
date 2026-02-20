import { baseStyles, sharedKeyframes } from "../theme";

const icons: Record<string, string> = {
  success: `<svg width="16" height="16" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="7" stroke="currentColor" stroke-width="1.5"/><path d="M5 8.5L7 10.5L11 5.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>`,
  error: `<svg width="16" height="16" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="7" stroke="currentColor" stroke-width="1.5"/><path d="M5.5 5.5L10.5 10.5M10.5 5.5L5.5 10.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>`,
  warning: `<svg width="16" height="16" viewBox="0 0 16 16" fill="none"><path d="M8 2L14.5 13H1.5L8 2Z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/><path d="M8 6.5V9" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><circle cx="8" cy="11" r="0.75" fill="currentColor"/></svg>`,
  info: `<svg width="16" height="16" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="7" stroke="currentColor" stroke-width="1.5"/><path d="M8 7V11" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><circle cx="8" cy="5" r="0.75" fill="currentColor"/></svg>`,
};

const statusColors: Record<string, string> = {
  success: "var(--haira-success)",
  error: "var(--haira-error)",
  warning: "var(--haira-warn)",
  info: "var(--haira-info)",
};

export class HairaStatusCard extends HTMLElement {
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
        .header {
          display: flex;
          align-items: center;
          gap: 0.4rem;
          padding: 0.45rem 0.75rem;
        }
        .icon { display: flex; align-items: center; flex-shrink: 0; }
        .icon svg { width: 14px; height: 14px; }
        .title { font-size: 0.78rem; font-weight: 600; }
        .message {
          font-size: 0.75rem;
          color: var(--haira-text-dim);
          padding: 0 0.75rem 0.45rem 2rem;
          line-height: 1.4;
        }
        .sections {
          border-top: 1px solid var(--haira-border);
        }
        .section {
          padding: 0.4rem 0.75rem;
          border-bottom: 1px solid var(--haira-border);
        }
        .section:last-child { border-bottom: none; }
        .section-label {
          font-size: 0.68rem;
          font-weight: 600;
          color: var(--haira-muted);
          text-transform: uppercase;
          letter-spacing: 0.04em;
          margin-bottom: 0.2rem;
        }
        .section-content {
          font-size: 0.75rem;
          color: var(--haira-text-dim);
          line-height: 1.4;
          white-space: pre-wrap;
        }
        .section-content.code {
          font-family: var(--haira-mono);
          font-size: 0.72rem;
          background: var(--haira-bg);
          padding: 0.35rem 0.6rem;
          border-radius: var(--haira-radius-sm);
          overflow-x: auto;
        }
        /* Inline variant: title + message on one line, no sections */
        .card.inline .header {
          padding: 0.35rem 0.65rem;
        }
        .card.inline .message {
          display: inline;
          padding: 0;
          margin-left: 0.15rem;
          font-weight: 400;
        }
        .card.inline .header-row {
          display: flex;
          align-items: center;
          gap: 0.4rem;
          flex-wrap: wrap;
        }
      </style>
      <div class="card" id="card">
        <div class="header" id="header">
          <span class="icon" id="icon"></span>
          <span class="title" id="title"></span>
        </div>
        <div class="message" id="message"></div>
        <div class="sections" id="sections" style="display:none"></div>
      </div>
    `;
  }

  setProps(props: Record<string, unknown>) {
    try {
      const status = (props.status as string) || "info";
      const color = statusColors[status] || statusColors.info;
      const sections = props.sections as
        | Array<Record<string, string>>
        | undefined;
      const hasSections = sections && sections.length > 0;
      const hasMessage = !!props.message;

      const card = this.shadowRoot!.getElementById("card")!;
      const header = this.shadowRoot!.getElementById("header")!;

      // Inline variant: simple status with title + message, no sections
      if (!hasSections && hasMessage) {
        card.classList.add("inline");
        header.innerHTML = `
          <div class="header-row">
            <span class="icon" id="icon"></span>
            <span class="title" id="title"></span>
            <span class="message" id="message"></span>
          </div>`;
        const icon = header.querySelector("#icon")! as HTMLElement;
        icon.innerHTML = icons[status] || icons.info;
        icon.style.color = color;
        const titleEl = header.querySelector("#title")! as HTMLElement;
        titleEl.textContent = (props.title as string) || "";
        titleEl.style.color = color;
        (header.querySelector("#message")! as HTMLElement).textContent =
          props.message as string;
        this.shadowRoot!.getElementById("message")!.style.display = "none";
        this.shadowRoot!.getElementById("sections")!.style.display = "none";
        card.style.borderLeft = `3px solid ${color.includes("var(") ? color : color}`;
        return;
      }

      card.classList.remove("inline");

      const icon = this.shadowRoot!.getElementById("icon")!;
      icon.innerHTML = icons[status] || icons.info;
      icon.style.color = color;

      const title = this.shadowRoot!.getElementById("title")!;
      title.textContent = (props.title as string) || "";
      title.style.color = color;

      const message = this.shadowRoot!.getElementById("message")!;
      if (hasMessage) {
        message.textContent = props.message as string;
        message.style.display = "";
      } else {
        message.style.display = "none";
      }

      card.style.borderLeft = `3px solid ${color.includes("var(") ? color : color}`;

      const sectionsEl = this.shadowRoot!.getElementById("sections")!;
      if (hasSections) {
        sectionsEl.style.display = "";
        sectionsEl.innerHTML = sections
          .map(
            (s) => `
            <div class="section">
              <div class="section-label">${this.esc(s.label || "")}</div>
              <div class="section-content ${s.style === "code" ? "code" : ""}">${this.esc(s.content || "")}</div>
            </div>`,
          )
          .join("");
      } else {
        sectionsEl.style.display = "none";
      }
    } catch {
      // Graceful fallback on bad props
    }
  }

  private esc(s: string): string {
    return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }
}
