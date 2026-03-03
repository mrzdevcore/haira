import { LitElement, html, css, nothing } from "lit";
import { customElement, property } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles, hljsStyles } from "../core/styles";
import { iconStrings } from "../core/icons";
import hljs from "highlight.js/lib/core";
import sql from "highlight.js/lib/languages/sql";
import javascript from "highlight.js/lib/languages/javascript";
import typescript from "highlight.js/lib/languages/typescript";
import python from "highlight.js/lib/languages/python";
import go from "highlight.js/lib/languages/go";
import json from "highlight.js/lib/languages/json";
import yaml from "highlight.js/lib/languages/yaml";
import bash from "highlight.js/lib/languages/bash";
import xml from "highlight.js/lib/languages/xml";
import rust from "highlight.js/lib/languages/rust";

if (!hljs.getLanguage("sql")) {
  hljs.registerLanguage("sql", sql);
  hljs.registerLanguage("javascript", javascript);
  hljs.registerLanguage("js", javascript);
  hljs.registerLanguage("typescript", typescript);
  hljs.registerLanguage("ts", typescript);
  hljs.registerLanguage("python", python);
  hljs.registerLanguage("py", python);
  hljs.registerLanguage("go", go);
  hljs.registerLanguage("json", json);
  hljs.registerLanguage("yaml", yaml);
  hljs.registerLanguage("yml", yaml);
  hljs.registerLanguage("bash", bash);
  hljs.registerLanguage("sh", bash);
  hljs.registerLanguage("shell", bash);
  hljs.registerLanguage("xml", xml);
  hljs.registerLanguage("html", xml);
  hljs.registerLanguage("rust", rust);
}

interface MarkdownProps {
  content: string;
  title?: string;
  _restored?: boolean;
}

@customElement("haira-ui-markdown")
export class HairaMarkdown extends LitElement {
  static styles = [
    baseStyles,
    hljsStyles,
    css`
      :host {
        display: block;
      }
      .wrapper {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        overflow: hidden;
      }
      .header {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.55rem 0.85rem;
        background: var(--haira-bg-elevated);
        border-bottom: 1px solid var(--haira-border);
        font-size: 0.82rem;
        font-weight: 700;
        color: var(--haira-text);
      }
      .content {
        padding: 0.75rem 0.85rem;
        font-size: 0.86rem;
        line-height: 1.6;
        color: var(--haira-text);
      }

      /* Markdown styles */
      .md p { margin: 0.4em 0; }
      .md p:first-child { margin-top: 0; }
      .md h1, .md h2, .md h3, .md h4 {
        margin: 1em 0 0.4em;
        font-weight: 700;
        color: var(--haira-text);
        line-height: 1.3;
      }
      .md h1 { font-size: 1.3em; }
      .md h2 { font-size: 1.15em; }
      .md h3 { font-size: 1.05em; }
      .md h4 { font-size: 0.95em; }
      .md strong { font-weight: 700; color: var(--haira-text); }
      .md em { font-style: italic; }
      .md a { color: var(--haira-accent); text-decoration: none; }
      .md a:hover { text-decoration: underline; }

      .md code {
        font-family: Consolas, Menlo, Monaco, "Courier New", monospace;
        font-size: 0.85em;
        background: var(--haira-bg-elevated);
        padding: 0.15em 0.4em;
        border-radius: 4px;
        color: var(--haira-text);
      }

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
      .copy-btn:hover { background: var(--haira-bg-elevated); color: var(--haira-text); }
      .copy-btn.copied { color: var(--haira-success); }
      .code-block pre {
        margin: 0;
        padding: 0.7rem 0.85rem;
        overflow-x: auto;
        font-family: Consolas, Menlo, Monaco, "Courier New", monospace;
        font-size: 0.82rem;
        line-height: 1.55;
        color: var(--hljs-base);
      }
      .code-block pre code {
        font-family: inherit;
        background: transparent;
        padding: 0;
        color: inherit;
      }

      .md blockquote {
        margin: 0.5em 0;
        padding: 0.4em 0.85em;
        border-left: 3px solid var(--haira-accent);
        background: var(--haira-accent-dim);
        border-radius: 0 var(--haira-radius-sm) var(--haira-radius-sm) 0;
        color: var(--haira-text-dim);
        font-style: italic;
      }
      .md ul, .md ol { margin: 0.4em 0; padding-left: 1.5em; }
      .md li { margin: 0.2em 0; }

      .md table {
        width: 100%;
        border-collapse: collapse;
        margin: 0.6em 0;
        font-size: 0.84rem;
      }
      .md th, .md td {
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
      .md tr:nth-child(even) td { background: var(--haira-bg-card); }

      .md hr {
        border: none;
        border-top: 1px solid var(--haira-border);
        margin: 0.8em 0;
      }
    `,
  ];

  @property({ type: String }) content: string = "";
  @property({ type: String }) title: string = "";

