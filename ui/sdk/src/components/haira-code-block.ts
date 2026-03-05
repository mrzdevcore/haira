import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles, animateInStyles, hljsStyles, popoverStyles, scrollbarStyles } from "../core/styles";
import { iconStrings } from "../core/icons";
import type { CodeBlockProps, CodeTabData } from "../core/types";

// highlight.js — core + cherry-picked languages only
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
import markdown from "highlight.js/lib/languages/markdown";
import rust from "highlight.js/lib/languages/rust";

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
hljs.registerLanguage("markdown", markdown);
hljs.registerLanguage("md", markdown);
hljs.registerLanguage("rust", rust);

@customElement("haira-ui-code-block")
export class HairaCodeBlock extends LitElement {
  static styles = [
    baseStyles,
    animateInStyles,
    css`
      .code-card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        overflow: hidden;
      }

      .code-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.45rem 0.85rem;
        border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg);
      }
      .code-header-left {
        display: flex;
        align-items: center;
        gap: 0.55rem;
      }
      .code-title {
        font-size: 0.85rem;
        font-weight: 700;
        color: var(--haira-text);
      }
      .code-lang {
        font-size: 0.68rem;
        font-family: var(--haira-mono);
        color: var(--haira-muted);
        text-transform: lowercase;
        padding: 0.1rem 0.4rem;
        background: var(--haira-bg-elevated);
        border-radius: 3px;
      }

      .copy-btn {
        display: flex;
        align-items: center;
        gap: 0.25rem;
        background: none;
        border: none;
        color: var(--haira-muted);
        cursor: pointer;
        font-size: 0.7rem;
        font-family: var(--haira-font);
        padding: 0.2rem 0.45rem;
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

      /* Tab bar */
      .tab-bar {
        display: flex;
        gap: 0;
        border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg);
        overflow-x: auto;
      }
      .tab-bar::-webkit-scrollbar {
        height: 0;
      }
      .tab {
        padding: 0.45rem 0.8rem;
        font-size: 0.76rem;
        font-weight: 600;
        color: var(--haira-muted);
        cursor: pointer;
        border-bottom: 2px solid transparent;
        transition: all 0.12s;
        white-space: nowrap;
        background: none;
        border-top: none;
        border-left: none;
        border-right: none;
        font-family: var(--haira-font);
      }
      .tab:hover {
        color: var(--haira-text-dim);
        background: var(--haira-bg-card);
      }
      .tab.active {
        color: var(--haira-accent);
        border-bottom-color: var(--haira-accent);
      }

      /* Code content */
      .code-content {
        max-height: 480px;
        overflow: auto;
      }
      .code-content::-webkit-scrollbar {
        width: 5px;
        height: 5px;
      }
      .code-content::-webkit-scrollbar-thumb {
        background: var(--haira-muted);
        border-radius: 3px;
      }
      pre {
        margin: 0;
        padding: 0.75rem 0.95rem;
        font-family: Consolas, Menlo, Monaco, "Courier New", monospace;
        font-size: 0.82rem;
        line-height: 1.55;
        background: var(--haira-bg-input);
        white-space: pre;
        overflow-x: auto;
      }
      pre code {
        font-family: inherit;
        font-size: inherit;
        background: transparent;
        padding: 0;
        color: var(--hljs-base);
      }

    `,
    hljsStyles,
    popoverStyles,
    scrollbarStyles,
  ];

  @property({ type: String }) title: string = "";
  @property({ type: String }) language: string = "";
  @property({ type: String }) code: string = "";
  @property({ type: Array }) tabs: CodeTabData[] = [];

  @state() private _activeTab: number = 0;
  @state() private _copied: boolean = false;
  @state() private _expanded: boolean = false;

  /** Set all props at once */
  public setProps(props: CodeBlockProps): void {
    this.title = props.title || "";
    this.language = props.language || "";
    this.code = props.code || "";
    this.tabs = props.tabs || [];
    this._activeTab = 0;
    this._copied = false;
  }

  private _currentCode(): string {
    if (this.tabs && this.tabs.length > 0) {
      return this.tabs[this._activeTab]?.code || "";
    }
    return this.code;
  }

