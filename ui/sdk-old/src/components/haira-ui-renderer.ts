import { baseCSS } from "../core";
import type { ToolRenderEvent } from "../core/types";

const COMPONENT_MAP: Record<string, string> = {
  table: "haira-ui-table",
  "status-card": "haira-ui-status-card",
  "code-block": "haira-ui-code-block",
  diff: "haira-ui-diff",
  "key-value": "haira-ui-key-value",
  progress: "haira-ui-progress-view",
  form: "haira-ui-form-view",
  confirm: "haira-ui-confirm",
  choices: "haira-ui-choices",
  chart: "haira-ui-chart",
};

const MAX_DEPTH = 3;

interface Renderable {
  setProps(props: Record<string, unknown>): void;
}

export class HairaUIRenderer extends HTMLElement {
  private pendingEvent: ToolRenderEvent | null = null;

  connectedCallback() {
    this.ensureShadow();
    if (this.pendingEvent) {
      this.doRender(this.pendingEvent);
      this.pendingEvent = null;
    }
  }

  private ensureShadow() {
    if (this.shadowRoot) return;
    this.attachShadow({ mode: "open" });
    this.shadowRoot!.innerHTML = `
      <style>
        ${baseCSS}
        :host {
          display: block;
          margin-left: 2.25rem;
          max-width: 560px;
        }
        @media (max-width: 640px) {
          :host { margin-left: 0; max-width: 100%; }
        }
        .group { display: flex; flex-direction: column; gap: 0.5rem; }
        .fallback {
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius); padding: 0.75rem 1rem;
          font-family: var(--haira-mono); font-size: 0.75rem;
          color: var(--haira-text-dim); white-space: pre-wrap; overflow-x: auto;
        }
      </style>
      <div id="container"></div>
    `;
  }

  render(event: ToolRenderEvent) {
    this.ensureShadow();
    if (!this.isConnected) {
      this.pendingEvent = event;
      return;
    }
    this.doRender(event);
  }

  private doRender(event: ToolRenderEvent) {
    const container = this.shadowRoot!.getElementById("container")!;
    container.innerHTML = "";

    try {
      const el = this.renderNode(event.component, event.props, 0);
      if (el) container.appendChild(el);
    } catch {
      const pre = document.createElement("div");
      pre.className = "fallback";
      pre.textContent = JSON.stringify(event.props, null, 2);
      container.appendChild(pre);
    }
  }

  private renderNode(
    component: string,
    props: Record<string, unknown>,
    depth: number,
  ): HTMLElement | null {
    if (depth > MAX_DEPTH) return null;

    if (component === "group") {
      const group = document.createElement("div");
      group.className = "group";
      const children = (props.children as Array<Record<string, unknown>>) || [];
      for (const child of children) {
        const childComponent = child.component as string;
        const childProps = child.props as Record<string, unknown>;
        if (childComponent && childProps) {
          const el = this.renderNode(childComponent, childProps, depth + 1);
          if (el) group.appendChild(el);
        }
      }
      return group;
    }

    const tagName = COMPONENT_MAP[component];
    if (!tagName) {
      const pre = document.createElement("div");
      pre.className = "fallback";
      pre.textContent = JSON.stringify(props, null, 2);
      return pre;
    }

    const el = document.createElement(tagName);
    const isRestored = this.hasAttribute("data-restored");
    const finalProps = isRestored ? { ...props, _restored: true } : props;
    Promise.resolve().then(() => {
      (el as unknown as Renderable).setProps(finalProps);
    });
    return el;
  }
}
