import { baseStyles, sharedKeyframes, scrollbarStyles } from "../theme";

const iconCopy = `<svg width="13" height="13" viewBox="0 0 16 16" fill="none"><rect x="5" y="5" width="9" height="9" rx="1.5" stroke="currentColor" stroke-width="1.5"/><path d="M11 5V3.5A1.5 1.5 0 0 0 9.5 2H3.5A1.5 1.5 0 0 0 2 3.5v6A1.5 1.5 0 0 0 3.5 11H5" stroke="currentColor" stroke-width="1.5"/></svg>`;
const iconCheck = `<svg width="13" height="13" viewBox="0 0 16 16" fill="none"><path d="M4 8.5L6.5 11L12 5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>`;

interface CodeTabData {
  name: string;
  language?: string;
  code: string;
}

export class HairaCodeBlock extends HTMLElement {
  private codeText = "";
  private tabs: CodeTabData[] = [];
  private activeTab = 0;

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
        .copy-btn:hover { color: var(--haira-accent); background: var(--haira-accent-dim); }
        .copy-btn.copied { color: var(--haira-success); }
        .tab-bar {
          display: flex;
          gap: 0;
          border-bottom: 1px solid var(--haira-border);
          background: var(--haira-bg);
          overflow-x: auto;
          scrollbar-width: none;
        }
        .tab-bar::-webkit-scrollbar { display: none; }
        .tab {
          padding: 0.4rem 0.75rem;
          font-size: 0.72rem;
          font-family: var(--haira-font);
          color: var(--haira-muted);
          background: none;
          border: none;
          border-bottom: 2px solid transparent;
          cursor: pointer;
          white-space: nowrap;
          transition: color 0.15s, border-color 0.15s;
          flex-shrink: 0;
        }
        .tab:hover { color: var(--haira-text); }
        .tab.active {
          color: var(--haira-accent);
          border-bottom-color: var(--haira-accent);
          font-weight: 600;
        }
        .code-scroll {
          max-height: 480px;
          overflow: auto;
          ${scrollbarStyles}
        }
        pre {
          margin: 0;
          padding: 0.75rem 1rem;
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
        <div class="tab-bar" id="tab-bar" style="display:none"></div>
        <div class="code-scroll" id="code-scroll">
          <pre><code id="code"></code></pre>
        </div>
      </div>
    `;

    this.shadowRoot!.getElementById("copy-btn")!.addEventListener(
      "click",
      () => {
        this.copyCode();
      },
    );
  }

  setProps(props: Record<string, unknown>) {
    try {
      const titleEl = this.shadowRoot!.getElementById("title")!;
      titleEl.textContent = (props.title as string) || "";

      const tabs = props.tabs as CodeTabData[] | undefined;

      if (tabs && tabs.length > 0) {
        this.tabs = tabs;
        this.activeTab = 0;
        this.renderTabBar();
        this.loadTab(0);
      } else {
        this.tabs = [];
        const langEl = this.shadowRoot!.getElementById("lang")!;
        langEl.textContent = (props.language as string) || "";

        const codeEl = this.shadowRoot!.getElementById("code")!;
        this.codeText = (props.code as string) || "";
        codeEl.textContent = this.codeText;
      }
    } catch {
      // Graceful fallback
    }
  }

  private renderTabBar() {
    const tabBar = this.shadowRoot!.getElementById("tab-bar")!;
    tabBar.style.display = "flex";
    tabBar.innerHTML = "";

    for (let i = 0; i < this.tabs.length; i++) {
      const tab = this.tabs[i];
      const btn = document.createElement("button");
      btn.className = `tab${i === this.activeTab ? " active" : ""}`;
      btn.textContent = tab.name;
      btn.addEventListener("click", () => this.loadTab(i));
      tabBar.appendChild(btn);
    }
  }

  private loadTab(index: number) {
    this.activeTab = index;
    const tab = this.tabs[index];

    // Update active class
    const tabBar = this.shadowRoot!.getElementById("tab-bar")!;
    tabBar.querySelectorAll(".tab").forEach((btn, i) => {
      btn.classList.toggle("active", i === index);
    });

    // Update language badge
    const langEl = this.shadowRoot!.getElementById("lang")!;
    langEl.textContent = tab.language || "";

    // Update code
    this.codeText = tab.code || "";
    const codeEl = this.shadowRoot!.getElementById("code")!;
    codeEl.textContent = this.codeText;

    // Scroll to top
    this.shadowRoot!.getElementById("code-scroll")!.scrollTop = 0;
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
