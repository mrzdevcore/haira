import { baseCSS } from "./styles";
import { esc, escAttr } from "./html";

/**
 * BaseComponent — abstract base class for all Haira web components.
 *
 * Eliminates boilerplate: shadow DOM setup, shared styles, HTML escaping,
 * event emission, and DOM querying. Components extend this and implement
 * `render()` + optionally `onMount()` and `onUpdate()`.
 *
 * Usage:
 *   class MyComponent extends BaseComponent<MyProps> {
 *     render() { return `<div>...</div>`; }
 *     styles() { return `.card { ... }`; }
 *   }
 */
export abstract class BaseComponent<P = Record<string, unknown>> extends HTMLElement {
  protected root!: ShadowRoot;
  protected props: P = {} as P;

  connectedCallback() {
    this.root = this.attachShadow({ mode: "open" });
    this.root.innerHTML = `<style>${baseCSS}\n${this.styles()}</style>${this.render()}`;
    this.onMount();
  }

  /**
   * Called by the renderer or parent to pass typed props.
   * Updates internal state and triggers onUpdate().
   */
  setProps(props: P) {
    this.props = props;
    this.onUpdate();
  }

  // --- Subclass hooks ---

  /** Return the component's HTML template string. */
  protected abstract render(): string;

  /** Return component-specific CSS (no need to include baseCSS). */
  protected styles(): string {
    return "";
  }

  /** Called once after the shadow DOM is created. Wire up event listeners here. */
  protected onMount() {}

  /** Called whenever props are updated via setProps(). Re-render or patch DOM here. */
  protected onUpdate() {}

  // --- Shared utilities ---

  /** Query an element by ID within the shadow DOM. */
  protected $(id: string): HTMLElement {
    return this.root.getElementById(id)!;
  }

  /** Query an element within the shadow DOM using a CSS selector. */
  protected $q<T extends HTMLElement = HTMLElement>(selector: string): T {
    return this.root.querySelector(selector) as T;
  }

  /** Query all elements matching a CSS selector within the shadow DOM. */
  protected $qa<T extends HTMLElement = HTMLElement>(selector: string): NodeListOf<T> {
    return this.root.querySelectorAll(selector) as NodeListOf<T>;
  }

  /** Escape HTML special characters. */
  protected esc(s: string): string {
    return esc(s);
  }

  /** Escape HTML including double quotes (for attributes). */
  protected escAttr(s: string): string {
    return escAttr(s);
  }

  /** Dispatch a custom event that crosses shadow DOM boundaries. */
  protected emit(name: string, detail?: unknown) {
    this.dispatchEvent(
      new CustomEvent(name, {
        detail,
        bubbles: true,
        composed: true,
      }),
    );
  }
}
