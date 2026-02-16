import { baseStyles, sharedKeyframes, iconCopy, iconCopyDone } from "../theme";

export class HairaMessage extends HTMLElement {
  connectedCallback() {
    this.render();
  }

  private render() {
    const role = this.getAttribute("role") || "user";
    const content = this.getAttribute("content") || "";

    const shadow = this.attachShadow({ mode: "open" });
    shadow.innerHTML = `
      <style>
        ${baseStyles}
        ${sharedKeyframes}
        :host {
          display: block;
          margin-bottom: 0.6rem;
          animation: fadeSlideUp 0.25s ease-out;
        }
        .bubble {
          max-width: 80%;
          padding: 0.75rem 1rem;
          border-radius: 12px;
          line-height: 1.55;
          font-size: 0.9rem;
          word-wrap: break-word;
        }
        .user {
          background: var(--haira-gold);
          color: #1a0e04;
          margin-left: auto;
          border-bottom-right-radius: 4px;
          box-shadow: 0 2px 8px rgba(232, 163, 23, 0.15);
        }
        .assistant {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          color: var(--haira-text);
          border-bottom-left-radius: 4px;
        }
        .assistant .code-wrapper {
          position: relative;
          margin: 0.4rem 0;
        }
        .assistant pre {
          background: var(--haira-bg);
          border: 1px solid var(--haira-border);
          padding: 0.5rem;
          border-radius: 4px;
          overflow-x: auto;
          font-size: 0.8rem;
          font-family: var(--haira-mono);
          margin: 0;
        }
        .assistant code {
          background: var(--haira-bg);
          padding: 0.1rem 0.3rem;
          border-radius: 3px;
          font-size: 0.82rem;
          font-family: var(--haira-mono);
        }
        .assistant pre code {
          background: none;
          padding: 0;
        }
        .assistant strong { font-weight: 700; }
        .assistant em { font-style: italic; }
        .assistant ul {
          margin: 0.3rem 0;
          padding-left: 1.2rem;
        }
        .assistant li {
          margin: 0.15rem 0;
        }
        .copy-code {
          position: absolute;
          top: 0.35rem;
          right: 0.35rem;
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: 3px;
          padding: 0.2rem;
          cursor: pointer;
          color: var(--haira-muted);
          display: flex;
          align-items: center;
          opacity: 0;
          transition: opacity 0.15s;
        }
        .code-wrapper:hover .copy-code { opacity: 1; }
        .copy-code:hover {
          color: var(--haira-gold);
          border-color: var(--haira-gold);
        }
      </style>
      <div class="bubble ${role}" id="bubble"></div>
    `;

    const bubble = shadow.getElementById("bubble")!;
    if (role === "assistant") {
      bubble.innerHTML = this.renderMarkdown(content);
      this.attachCopyHandlers(shadow);
    } else {
      bubble.textContent = content;
    }
  }

  updateContent(text: string) {
    const bubble = this.shadowRoot?.getElementById("bubble");
    if (bubble) {
      bubble.innerHTML = this.renderMarkdown(text);
      this.attachCopyHandlers(this.shadowRoot!);
    }
  }

  private attachCopyHandlers(root: ShadowRoot) {
    root.querySelectorAll(".copy-code").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const wrapper = btn.closest(".code-wrapper");
        const code = wrapper?.querySelector("code");
        if (!code) return;
        try {
          await navigator.clipboard.writeText(code.textContent || "");
          btn.innerHTML = iconCopyDone;
          setTimeout(() => {
            btn.innerHTML = iconCopy;
          }, 1500);
        } catch {
          /* clipboard unavailable */
        }
      });
    });
  }

  private renderMarkdown(text: string): string {
    let t = this.esc(text);
    // Code blocks with copy button
    t = t.replace(
      /```(\w*)\n([\s\S]*?)```/g,
      (_: string, _lang: string, code: string) =>
        `<div class="code-wrapper"><pre><code>${code.trim()}</code></pre><button class="copy-code" title="Copy code">${iconCopy}</button></div>`,
    );
    // Inline code
    t = t.replace(/`([^`]+)`/g, "<code>$1</code>");
    // Bold
    t = t.replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>");
    // Italic
    t = t.replace(/(?<!\w)_(.+?)_(?!\w)/g, "<em>$1</em>");
    // Bulleted lists: lines starting with "- "
    t = t.replace(
      /(^|\n)- (.+?)(?=\n|$)/g,
      (_: string, pre: string, item: string) => `${pre}<li>${item}</li>`,
    );
    // Wrap consecutive <li> in <ul>
    t = t.replace(/((?:<li>.*?<\/li>\n?)+)/g, "<ul>$1</ul>");
    // Remaining newlines
    t = t.replace(/\n/g, "<br>");
    return t;
  }

  private esc(s: string): string {
    return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }
}
