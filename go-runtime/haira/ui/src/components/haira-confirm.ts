import { baseStyles, sharedKeyframes } from "../theme";

export class HairaConfirm extends HTMLElement {
  private answered = false;

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
          border-left: 3px solid var(--haira-info);
          border-radius: var(--haira-radius);
          padding: 0.55rem 0.75rem;
        }
        .title {
          font-size: 0.78rem;
          font-weight: 600;
          color: var(--haira-text);
          margin-bottom: 0.15rem;
        }
        .message {
          font-size: 0.73rem;
          color: var(--haira-text-dim);
          line-height: 1.4;
          margin-bottom: 0.5rem;
        }
        .actions {
          display: flex;
          gap: 0.4rem;
        }
        button {
          font-family: var(--haira-font);
          font-size: 0.73rem;
          font-weight: 600;
          padding: 0.32rem 0.85rem;
          border-radius: 6px;
          cursor: pointer;
          transition: all 0.15s;
        }
        .confirm-btn {
          background: var(--haira-accent);
          color: #1a0e04;
          border: none;
        }
        .confirm-btn:hover { background: var(--haira-accent-light); }
        .deny-btn {
          background: transparent;
          color: var(--haira-text-dim);
          border: 1px solid var(--haira-border);
        }
        .deny-btn:hover {
          border-color: var(--haira-text-dim);
          color: var(--haira-text);
        }
        button:disabled {
          opacity: 0.4;
          cursor: default;
          pointer-events: none;
        }
        button.selected {
          opacity: 1;
        }
        .selected-indicator {
          display: none;
          font-size: 0.68rem;
          color: var(--haira-muted);
          margin-top: 0.35rem;
        }
        .selected-indicator.visible { display: block; }
      </style>
      <div class="card" id="card">
        <div class="title" id="title"></div>
        <div class="message" id="message"></div>
        <div class="actions" id="actions">
          <button class="confirm-btn" id="confirm-btn"></button>
          <button class="deny-btn" id="deny-btn"></button>
        </div>
        <div class="selected-indicator" id="indicator"></div>
      </div>
    `;
  }

  setProps(props: Record<string, unknown>) {
    try {
      const title = this.shadowRoot!.getElementById("title")!;
      title.textContent = (props.title as string) || "Confirm";

      const message = this.shadowRoot!.getElementById("message")!;
      if (props.message) {
        message.textContent = props.message as string;
        message.style.display = "";
      } else {
        message.style.display = "none";
      }

      const confirmLabel = (props.confirm_label as string) || "Confirm";
      const denyLabel = (props.deny_label as string) || "Cancel";

      const confirmBtn = this.shadowRoot!.getElementById(
        "confirm-btn",
      ) as HTMLButtonElement;
      const denyBtn = this.shadowRoot!.getElementById(
        "deny-btn",
      ) as HTMLButtonElement;
      confirmBtn.textContent = confirmLabel;
      denyBtn.textContent = denyLabel;

      if (props._restored) {
        this.answered = true;
        confirmBtn.disabled = true;
        denyBtn.disabled = true;
      } else {
        confirmBtn.onclick = () => this.select(confirmLabel, confirmBtn, denyBtn);
        denyBtn.onclick = () => this.select(denyLabel, denyBtn, confirmBtn);
      }
    } catch {
      // Graceful fallback
    }
  }

  private select(
    text: string,
    chosen: HTMLButtonElement,
    other: HTMLButtonElement,
  ) {
    if (this.answered) return;
    this.answered = true;

    chosen.classList.add("selected");
    chosen.disabled = true;
    other.disabled = true;

    const indicator = this.shadowRoot!.getElementById("indicator")!;
    indicator.textContent = `Selected: ${text}`;
    indicator.classList.add("visible");

    // Include context so the LLM knows this is a UI response, not a new request
    const title = this.shadowRoot!.getElementById("title")?.textContent || "";
    const contextMsg = `[User clicked "${text}" on confirmation: ${title}]`;

    this.dispatchEvent(
      new CustomEvent("haira-chat-input", {
        detail: { text: contextMsg },
        bubbles: true,
        composed: true,
      }),
    );
  }
}
