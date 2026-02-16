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
          animation: fadeSlideUp 0.3s ease-out;
        }
        .card {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          padding: 1rem;
          position: relative;
        }
        .header {
          display: flex;
          align-items: center;
          gap: 0.4rem;
          font-weight: 600;
          font-size: 0.82rem;
          color: var(--haira-muted);
          margin-bottom: 0.5rem;
        }
        .dot {
          width: 7px;
          height: 7px;
          border-radius: 50%;
          flex-shrink: 0;
        }
        .dot.success { background: var(--haira-success); }
        .dot.error { background: var(--haira-error); }
        .body-wrap {
          position: relative;
        }
        .body {
          background: var(--haira-bg);
          border: 1px solid var(--haira-border);
          border-radius: 6px;
          padding: 0.85rem;
          font-family: var(--haira-mono);
          font-size: 0.82rem;
          line-height: 1.55;
          white-space: pre-wrap;
          word-break: break-word;
          color: var(--haira-text-dim);
          max-height: 500px;
          overflow-y: auto;
          ${scrollbarStyles}
        }
        .body.success { border-left: 3px solid var(--haira-success); }
        .body.error { border-left: 3px solid var(--haira-error); }
        .copy-btn {
          position: absolute;
          top: 0.5rem;
          right: 0.5rem;
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: 4px;
          padding: 0.3rem;
          cursor: pointer;
          color: var(--haira-muted);
          display: flex;
          align-items: center;
          justify-content: center;
          transition: all 0.15s;
          opacity: 0;
        }
        .body-wrap:hover .copy-btn { opacity: 1; }
        .copy-btn:hover {
          color: var(--haira-gold);
          border-color: var(--haira-gold);
        }
      </style>
      <div class="card">
        <div class="header">
          <span class="dot" id="dot"></span>
          <span>Result</span>
        </div>
        <div class="body-wrap">
          <div class="body" id="body"></div>
          <button class="copy-btn" id="copy-btn" title="Copy to clipboard">${iconCopy}</button>
        </div>
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
    body.className = `body ${isError ? "error" : "success"}`;
    dot.className = `dot ${isError ? "error" : "success"}`;

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
