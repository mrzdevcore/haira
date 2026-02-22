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

  private _getWorkflowPath(): string {
    const route = currentRoute();
    if (route.page === "workbench") {
      return (route as { page: "workbench"; path: string }).path;
    }
    return this.meta?.path || "";
  }

  private _getWorkflowMeta(): WorkflowMeta | null {
    if (!this.meta) return null;

    const path = this._getWorkflowPath();
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

    if (mode === "chat") {
      return html`<haira-chat .meta=${wfMeta}></haira-chat>`;
    }

    return html`<haira-form .meta=${wfMeta}></haira-form>`;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-page-workbench": PageWorkbench;
  }
}
