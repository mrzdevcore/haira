import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles, keyframes, animateInStyles } from "../core/styles";
import { iconStrings } from "../core/icons";

@customElement("haira-message")
export class HairaMessage extends LitElement {
  static styles = [
    baseStyles,
    css`
      :host {
        display: block;
        animation: fadeSlideUp 0.2s ease-out;
      }

      /* ---- User message ---- */
      .msg-user {
        display: flex;
        justify-content: flex-end;
        padding: 0.35rem 0;
      }
      .msg-user .bubble {
        background: var(--haira-bg-elevated);
        color: var(--haira-text);
        padding: 0.65rem 1rem;
        border-radius: var(--haira-radius) var(--haira-radius) 4px
          var(--haira-radius);
        max-width: 75%;
        font-size: 0.88rem;
        line-height: 1.55;
        word-break: break-word;
        white-space: pre-wrap;
      }
      .file-badge {
        display: inline-flex;
        align-items: center;
        gap: 0.3rem;
        margin-top: 0.35rem;
        padding: 0.2rem 0.5rem;
        font-size: 0.72rem;
        font-family: var(--haira-mono);
        background: rgba(232, 163, 23, 0.08);
        color: var(--haira-accent);
        border-radius: 4px;
      }
      .file-badge svg {
        flex-shrink: 0;
      }

      /* ---- Assistant message ---- */
      .msg-assistant {
        display: flex;
        gap: 0.65rem;
        padding: 0.35rem 0;
        align-items: flex-start;
      }
      .avatar {
        width: 28px;
        height: 28px;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
        font-size: 0.82rem;
        font-weight: 700;
        color: var(--haira-bg);
        background: var(--haira-accent);
        overflow: hidden;
      }
      .avatar img {
        width: 100%;
        height: 100%;
        object-fit: cover;
      }
      .body {
        flex: 1;
        min-width: 0;
        font-size: 0.88rem;
        line-height: 1.65;
        color: var(--haira-text);
      }

      /* ---- Markdown rendered content ---- */
      .md p {
        margin: 0.4em 0;
      }
      .md p:first-child {
        margin-top: 0;
      }
      .md h1,
      .md h2,
      .md h3,
      .md h4 {
        margin: 1em 0 0.4em;
        font-weight: 700;
        color: var(--haira-text);
        line-height: 1.3;
      }
      .md h1 {
        font-size: 1.3em;
      }
      .md h2 {
        font-size: 1.15em;
      }
      .md h3 {
        font-size: 1.05em;
      }
      .md h4 {
        font-size: 0.95em;
      }
      .md strong {
        font-weight: 700;
        color: var(--haira-text);
      }
      .md em {
        font-style: italic;
      }
      .md a {
        color: var(--haira-accent);
        text-decoration: none;
      }
      .md a:hover {
        text-decoration: underline;
      }

      /* Inline code */
      .md code {
        font-family: var(--haira-mono);
        font-size: 0.85em;
        background: var(--haira-bg-elevated);
        padding: 0.15em 0.4em;
        border-radius: 4px;
        color: var(--haira-accent-light);
      }

      /* Code blocks */
      .code-block {
        position: relative;
        margin: 0.6em 0;
        background: var(--haira-bg-input);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius-sm);
        overflow: hidden;
      }
      .code-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.3rem 0.6rem;
        background: var(--haira-bg-card);
        border-bottom: 1px solid var(--haira-border);
      }
      .code-lang {
        font-size: 0.68rem;
        font-family: var(--haira-mono);
        color: var(--haira-muted);
        text-transform: lowercase;
      }
      .copy-btn {
        display: flex;
        align-items: center;
        gap: 0.25rem;
        background: none;
        border: none;
        color: var(--haira-muted);
        cursor: pointer;
        font-size: 0.68rem;
        font-family: var(--haira-font);
        padding: 0.15rem 0.35rem;
        border-radius: 4px;
        transition: all 0.15s;
      }
      .copy-btn:hover {
        background: var(--haira-bg-elevated);
        color: var(--haira-text);
      }
      .copy-btn.copied {
        color: var(--haira-success);
      }
      .code-block pre {
        margin: 0;
        padding: 0.7rem 0.85rem;
        overflow-x: auto;
        font-family: var(--haira-mono);
        font-size: 0.82rem;
        line-height: 1.55;
        color: var(--haira-text);
      }
      .code-block pre::-webkit-scrollbar {
        height: 4px;
      }
      .code-block pre::-webkit-scrollbar-thumb {
        background: var(--haira-muted);
        border-radius: 2px;
      }

      /* Blockquotes */
      .md blockquote {
        margin: 0.5em 0;
        padding: 0.4em 0.85em;
        border-left: 3px solid var(--haira-accent);
        background: var(--haira-accent-dim);
        border-radius: 0 var(--haira-radius-sm) var(--haira-radius-sm) 0;
        color: var(--haira-text-dim);
        font-style: italic;
      }

      /* Lists */
      .md ul,
      .md ol {
        margin: 0.4em 0;
        padding-left: 1.5em;
      }
      .md li {
        margin: 0.2em 0;
      }

      /* Tables */
      .md table {
        width: 100%;
        border-collapse: collapse;
        margin: 0.6em 0;
        font-size: 0.84rem;
      }
      .md th,
      .md td {
        padding: 0.45rem 0.65rem;
        border: 1px solid var(--haira-border);
        text-align: left;
      }
      .md th {
        background: var(--haira-bg-card);
        font-weight: 600;
        font-size: 0.78rem;
        color: var(--haira-text-dim);
        text-transform: uppercase;
        letter-spacing: 0.02em;
      }
      .md tr:nth-child(even) td {
        background: var(--haira-bg-card);
      }

      /* Cursor for streaming */
      .cursor {
        display: inline-block;
        width: 2px;
        height: 1em;
        background: var(--haira-accent);
        margin-left: 2px;
        vertical-align: text-bottom;
        animation: blink 1s step-end infinite;
      }
    `,
  ];

