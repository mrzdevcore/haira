import { baseStyles, sharedKeyframes, scrollbarStyles } from "../theme";

const iconCopy = `<svg width="13" height="13" viewBox="0 0 16 16" fill="none"><rect x="5" y="5" width="9" height="9" rx="1.5" stroke="currentColor" stroke-width="1.5"/><path d="M11 5V3.5A1.5 1.5 0 0 0 9.5 2H3.5A1.5 1.5 0 0 0 2 3.5v6A1.5 1.5 0 0 0 3.5 11H5" stroke="currentColor" stroke-width="1.5"/></svg>`;
const iconCheck = `<svg width="13" height="13" viewBox="0 0 16 16" fill="none"><path d="M4 8.5L6.5 11L12 5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>`;

export class HairaCodeBlock extends HTMLElement {
  private codeText = "";

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
          justify-content: space-between;
          padding: 0.45rem 0.75rem;
          border-bottom: 1px solid var(--haira-border);
          background: var(--haira-bg);
        }
        .title {
          font-size: 0.78rem;
          font-weight: 600;
          color: var(--haira-text);
        }
        .lang {
          font-size: 0.68rem;
          color: var(--haira-muted);
          font-family: var(--haira-mono);
        }
        .actions {
          display: flex;
          align-items: center;
          gap: 0.5rem;
        }
        .copy-btn {
          background: none;
          border: none;
          color: var(--haira-muted);
          cursor: pointer;
          display: flex;
          align-items: center;
          gap: 0.3rem;
          font-size: 0.7rem;
          font-family: var(--haira-font);
          padding: 0.2rem 0.4rem;
          border-radius: 4px;
          transition: all 0.15s;
        }
        .copy-btn:hover { color: var(--haira-gold); background: var(--haira-gold-dim); }
        .copy-btn.copied { color: var(--haira-success); }
        pre {
          margin: 0;
          padding: 0.75rem 1rem;
          overflow-x: auto;
          ${scrollbarStyles}
        }
        code {
          font-family: var(--haira-mono);
          font-size: 0.78rem;
          color: var(--haira-text-dim);
          line-height: 1.6;
          white-space: pre;
        }
      </style>
      <div class="card">
        <div class="header">
          <div style="display:flex;align-items:center;gap:0.5rem">
            <span class="title" id="title"></span>
            <span class="lang" id="lang"></span>
          </div>
          <div class="actions">
            <button class="copy-btn" id="copy-btn">${iconCopy} Copy</button>
          </div>
        </div>
        <pre><code id="code"></code></pre>
      </div>
    `;

    this.shadowRoot!.getElementById("copy-btn")!.addEventListener("click", () => {
      this.copyCode();
    });
  }

  setProps(props: Record<string, unknown>) {
    try {
      const titleEl = this.shadowRoot!.getElementById("title")!;
      titleEl.textContent = (props.title as string) || "";

      const langEl = this.shadowRoot!.getElementById("lang")!;
      const lang = (props.language as string) || "";
      langEl.textContent = lang;

      const codeEl = this.shadowRoot!.getElementById("code")!;
      this.codeText = (props.code as string) || "";
      codeEl.textContent = this.codeText;
    } catch {
      // Graceful fallback
    }
  }

  private copyCode() {
    navigator.clipboard.writeText(this.codeText).then(() => {
      const btn = this.shadowRoot!.getElementById("copy-btn")!;
      btn.innerHTML = `${iconCheck} Copied`;
      btn.classList.add("copied");
      setTimeout(() => {
        btn.innerHTML = `${iconCopy} Copy`;
        btn.classList.remove("copied");
      }, 2000);
    });
  }
}
