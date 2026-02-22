import { BaseComponent, animateInCSS, scrollbarCSS, cardCSS, icons } from "../core";
import type { TableProps, TabData } from "../core/types";

const SCROLL_THRESHOLD = 15;

export class HairaTable extends BaseComponent<TableProps> {
  private allRows: string[][] = [];
  private filteredRows: { row: string[]; idx: number }[] = [];
  private headers: string[] = [];
  private highlight: Set<number> = new Set();
  private searchTerm = "";
  private tabs: TabData[] = [];
  private activeTab = 0;

  protected render() {
    return `
      <div class="card">
        <div class="toolbar" id="toolbar" style="display:none">
          <span class="toolbar-title" id="title"></span>
          <span class="row-count" id="row-count"></span>
          <div class="search-wrap" id="search-wrap" style="display:none">
            ${icons.search}
            <input class="search" id="search" type="text" placeholder="Filter rows..." />
          </div>
        </div>
        <div class="tab-bar" id="tab-bar" style="display:none"></div>
        <div class="table-scroll" id="scroll">
          <table>
            <thead id="thead"></thead>
            <tbody id="tbody"></tbody>
          </table>
        </div>
        <div class="footer" id="footer" style="display:none"></div>
      </div>`;
  }

  protected styles() {
    return `
      ${animateInCSS}
      .card { ${cardCSS} }
      .toolbar {
        display: flex; align-items: center; gap: 0.5rem;
        padding: 0.5rem 0.75rem; border-bottom: 1px solid var(--haira-border);
      }
      .toolbar-title { font-size: 0.78rem; font-weight: 600; color: var(--haira-text); white-space: nowrap; }
      .row-count {
        font-size: 0.68rem; color: var(--haira-muted); background: var(--haira-bg);
        padding: 0.15rem 0.45rem; border-radius: 9px; white-space: nowrap; flex-shrink: 0;
      }
      .search-wrap { margin-left: auto; position: relative; flex-shrink: 0; }
      .search-wrap svg { position: absolute; left: 0.45rem; top: 50%; transform: translateY(-50%); color: var(--haira-muted); pointer-events: none; }
      .search {
        background: var(--haira-bg); border: 1px solid var(--haira-border);
        color: var(--haira-text); font-size: 0.72rem; font-family: var(--haira-font);
        padding: 0.28rem 0.5rem 0.28rem 1.6rem; border-radius: 6px; width: 160px;
        outline: none; transition: border-color 0.15s;
      }
      .search:focus { border-color: var(--haira-accent); }
      .search::placeholder { color: var(--haira-muted); }
      .tab-bar {
        display: flex; gap: 0; border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg); overflow-x: auto; scrollbar-width: none;
      }
      .tab-bar::-webkit-scrollbar { display: none; }
      .tab {
        padding: 0.4rem 0.75rem; font-size: 0.72rem; font-family: var(--haira-font);
        color: var(--haira-muted); background: none; border: none;
        border-bottom: 2px solid transparent; cursor: pointer;
        white-space: nowrap; transition: color 0.15s, border-color 0.15s; flex-shrink: 0;
      }
      .tab:hover { color: var(--haira-text); }
      .tab.active { color: var(--haira-accent); border-bottom-color: var(--haira-accent); font-weight: 600; }
      .table-scroll { overflow: auto; ${scrollbarCSS} }
      .table-scroll.capped { max-height: 420px; }
      table { width: 100%; border-collapse: collapse; font-size: 0.74rem; }
      th {
        text-align: left; padding: 0.35rem 0.65rem; font-weight: 600;
        font-size: 0.68rem; color: var(--haira-muted); text-transform: uppercase;
        letter-spacing: 0.04em; background: var(--haira-bg);
        border-bottom: 1px solid var(--haira-border); white-space: nowrap;
        position: sticky; top: 0; z-index: 1;
      }
      td {
        padding: 0.28rem 0.65rem; color: var(--haira-text-dim);
        border-bottom: 1px solid var(--haira-border); line-height: 1.35;
        max-width: 320px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
      }
      td:hover { white-space: normal; word-break: break-all; }
      tr:last-child td { border-bottom: none; }
      tr:hover td { background: var(--haira-bg-card-hover); }
      tr.highlight td { background: rgba(232, 163, 23, 0.06); }
      .no-results { padding: 1.5rem; text-align: center; color: var(--haira-muted); font-size: 0.75rem; }
      .footer {
        padding: 0.3rem 0.75rem; border-top: 1px solid var(--haira-border);
        font-size: 0.68rem; color: var(--haira-muted); text-align: right;
      }
      @media (max-width: 640px) {
        .toolbar { flex-wrap: wrap; }
        .search-wrap { margin-left: 0; width: 100%; }
        .search { width: 100%; }
        td { max-width: 200px; }
      }`;
  }

