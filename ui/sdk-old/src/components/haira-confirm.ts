import { BaseComponent, animateInCSS, cardCSS } from "../core";
import type { ConfirmProps } from "../core/types";

export class HairaConfirm extends BaseComponent<ConfirmProps> {
  private answered = false;

  protected render() {
    return `
      <div class="card" id="card">
        <div class="title" id="title"></div>
        <div class="message" id="message"></div>
        <div class="actions" id="actions">
          <button class="confirm-btn" id="confirm-btn"></button>
          <button class="deny-btn" id="deny-btn"></button>
        </div>
        <div class="selected-indicator" id="indicator"></div>
      </div>`;
  }

  protected styles() {
    return `
      ${animateInCSS}
      .card {
        ${cardCSS}
        border-left: 3px solid var(--haira-info);
        padding: 0.55rem 0.75rem;
      }
      .title { font-size: 0.78rem; font-weight: 600; color: var(--haira-text); margin-bottom: 0.15rem; }
      .message { font-size: 0.73rem; color: var(--haira-text-dim); line-height: 1.4; margin-bottom: 0.5rem; }
      .actions { display: flex; gap: 0.4rem; }
      button {
        font-family: var(--haira-font); font-size: 0.73rem; font-weight: 600;
        padding: 0.32rem 0.85rem; border-radius: 6px; cursor: pointer; transition: all 0.15s;
      }
      .confirm-btn { background: var(--haira-accent); color: #1a0e04; border: none; }
      .confirm-btn:hover { background: var(--haira-accent-light); }
      .deny-btn { background: transparent; color: var(--haira-text-dim); border: 1px solid var(--haira-border); }
      .deny-btn:hover { border-color: var(--haira-text-dim); color: var(--haira-text); }
      button:disabled { opacity: 0.4; cursor: default; pointer-events: none; }
      button.selected { opacity: 1; }
      .selected-indicator { display: none; font-size: 0.68rem; color: var(--haira-muted); margin-top: 0.35rem; }
      .selected-indicator.visible { display: block; }`;
  }

  protected onUpdate() {
    const { title = "Confirm", message, confirm_label = "Confirm", deny_label = "Cancel", _restored } = this.props;

    this.$("title").textContent = title;

    const messageEl = this.$("message");
    if (message) {
      messageEl.textContent = message;
      messageEl.style.display = "";
    } else {
      messageEl.style.display = "none";
    }

    const confirmBtn = this.$("confirm-btn") as HTMLButtonElement;
    const denyBtn = this.$("deny-btn") as HTMLButtonElement;
    confirmBtn.textContent = confirm_label;
    denyBtn.textContent = deny_label;

    if (_restored) {
      this.answered = true;
      confirmBtn.disabled = true;
      denyBtn.disabled = true;
    } else {
      confirmBtn.onclick = () => this.select(confirm_label, confirmBtn, denyBtn);
      denyBtn.onclick = () => this.select(deny_label, denyBtn, confirmBtn);
    }
  }

  private select(text: string, chosen: HTMLButtonElement, other: HTMLButtonElement) {
    if (this.answered) return;
    this.answered = true;

    chosen.classList.add("selected");
    chosen.disabled = true;
    other.disabled = true;

    const indicator = this.$("indicator");
    indicator.textContent = `Selected: ${text}`;
    indicator.classList.add("visible");

    const title = this.$("title")?.textContent || "";
    this.emit("haira-chat-input", { text: `[User clicked "${text}" on confirmation: ${title}]` });
  }
}
