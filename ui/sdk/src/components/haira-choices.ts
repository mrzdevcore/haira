import { BaseComponent, animateInCSS, cardCSS } from "../core";
import type { ChoicesProps } from "../core/types";

export class HairaChoices extends BaseComponent<ChoicesProps> {
  private answered = false;

  protected render() {
    return `
      <div class="card">
        <div class="title" id="title"></div>
        <div id="options"></div>
      </div>`;
  }

  protected styles() {
    return `
      ${animateInCSS}
      .card { ${cardCSS} padding: 0.55rem 0.75rem; }
      .title { font-size: 0.78rem; font-weight: 600; color: var(--haira-text); margin-bottom: 0.45rem; }
      .options-buttons { display: flex; flex-wrap: wrap; gap: 0.35rem; }
      .opt-btn {
        background: transparent; border: 1px solid var(--haira-border);
        color: var(--haira-text-dim); font-family: var(--haira-font);
        font-size: 0.73rem; padding: 0.3rem 0.7rem; border-radius: 16px;
        cursor: pointer; transition: all 0.15s;
      }
      .opt-btn:hover { border-color: var(--haira-accent); color: var(--haira-accent); background: var(--haira-accent-dim); }
      .opt-btn:disabled { opacity: 0.35; cursor: default; pointer-events: none; }
      .opt-btn.selected { opacity: 1; background: var(--haira-accent); color: #1a0e04; border-color: var(--haira-accent); }
      .options-list { display: flex; flex-direction: column; gap: 0.15rem; }
      .opt-row {
        display: flex; align-items: center; gap: 0.45rem;
        padding: 0.35rem 0.5rem; border-radius: 6px; cursor: pointer;
        transition: background 0.15s; font-size: 0.75rem; color: var(--haira-text-dim);
      }
      .opt-row:hover { background: var(--haira-bg-card-hover); }
      .opt-radio {
        width: 14px; height: 14px; border-radius: 50%;
        border: 2px solid var(--haira-border); flex-shrink: 0;
        transition: all 0.15s; display: flex; align-items: center; justify-content: center;
      }
      .opt-row:hover .opt-radio { border-color: var(--haira-accent); }
      .opt-row.selected .opt-radio { border-color: var(--haira-accent); background: var(--haira-accent); }
      .opt-row.selected .opt-radio::after { content: ""; width: 5px; height: 5px; border-radius: 50%; background: #1a0e04; }
      .opt-row.disabled { opacity: 0.35; cursor: default; pointer-events: none; }
      .opt-row.selected.disabled { opacity: 1; }`;
  }

  protected onUpdate() {
    const { title = "Choose an option", options = [], style = "buttons", _restored } = this.props;

    this.$("title").textContent = title;
    const container = this.$("options");

    if (_restored) this.answered = true;

    if (style === "list") {
      container.className = "options-list";
      container.innerHTML = options
        .map((opt) => `<div class="opt-row${_restored ? " disabled" : ""}" data-value="${this.escAttr(opt)}">
          <span class="opt-radio"></span><span>${this.esc(opt)}</span>
        </div>`)
        .join("");

      if (!_restored) {
        container.querySelectorAll(".opt-row").forEach((row) => {
          row.addEventListener("click", () => this.selectOption((row as HTMLElement).dataset.value || "", container, "list"));
        });
      }
    } else {
      container.className = "options-buttons";
      container.innerHTML = options
        .map((opt) => `<button class="opt-btn" data-value="${this.escAttr(opt)}"${_restored ? " disabled" : ""}>${this.esc(opt)}</button>`)
        .join("");

      if (!_restored) {
        container.querySelectorAll(".opt-btn").forEach((btn) => {
          btn.addEventListener("click", () => this.selectOption((btn as HTMLElement).dataset.value || "", container, "buttons"));
        });
      }
    }
  }

  private selectOption(value: string, container: HTMLElement, style: string) {
    if (this.answered) return;
    this.answered = true;

    if (style === "list") {
      container.querySelectorAll(".opt-row").forEach((row) => {
        const el = row as HTMLElement;
        if (el.dataset.value === value) el.classList.add("selected", "disabled");
        else el.classList.add("disabled");
      });
    } else {
      container.querySelectorAll(".opt-btn").forEach((btn) => {
        const el = btn as HTMLButtonElement;
        if (el.dataset.value === value) el.classList.add("selected");
        el.disabled = true;
      });
    }

    const title = this.$("title")?.textContent || "";
    this.emit("haira-chat-input", { text: `[User selected "${value}" from choices: ${title}]` });
  }
}
