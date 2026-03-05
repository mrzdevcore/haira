import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles, keyframes, animateInStyles } from "../core/styles";
import { iconStrings } from "../core/icons";
import type { TableProps, TabData } from "../core/types";

/** Threshold above which virtual scrolling kicks in */
const VIRTUAL_THRESHOLD = 200;
/** Fixed row height in px for virtual rows */
const ROW_HEIGHT = 32;
/** Extra rows rendered above/below the viewport */
const OVERSCAN = 15;
/** Viewport height for the scroll area */
const VIEWPORT_HEIGHT = 420;

type SortDir = "asc" | "desc" | null;

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

      /* ── Standard (non-virtual) table scroll ── */
      .table-scroll {
        max-height: ${VIEWPORT_HEIGHT}px;
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
        user-select: none;
        position: relative;
      }
      /* Sortable header */
      th.sortable {
        cursor: pointer;
      }
      th.sortable:hover {
        color: var(--haira-text);
      }
      .sort-indicator {
        display: inline-block;
        margin-left: 0.3rem;
        font-size: 0.6rem;
        vertical-align: middle;
        opacity: 0.4;
      }
      th.sort-active .sort-indicator {
        opacity: 1;
        color: var(--haira-accent);
      }
      /* Drag handle on header */
      th.drag-over-left {
        box-shadow: inset 3px 0 0 0 var(--haira-accent);
      }
      th.drag-over-right {
        box-shadow: inset -3px 0 0 0 var(--haira-accent);
      }
      th.dragging {
        opacity: 0.4;
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

      /* ── Virtual scroll layout ── */
      .v-container {
        display: flex;
        flex-direction: column;
        overflow: hidden;
      }
      .v-head {
        flex: 0 0 auto;
        overflow: hidden;
      }
      .v-head table {
        table-layout: fixed;
      }
      .v-viewport {
        height: ${VIEWPORT_HEIGHT}px;
        overflow-y: auto;
        overflow-x: hidden;
      }
      .v-viewport::-webkit-scrollbar {
        width: 5px;
      }
      .v-viewport::-webkit-scrollbar-thumb {
        background: var(--haira-muted);
        border-radius: 3px;
      }
      .v-runway {
        position: relative;
        overflow: hidden;
      }
      .v-body {
        position: absolute;
        left: 0;
        right: 0;
      }
      .v-body table {
        table-layout: fixed;
      }
      .v-body td {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        height: ${ROW_HEIGHT}px;
        box-sizing: border-box;
      }

      .no-results {
        padding: 1.5rem;
        text-align: center;
        font-size: 0.82rem;
        color: var(--haira-muted);
      }

      /* Row count badge */
      .row-count {
        font-size: 0.72rem;
        color: var(--haira-muted);
        font-weight: 400;
        margin-left: 0.5rem;
      }

      /* Hide empty toggle */
      .toggle-empty {
        display: flex;
        align-items: center;
        gap: 0.35rem;
        padding: 0.25rem 0.55rem;
        font-size: 0.72rem;
        font-weight: 600;
        font-family: var(--haira-font);
        color: var(--haira-muted);
        background: none;
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius-sm);
        cursor: pointer;
        transition: all 0.12s;
        white-space: nowrap;
      }
      .toggle-empty:hover {
        color: var(--haira-text-dim);
        border-color: var(--haira-text-dim);
      }
      .toggle-empty.active {
        color: var(--haira-accent);
        border-color: var(--haira-accent);
        background: rgba(232, 163, 23, 0.06);
      }

      /* Reset sort button */
      .reset-sort {
        display: flex;
        align-items: center;
        gap: 0.25rem;
        padding: 0.25rem 0.55rem;
        font-size: 0.72rem;
        font-weight: 600;
        font-family: var(--haira-font);
        color: var(--haira-accent);
        background: rgba(232, 163, 23, 0.06);
        border: 1px solid var(--haira-accent);
        border-radius: var(--haira-radius-sm);
        cursor: pointer;
        transition: all 0.12s;
        white-space: nowrap;
      }
      .reset-sort:hover {
        background: rgba(232, 163, 23, 0.12);
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
  @state() private _hideEmpty: boolean = false;
  @state() private _scrollTop: number = 0;

  // Sort state
  @state() private _sortCol: number = -1;
  @state() private _sortDir: SortDir = null;

  // Column reorder state — maps tab index to column order
  // key: tab index (or -1 for non-tab), value: array of original column indices
  private _columnOrders: Map<number, number[]> = new Map();

  // Drag state (not reactive — used only during drag)
  private _dragColIdx: number = -1;
  private _dragOverColIdx: number = -1;
  private _dragOverSide: "left" | "right" = "left";

  private _viewportEl: HTMLElement | null = null;
  private _rafId: number = 0;

  /** Set all props at once */
  public setProps(props: TableProps): void {
    this.title = props.title || "";
    this.headers = (props.headers || []) as string[];
    this.rows = (props.rows || []) as string[][];
    this.tabs = (props.tabs || []) as TabData[];
    this.highlight = (props.highlight || []) as number[];
    this._activeTab = 0;
    this._search = "";
    this._scrollTop = 0;
    this._sortCol = -1;
    this._sortDir = null;
    this._columnOrders = new Map();
    this.requestUpdate();
  }

  protected updated() {
    this._attachScrollListener();
  }

  private _attachScrollListener() {
    const el = this.renderRoot.querySelector(".v-viewport") as HTMLElement;
    if (el && el !== this._viewportEl) {
      this._viewportEl?.removeEventListener("scroll", this._onVScroll);
      this._viewportEl = el;
      el.addEventListener("scroll", this._onVScroll, { passive: true });
    }
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    this._viewportEl?.removeEventListener("scroll", this._onVScroll);
    if (this._rafId) cancelAnimationFrame(this._rafId);
  }

  private _onVScroll = () => {
    if (this._rafId) return;
    this._rafId = requestAnimationFrame(() => {
      this._rafId = 0;
      this._scrollTop = this._viewportEl?.scrollTop ?? 0;
    });
  };

  // ── Column order ──

  private _tabKey(): number {
    return this.tabs?.length > 0 ? this._activeTab : -1;
  }

  private _getColumnOrder(): number[] {
    const key = this._tabKey();
    if (this._columnOrders.has(key)) return this._columnOrders.get(key)!;
    const hdrs = this._rawHeaders();
    return hdrs.map((_, i) => i);
  }

  private _setColumnOrder(order: number[]) {
    const key = this._tabKey();
    this._columnOrders.set(key, order);
    this.requestUpdate();
  }

  // ── Data accessors ──

  private _rawHeaders(): string[] {
    if (this.tabs && this.tabs.length > 0) {
      return this.tabs[this._activeTab]?.headers || [];
    }
    return this.headers;
  }

  /** Headers in current column order */
  private _currentHeaders(): string[] {
    const raw = this._rawHeaders();
    const order = this._getColumnOrder();
    return order.map((i) => raw[i]);
  }

  private _rawRows(): string[][] {
    if (this.tabs && this.tabs.length > 0) {
      return this.tabs[this._activeTab]?.rows || [];
    }
    return this.rows;
  }

  /** Rows reordered by column order */
  private _currentRows(): string[][] {
    const raw = this._rawRows();
    const order = this._getColumnOrder();
    return raw.map((row) => order.map((i) => row[i]));
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

  /** Sort rows by current sort column. Returns a new array (doesn't mutate). */
  private _sortedRows(rows: string[][]): string[][] {
    if (this._sortCol < 0 || !this._sortDir) return rows;
    const col = this._sortCol;
    const dir = this._sortDir === "asc" ? 1 : -1;
    return [...rows].sort((a, b) => {
      const av = String(a[col] ?? "");
      const bv = String(b[col] ?? "");
      // Try numeric comparison first
      const an = Number(av);
      const bn = Number(bv);
      if (!isNaN(an) && !isNaN(bn) && av !== "" && bv !== "") {
        return (an - bn) * dir;
      }
      return av.localeCompare(bv, undefined, { numeric: true, sensitivity: "base" }) * dir;
    });
  }

  private _processedRows(): string[][] {
    return this._sortedRows(this._filteredRows());
  }

  private _showSearch(): boolean {
    return this._rawRows().length >= 15;
  }

  private _isVirtual(rowCount: number): boolean {
    return rowCount > VIRTUAL_THRESHOLD;
  }

  // ── Sort handlers ──

  private _onSort(colIndex: number) {
    if (this._sortCol === colIndex) {
      // Cycle: asc -> desc -> none
      if (this._sortDir === "asc") {
        this._sortDir = "desc";
      } else {
        this._sortCol = -1;
        this._sortDir = null;
      }
    } else {
      this._sortCol = colIndex;
      this._sortDir = "asc";
    }
    this._resetScroll();
  }

  private _onResetSort() {
    this._sortCol = -1;
    this._sortDir = null;
  }

  // ── Drag & drop column reorder ──

  private _onDragStart(e: DragEvent, colIdx: number) {
    this._dragColIdx = colIdx;
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = "move";
      e.dataTransfer.setData("text/plain", String(colIdx));
    }
    // Mark dragging column after a tick so the drag image captures first
    requestAnimationFrame(() => this.requestUpdate());
  }

  private _onDragOver(e: DragEvent, colIdx: number) {
    e.preventDefault();
    if (e.dataTransfer) e.dataTransfer.dropEffect = "move";
    if (colIdx === this._dragColIdx) return;

    // Determine left/right side
    const th = (e.currentTarget as HTMLElement);
    const rect = th.getBoundingClientRect();
    const mid = rect.left + rect.width / 2;
    const side = e.clientX < mid ? "left" : "right";

    if (this._dragOverColIdx !== colIdx || this._dragOverSide !== side) {
      this._dragOverColIdx = colIdx;
      this._dragOverSide = side;
      this.requestUpdate();
    }
  }

  private _onDragLeave(e: DragEvent, colIdx: number) {
    // Only clear if actually leaving this column
    const related = e.relatedTarget as HTMLElement | null;
    if (related && (e.currentTarget as HTMLElement).contains(related)) return;
    if (this._dragOverColIdx === colIdx) {
      this._dragOverColIdx = -1;
      this.requestUpdate();
    }
  }

  private _onDrop(e: DragEvent, dropColIdx: number) {
    e.preventDefault();
    const fromIdx = this._dragColIdx;
    if (fromIdx < 0 || fromIdx === dropColIdx) {
      this._clearDrag();
      return;
    }

    const order = [...this._getColumnOrder()];
    const item = order.splice(fromIdx, 1)[0];

    // Calculate insert position based on side
    let insertIdx = dropColIdx;
    if (fromIdx < dropColIdx) {
      insertIdx = this._dragOverSide === "right" ? dropColIdx : dropColIdx - 1;
    } else {
      insertIdx = this._dragOverSide === "left" ? dropColIdx : dropColIdx + 1;
    }
    insertIdx = Math.max(0, Math.min(order.length, insertIdx));
    order.splice(insertIdx, 0, item);

    // Adjust sort column to follow the moved column
    if (this._sortCol >= 0) {
      if (this._sortCol === fromIdx) {
        this._sortCol = insertIdx;
      } else {
        // Recalculate: find where old sortCol ended up
        const oldOrder = this._getColumnOrder();
        const origSortIdx = oldOrder[this._sortCol];
        this._sortCol = order.indexOf(origSortIdx);
      }
    }

    this._setColumnOrder(order);
    this._clearDrag();
  }

  private _onDragEnd() {
    this._clearDrag();
  }

  private _clearDrag() {
    this._dragColIdx = -1;
    this._dragOverColIdx = -1;
    this.requestUpdate();
  }

  // ── Event handlers ──

  private _resetScroll() {
    this._scrollTop = 0;
    if (this._viewportEl) this._viewportEl.scrollTop = 0;
  }

  private _onSearch(e: Event) {
    this._search = (e.target as HTMLInputElement).value;
    this._resetScroll();
  }

  private _isEmptyTab(tab: TabData): boolean {
    return /\(0\)\s*$/.test(tab.name) || (tab.rows?.length ?? 0) === 0;
  }

  private _visibleTabs(): { tab: TabData; originalIndex: number }[] {
    if (!this.tabs?.length) return [];
    return this.tabs
      .map((tab, i) => ({ tab, originalIndex: i }))
      .filter(({ tab }) => !this._hideEmpty || !this._isEmptyTab(tab));
  }

  private _hasEmptyTabs(): boolean {
    return this.tabs?.some((tab) => this._isEmptyTab(tab)) ?? false;
  }

  private _onToggleEmpty() {
    this._hideEmpty = !this._hideEmpty;
    const visible = this._visibleTabs();
    if (visible.length && !visible.some((v) => v.originalIndex === this._activeTab)) {
      this._activeTab = visible[0].originalIndex;
      this._search = "";
    }
  }

  private _onTabClick(index: number) {
    this._activeTab = index;
    this._search = "";
    this._sortCol = -1;
    this._sortDir = null;
    this._resetScroll();
  }

  // ── Render helpers ──

  private _renderTh(headerText: string, colIdx: number) {
    const isSorted = this._sortCol === colIdx;
    const isDragging = this._dragColIdx === colIdx;
    const isDragOver = this._dragOverColIdx === colIdx && this._dragColIdx !== colIdx;

    let cls = "sortable";
    if (isSorted) cls += " sort-active";
    if (isDragging) cls += " dragging";
    if (isDragOver) cls += ` drag-over-${this._dragOverSide}`;

    let arrow = "\u2195"; // ↕ default
    if (isSorted && this._sortDir === "asc") arrow = "\u2191"; // ↑
    if (isSorted && this._sortDir === "desc") arrow = "\u2193"; // ↓

    return html`
      <th
        class="${cls}"
        draggable="true"
        @click=${() => this._onSort(colIdx)}
        @dragstart=${(e: DragEvent) => this._onDragStart(e, colIdx)}
        @dragover=${(e: DragEvent) => this._onDragOver(e, colIdx)}
        @dragleave=${(e: DragEvent) => this._onDragLeave(e, colIdx)}
        @drop=${(e: DragEvent) => this._onDrop(e, colIdx)}
        @dragend=${() => this._onDragEnd()}
      >
        ${headerText}<span class="sort-indicator">${arrow}</span>
      </th>
    `;
  }

  private _renderHeader(filtered: string[][]) {
    const hasSortOrReorder = this._sortDir !== null;
    return this.title || this._showSearch()
      ? html`
          <div class="table-header">
            <span class="table-title">
              ${this.title}
              ${filtered.length > VIRTUAL_THRESHOLD
                ? html`<span class="row-count"
                    >${filtered.length.toLocaleString()} rows</span
                  >`
                : nothing}
            </span>
            <div style="display:flex;align-items:center;gap:0.5rem;">
              ${hasSortOrReorder
                ? html`
                    <button class="reset-sort" @click=${this._onResetSort}>
                      Reset sort
                    </button>
                  `
                : nothing}
              ${this._hasEmptyTabs()
                ? html`
                    <button
                      class="toggle-empty ${this._hideEmpty ? "active" : ""}"
                      @click=${this._onToggleEmpty}
                    >
                      ${this._hideEmpty ? "Show empty" : "Hide empty"}
                    </button>
                  `
                : nothing}
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
          </div>
        `
      : nothing;
  }

  private _renderTabs() {
    return this.tabs && this.tabs.length > 0
      ? html`
          <div class="tab-bar">
            ${this._visibleTabs().map(
              ({ tab, originalIndex }) => html`
                <button
                  class="tab ${this._activeTab === originalIndex ? "active" : ""}"
                  @click=${() => this._onTabClick(originalIndex)}
                >
                  ${tab.name}
                </button>
              `
            )}
          </div>
        `
      : nothing;
  }

  /** Standard table for small datasets */
  private _renderStandardTable(
    hdrs: string[],
    rows: string[][],
    hl: Set<number>
  ) {
    return html`
      <div class="table-scroll">
        ${rows.length > 0
          ? html`
              <table>
                <thead>
                  <tr>
                    ${hdrs.map((h, i) => this._renderTh(h, i))}
                  </tr>
                </thead>
                <tbody>
                  ${rows.map(
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
    `;
  }

  /** Virtual-scrolled table for large datasets */
  private _renderVirtualTable(
    hdrs: string[],
    rows: string[][],
    hl: Set<number>
  ) {
    const totalRows = rows.length;
    if (totalRows === 0) {
      return html`<div class="no-results">
        ${this._search ? "No matching rows." : "No data."}
      </div>`;
    }

    const totalHeight = totalRows * ROW_HEIGHT;
    const visibleCount = Math.ceil(VIEWPORT_HEIGHT / ROW_HEIGHT);
    const startIndex = Math.max(
      0,
      Math.floor(this._scrollTop / ROW_HEIGHT) - OVERSCAN
    );
    const endIndex = Math.min(
      totalRows,
      startIndex + visibleCount + OVERSCAN * 2
    );
    const offsetY = startIndex * ROW_HEIGHT;
    const visibleRows = rows.slice(startIndex, endIndex);

    return html`
      <div class="v-container">
        <div class="v-head">
          <table>
            <thead>
              <tr>
                ${hdrs.map((h, i) => this._renderTh(h, i))}
              </tr>
            </thead>
          </table>
        </div>
        <div class="v-viewport">
          <div class="v-runway" style="height:${totalHeight}px">
            <div class="v-body" style="transform:translateY(${offsetY}px)">
              <table>
                <tbody>
                  ${visibleRows.map(
                    (row, i) => html`
                      <tr
                        class="${hl.has(startIndex + i) ? "highlighted" : ""}"
                      >
                        ${row.map((cell) => html`<td>${cell}</td>`)}
                      </tr>
                    `
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    `;
  }

  render() {
    const hdrs = this._currentHeaders();
    const rows = this._processedRows();
    const hl = new Set(this._currentHighlight());
    const useVirtual = this._isVirtual(rows.length);

    return html`
      <div class="table-card">
        ${this._renderHeader(rows)} ${this._renderTabs()}
        ${useVirtual
          ? this._renderVirtualTable(hdrs, rows, hl)
          : this._renderStandardTable(hdrs, rows, hl)}
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-ui-table": HairaTable;
  }
}
