import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { baseStyles, animateInStyles, popoverStyles } from "../core/styles";
import { iconStrings } from "../core/icons";
import type { ImageProps } from "../core/types";

@customElement("haira-ui-image")
export class HairaImage extends LitElement {
  static styles = [
    baseStyles,
    animateInStyles,
    popoverStyles,
    css`
      .image-card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius-md);
        overflow: hidden;
        animation: var(--haira-animate-in);
      }

      .image-header {
        padding: 0.75rem 1rem;
        border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg);
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 0.5rem;
      }

      .image-title {
        font-size: 0.85rem;
        font-weight: 500;
        color: var(--haira-text);
        margin: 0;
        flex: 1;
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .image-actions {
        display: flex;
        align-items: center;
        gap: 0.25rem;
      }

      .action-btn {
        background: none;
        border: none;
        padding: 0.25rem;
        border-radius: var(--haira-radius-sm);
        color: var(--haira-muted);
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        transition: all 0.15s ease;
      }

      .action-btn:hover {
        background: var(--haira-bg-hover);
        color: var(--haira-text);
      }

      .action-btn svg {
        width: 14px;
        height: 14px;
      }

      .image-container {
        position: relative;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--haira-bg);
        min-height: 200px;
      }

      .image-container.has-dimensions {
        min-height: auto;
      }

      .main-image {
        max-width: 100%;
        height: auto;
        display: block;
        border-radius: 0;
        transition: transform 0.2s ease;
        cursor: zoom-in;
      }

      .main-image.clickable:hover {
        transform: scale(1.02);
      }

      .image-error {
        padding: 2rem;
        text-align: center;
        color: var(--haira-muted);
        font-size: 0.85rem;
      }

      .image-error svg {
        width: 32px;
        height: 32px;
        margin-bottom: 0.5rem;
        opacity: 0.5;
      }

      .image-loading {
        padding: 2rem;
        text-align: center;
        color: var(--haira-muted);
        font-size: 0.85rem;
      }

      .image-caption {
        padding: 0.75rem 1rem;
        border-top: 1px solid var(--haira-border);
        background: var(--haira-bg);
        font-size: 0.8rem;
        color: var(--haira-muted);
        line-height: 1.4;
      }

      /* Expanded overlay */
      .expanded-overlay {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: rgba(0, 0, 0, 0.9);
        z-index: 1000;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 2rem;
        cursor: zoom-out;
      }

      .expanded-image {
        max-width: 100%;
        max-height: 100%;
        object-fit: contain;
        border-radius: var(--haira-radius-md);
        box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
      }

      .expanded-close {
        position: absolute;
        top: 1rem;
        right: 1rem;
        background: rgba(0, 0, 0, 0.7);
        border: none;
        color: white;
        padding: 0.5rem;
        border-radius: var(--haira-radius-sm);
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        transition: background 0.15s ease;
      }

      .expanded-close:hover {
        background: rgba(0, 0, 0, 0.9);
      }

      .expanded-close svg {
        width: 20px;
        height: 20px;
      }

      /* Object fit styles */
      .fit-contain { object-fit: contain; }
      .fit-cover { object-fit: cover; }
      .fit-fill { object-fit: fill; }
      .fit-scale-down { object-fit: scale-down; }
      .fit-none { object-fit: none; }
    `,
  ];

  @property({ type: String }) title?: string;
  @property({ type: String }) src: string = "";
  @property({ type: String }) alt?: string;
  @property({ type: Number }) width?: number;
  @property({ type: Number }) height?: number;
  @property({ type: String }) caption?: string;
  @property({ type: String }) fit: "contain" | "cover" | "fill" | "scale-down" | "none" = "contain";

  @state() private _loading: boolean = true;
  @state() private _error: boolean = false;
  @state() private _expanded: boolean = false;

