import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { baseStyles } from "../core/styles";
import type { ToolRenderEvent } from "../core/types";

/**
 * Maps logical component names to registered custom element tag names.
 */
const COMPONENT_MAP: Record<string, string> = {
  "table": "haira-ui-table",
  "status-card": "haira-ui-status-card",
  "code-block": "haira-ui-code-block",
  "diff": "haira-ui-diff",
  "key-value": "haira-ui-key-value",
  "progress": "haira-ui-progress-view",
  "form": "haira-ui-form-view",
  "confirm": "haira-ui-confirm",
  "choices": "haira-ui-choices",
  "chart": "haira-ui-chart",
  "product-cards": "haira-ui-product-cards",
};

@customElement("haira-ui-renderer")
export class HairaUiRenderer extends LitElement {
  static styles = [
    baseStyles,
    css`
      :host {
        display: block;
      }
      .fallback {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius-sm);
        padding: 0.65rem 0.85rem;
        font-family: var(--haira-mono);
        font-size: 0.78rem;
        color: var(--haira-text-dim);
        white-space: pre-wrap;
        word-break: break-word;
        max-height: 300px;
        overflow-y: auto;
      }
      .group {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
      }
    `,
  ];

  @property({ type: Object }) event: ToolRenderEvent | null = null;

  private _rendered: HTMLElement | null = null;

  updated(changed: Map<string, unknown>) {
    if (changed.has("event") && this.event) {
      // Clear previous content
      const container = this.renderRoot.querySelector(".renderer-slot");
      if (container) {
        container.innerHTML = "";
        const el = this._createComponent(this.event, 0);
        if (el) container.appendChild(el);
      }
    }
  }

  /**
   * Imperatively render a ToolRenderEvent and return the created element.
   */
  public renderEvent(event: ToolRenderEvent): HTMLElement | null {
    return this._createComponent(event, 0);
  }

  private _createComponent(
    event: ToolRenderEvent,
    depth: number
  ): HTMLElement | null {
    // Handle group component with recursive rendering (max depth 3)
    if (event.component === "group") {
      if (depth >= 3) return null;
      const wrapper = document.createElement("div");
      wrapper.style.display = "flex";
      wrapper.style.flexDirection = "column";
      wrapper.style.gap = "0.5rem";

      const children = (event.props.children || []) as ToolRenderEvent[];
      for (const child of children) {
        const childEl = this._createComponent(child, depth + 1);
        if (childEl) wrapper.appendChild(childEl);
      }
      return wrapper;
    }

    const tagName = COMPONENT_MAP[event.component];

    if (!tagName) {
      // Fallback: render unknown components as JSON
      const fallback = document.createElement("div");
      fallback.className = "fallback";
      fallback.textContent = JSON.stringify(
        { component: event.component, props: event.props },
        null,
        2
      );
      return fallback;
    }

    const el = document.createElement(tagName) as HTMLElement & {
      setProps?: (props: Record<string, unknown>) => void;
    };

    // Support data-restored attribute for _restored prop
    if (event.props._restored) {
      el.setAttribute("data-restored", "");
    }

    // Set props via the setProps method if available, otherwise assign directly
    if (typeof el.setProps === "function") {
      el.setProps(event.props);
    } else {
      // The element may not be upgraded yet; wait for upgrade then set props
      customElements.whenDefined(tagName).then(() => {
        if (typeof (el as any).setProps === "function") {
          (el as any).setProps(event.props);
        }
      });
    }

    return el;
  }

  render() {
    return html`<div class="renderer-slot"></div>`;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-ui-renderer": HairaUiRenderer;
  }
}
