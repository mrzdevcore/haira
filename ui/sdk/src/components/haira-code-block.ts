import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles, keyframes, animateInStyles } from "../core/styles";
import { iconStrings } from "../core/icons";
import type { CodeBlockProps, CodeTabData } from "../core/types";

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
        font-family: var(--haira-mono);
        font-size: 0.82rem;
        line-height: 1.55;
        color: var(--haira-text);
        background: var(--haira-bg-input);
        white-space: pre;
        overflow-x: auto;
      }
    `,
  ];

  @property({ type: String }) title: string = "";
  @property({ type: String }) language: string = "";
  @property({ type: String }) code: string = "";
  @property({ type: Array }) tabs: CodeTabData[] = [];

  @state() private _activeTab: number = 0;
  @state() private _copied: boolean = false;

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

  private _escapeHtml(str: string): string {
    return str
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

  render() {
    const lang = this._currentLanguage();
    const code = this._currentCode();

    return html`
      <div class="code-card">
        <div class="code-header">
          <div class="code-header-left">
            ${this.title
              ? html`<span class="code-title">${this.title}</span>`
              : nothing}
            ${lang ? html`<span class="code-lang">${lang}</span>` : nothing}
          </div>
          <button
            class="copy-btn ${this._copied ? "copied" : ""}"
            @click=${this._copy}
          >
            ${this._copied
              ? html`${unsafeHTML(iconStrings.copyDone)} Copied`
              : html`${unsafeHTML(iconStrings.copy)} Copy`}
          </button>
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
          <pre>${this._escapeHtml(code)}</pre>
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-ui-code-block": HairaCodeBlock;
  }
}
