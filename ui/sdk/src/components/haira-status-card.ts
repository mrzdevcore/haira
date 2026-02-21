import { BaseComponent, animateInCSS, cardCSS, icons } from "../core";
import type { StatusCardProps } from "../core/types";

const statusIcons: Record<string, string> = {
  success: icons.statusSuccess,
  error: icons.statusError,
  warning: icons.statusWarning,
  info: icons.statusInfo,
};

const statusColors: Record<string, string> = {
  success: "var(--haira-success)",
  error: "var(--haira-error)",
  warning: "var(--haira-warn)",
  info: "var(--haira-info)",
};

export class HairaStatusCard extends BaseComponent<StatusCardProps> {
  protected render() {
    return `
      <div class="card" id="card">
        <div class="header" id="header">
          <span class="icon" id="icon"></span>
          <span class="title" id="title"></span>
        </div>
        <div class="message" id="message"></div>
        <div class="sections" id="sections" style="display:none"></div>
      </div>`;
  }

  protected styles() {
    return `
      ${animateInCSS}
      .card { ${cardCSS} }
      .header {
        display: flex; align-items: center; gap: 0.4rem;
        padding: 0.45rem 0.75rem;
      }
      .icon { display: flex; align-items: center; flex-shrink: 0; }
      .icon svg { width: 14px; height: 14px; }
      .title { font-size: 0.78rem; font-weight: 600; }
      .message {
        font-size: 0.75rem; color: var(--haira-text-dim);
        padding: 0 0.75rem 0.45rem 2rem; line-height: 1.4;
      }
      .sections { border-top: 1px solid var(--haira-border); }
      .section {
        padding: 0.4rem 0.75rem;
        border-bottom: 1px solid var(--haira-border);
      }
      .section:last-child { border-bottom: none; }
      .section-label {
        font-size: 0.68rem; font-weight: 600; color: var(--haira-muted);
        text-transform: uppercase; letter-spacing: 0.04em; margin-bottom: 0.2rem;
      }
      .section-content {
        font-size: 0.75rem; color: var(--haira-text-dim);
        line-height: 1.4; white-space: pre-wrap;
      }
      .section-content.code {
        font-family: var(--haira-mono); font-size: 0.72rem;
        background: var(--haira-bg); padding: 0.35rem 0.6rem;
        border-radius: var(--haira-radius-sm); overflow-x: auto;
      }
      .card.inline .header { padding: 0.35rem 0.65rem; }
      .card.inline .message { display: inline; padding: 0; margin-left: 0.15rem; font-weight: 400; }
      .card.inline .header-row { display: flex; align-items: center; gap: 0.4rem; flex-wrap: wrap; }`;
  }

  protected onUpdate() {
    const { status = "info", title = "", message, sections } = this.props;
    const color = statusColors[status] || statusColors.info;
    const hasSections = sections && sections.length > 0;
    const hasMessage = !!message;

    const card = this.$("card");
    const header = this.$("header");

    // Inline variant: simple status with title + message, no sections
    if (!hasSections && hasMessage) {
      card.classList.add("inline");
      header.innerHTML = `
        <div class="header-row">
          <span class="icon" id="icon"></span>
          <span class="title" id="title"></span>
          <span class="message" id="message"></span>
        </div>`;
      const iconEl = header.querySelector("#icon")! as HTMLElement;
      iconEl.innerHTML = statusIcons[status] || statusIcons.info;
      iconEl.style.color = color;
      const titleEl = header.querySelector("#title")! as HTMLElement;
      titleEl.textContent = title;
      titleEl.style.color = color;
      (header.querySelector("#message")! as HTMLElement).textContent = message!;
      this.$("message").style.display = "none";
      this.$("sections").style.display = "none";
      card.style.borderLeft = `3px solid ${color}`;
      return;
    }

    card.classList.remove("inline");

    const iconEl = this.$("icon");
    iconEl.innerHTML = statusIcons[status] || statusIcons.info;
    iconEl.style.color = color;

    const titleEl = this.$("title");
    titleEl.textContent = title;
    titleEl.style.color = color;

    const messageEl = this.$("message");
    if (hasMessage) {
      messageEl.textContent = message!;
      messageEl.style.display = "";
    } else {
      messageEl.style.display = "none";
    }

    card.style.borderLeft = `3px solid ${color}`;

    const sectionsEl = this.$("sections");
    if (hasSections) {
      sectionsEl.style.display = "";
      sectionsEl.innerHTML = sections!
        .map((s) => `
          <div class="section">
            <div class="section-label">${this.esc(s.label || "")}</div>
            <div class="section-content ${s.style === "code" ? "code" : ""}">${this.esc(s.content || "")}</div>
          </div>`)
        .join("");
    } else {
      sectionsEl.style.display = "none";
    }
  }
}
