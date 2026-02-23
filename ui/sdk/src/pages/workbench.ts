import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { baseStyles } from "../core/styles";
import type { WorkflowMeta } from "../core/types";
import { currentRoute } from "../core/router";

@customElement("haira-page-workbench")
export class PageWorkbench extends LitElement {
  static styles = [
    baseStyles,
    css`
      :host {
        display: flex;
        flex: 1;
        overflow: hidden;
      }
    `,
  ];

  @property({ type: Object }) meta: WorkflowMeta | null = null;

  /** Track the current workflow path so we re-render on hash changes */
  @state() private _currentPath = "";

  connectedCallback() {
    super.connectedCallback();
    this._updatePath();
    window.addEventListener("popstate", this._onRouteChange);
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    window.removeEventListener("popstate", this._onRouteChange);
  }

  private _onRouteChange = () => {
    this._updatePath();
  };

  private _updatePath() {
    const route = currentRoute();
    if (route.page === "workbench") {
      this._currentPath = (route as { page: "workbench"; path: string }).path;
    }
  }

  private _getWorkflowMeta(): WorkflowMeta | null {
    if (!this.meta) return null;

    const path = this._currentPath;
    const workflows = this.meta.workflows || [];
    const wf = workflows.find((w) => w.path === path);

    if (wf) {
      // Build a meta for this specific workflow
      return {
        ...this.meta,
        mode: wf.uiType?.toLowerCase() === "chat" ? "chat" : "form",
        name: wf.name,
        path: wf.path,
        method: wf.method,
        title: wf.title || wf.name,
        description: wf.description || this.meta.description,
        hasFile: wf.hasFile ?? false,
        params: wf.params || [],
        chatParam: wf.chatParam,
        fileParam: wf.fileParam,
        suggestions: wf.suggestions,
        accent: wf.accent || this.meta.accent,
        logo: wf.logo || this.meta.logo,
        theme: wf.theme || this.meta.theme,
        avatar: wf.avatar || this.meta.avatar,
        arpUrl: wf.arpUrl,
        backend: wf.backend,
      };
    }

    // If meta.path matches directly (single-workflow mode)
    if (this.meta.path === path || this.meta.mode !== "index") {
      return this.meta;
    }

    return null;
  }

  render() {
    const wfMeta = this._getWorkflowMeta();

    if (!wfMeta) {
      return html`<div
        style="padding:2rem;color:var(--haira-muted);text-align:center"
      >
        Workflow not found.
      </div>`;
    }

    const mode = wfMeta.mode === "chat" ? "chat" : "form";

    // Use keyed rendering so Lit creates a new element when the workflow changes.
    // Without this, switching from one chat workflow to another reuses the same
    // <haira-chat> element — which never re-runs connectedCallback/initSession.
    if (mode === "chat") {
      return html`<haira-chat .meta=${wfMeta} .key=${wfMeta.path}></haira-chat>`;
    }

    return html`<haira-form .meta=${wfMeta} .key=${wfMeta.path}></haira-form>`;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-page-workbench": PageWorkbench;
  }
}
