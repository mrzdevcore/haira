import { baseStyles, sharedKeyframes, scrollbarStyles } from "../theme";

const SCROLL_THRESHOLD = 15;

export class HairaTable extends HTMLElement {
  private allRows: string[][] = [];
  private filteredRows: { row: string[]; idx: number }[] = [];
  private headers: string[] = [];
  private highlight: Set<number> = new Set();
  private searchTerm = "";

  connectedCallback() {
    this.attachShadow({ mode: "open" });
    this.shadowRoot!.innerHTML = `
      <style>
        ${baseStyles}
        ${sharedKeyframes}
        :host {
          display: block;
          animation: fadeSlideUp 0.25s ease-out;
        }
        .card {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          overflow: hidden;
        }
        .toolbar {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          padding: 0.5rem 0.75rem;
          border-bottom: 1px solid var(--haira-border);
        }
        .toolbar-title {
          font-size: 0.78rem;
          font-weight: 600;
          color: var(--haira-text);
          white-space: nowrap;
        }
        .row-count {
          font-size: 0.68rem;
          color: var(--haira-muted);
          background: var(--haira-bg);
          padding: 0.15rem 0.45rem;
          border-radius: 9px;
          white-space: nowrap;
          flex-shrink: 0;
        }
        .search-wrap {
          margin-left: auto;
          position: relative;
          flex-shrink: 0;
        }
        .search-wrap svg {
          position: absolute;
          left: 0.45rem;
          top: 50%;
          transform: translateY(-50%);
          color: var(--haira-muted);
          pointer-events: none;
        }
        .search {
          background: var(--haira-bg);
          border: 1px solid var(--haira-border);
          color: var(--haira-text);
          font-size: 0.72rem;
          font-family: var(--haira-font);
          padding: 0.28rem 0.5rem 0.28rem 1.6rem;
          border-radius: 6px;
          width: 160px;
          outline: none;
          transition: border-color 0.15s;
        }
        .search:focus { border-color: var(--haira-accent); }
        .search::placeholder { color: var(--haira-muted); }
        .table-scroll {
          overflow: auto;
          ${scrollbarStyles}
        }
        .table-scroll.capped {
          max-height: 420px;
        }
        table {
          width: 100%;
          border-collapse: collapse;
          font-size: 0.74rem;
        }
        th {
          text-align: left;
          padding: 0.35rem 0.65rem;
          font-weight: 600;
          font-size: 0.68rem;
          color: var(--haira-muted);
          text-transform: uppercase;
          letter-spacing: 0.04em;
          background: var(--haira-bg);
          border-bottom: 1px solid var(--haira-border);
          white-space: nowrap;
          position: sticky;
          top: 0;
          z-index: 1;
        }
        td {
          padding: 0.28rem 0.65rem;
          color: var(--haira-text-dim);
          border-bottom: 1px solid var(--haira-border);
          line-height: 1.35;
          max-width: 320px;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
        td:hover { white-space: normal; word-break: break-all; }
        tr:last-child td { border-bottom: none; }
        tr:hover td { background: var(--haira-bg-card-hover); }
        tr.highlight td { background: rgba(232, 163, 23, 0.06); }
        .no-results {
          padding: 1.5rem;
          text-align: center;
          color: var(--haira-muted);
          font-size: 0.75rem;
        }
        .footer {
          padding: 0.3rem 0.75rem;
          border-top: 1px solid var(--haira-border);
          font-size: 0.68rem;
          color: var(--haira-muted);
          text-align: right;
        }
        @media (max-width: 640px) {
          .toolbar { flex-wrap: wrap; }
          .search-wrap { margin-left: 0; width: 100%; }
          .search { width: 100%; }
          td { max-width: 200px; }
        }
      </style>
      <div class="card">
        <div class="toolbar" id="toolbar" style="display:none">
          <span class="toolbar-title" id="title"></span>
          <span class="row-count" id="row-count"></span>
          <div class="search-wrap" id="search-wrap" style="display:none">
            <svg width="13" height="13" viewBox="0 0 16 16" fill="none"><circle cx="6.5" cy="6.5" r="5" stroke="currentColor" stroke-width="1.5"/><path d="M10.5 10.5L14.5 14.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
            <input class="search" id="search" type="text" placeholder="Filter rows..." />
          </div>
        </div>
        <div class="table-scroll" id="scroll">
          <table>
            <thead id="thead"></thead>
            <tbody id="tbody"></tbody>
          </table>
        </div>
        <div class="footer" id="footer" style="display:none"></div>
      </div>
    `;

    this.shadowRoot!.getElementById("search")!.addEventListener(
      "input",
      (e) => {
        this.searchTerm = (e.target as HTMLInputElement).value.toLowerCase();
        this.applyFilter();
      },
    );
  }

  setProps(props: Record<string, unknown>) {
    try {
      const title = props.title as string | undefined;
      this.headers = (props.headers as string[]) || [];
      this.allRows = (props.rows as string[][]) || [];
      this.highlight = new Set((props.highlight as number[]) || []);

      // Toolbar: show if title or enough rows
      const toolbar = this.shadowRoot!.getElementById("toolbar")!;
      const titleEl = this.shadowRoot!.getElementById("title")!;
      const countEl = this.shadowRoot!.getElementById("row-count")!;
      const searchWrap = this.shadowRoot!.getElementById("search-wrap")!;

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

      // Sticky scroll container
      const scroll = this.shadowRoot!.getElementById("scroll")!;
      if (hasMany) {
        scroll.classList.add("capped");
      } else {
        scroll.classList.remove("capped");
      }

      // Render headers
      const thead = this.shadowRoot!.getElementById("thead")!;
      thead.innerHTML = `<tr>${this.headers.map((h) => `<th>${this.esc(h)}</th>`).join("")}</tr>`;

      this.searchTerm = "";
      (this.shadowRoot!.getElementById("search") as HTMLInputElement).value =
        "";
      this.applyFilter();
    } catch {
      // Graceful fallback
    }
  }

  private applyFilter() {
    if (this.searchTerm) {
      this.filteredRows = this.allRows
        .map((row, idx) => ({ row, idx }))
        .filter(({ row }) =>
          row.some((cell) =>
            String(cell).toLowerCase().includes(this.searchTerm),
          ),
        );
    } else {
      this.filteredRows = this.allRows.map((row, idx) => ({ row, idx }));
    }
    this.renderRows();
  }

  private renderRows() {
    const tbody = this.shadowRoot!.getElementById("tbody")!;
    const footer = this.shadowRoot!.getElementById("footer")!;
    const countEl = this.shadowRoot!.getElementById("row-count")!;

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

    // Update count badge
    if (this.searchTerm) {
      countEl.textContent = `${this.filteredRows.length} / ${this.allRows.length} rows`;
      footer.style.display = "";
      footer.textContent = `Showing ${this.filteredRows.length} of ${this.allRows.length} rows`;
    } else {
      countEl.textContent = `${this.allRows.length} rows`;
      footer.style.display = "none";
    }
  }

  private esc(s: string): string {
    return s
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }
}