  @property({ type: String }) role: string = "assistant";
  @property({ type: String }) content: string = "";
  @property({ type: String }) file: string = "";
  @property({ type: String }) avatar: string = "";

  /** External API: update content and trigger re-render */
  public updateContent(text: string): void {
    this.content = text;
  }

  private _escapeHtml(str: string): string {
    return str
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  /**
   * Lightweight Markdown renderer.
   * Supports: code blocks (``` with language + copy), tables, inline code,
   * headings, bold, italic, links, numbered/bulleted lists, blockquotes, paragraphs.
   */
  private _renderMarkdown(src: string): string {
    // Collect fenced code blocks into placeholders
    const codeBlocks: string[] = [];
    let text = src.replace(
      /```(\w*)\n([\s\S]*?)```/g,
      (_match, lang: string, code: string) => {
        const idx = codeBlocks.length;
        const escaped = this._escapeHtml(code.replace(/\n$/, ""));
        const header =
          `<div class="code-header">` +
          `<span class="code-lang">${this._escapeHtml(lang) || "text"}</span>` +
          `<button class="copy-btn" data-code="${btoa(encodeURIComponent(code))}" onclick="this.getRootNode().host._handleCopyClick(this)">` +
          `${iconStrings.copy} Copy</button></div>`;
        codeBlocks.push(
          `<div class="code-block">${header}<pre>${escaped}</pre></div>`
        );
        return `\x00CODE${idx}\x00`;
      }
    );

    // Tables: detect lines starting with |
    text = text.replace(
      /(?:^|\n)(\|.+\|\n\|[-| :]+\|\n(?:\|.+\|\n?)+)/g,
      (_match, block: string) => {
        const lines = block.trim().split("\n");
        if (lines.length < 2) return block;

        const headerCells = lines[0]
          .split("|")
          .filter((c) => c.trim() !== "")
          .map((c) => `<th>${this._escapeHtml(c.trim())}</th>`)
          .join("");

        const bodyRows = lines
          .slice(2)
          .map((row) => {
            const cells = row
              .split("|")
              .filter((c) => c.trim() !== "")
              .map((c) => `<td>${this._escapeHtml(c.trim())}</td>`)
              .join("");
            return `<tr>${cells}</tr>`;
          })
          .join("");

        return `<table><thead><tr>${headerCells}</tr></thead><tbody>${bodyRows}</tbody></table>`;
      }
    );

    // Blockquotes
    text = text.replace(/(?:^|\n)(?:> (.+?)(?:\n|$))+/g, (match) => {
      const content = match
        .replace(/(?:^|\n)> /g, "\n")
        .trim();
      return `<blockquote>${content}</blockquote>`;
    });

    // Headings
    text = text.replace(/^#### (.+)$/gm, "<h4>$1</h4>");
    text = text.replace(/^### (.+)$/gm, "<h3>$1</h3>");
    text = text.replace(/^## (.+)$/gm, "<h2>$1</h2>");
    text = text.replace(/^# (.+)$/gm, "<h1>$1</h1>");

    // Ordered lists
    text = text.replace(
      /(?:^|\n)((?:\d+\. .+(?:\n|$))+)/g,
      (_match, block: string) => {
        const items = block
          .trim()
          .split("\n")
          .map((li) => `<li>${li.replace(/^\d+\.\s+/, "")}</li>`)
          .join("");
        return `<ol>${items}</ol>`;
      }
    );

    // Unordered lists
    text = text.replace(
      /(?:^|\n)((?:[-*] .+(?:\n|$))+)/g,
      (_match, block: string) => {
        const items = block
          .trim()
          .split("\n")
          .map((li) => `<li>${li.replace(/^[-*]\s+/, "")}</li>`)
          .join("");
        return `<ul>${items}</ul>`;
      }
    );

    // Inline formatting (order matters)
    // Bold + italic
    text = text.replace(
      /\*\*\*(.+?)\*\*\*/g,
      "<strong><em>$1</em></strong>"
    );
    // Bold
    text = text.replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>");
    // Italic
    text = text.replace(/\*(.+?)\*/g, "<em>$1</em>");
    // Inline code
    text = text.replace(/`([^`]+)`/g, "<code>$1</code>");
    // Links
    text = text.replace(
      /\[([^\]]+)\]\(([^)]+)\)/g,
      '<a href="$2" target="_blank" rel="noopener">$1</a>'
    );

    // Paragraphs: split on double newlines (but not inside block-level elements)
    text = text
      .split(/\n{2,}/)
      .map((block) => {
        const trimmed = block.trim();
        if (!trimmed) return "";
        // Don't wrap block-level elements
        if (
          /^<(h[1-4]|ul|ol|blockquote|table|div)/.test(trimmed) ||
          /^\x00CODE/.test(trimmed)
        ) {
          return trimmed;
        }
        return `<p>${trimmed.replace(/\n/g, "<br>")}</p>`;
      })
      .join("");

    // Restore code blocks
    text = text.replace(/\x00CODE(\d+)\x00/g, (_m, idx: string) => {
      return codeBlocks[parseInt(idx, 10)];
    });

    return text;
  }

  /** Click handler for copy buttons inside rendered markdown */
  public _handleCopyClick(btn: HTMLButtonElement): void {
    const encoded = btn.getAttribute("data-code") || "";
    const code = decodeURIComponent(atob(encoded));
    navigator.clipboard.writeText(code).then(() => {
      btn.classList.add("copied");
      btn.innerHTML = `${iconStrings.copyDone} Copied!`;
      setTimeout(() => {
        btn.classList.remove("copied");
        btn.innerHTML = `${iconStrings.copy} Copy`;
      }, 1500);
    });
  }

  private _renderAvatar() {
    const av = this.avatar;
    if (av && (av.startsWith("http://") || av.startsWith("https://"))) {
      return html`<div class="avatar">
        <img src=${av} alt="avatar" />
      </div>`;
    }
    const char = av ? av.charAt(0).toUpperCase() : "A";
    return html`<div class="avatar">${char}</div>`;
  }

  render() {
    if (this.role === "user") {
      return html`
        <div class="msg-user">
          <div class="bubble">${this.content}${this.file
              ? html`<div class="file-badge">
                  ${unsafeHTML(iconStrings.file)} ${this.file}
                </div>`
              : nothing}</div>
        </div>
      `;
    }

    // Assistant message
    const rendered = this._renderMarkdown(this.content);
    return html`
      <div class="msg-assistant">
        ${this._renderAvatar()}
        <div class="body md">${unsafeHTML(rendered)}</div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-message": HairaMessage;
  }
}
