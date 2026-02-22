import { LitElement, html, css, nothing } from "lit";
import { customElement, state } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles, scrollbarStyles } from "../core/styles";
import { iconStrings, logoSvgStr } from "../core/icons";
import { applyTheme } from "../services/theme-manager";
import type { WorkflowMeta } from "../core/types";
import type { Route } from "../core/router";
import { parseRoute, navigate, currentRoute } from "../core/router";

@customElement("haira-app")
export class HairaApp extends LitElement {
  static styles = [
    baseStyles,
    scrollbarStyles,
    css`
      :host {
        display: block;
        height: 100vh;
        overflow: hidden;
        background: var(--haira-bg);
      }
      .shell {
        height: 100%;
        display: flex;
        overflow: hidden;
      }

      /* Sidebar navigation */
      .sidebar {
        width: 220px;
        flex-shrink: 0;
        display: flex;
        flex-direction: column;
        border-right: 1px solid var(--haira-border);
        background: var(--haira-bg);
        overflow: hidden;
      }
      .sidebar-brand {
        padding: 0.85rem 1rem;
        display: flex;
        align-items: center;
        gap: 0.5rem;
        border-bottom: 1px solid var(--haira-border);
        text-decoration: none;
        cursor: pointer;
      }
      .sidebar-brand:hover {
        background: var(--haira-bg-card);
      }
      .brand-icon {
        display: flex;
        align-items: center;
      }
      .brand-text {
        font-weight: 700;
        font-size: 0.95rem;
        color: var(--haira-text);
        letter-spacing: -0.01em;
      }
      .brand-text .ai {
        color: var(--haira-accent);
      }

      .sidebar-nav {
        flex: 1;
        padding: 0.5rem;
        display: flex;
        flex-direction: column;
        gap: 2px;
        overflow-y: auto;
      }
      .nav-section {
        margin-top: 0.75rem;
        margin-bottom: 0.25rem;
        padding: 0 0.5rem;
        font-size: 0.65rem;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.05em;
        color: var(--haira-muted);
      }
      .nav-section:first-child {
        margin-top: 0.25rem;
      }
      .nav-item {
        display: flex;
        align-items: center;
        gap: 0.55rem;
        padding: 0.5rem 0.65rem;
        border-radius: 6px;
        cursor: pointer;
        transition: all 0.12s;
        color: var(--haira-text-dim);
        font-size: 0.82rem;
        font-weight: 500;
        text-decoration: none;
      }
      .nav-item:hover {
        background: var(--haira-bg-card);
        color: var(--haira-text);
      }
      .nav-item.active {
        background: var(--haira-accent-dim);
        color: var(--haira-accent);
      }
      .nav-icon {
        display: flex;
        align-items: center;
        flex-shrink: 0;
        opacity: 0.7;
      }
      .nav-item.active .nav-icon {
        opacity: 1;
      }

      .sidebar-footer {
        padding: 0.5rem;
        border-top: 1px solid var(--haira-border);
      }

      /* Main content */
      .main {
        flex: 1;
        display: flex;
        flex-direction: column;
        overflow: hidden;
        min-width: 0;
      }
      .topbar {
        padding: 0.6rem 1.25rem;
        border-bottom: 1px solid var(--haira-border);
        display: flex;
        align-items: center;
        gap: 0.6rem;
        background: var(--haira-bg);
        flex-shrink: 0;
      }
      .topbar-title {
        font-size: 0.9rem;
        font-weight: 600;
        color: var(--haira-text);
      }
      .content {
        flex: 1;
        overflow: hidden;
        display: flex;
        flex-direction: column;
      }
      .content.scrollable {
        overflow-y: auto;
      }

      /* Mobile toggle */
      .mobile-toggle {
        display: none;
        position: fixed;
        bottom: 1rem;
        left: 1rem;
        z-index: 200;
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        color: var(--haira-muted);
        cursor: pointer;
        padding: 0.5rem;
        border-radius: 8px;
        box-shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
      }
      @media (max-width: 768px) {
        .sidebar {
          position: fixed;
          left: 0;
          top: 0;
          bottom: 0;
          z-index: 100;
          box-shadow: 4px 0 24px rgba(0, 0, 0, 0.3);
          transform: translateX(-100%);
          transition: transform 0.2s ease;
        }
        .sidebar.open {
          transform: translateX(0);
        }
        .mobile-toggle {
          display: flex;
        }
      }
    `,
  ];