  connectedCallback() {
    super.connectedCallback();
    document.addEventListener("keydown", this._onEscKey);
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    document.removeEventListener("keydown", this._onEscKey);
  }

  setProps(props: ImageProps) {
    this.title = props.title;
    this.src = props.src;
    this.alt = props.alt;
    this.width = props.width;
    this.height = props.height;
    this.caption = props.caption;
    this.fit = props.fit || "contain";
    
    // Reset state when src changes
    this._loading = true;
    this._error = false;
  }

  private _onImageLoad() {
    this._loading = false;
    this._error = false;
  }

  private _onImageError() {
    this._loading = false;
    this._error = true;
  }

  private _onExpand() {
    if (!this._error && !this._loading) {
      this._expanded = true;
    }
  }

  private _onCollapse() {
    this._expanded = false;
  }

  private _onEscKey = (e: KeyboardEvent) => {
    if (e.key === "Escape" && this._expanded) {
      this._onCollapse();
    }
  };

  private _onOverlayClick(e: MouseEvent) {
    if (e.target === e.currentTarget) {
      this._onCollapse();
    }
  }

  private _copyImageUrl() {
    if (navigator.clipboard && this.src) {
      navigator.clipboard.writeText(this.src).catch(() => {
        // Fallback: create a temporary input element
        const input = document.createElement("input");
        input.value = this.src;
        document.body.appendChild(input);
        input.select();
        document.execCommand("copy");
        document.body.removeChild(input);
      });
    }
  }

  private _openInNewTab() {
    if (this.src) {
      window.open(this.src, "_blank");
    }
  }

  render() {
    const hasTitle = this.title && this.title.trim() !== "";
    const hasCaption = this.caption && this.caption.trim() !== "";
    const hasDimensions = this.width || this.height;

    const imageStyle = {
      ...(this.width && { width: `${this.width}px` }),
      ...(this.height && { height: `${this.height}px` }),
    };

    return html`
      <div class="image-card">
        ${hasTitle
          ? html`
              <div class="image-header">
                <h3 class="image-title">${this.title}</h3>
                <div class="image-actions">
                  <button
                    class="action-btn"
                    title="Copy image URL"
                    @click=${this._copyImageUrl}
                  >
                    ${iconStrings.copy}
                  </button>
                  <button
                    class="action-btn"
                    title="Open in new tab"
                    @click=${this._openInNewTab}
                  >
                    ${iconStrings.external_link}
                  </button>
                </div>
              </div>
            `
          : nothing}

        <div class="image-container ${hasDimensions ? "has-dimensions" : ""}">
          ${this._loading
            ? html`
                <div class="image-loading">
                  <div>Loading image...</div>
                </div>
              `
            : this._error
            ? html`
                <div class="image-error">
                  ${iconStrings.image}
                  <div>Failed to load image</div>
                  <div style="font-size: 0.75rem; margin-top: 0.25rem; opacity: 0.7;">
                    ${this.src}
                  </div>
                </div>
              `
            : html`
                <img
                  class="main-image clickable fit-${this.fit}"
                  src=${this.src}
                  alt=${this.alt || this.title || "Image"}
                  style=${Object.entries(imageStyle)
                    .map(([key, value]) => `${key}: ${value}`)
                    .join("; ")}
                  @load=${this._onImageLoad}
                  @error=${this._onImageError}
                  @click=${this._onExpand}
                />
              `}
        </div>

        ${hasCaption
          ? html`
              <div class="image-caption">
                ${this.caption}
              </div>
            `
          : nothing}
      </div>

      ${this._expanded
        ? html`
            <div class="expanded-overlay" @click=${this._onOverlayClick}>
              <button class="expanded-close" @click=${this._onCollapse}>
                ${iconStrings.x}
              </button>
              <img
                class="expanded-image"
                src=${this.src}
                alt=${this.alt || this.title || "Image"}
              />
            </div>
          `
        : nothing}
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-ui-image": HairaImage;
  }
}