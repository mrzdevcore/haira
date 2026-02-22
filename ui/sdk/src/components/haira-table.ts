import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles, keyframes, animateInStyles } from "../core/styles";
import { iconStrings } from "../core/icons";
import type { TableProps, TabData } from "../core/types";

@customElement("haira-ui-table")
export class HairaTable extends LitElement {
  static styles = [
    baseStyles,
    animateInStyles,
    css`
      .table-card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        overflow: hidden;
      }

      .table-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.55rem 0.85rem;
        border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg);
      }
      .table-title {
        font-size: 0.85rem;
        font-weight: 700;
        color: var(--haira-text);
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

      /* Tab bar */
      .tab-bar {
        display: flex;
        gap: 0;
        border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg);
        overflow-x: auto;
      }
      .tab-bar::-webkit-scrollbar {
        height: 0;
      }
      .tab {
        padding: 0.5rem 0.85rem;
        font-size: 0.78rem;
        font-weight: 600;
        color: var(--haira-muted);
        cursor: pointer;
        border-bottom: 2px solid transparent;
        transition: all 0.12s;
        white-space: nowrap;
        background: none;
        border-top: none;
        border-left: none;
        border-right: none;
        font-family: var(--haira-font);
      }
      .tab:hover {
        color: var(--haira-text-dim);
        background: var(--haira-bg-card);
      }
      .tab.active {
        color: var(--haira-accent);
        border-bottom-color: var(--haira-accent);
      }

      /* Scrollable table area */
      .table-scroll {
        max-height: 420px;
        overflow: auto;
      }
      .table-scroll::-webkit-scrollbar {
        width: 5px;
        height: 5px;
      }
      .table-scroll::-webkit-scrollbar-thumb {
        background: var(--haira-muted);
        border-radius: 3px;
      }

      table {
        width: 100%;
        border-collapse: collapse;
        font-size: 0.82rem;
      }
      thead {
        position: sticky;
        top: 0;
        z-index: 1;
      }
      th {
        padding: 0.5rem 0.7rem;
        background: var(--haira-bg-card);
        border-bottom: 1px solid var(--haira-border);
        font-weight: 600;
        font-size: 0.74rem;
        color: var(--haira-text-dim);
        text-transform: uppercase;
        letter-spacing: 0.03em;
        text-align: left;
        white-space: nowrap;
      }
      td {
        padding: 0.45rem 0.7rem;
        border-bottom: 1px solid var(--haira-border);
        color: var(--haira-text);
        word-break: break-word;
      }
      tr:last-child td {
        border-bottom: none;
      }
      tr.highlighted td {
        background: rgba(232, 163, 23, 0.06);
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
  @property({ type: Array }) headers: string[] = [];
  @property({ type: Array }) rows: string[][] = [];
  @property({ type: Array }) tabs: TabData[] = [];
  @property({ type: Array }) highlight: number[] = [];

  @state() private _activeTab: number = 0;
  @state() private _search: string = "";

  /** Set all props at once */
  public setProps(props: TableProps): void {
    this.title = props.title || "";
    this.headers = (props.headers || []) as string[];
    this.rows = (props.rows || []) as string[][];
    this.tabs = (props.tabs || []) as TabData[];
    this.highlight = (props.highlight || []) as number[];
    this._activeTab = 0;
    this._search = "";
    this.requestUpdate();
  }

  private _currentHeaders(): string[] {
    if (this.tabs && this.tabs.length > 0) {
      return this.tabs[this._activeTab]?.headers || [];
    }
    return this.headers;
  }

  private _currentRows(): string[][] {
    if (this.tabs && this.tabs.length > 0) {
      return this.tabs[this._activeTab]?.rows || [];
    }
    return this.rows;
  }

  private _currentHighlight(): number[] {
    if (this.tabs && this.tabs.length > 0) {
      return this.tabs[this._activeTab]?.highlight || [];
    }
    return this.highlight;
  }

  private _filteredRows(): string[][] {
    const rows = this._currentRows();
    if (!this._search) return rows;
    const q = this._search.toLowerCase();
    return rows.filter((row) =>
      row.some((cell) => String(cell ?? "").toLowerCase().includes(q))
    );
  }

  private _showSearch(): boolean {
    return this._currentRows().length >= 15;
  }

  private _onSearch(e: Event) {
    this._search = (e.target as HTMLInputElement).value;
  }

  private _onTabClick(index: number) {
    this._activeTab = index;
    this._search = "";
  }

  render() {
    const hdrs = this._currentHeaders();
    const filtered = this._filteredRows();
    const hl = new Set(this._currentHighlight());

    return html`
      <div class="table-card">
        ${this.title || this._showSearch()
          ? html`
              <div class="table-header">
                <span class="table-title">${this.title}</span>
                ${this._showSearch()
                  ? html`
                      <div class="search-wrap">
                        <span class="search-icon"
                          >${unsafeHTML(iconStrings.search)}</span
                        >
                        <input
                          class="search-input"
                          type="text"
                          placeholder="Filter rows..."
                          .value=${this._search}
                          @input=${this._onSearch}
                        />
                      </div>
                    `
                  : nothing}
              </div>
            `
          : nothing}
        ${this.tabs && this.tabs.length > 0
          ? html`
              <div class="tab-bar">
                ${this.tabs.map(
                  (tab, i) => html`
                    <button
                      class="tab ${this._activeTab === i ? "active" : ""}"
                      @click=${() => this._onTabClick(i)}
                    >
                      ${tab.name}
                    </button>
                  `
                )}
              </div>
            `
          : nothing}

        <div class="table-scroll">
          ${filtered.length > 0
            ? html`
                <table>
                  <thead>
                    <tr>
                      ${hdrs.map((h) => html`<th>${h}</th>`)}
                    </tr>
                  </thead>
                  <tbody>
                    ${filtered.map(
                      (row, ri) => html`
                        <tr class="${hl.has(ri) ? "highlighted" : ""}">
                          ${row.map((cell) => html`<td>${cell}</td>`)}
                        </tr>
                      `
                    )}
                  </tbody>
                </table>
              `
            : html`<div class="no-results">
                ${this._search ? "No matching rows." : "No data."}
              </div>`}
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-ui-table": HairaTable;
  }
}
