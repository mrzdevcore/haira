import {
  baseStyles,
  sharedKeyframes,
  scrollbarStyles,
  iconCopy,
  iconCopyDone,
} from "../theme";

export class HairaResult extends HTMLElement {
  connectedCallback() {
    const shadow = this.attachShadow({ mode: "open" });
    shadow.innerHTML = `
      <style>
        ${baseStyles}
        ${sharedKeyframes}
        :host { display: none; margin-top: 0.75rem; }
        :host([visible]) {
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
          padding: 0.6rem 0.85rem;
          border-bottom: 1px solid var(--haira-border);
        }
        .header-left {
          display: flex;
          align-items: center;
          gap: 0.4rem;
          font-weight: 600;
          font-size: 0.78rem;
          color: var(--haira-muted);
        }
        .dot {
          width: 6px;
          height: 6px;
          border-radius: 50%;
          flex-shrink: 0;
        }
        .dot.success { background: var(--haira-success); }
        .dot.error { background: var(--haira-error); }
        .copy-btn {
          background: none;
          border: 1px solid transparent;
          border-radius: 4px;
          padding: 0.25rem;
          cursor: pointer;
          color: var(--haira-muted);
          display: flex;
          align-items: center;
          justify-content: center;
          transition: all 0.15s;
        }
        .copy-btn:hover {
          color: var(--haira-gold);
          border-color: var(--haira-border);
          background: var(--haira-bg-elevated);
        }
        .body {
          padding: 0.85rem;
          font-family: var(--haira-mono);
          font-size: 0.8rem;
          line-height: 1.6;
          white-space: pre-wrap;
          word-break: break-word;
          color: var(--haira-text-dim);
          max-height: 400px;
          overflow-y: auto;
          ${scrollbarStyles}
        }
      </style>
      <div class="card">
        <div class="header">
          <div class="header-left">
            <span class="dot" id="dot"></span>
            <span id="label">Result</span>
          </div>
          <button class="copy-btn" id="copy-btn" title="Copy to clipboard">${iconCopy}</button>
        </div>
        <div class="body" id="body"></div>
      </div>
    `;

    shadow
      .getElementById("copy-btn")!
      .addEventListener("click", () => this.copyResult());
  }

  show(data: unknown, isError: boolean) {
    this.setAttribute("visible", "");
    const body = this.shadowRoot!.getElementById("body")!;
    const dot = this.shadowRoot!.getElementById("dot")!;
    const label = this.shadowRoot!.getElementById("label")!;
    dot.className = `dot ${isError ? "error" : "success"}`;
    label.textContent = isError ? "Error" : "Result";

    if (typeof data === "string") {
      body.textContent = data;
    } else {
      body.textContent = JSON.stringify(data, null, 2);
    }
  }

  hide() {
    this.removeAttribute("visible");
  }

  private async copyResult() {
    const body = this.shadowRoot?.getElementById("body");
    const btn = this.shadowRoot?.getElementById("copy-btn");
    if (!body || !btn) return;
    try {
      await navigator.clipboard.writeText(body.textContent || "");
      btn.innerHTML = iconCopyDone;
      setTimeout(() => {
        btn.innerHTML = iconCopy;
      }, 1500);
    } catch {
      // clipboard API not available
    }
  }
}