  @state() private _route: Route = currentRoute();
  @state() private _sidebarOpen = false;

  private meta: WorkflowMeta | null = null;

  connectedCallback() {
    super.connectedCallback();

    const metaEl = document.getElementById("haira-meta");
    if (metaEl) {
      try {
        this.meta = JSON.parse(metaEl.textContent || "{}");
      } catch {
        this.meta = null;
      }
    }

    if (this.meta) {
      const name = this.meta.title || this.meta.name;
      if (this.meta.mode === "orchestrator") {
        document.title = "Haira Orchestrator";
      } else {
        document.title = name ? `${name} — Haira` : "Haira Console";
      }

      applyTheme(this, {
        theme: this.meta.theme,
        accent: this.meta.accent,
      });
    }

    // If mode is "form" or "chat", route directly to workbench
    if (this.meta?.mode === "chat" || this.meta?.mode === "form") {
      if (!window.location.hash || window.location.hash === "#/") {
        navigate({ page: "workbench", path: this.meta.path });
      }
    }

    this._route = currentRoute();
    window.addEventListener("hashchange", this._onHashChange);
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    window.removeEventListener("hashchange", this._onHashChange);
  }

  private _onHashChange = () => {
    this._route = currentRoute();
    this._sidebarOpen = false;
  };

  private _pageTitle(): string {
    switch (this._route.page) {
      case "home":
        return "Overview";
      case "workbench":
        return this.meta?.title || this.meta?.name || "Workbench";
      case "workflows":
        return "Workflows";
      case "agents":
        return "Agents";
      case "observe":
        return "Observability";
      case "settings":
        return "Settings";
      case "deployments":
        return "Deployments";
      default:
        return "Haira";
    }
  }

  private _isActive(page: string): boolean {
    return this._route.page === page;
  }

  private _navClick(page: Route["page"]) {
    return (e: Event) => {
      e.preventDefault();
      if (page === "home") navigate({ page: "home" });
      else if (page === "workflows") navigate({ page: "workflows" });
      else if (page === "agents") navigate({ page: "agents" });
      else if (page === "observe") navigate({ page: "observe" });
      else if (page === "settings") navigate({ page: "settings" });
    };
  }

  private _renderPage() {
    switch (this._route.page) {
      case "home":
        return html`<haira-page-home
          .meta=${this.meta}
        ></haira-page-home>`;
      case "workbench":
        return html`<haira-page-workbench
          .meta=${this.meta}
        ></haira-page-workbench>`;
      case "workflows":
        return html`<haira-page-workflows
          .meta=${this.meta}
        ></haira-page-workflows>`;
      case "agents":
        return html`<haira-page-agents
          .meta=${this.meta}
        ></haira-page-agents>`;
      case "observe":
        return html`<haira-page-observe></haira-page-observe>`;
      case "settings":
        return html`<haira-page-settings
          .meta=${this.meta}
        ></haira-page-settings>`;
      case "deployments":
        return html`<haira-page-deployments
          .meta=${this.meta}
        ></haira-page-deployments>`;
      default:
        return html`<haira-page-home
          .meta=${this.meta}
        ></haira-page-home>`;
    }
  }

  private _isScrollable(): boolean {
    const r = this._route;
    return r.page !== "workbench";
  }

  private _renderOrchestratorNav() {
    const deployments = this.meta?.deployments || [];
    return html`
      <div class="nav-section">General</div>
      <a
        class="nav-item ${this._isActive("home") ? "active" : ""}"
        href="#/"
        @click=${this._navClick("home")}
      >
        <span class="nav-icon">${unsafeHTML(iconStrings.home)}</span>
        Overview
      </a>
      <a
        class="nav-item ${this._isActive("deployments") ? "active" : ""}"
        href="#/deployments"
        @click=${(e: Event) => {
          e.preventDefault();
          navigate({ page: "deployments" });
        }}
      >
        <span class="nav-icon">${unsafeHTML(iconStrings.workflow)}</span>
        Deployments
      </a>

      ${deployments.length > 0
        ? html`
            <div class="nav-section">Deployed</div>
            ${deployments.map(
              (dep) => html`
                <a
                  class="nav-item"
                  href="${dep.url}_ui/"
                  target="_blank"
                  rel="noopener"
                >
                  <span class="nav-icon">${unsafeHTML(
                    dep.status === "running"
                      ? iconStrings.check
                      : dep.status === "crashed"
                        ? iconStrings.x
                        : iconStrings.pending
                  )}</span>
                  ${dep.name}
                </a>
              `
            )}
          `
        : nothing}
    `;
  }