  private _currentLanguage(): string {
    if (this.tabs && this.tabs.length > 0) {
      return this.tabs[this._activeTab]?.language || "";
    }
    return this.language;
  }

  private _highlight(code: string, lang: string): string {
    const language = lang.toLowerCase().trim();
    try {
      if (language && hljs.getLanguage(language)) {
        return hljs.highlight(code, { language }).value;
      }
      // auto-detect as fallback (capped relevance to avoid false positives)
      const result = hljs.highlightAuto(code);
      if (result.relevance > 5) return result.value;
    } catch {
      // fall through to plain escape
    }
    return code
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;");
  }

  private _onTabClick(index: number) {
    this._activeTab = index;
    this._copied = false;
  }

  private async _copy() {
    try {
      await navigator.clipboard.writeText(this._currentCode());
      this._copied = true;
      setTimeout(() => {
        this._copied = false;
      }, 1500);
    } catch {
      // Clipboard write failed
    }
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    document.removeEventListener("keydown", this._onEscKey);
  }

  private _onExpand() {
    this._expanded = true;
    document.addEventListener("keydown", this._onEscKey);
  }

  private _onCollapse() {
    this._expanded = false;
    document.removeEventListener("keydown", this._onEscKey);
  }

  private _onEscKey = (e: KeyboardEvent) => {
    if (e.key === "Escape") this._onCollapse();
  };

  private _onBackdropClick(e: MouseEvent) {
    if ((e.target as HTMLElement).classList.contains("popover-backdrop")) {
      this._onCollapse();
    }
  }

  render() {
    const lang = this._currentLanguage();
    const code = this._currentCode();
    const highlighted = this._highlight(code, lang);

    return html`
      <div class="code-card">
        <div class="code-header">
          <div class="code-header-left">
            ${this.title
              ? html`<span class="code-title">${this.title}</span>`
              : nothing}
            ${lang ? html`<span class="code-lang">${lang}</span>` : nothing}
          </div>
          <div style="display:flex;align-items:center;gap:0.3rem">
            <button
              class="copy-btn ${this._copied ? "copied" : ""}"
              @click=${this._copy}
            >
              ${this._copied
                ? html`${unsafeHTML(iconStrings.copyDone)} Copied`
                : html`${unsafeHTML(iconStrings.copy)} Copy`}
            </button>
            <button class="expand-btn" title="Expand" @click=${this._onExpand}>
              ${unsafeHTML(iconStrings.expand)}
            </button>
          </div>
        </div>

        ${this.tabs && this.tabs.length > 0
          ? html`
              <div class="tab-bar">
                ${this.tabs.map(
                  (tab, i) => html`
                    <button
                      class="tab ${this._activeTab === i ? "active" : ""}"
                      @click=${() => this._onTabClick(i)}
                    >
                      ${tab.name}
                    </button>
                  `
                )}
              </div>
            `
          : nothing}

        <div class="code-content">
          <pre><code class="language-${lang}">${unsafeHTML(highlighted)}</code></pre>
        </div>
      </div>
      ${this._expanded
        ? html`
          <div class="popover-backdrop" @click=${this._onBackdropClick}>
            <div class="popover-card">
              <div class="popover-header">
                <span class="popover-title">${this.title || "Code"}${lang ? html` <span style="font-size:0.72rem;font-weight:400;color:var(--haira-muted);margin-left:0.5rem">${lang}</span>` : nothing}</span>
                <button class="popover-close" @click=${this._onCollapse}>
                  ${unsafeHTML(iconStrings.x)}
                </button>
              </div>
              ${this.tabs && this.tabs.length > 0
                ? html`
                  <div class="tab-bar" style="flex-shrink:0;position:relative;z-index:3">
                    ${this.tabs.map(
                      (tab, i) => html`
                        <button
                          class="tab ${this._activeTab === i ? "active" : ""}"
                          @click=${() => this._onTabClick(i)}
                        >
                          ${tab.name}
                        </button>
                      `
                    )}
                  </div>
                `
                : nothing}
              <div class="popover-body">
                <pre style="max-height:none"><code class="language-${lang}">${unsafeHTML(highlighted)}</code></pre>
              </div>
            </div>
          </div>
        `
        : nothing}
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-ui-code-block": HairaCodeBlock;
  }
}
