import { baseStyles, sharedKeyframes, iconCopy, iconCopyDone } from "../theme";

const iconFile = `<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>`;

export class HairaMessage extends HTMLElement {
  connectedCallback() {
    this.render();
  }

  private render() {
    const role = this.getAttribute("role") || "user";
    const content = this.getAttribute("content") || "";
    const file = this.getAttribute("file") || "";

    const shadow = this.attachShadow({ mode: "open" });
    shadow.innerHTML = `
      <style>
        ${baseStyles}
        ${sharedKeyframes}
        :host {
          display: block;
          animation: fadeSlideUp 0.2s ease-out;
        }
        .row {
          display: flex;
          gap: 0.6rem;
          align-items: flex-start;
        }
        .row.user {
          justify-content: flex-end;
        }
        .row.assistant {
          justify-content: flex-start;
        }

        /* Avatar */
        .avatar {
          width: 26px;
          height: 26px;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          flex-shrink: 0;
          font-size: 0.65rem;
          font-weight: 700;
          margin-top: 2px;
        }
        .avatar.assistant {
          background: var(--haira-gold-dim);
          border: 1px solid rgba(232, 163, 23, 0.2);
          color: var(--haira-gold);
        }
        .avatar.user {
          display: none;
        }

        /* Bubble */
        .bubble {
          padding: 0.7rem 0.9rem;
          border-radius: 12px;
          line-height: 1.6;
          font-size: 0.88rem;
          word-wrap: break-word;
          min-width: 40px;
        }
        .bubble.user {
          background: var(--haira-bg-elevated);
          border: 1px solid var(--haira-border);
          color: var(--haira-text);
          border-bottom-right-radius: 4px;
          max-width: 85%;
        }
        .bubble.assistant {
          background: transparent;
          color: var(--haira-text);
          padding: 0.4rem 0;
          flex: 1;
        }

        /* File chip in user message */
        .file-tag {
          display: inline-flex;
          align-items: center;
          gap: 0.3rem;
          background: rgba(0,0,0,0.12);
          padding: 0.2rem 0.5rem;
          border-radius: 5px;
          font-size: 0.75rem;
          margin-bottom: 0.3rem;
          font-weight: 500;
        }
        .file-tag svg { opacity: 0.7; }

        /* Assistant markdown */
        .bubble.assistant .code-wrapper {
          position: relative;
          margin: 0.5rem 0;
        }
        .bubble.assistant pre {
          background: var(--haira-bg);
          border: 1px solid var(--haira-border);
          padding: 0.6rem 0.75rem;
          border-radius: 6px;
          overflow-x: auto;
          font-size: 0.78rem;
          font-family: var(--haira-mono);
          line-height: 1.5;
          margin: 0;
        }
        .bubble.assistant code {
          background: var(--haira-bg-elevated);
          border: 1px solid var(--haira-border);
          padding: 0.1rem 0.3rem;
          border-radius: 3px;
          font-size: 0.8rem;
          font-family: var(--haira-mono);
          color: var(--haira-gold-light);
        }
        .bubble.assistant pre code {
          background: none;
          border: none;
          padding: 0;
          color: var(--haira-text);
        }
        .bubble.assistant strong { font-weight: 700; }
        .bubble.assistant em { font-style: italic; color: var(--haira-text-dim); }
        .bubble.assistant p {
          margin: 0.3rem 0;
        }
        .bubble.assistant p:first-child { margin-top: 0; }
        .bubble.assistant p:last-child { margin-bottom: 0; }
        .bubble.assistant ul, .bubble.assistant ol {
          margin: 0.35rem 0;
          padding-left: 1.3rem;
        }
        .bubble.assistant li {
          margin: 0.2rem 0;
        }
        .bubble.assistant h1, .bubble.assistant h2, .bubble.assistant h3 {
          font-size: 0.9rem;
          font-weight: 700;
          margin: 0.6rem 0 0.25rem;
          color: var(--haira-text);
        }
        .bubble.assistant h1:first-child,
        .bubble.assistant h2:first-child,
        .bubble.assistant h3:first-child {
          margin-top: 0;
        }
        .bubble.assistant hr {
          border: none;
          border-top: 1px solid var(--haira-border);
          margin: 0.5rem 0;
        }
        .bubble.assistant a {
          color: var(--haira-gold);
          text-decoration: none;
        }
        .bubble.assistant a:hover { text-decoration: underline; }
        .bubble.assistant blockquote {
          border-left: 3px solid var(--haira-gold);
          margin: 0.4rem 0;
          padding: 0.2rem 0.6rem;
          color: var(--haira-text-dim);
        }

        /* Copy button on code blocks */
        .copy-code {
          position: absolute;
          top: 0.4rem;
          right: 0.4rem;
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: 4px;
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
      <div class="row ${role}">
        ${role === "assistant" ? `<div class="avatar assistant">H</div>` : ""}
        <div class="bubble ${role}" id="bubble"></div>
      </div>
    `;

    const bubble = shadow.getElementById("bubble")!;
    if (role === "assistant") {
      bubble.innerHTML = this.renderMarkdown(content);
      this.attachCopyHandlers(shadow);
    } else {
      let html = "";
      if (file) {
        html += `<div class="file-tag">${iconFile} ${this.esc(file)}</div><br>`;
      }
      if (content) {
        html += this.esc(content);
      }
      bubble.innerHTML = html;
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

    // Headings (### h3, ## h2, # h1)
    t = t.replace(/^### (.+)$/gm, "<h3>$1</h3>");
    t = t.replace(/^## (.+)$/gm, "<h2>$1</h2>");
    t = t.replace(/^# (.+)$/gm, "<h1>$1</h1>");

    // Horizontal rule
    t = t.replace(/^---$/gm, "<hr>");

    // Bold
    t = t.replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>");

    // Italic
    t = t.replace(/(?<!\w)\*(.+?)\*(?!\w)/g, "<em>$1</em>");

    // Numbered lists
    t = t.replace(
      /(^|\n)(\d+)\. (.+?)(?=\n|$)/g,
      (_: string, pre: string, _n: string, item: string) =>
        `${pre}<li>${item}</li>`,
    );

    // Bulleted lists
    t = t.replace(
      /(^|\n)- (.+?)(?=\n|$)/g,
      (_: string, pre: string, item: string) => `${pre}<li>${item}</li>`,
    );

    // Wrap consecutive <li> in <ul>
    t = t.replace(/((?:<li>.*?<\/li>\n?)+)/g, "<ul>$1</ul>");

    // Paragraphs: double newlines
    t = t.replace(/\n\n/g, "</p><p>");

    // Single newlines (not inside code blocks)
    t = t.replace(/\n/g, "<br>");

    // Wrap in paragraph
    if (!t.startsWith("<")) {
      t = `<p>${t}</p>`;
    }

    return t;
  }

  private esc(s: string): string {
    return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }
}