  protected onMount() {
    this.$("search").addEventListener("input", (e) => {
      this.searchTerm = (e.target as HTMLInputElement).value.toLowerCase();
      this.applyFilter();
    });
  }

  protected onUpdate() {
    try {
      const { title, tabs } = this.props;

      if (tabs && tabs.length > 0) {
        this.tabs = tabs;
        this.activeTab = 0;
        this.$("toolbar").style.display = "";
        this.$("title").textContent = title || "Table";
        this.renderTabBar();
        this.loadTab(0);
      } else {
        this.tabs = [];
        this.loadSingleTable();
      }
    } catch {
      // Graceful fallback
    }
  }

  private loadSingleTable() {
    const { title, headers = [], rows = [], highlight = [] } = this.props;
    this.headers = headers;
    this.allRows = rows;
    this.highlight = new Set(highlight);

    const toolbar = this.$("toolbar");
    const titleEl = this.$("title");
    const countEl = this.$("row-count");
    const searchWrap = this.$("search-wrap");

    const hasTitle = !!title;
    const hasMany = this.allRows.length >= SCROLL_THRESHOLD;

    if (hasTitle || hasMany) {
      toolbar.style.display = "";
      titleEl.textContent = title || "Table";
      countEl.textContent = `${this.allRows.length} rows`;
    }

    if (hasMany) {
      searchWrap.style.display = "";
    }

    this.$("scroll").classList.toggle("capped", hasMany);
    this.$("thead").innerHTML = `<tr>${this.headers.map((h) => `<th>${this.esc(h)}</th>`).join("")}</tr>`;

    this.searchTerm = "";
    (this.$("search") as HTMLInputElement).value = "";
    this.applyFilter();
  }

  private renderTabBar() {
    const tabBar = this.$("tab-bar");
    tabBar.style.display = "flex";
    tabBar.innerHTML = "";

    for (let i = 0; i < this.tabs.length; i++) {
      const tab = this.tabs[i];
      const btn = document.createElement("button");
      btn.className = `tab${i === this.activeTab ? " active" : ""}`;
      btn.textContent = `${tab.name} (${tab.rows.length})`;
      btn.addEventListener("click", () => this.loadTab(i));
      tabBar.appendChild(btn);
    }
  }

  private loadTab(index: number) {
    this.activeTab = index;
    const tab = this.tabs[index];

    this.headers = tab.headers || [];
    this.allRows = tab.rows || [];
    this.highlight = new Set((tab.highlight as number[]) || []);

    this.$("tab-bar").querySelectorAll(".tab").forEach((btn, i) => {
      btn.classList.toggle("active", i === index);
    });

    this.$("row-count").textContent = `${this.allRows.length} rows`;

    const hasMany = this.allRows.length >= SCROLL_THRESHOLD;
    this.$("search-wrap").style.display = hasMany ? "" : "none";
    this.$("scroll").classList.toggle("capped", hasMany);
    this.$("thead").innerHTML = `<tr>${this.headers.map((h) => `<th>${this.esc(h)}</th>`).join("")}</tr>`;

    this.searchTerm = "";
    (this.$("search") as HTMLInputElement).value = "";
    this.applyFilter();
  }

  private applyFilter() {
    if (this.searchTerm) {
      this.filteredRows = this.allRows
        .map((row, idx) => ({ row, idx }))
        .filter(({ row }) =>
          row.some((cell) => String(cell).toLowerCase().includes(this.searchTerm)),
        );
    } else {
      this.filteredRows = this.allRows.map((row, idx) => ({ row, idx }));
    }
    this.renderRows();
  }

  private renderRows() {
    const tbody = this.$("tbody");
    const footer = this.$("footer");
    const countEl = this.$("row-count");

    if (this.filteredRows.length === 0 && this.searchTerm) {
      tbody.innerHTML = `<tr><td colspan="${this.headers.length || 1}" class="no-results">No matching rows</td></tr>`;
    } else {
      tbody.innerHTML = this.filteredRows
        .map(
          ({ row, idx }) =>
            `<tr class="${this.highlight.has(idx) ? "highlight" : ""}">${row.map((c) => `<td title="${this.esc(String(c))}">${this.esc(String(c))}</td>`).join("")}</tr>`,
        )
        .join("");
    }

    if (this.searchTerm) {
      countEl.textContent = `${this.filteredRows.length} / ${this.allRows.length} rows`;
      footer.style.display = "";
      footer.textContent = `Showing ${this.filteredRows.length} of ${this.allRows.length} rows`;
    } else {
      countEl.textContent = `${this.allRows.length} rows`;
      footer.style.display = "none";
    }
  }
}