  public setProps(props: MarkdownProps): void {
    this.content = props.content || "";
    this.title = props.title || "";
  }

  private _escapeHtml(str: string): string {
    return str
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  private _renderMarkdown(src: string): string {
    const codeBlocks: string[] = [];
    let text = src.replace(
      /```(\w*)\n([\s\S]*?)```/g,
      (_match, lang: string, code: string) => {
        const idx = codeBlocks.length;
        const trimmed = code.replace(/\n$/, "");
        const language = lang.toLowerCase().trim();
        let highlighted: string;
        try {
          if (language && hljs.getLanguage(language)) {
            highlighted = hljs.highlight(trimmed, { language }).value;
          } else {
            const result = hljs.highlightAuto(trimmed);
            highlighted = result.relevance > 5 ? result.value : this._escapeHtml(trimmed);
          }
        } catch {
          highlighted = this._escapeHtml(trimmed);
        }
        const header =
          `<div class="code-header">` +
          `<span class="code-lang">${this._escapeHtml(lang) || "text"}</span>` +
          `<button class="copy-btn" data-code="${btoa(encodeURIComponent(code))}" onclick="this.getRootNode().host._handleCopyClick(this)">` +
          `${iconStrings.copy} Copy</button></div>`;
        codeBlocks.push(
          `<div class="code-block">${header}<pre><code class="language-${this._escapeHtml(language)}">${highlighted}</code></pre></div>`
        );
        return `\x00CODE${idx}\x00`;
      }
    );

    // Tables
    text = text.replace(
      /(?:^|\n)(\|.+\|\n\|[-| :]+\|\n(?:\|.+\|\n?)+)/g,
      (_match, block: string) => {
        const lines = block.trim().split("\n");
        if (lines.length < 2) return block;
        const headerCells = lines[0].split("|").filter((c) => c.trim() !== "").map((c) => `<th>${this._escapeHtml(c.trim())}</th>`).join("");
        const bodyRows = lines.slice(2).map((row) => {
          const cells = row.split("|").filter((c) => c.trim() !== "").map((c) => `<td>${this._escapeHtml(c.trim())}</td>`).join("");
          return `<tr>${cells}</tr>`;
        }).join("");
        return `<table><thead><tr>${headerCells}</tr></thead><tbody>${bodyRows}</tbody></table>`;
      }
    );

    // Blockquotes
    text = text.replace(/(?:^|\n)(?:> (.+?)(?:\n|$))+/g, (match) => {
      const content = match.replace(/(?:^|\n)> /g, "\n").trim();
      return `<blockquote>${content}</blockquote>`;
    });

    // Headings
    text = text.replace(/^#### (.+)$/gm, "<h4>$1</h4>");
    text = text.replace(/^### (.+)$/gm, "<h3>$1</h3>");
    text = text.replace(/^## (.+)$/gm, "<h2>$1</h2>");
    text = text.replace(/^# (.+)$/gm, "<h1>$1</h1>");

    // Ordered lists
    text = text.replace(/(?:^|\n)((?:\d+\. .+(?:\n|$))+)/g, (_match, block: string) => {
      const items = block.trim().split("\n").map((li) => `<li>${li.replace(/^\d+\.\s+/, "")}</li>`).join("");
      return `<ol>${items}</ol>`;
    });

    // Unordered lists
    text = text.replace(/(?:^|\n)((?:[-*] .+(?:\n|$))+)/g, (_match, block: string) => {
      const items = block.trim().split("\n").map((li) => `<li>${li.replace(/^[-*]\s+/, "")}</li>`).join("");
      return `<ul>${items}</ul>`;
    });

    // Inline formatting
    text = text.replace(/\*\*\*(.+?)\*\*\*/g, "<strong><em>$1</em></strong>");
    text = text.replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>");
    text = text.replace(/\*(.+?)\*/g, "<em>$1</em>");
    text = text.replace(/`([^`]+)`/g, "<code>$1</code>");
    text = text.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener">$1</a>');

    // Paragraphs
    text = text.split(/\n{2,}/).map((block) => {
      const trimmed = block.trim();
      if (!trimmed) return "";
      if (/^<(h[1-4]|ul|ol|blockquote|table|div)/.test(trimmed) || /^\x00CODE/.test(trimmed)) {
        return trimmed;
      }
      return `<p>${trimmed.replace(/\n/g, "<br>")}</p>`;
    }).join("");

    // Restore code blocks
    text = text.replace(/\x00CODE(\d+)\x00/g, (_m, idx: string) => codeBlocks[parseInt(idx, 10)]);

    return text;
  }

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

  render() {
    const rendered = this._renderMarkdown(this.content);
    return html`
      <div class="wrapper">
        ${this.title ? html`<div class="header">${this.title}</div>` : nothing}
        <div class="content md">${unsafeHTML(rendered)}</div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-ui-markdown": HairaMarkdown;
  }
}