  private _renderDefaultNav() {
    const workflows = this.meta?.workflows || [];
    return html`
      <div class="nav-section">General</div>
      <a
        class="nav-item ${this._isActive("home") ? "active" : ""}"
        href="#/"
        @click=${this._navClick("home")}
      >
        <span class="nav-icon">${unsafeHTML(iconStrings.home)}</span>
        Overview
      </a>
      <a
        class="nav-item ${this._isActive("workflows") ? "active" : ""}"
        href="#/workflows"
        @click=${this._navClick("workflows")}
      >
        <span class="nav-icon">${unsafeHTML(iconStrings.workflow)}</span>
        Workflows
      </a>
      <a
        class="nav-item ${this._isActive("agents") ? "active" : ""}"
        href="#/agents"
        @click=${this._navClick("agents")}
      >
        <span class="nav-icon">${unsafeHTML(iconStrings.agent)}</span>
        Agents
      </a>

      ${workflows.length > 0
        ? html`
            <div class="nav-section">Workbench</div>
            ${workflows.map(
              (wf) => html`
                <a
                  class="nav-item ${this._route.page === "workbench" &&
                  (this._route as { path: string }).path === wf.path
                    ? "active"
                    : ""}"
                  href="#/workbench${wf.path}"
                  @click=${(e: Event) => {
                    e.preventDefault();
                    navigate({ page: "workbench", path: wf.path });
                  }}
                >
                  <span class="nav-icon"
                    >${unsafeHTML(
                      wf.uiType === "Chat"
                        ? iconStrings.chat
                        : iconStrings.workflow
                    )}</span
                  >
                  ${wf.title || wf.name}
                </a>
              `
            )}
          `
        : nothing}

      <div class="nav-section">System</div>
      <a
        class="nav-item ${this._isActive("observe") ? "active" : ""}"
        href="#/observe"
        @click=${this._navClick("observe")}
      >
        <span class="nav-icon">${unsafeHTML(iconStrings.observe)}</span>
        Observability
      </a>
      <a
        class="nav-item ${this._isActive("settings") ? "active" : ""}"
        href="#/settings"
        @click=${this._navClick("settings")}
      >
        <span class="nav-icon">${unsafeHTML(iconStrings.settings)}</span>
        Settings
      </a>
    `;
  }

  render() {
    const isOrchestrator = this.meta?.mode === "orchestrator";
    return html`
      <div class="shell">
        <nav class="sidebar ${this._sidebarOpen ? "open" : ""}">
          <a
            class="sidebar-brand"
            href="#/"
            @click=${this._navClick("home")}
          >
            <span class="brand-icon">${unsafeHTML(this.meta?.logo ? `<img src="${this.meta.logo}" alt="" width="22" height="22" style="object-fit:contain">` : logoSvgStr)}</span>
            <span class="brand-text"
              >h<span class="ai">ai</span>ra</span
            >
          </a>

          <div class="sidebar-nav">
            ${isOrchestrator ? this._renderOrchestratorNav() : this._renderDefaultNav()}
          </div>
        </nav>

        <div class="main">
          <div class="topbar">
            <span class="topbar-title">${this._pageTitle()}</span>
          </div>
          <div class="content ${this._isScrollable() ? "scrollable" : ""}">
            ${this._renderPage()}
          </div>
        </div>

        <button
          class="mobile-toggle"
          @click=${() => (this._sidebarOpen = !this._sidebarOpen)}
        >
          ${unsafeHTML(iconStrings.sidebar)}
        </button>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-app": HairaApp;
  }
}
