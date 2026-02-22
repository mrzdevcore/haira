import { BaseComponent, animateInCSS, cardCSS, scrollbarCSS, icons } from "../core";
import type { CodeBlockProps, CodeTabData } from "../core/types";

export class HairaCodeBlock extends BaseComponent<CodeBlockProps> {
  private codeText = "";
  private tabs: CodeTabData[] = [];
  private activeTab = 0;

  protected render() {
    return `
      <div class="card">
        <div class="header">
          <div style="display:flex;align-items:center;gap:0.5rem">
            <span class="title" id="title"></span>
            <span class="lang" id="lang"></span>
          </div>
          <div class="actions">
            <button class="copy-btn" id="copy-btn">${icons.copy} Copy</button>
          </div>
        </div>
        <div class="tab-bar" id="tab-bar" style="display:none"></div>
        <div class="code-scroll" id="code-scroll">
          <pre><code id="code"></code></pre>
        </div>
      </div>`;
  }

  protected styles() {
    return `
      ${animateInCSS}
      .card { ${cardCSS} }
      .header {
        display: flex; align-items: center; justify-content: space-between;
        padding: 0.45rem 0.75rem; border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg);
      }
      .title { font-size: 0.78rem; font-weight: 600; color: var(--haira-text); }
      .lang { font-size: 0.68rem; color: var(--haira-muted); font-family: var(--haira-mono); }
      .actions { display: flex; align-items: center; gap: 0.5rem; }
      .copy-btn {
        background: none; border: none; color: var(--haira-muted);
        cursor: pointer; display: flex; align-items: center; gap: 0.3rem;
        font-size: 0.7rem; font-family: var(--haira-font);
        padding: 0.2rem 0.4rem; border-radius: 4px; transition: all 0.15s;
      }
      .copy-btn:hover { color: var(--haira-accent); background: var(--haira-accent-dim); }
      .copy-btn.copied { color: var(--haira-success); }
      .tab-bar {
        display: flex; gap: 0; border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg); overflow-x: auto; scrollbar-width: none;
      }
      .tab-bar::-webkit-scrollbar { display: none; }
      .tab {
        padding: 0.4rem 0.75rem; font-size: 0.72rem; font-family: var(--haira-font);
        color: var(--haira-muted); background: none; border: none;
        border-bottom: 2px solid transparent; cursor: pointer;
        white-space: nowrap; transition: color 0.15s, border-color 0.15s; flex-shrink: 0;
      }
      .tab:hover { color: var(--haira-text); }
      .tab.active { color: var(--haira-accent); border-bottom-color: var(--haira-accent); font-weight: 600; }
      .code-scroll { max-height: 480px; overflow: auto; ${scrollbarCSS} }
      pre { margin: 0; padding: 0.75rem 1rem; }
      code {
        font-family: var(--haira-mono); font-size: 0.78rem;
        color: var(--haira-text-dim); line-height: 1.6; white-space: pre;
      }`;
  }

  protected onMount() {
    this.$("copy-btn").addEventListener("click", () => this.copyCode());
  }

  protected onUpdate() {
    const { title, language, code, tabs } = this.props;

    this.$("title").textContent = title || "";

    if (tabs && tabs.length > 0) {
      this.tabs = tabs;
      this.activeTab = 0;
      this.renderTabBar();
      this.loadTab(0);
    } else {
      this.tabs = [];
      this.$("lang").textContent = language || "";
      this.codeText = code || "";
      this.$("code").textContent = this.codeText;
    }
  }

  private renderTabBar() {
    const tabBar = this.$("tab-bar");
    tabBar.style.display = "flex";
    tabBar.innerHTML = "";
    for (let i = 0; i < this.tabs.length; i++) {
      const btn = document.createElement("button");
      btn.className = `tab${i === this.activeTab ? " active" : ""}`;
      btn.textContent = this.tabs[i].name;
      btn.addEventListener("click", () => this.loadTab(i));
      tabBar.appendChild(btn);
    }
  }

  private loadTab(index: number) {
    this.activeTab = index;
    const tab = this.tabs[index];
    this.$("tab-bar").querySelectorAll(".tab").forEach((btn, i) => {
      btn.classList.toggle("active", i === index);
    });
    this.$("lang").textContent = tab.language || "";
    this.codeText = tab.code || "";
    this.$("code").textContent = this.codeText;
    this.$("code-scroll").scrollTop = 0;
  }

  private copyCode() {
    navigator.clipboard.writeText(this.codeText).then(() => {
      const btn = this.$("copy-btn");
      btn.innerHTML = `${icons.copyDone} Copied`;
      btn.classList.add("copied");
      setTimeout(() => {
        btn.innerHTML = `${icons.copy} Copy`;
        btn.classList.remove("copied");
      }, 2000);
    });
  }
}
