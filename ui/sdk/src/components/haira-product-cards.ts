import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { baseStyles, animateInStyles } from "../core/styles";
import { iconStrings } from "../core/icons";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import type { ProductCardsProps, ProductCardItem } from "../core/types";

@customElement("haira-ui-product-cards")
export class HairaProductCards extends LitElement {
  static styles = [
    baseStyles,
    animateInStyles,
    css`
      .product-cards-container {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        overflow: hidden;
      }

      .cards-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.55rem 0.85rem;
        border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg);
      }
      .cards-title {
        font-size: 0.85rem;
        font-weight: 700;
        color: var(--haira-text);
      }
      .cards-count {
        font-size: 0.74rem;
        color: var(--haira-muted);
      }

      /* Search */
      .search-wrap {
        position: relative;
        display: flex;
        align-items: center;
      }
      .search-icon {
        position: absolute;
        left: 0.5rem;
        display: flex;
        align-items: center;
        color: var(--haira-muted);
        pointer-events: none;
      }
      .search-input {
        padding: 0.35rem 0.55rem 0.35rem 1.7rem;
        background: var(--haira-bg-input);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius-sm);
        color: var(--haira-text);
        font-family: var(--haira-font);
        font-size: 0.78rem;
        outline: none;
        width: 180px;
        transition: border-color 0.15s;
      }
      .search-input:focus {
        border-color: var(--haira-border-focus);
      }
      .search-input::placeholder {
        color: var(--haira-muted);
      }

      /* Grid */
      .cards-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(190px, 1fr));
        gap: 0.75rem;
        padding: 0.85rem;
        max-height: 520px;
        overflow-y: auto;
      }
      .cards-grid::-webkit-scrollbar {
        width: 5px;
      }
      .cards-grid::-webkit-scrollbar-thumb {
        background: var(--haira-muted);
        border-radius: 3px;
      }

      /* Card */
      .product-card {
        background: var(--haira-bg);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius-sm);
        overflow: hidden;
        transition: border-color 0.15s, box-shadow 0.15s;
        display: flex;
        flex-direction: column;
      }
      .product-card:hover {
        border-color: var(--haira-accent);
        box-shadow: 0 0 0 1px rgba(232, 163, 23, 0.15);
      }

      /* Image */
      .card-image-wrap {
        position: relative;
        width: 100%;
        aspect-ratio: 1 / 1;
        background: var(--haira-bg-elevated);
        overflow: hidden;
        display: flex;
        align-items: center;
        justify-content: center;
      }
      .card-image {
        width: 100%;
        height: 100%;
        object-fit: cover;
        transition: transform 0.2s;
      }
      .product-card:hover .card-image {
        transform: scale(1.03);
      }
      .card-placeholder {
        color: var(--haira-muted);
        font-size: 1.8rem;
        opacity: 0.4;
      }

      /* Badge */
      .card-badge {
        position: absolute;
        top: 0.4rem;
        left: 0.4rem;
        background: var(--haira-accent);
        color: #000;
        font-size: 0.65rem;
        font-weight: 700;
        padding: 0.15rem 0.4rem;
        border-radius: 3px;
        text-transform: uppercase;
        letter-spacing: 0.02em;
      }

      /* Content */
      .card-body {
        padding: 0.6rem 0.65rem;
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 0.2rem;
      }
      .card-brand {
        font-size: 0.68rem;
        font-weight: 600;
        color: var(--haira-muted);
        text-transform: uppercase;
        letter-spacing: 0.03em;
      }
      .card-name {
        font-size: 0.8rem;
        font-weight: 600;
        color: var(--haira-text);
        line-height: 1.3;
        display: -webkit-box;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
      }
      .card-description {
        font-size: 0.72rem;
        color: var(--haira-text-dim);
        line-height: 1.35;
        display: -webkit-box;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
        margin-top: 0.1rem;
      }
      .card-price {
        font-size: 0.88rem;
        font-weight: 700;
        color: var(--haira-accent);
        margin-top: auto;
        padding-top: 0.3rem;
      }

      .no-results {
        padding: 1.5rem;
        text-align: center;
        font-size: 0.82rem;
        color: var(--haira-muted);
      }
    `,
  ];

  @property({ type: String }) title: string = "";
  @property({ type: Array }) cards: ProductCardItem[] = [];

  @state() private _search: string = "";

  public setProps(props: ProductCardsProps): void {
    this.title = props.title || "";
    this.cards = (props.cards || []) as ProductCardItem[];
    this._search = "";
    this.requestUpdate();
  }

  private _filteredCards(): ProductCardItem[] {
    if (!this._search) return this.cards;
    const q = this._search.toLowerCase();
    return this.cards.filter(
      (c) =>
        (c.name || "").toLowerCase().includes(q) ||
        (c.brand || "").toLowerCase().includes(q) ||
        (c.description || "").toLowerCase().includes(q) ||
        (c.price || "").toLowerCase().includes(q)
    );
  }

  private _showSearch(): boolean {
    return this.cards.length >= 12;
  }

  private _onSearch(e: Event) {
    this._search = (e.target as HTMLInputElement).value;
  }

  private _onImageError(e: Event) {
    const img = e.target as HTMLImageElement;
    img.style.display = "none";
    const placeholder = img.parentElement?.querySelector(
      ".card-placeholder"
    ) as HTMLElement;
    if (placeholder) placeholder.style.display = "flex";
  }

  render() {
    const filtered = this._filteredCards();

    return html`
      <div class="product-cards-container">
        ${this.title || this._showSearch()
          ? html`
              <div class="cards-header">
                <div>
                  <span class="cards-title">${this.title}</span>
                  <span class="cards-count"
                    >${this.cards.length} product${this.cards.length !== 1
                      ? "s"
                      : ""}</span
                  >
                </div>
                ${this._showSearch()
                  ? html`
                      <div class="search-wrap">
                        <span class="search-icon"
                          >${unsafeHTML(iconStrings.search)}</span
                        >
                        <input
                          class="search-input"
                          type="text"
                          placeholder="Filter products..."
                          .value=${this._search}
                          @input=${this._onSearch}
                        />
                      </div>
                    `
                  : nothing}
              </div>
            `
          : nothing}

        <div class="cards-grid">
          ${filtered.length > 0
            ? filtered.map(
                (card) => html`
                  <div class="product-card">
                    <div class="card-image-wrap">
                      ${card.image
                        ? html`
                            <img
                              class="card-image"
                              src="${card.image}"
                              alt="${card.name}"
                              loading="lazy"
                              @error=${this._onImageError}
                            />
                            <div
                              class="card-placeholder"
                              style="display:none;position:absolute;"
                            >
                              &#x1f4e6;
                            </div>
                          `
                        : html`<div class="card-placeholder">&#x1f4e6;</div>`}
                      ${card.badge
                        ? html`<span class="card-badge">${card.badge}</span>`
                        : nothing}
                    </div>
                    <div class="card-body">
                      ${card.brand
                        ? html`<div class="card-brand">${card.brand}</div>`
                        : nothing}
                      <div class="card-name">${card.name}</div>
                      ${card.description
                        ? html`<div class="card-description">
                            ${card.description}
                          </div>`
                        : nothing}
                      <div class="card-price">${card.price}</div>
                    </div>
                  </div>
                `
              )
            : html`<div class="no-results">
                ${this._search ? "No matching products." : "No products."}
              </div>`}
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-ui-product-cards": HairaProductCards;
  }
}
