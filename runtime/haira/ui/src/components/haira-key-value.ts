import { baseStyles, sharedKeyframes } from "../theme";

const styleColors: Record<string, string> = {
  success: "var(--haira-success)",
  error: "var(--haira-error)",
  warning: "var(--haira-warn)",
  info: "var(--haira-info)",
  code: "inherit",
};

export class HairaKeyValue extends HTMLElement {
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
        .items {
          padding: 0.5rem 0;
        }
        .item {
          display: flex;
          align-items: baseline;
          padding: 0.3rem 1rem;
          gap: 0.75rem;
        }
        .key {
          font-size: 0.75rem;
          font-weight: 600;
          color: var(--haira-muted);
          min-width: 100px;
          flex-shrink: 0;
        }
        .value {
          font-size: 0.8rem;
          color: var(--haira-text-dim);
          word-break: break-word;
        }
        .value.code {
          font-family: var(--haira-mono);
          font-size: 0.75rem;
          background: var(--haira-bg);
          padding: 0.15rem 0.4rem;
          border-radius: 4px;
        }
      </style>
      <div class="card">
        <div class="title-bar" id="title"></div>
        <div class="items" id="items"></div>
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

      const items = (props.items as Array<Record<string, string>>) || [];
      const container = this.shadowRoot!.getElementById("items")!;
      container.innerHTML = items
        .map((item) => {
          const color = styleColors[item.style] || "";
          const colorStyle = color && color !== "inherit" ? `color:${color}` : "";
          const isCode = item.style === "code";
          return `<div class="item">
            <span class="key">${this.esc(item.key || "")}</span>
            <span class="value ${isCode ? "code" : ""}" style="${colorStyle}">${this.esc(item.value || "")}</span>
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
