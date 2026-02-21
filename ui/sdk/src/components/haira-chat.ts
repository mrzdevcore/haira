import { baseCSS, scrollbarCSS, logoSvg, icons, esc, formatBytes } from "../core";
import { streamSSE } from "../services/sse-client";
import type { WorkflowMeta, ToolRenderEvent, ChatSessionSummary, ChatSessionDetail } from "../core/types";
import type { HairaMessage } from "./haira-message";
import type { HairaToolCard } from "./haira-tool-card";
import type { HairaUIRenderer } from "./haira-ui-renderer";

export class HairaChat extends HTMLElement {
  private meta!: WorkflowMeta;
  private sessionId = "";
  private attachedFile: File | null = null;
  private streamAbort: AbortController | null = null;

  connectedCallback() {
    this.meta = JSON.parse(this.getAttribute("data-meta") || "{}");

    const url = new URL(window.location.href);
    const urlSession = url.searchParams.get("session");
    if (urlSession) {
      this.sessionId = urlSession;
      this.renderChat();
    } else {
      this.initWithLatestSession(url);
    }
  }

  private async initWithLatestSession(url: URL) {
    try {
      const resp = await fetch(
        `/_api/chats?workflow=${encodeURIComponent(this.meta.path)}`,
      );
      if (resp.ok) {
        const sessions = await resp.json();
        if (sessions && sessions.length > 0) {
          this.sessionId = sessions[0].id;
          url.searchParams.set("session", this.sessionId);
          window.history.replaceState({}, "", url.toString());
          this.renderChat();
          return;
        }
      }
    } catch {
      // API unavailable
    }
    this.sessionId = crypto.randomUUID();
    url.searchParams.set("session", this.sessionId);
    window.history.replaceState({}, "", url.toString());
    this.renderChat();
  }

  private renderChat() {
    const m = this.meta;
    const shadow = this.shadowRoot || this.attachShadow({ mode: "open" });
    shadow.innerHTML = `
      <style>
        ${baseCSS}
        :host { display: flex; flex-direction: row; flex: 1; overflow: hidden; position: relative; }

        /* Sidebar */
        .sidebar {
          width: 240px; flex-shrink: 0; display: flex; flex-direction: column;
          border-right: 1px solid var(--haira-border); background: var(--haira-bg);
          overflow: hidden; transition: width 0.2s, opacity 0.2s;
        }
        .sidebar.collapsed { width: 0; opacity: 0; pointer-events: none; }
        .sidebar-header {
          display: flex; align-items: center; gap: 0.4rem;
          padding: 0.55rem 0.65rem; border-bottom: 1px solid var(--haira-border); flex-shrink: 0;
        }
        .sidebar-title { font-size: 0.75rem; font-weight: 600; color: var(--haira-text-dim); flex: 1; }
        .sidebar-btn {
          background: none; border: none; color: var(--haira-muted); cursor: pointer;
          display: flex; align-items: center; justify-content: center; padding: 0.25rem;
          border-radius: 4px; transition: all 0.15s;
        }
        .sidebar-btn:hover { color: var(--haira-accent); background: var(--haira-accent-dim); }
        .sidebar-list {
          flex: 1; overflow-y: auto; padding: 0.35rem;
          display: flex; flex-direction: column; gap: 1px; ${scrollbarCSS}
        }
        .session-item {
          display: flex; align-items: center; gap: 0.4rem; padding: 0.45rem 0.5rem;
          border-radius: 6px; cursor: pointer; transition: all 0.12s; text-decoration: none;
          color: var(--haira-text-dim); font-size: 0.78rem; line-height: 1.35; position: relative;
        }
        .session-item:hover { background: var(--haira-bg-card); color: var(--haira-text); }
        .session-item.active { background: var(--haira-accent-dim); color: var(--haira-accent); }
        .session-icon { display: flex; flex-shrink: 0; opacity: 0.5; }
        .session-item.active .session-icon { opacity: 1; }
        .session-title { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; min-width: 0; }
        .session-delete {
          display: none; background: none; border: none; color: var(--haira-muted); cursor: pointer;
          padding: 0.15rem; border-radius: 3px; flex-shrink: 0; align-items: center; justify-content: center;
        }
        .session-item:hover .session-delete { display: flex; }
        .session-delete:hover { color: var(--haira-error); background: rgba(239, 68, 68, 0.1); }
        .sidebar-empty { padding: 1rem; text-align: center; font-size: 0.75rem; color: var(--haira-muted); opacity: 0.5; }

        .sidebar-toggle {
          position: absolute; top: 0.5rem; left: 0.5rem; z-index: 10;
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          color: var(--haira-muted); cursor: pointer; display: none;
          align-items: center; justify-content: center; padding: 0.35rem;
          border-radius: 6px; transition: all 0.15s;
        }
        .sidebar-toggle.visible { display: flex; }
        .sidebar-toggle:hover { color: var(--haira-accent); border-color: var(--haira-accent); background: var(--haira-accent-dim); }

        /* Chat main */
        .chat-main {
          flex: 1; display: flex; flex-direction: column; overflow: hidden;
          position: relative; min-width: 0; height: 100%;
        }

        /* Welcome */
        .welcome {
          flex: 1; display: flex; flex-direction: column; align-items: center;
          justify-content: center; gap: 1rem; padding: 2rem; opacity: 1; transition: opacity 0.3s;
        }
        .welcome.hidden { display: none; }
        .welcome-icon { opacity: 0.15; }
        .welcome-icon img { width: 56px; height: 56px; object-fit: contain; opacity: 1; }
        .welcome h2 { font-size: 1.1rem; font-weight: 600; color: var(--haira-text); }
        .welcome p { font-size: 0.85rem; color: var(--haira-muted); text-align: center; max-width: 420px; line-height: 1.5; }
        .suggestions { display: flex; flex-wrap: wrap; gap: 0.5rem; justify-content: center; margin-top: 0.5rem; max-width: 540px; }
        .suggestion {
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          color: var(--haira-text-dim); padding: 0.45rem 0.85rem; border-radius: 20px;
          font-size: 0.78rem; font-family: var(--haira-font); cursor: pointer; transition: all 0.15s;
        }
        .suggestion:hover { border-color: var(--haira-accent); color: var(--haira-accent); background: var(--haira-accent-dim); }

        /* Messages */
        .messages { flex: 1; min-height: 0; overflow-y: auto; display: none; flex-direction: column; ${scrollbarCSS} }
        .messages.active { display: flex; }
        .messages-inner {
          max-width: 768px; width: 100%; margin: 0 auto; padding: 1.5rem 1.25rem;
          display: flex; flex-direction: column; gap: 0.75rem;
        }

        /* Typing */
        .typing {
          display: none; padding: 0.25rem 0; align-items: center; gap: 0.4rem;
          font-size: 0.75rem; color: var(--haira-muted); margin-left: 2.25rem;
        }
        .typing.visible { display: flex; }
        .typing-dots { display: flex; gap: 0.2rem; align-items: center; }
        .typing-dot {
          display: inline-block; width: 5px; height: 5px; border-radius: 50%;
          background: var(--haira-accent); animation: bounce 1.4s ease-in-out infinite;
        }
        .typing-dot:nth-child(2) { animation-delay: 0.2s; }
        .typing-dot:nth-child(3) { animation-delay: 0.4s; }

        /* Drop overlay */
        .drop-overlay {
          display: none; position: absolute; inset: 0;
          background: rgba(9, 9, 11, 0.85); z-index: 200;
          align-items: center; justify-content: center; flex-direction: column; gap: 0.75rem;
          border: 2px dashed var(--haira-accent); border-radius: var(--haira-radius); margin: 0.5rem;
        }
        .drop-overlay.visible { display: flex; }
        .drop-overlay-icon { color: var(--haira-accent); opacity: 0.7; }
        .drop-overlay-text { color: var(--haira-accent); font-size: 0.9rem; font-weight: 600; }

        /* Input */
        .input-area {
          padding: 0.75rem 1rem 1rem; flex-shrink: 0;
          background: var(--haira-bg); border-top: 1px solid var(--haira-border);
        }
        .input-card {
          display: flex; flex-direction: column; background: var(--haira-bg-card);
          border: 1px solid var(--haira-border); border-radius: var(--haira-radius);
          transition: border-color 0.15s; max-width: 768px; margin: 0 auto;
        }
        .input-card:focus-within { border-color: var(--haira-border-focus); }
        .file-chip { display: none; align-items: center; gap: 0.4rem; padding: 0.4rem 0.6rem 0; margin: 0 0.5rem; }
        .file-chip.visible { display: flex; }
        .file-chip-inner {
          display: flex; align-items: center; gap: 0.35rem;
          background: var(--haira-bg-elevated); border: 1px solid var(--haira-border);
          border-radius: 6px; padding: 0.25rem 0.5rem; font-size: 0.75rem; color: var(--haira-text-dim);
        }
        .file-chip-icon { color: var(--haira-accent); display: flex; }
        .file-chip-name { max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .file-chip-size { color: var(--haira-muted); font-size: 0.7rem; }
        .file-chip-remove {
          background: none; border: none; color: var(--haira-muted); cursor: pointer;
          display: flex; padding: 0.1rem; border-radius: 3px; transition: all 0.15s;
        }
        .file-chip-remove:hover { color: var(--haira-error); background: rgba(239, 68, 68, 0.1); }
        .input-row { display: flex; align-items: flex-end; gap: 0.35rem; padding: 0.35rem; }
        .attach-btn {
          background: none; border: none; color: var(--haira-muted); cursor: pointer;
          display: flex; align-items: center; justify-content: center; padding: 0.45rem;
          border-radius: 6px; transition: all 0.15s; flex-shrink: 0;
        }
        .attach-btn:hover { color: var(--haira-accent); background: var(--haira-accent-dim); }
        textarea {
          flex: 1; background: transparent; border: none; color: var(--haira-text);
          padding: 0.5rem 0.35rem; font-size: 0.9rem; font-family: var(--haira-font);
          resize: none; min-height: 44px; max-height: 200px; outline: none; line-height: 1.5;
        }
        textarea::placeholder { color: var(--haira-muted); }
        .send-btn {
          background: var(--haira-accent); color: #1a0e04; border: none;
          width: 34px; height: 34px; border-radius: 8px; cursor: pointer;
          display: flex; align-items: center; justify-content: center;
          transition: all 0.15s; flex-shrink: 0;
        }
        .send-btn:hover { background: var(--haira-accent-light); box-shadow: 0 2px 12px rgba(232, 163, 23, 0.25); }
        .send-btn:disabled { opacity: 0.35; cursor: not-allowed; box-shadow: none; }
        .input-hint { text-align: center; font-size: 0.68rem; color: var(--haira-muted); opacity: 0.5; padding-top: 0.35rem; }

        /* Activity panel */
        .activity-panel {
          position: absolute; right: 0; top: 0; bottom: 0; width: 280px; z-index: 50;
          display: flex; flex-direction: column; border-left: 1px solid var(--haira-border);
          background: var(--haira-bg); overflow: hidden; box-shadow: -4px 0 24px rgba(0, 0, 0, 0.25);
        }
        .activity-panel.collapsed { display: none; }
        .panel-header {
          display: flex; align-items: center; gap: 0.5rem;
          padding: 0.55rem 0.75rem; border-bottom: 1px solid var(--haira-border); flex-shrink: 0;
        }
        .panel-header-icon { display: flex; color: var(--haira-muted); }
        .panel-title { font-size: 0.78rem; font-weight: 600; color: var(--haira-text-dim); flex: 1; }
        .panel-count { font-size: 0.68rem; color: var(--haira-muted); font-family: var(--haira-mono); }
        .panel-close {
          background: none; border: none; color: var(--haira-muted); cursor: pointer;
          display: flex; align-items: center; justify-content: center; padding: 0.2rem;
          border-radius: 4px; transition: all 0.15s;
        }
        .panel-close:hover { color: var(--haira-text); background: var(--haira-bg-elevated); }
        .panel-body {
          flex: 1; overflow-y: auto; padding: 0.5rem;
          display: flex; flex-direction: column; gap: 0.4rem; ${scrollbarCSS}
        }
        .panel-empty { display: flex; align-items: center; justify-content: center; flex: 1; font-size: 0.75rem; color: var(--haira-muted); opacity: 0.5; }
        .activity-toggle {
          position: absolute; top: 0.5rem; right: 0.5rem; z-index: 10;
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          color: var(--haira-muted); cursor: pointer; display: flex;
          align-items: center; justify-content: center; padding: 0.35rem;
          border-radius: 6px; transition: all 0.15s; gap: 0.3rem;
        }
        .activity-toggle:hover { color: var(--haira-accent); border-color: var(--haira-accent); background: var(--haira-accent-dim); }
        .activity-toggle .badge {
          display: none; min-width: 16px; height: 16px; padding: 0 4px;
          border-radius: 8px; background: var(--haira-accent); color: #1a0e04;
          font-size: 0.62rem; font-weight: 700; line-height: 16px; text-align: center;
        }
        .activity-toggle .badge.visible { display: inline-block; }

        @media (max-width: 640px) {
          .sidebar {
            position: absolute; left: 0; top: 0; bottom: 0; z-index: 100;
            box-shadow: 4px 0 24px rgba(0, 0, 0, 0.3);
          }
          .messages-inner { padding: 1rem 0.75rem; }
          .input-area { padding: 0.5rem 0.5rem 0.75rem; padding-bottom: max(0.75rem, env(safe-area-inset-bottom)); }
          .welcome { padding: 1.5rem 1rem; }
          .suggestions { max-width: 100%; }
        }
      </style>

      <div class="sidebar" id="sidebar">
        <div class="sidebar-header">
          <span class="sidebar-title">Chats</span>
          <button class="sidebar-btn" id="new-chat-btn" title="New chat">${icons.plus}</button>
          <button class="sidebar-btn" id="sidebar-close-btn" title="Close sidebar">${icons.chevronLeft}</button>
        </div>
        <div class="sidebar-list" id="sidebar-list">
          <div class="sidebar-empty" id="sidebar-empty">No chats yet</div>
        </div>
      </div>

      <div class="chat-main">
        <button class="sidebar-toggle" id="sidebar-open-btn" title="Show chats">${icons.sidebar}</button>

        <div class="welcome" id="welcome">
          <span class="welcome-icon">${m.logo ? `<img src="${esc(m.logo)}" alt="">` : logoSvg.replace(/width="22" height="22"/, 'width="56" height="56"')}</span>
          <h2>${esc(m.title || m.name || "Chat")}</h2>
          ${m.description ? `<p>${esc(m.description)}</p>` : ""}
          <div class="suggestions" id="suggestions"></div>
        </div>

        <div class="messages" id="messages">
          <div class="messages-inner" id="messages-inner">
            <div class="typing" id="typing">
              <div class="typing-dots">
                <span class="typing-dot"></span>
                <span class="typing-dot"></span>
                <span class="typing-dot"></span>
              </div>
              <span>Thinking...</span>
            </div>
          </div>
        </div>

        <div class="drop-overlay" id="drop-overlay">
          <span class="drop-overlay-icon">${icons.attach}</span>
          <span class="drop-overlay-text">Drop file to attach</span>
        </div>

        <button class="activity-toggle" id="activity-toggle" title="Toggle activity panel">
          ${icons.activity}
          <span class="badge" id="toggle-badge">0</span>
        </button>

        <div class="input-area">
          <div class="input-card" id="input-card">
            <div class="file-chip" id="file-chip">
              <div class="file-chip-inner">
                <span class="file-chip-icon">${icons.file}</span>
                <span class="file-chip-name" id="file-name"></span>
                <span class="file-chip-size" id="file-size"></span>
                <button class="file-chip-remove" id="file-remove" title="Remove file">${icons.xSmall}</button>
              </div>
            </div>
            <div class="input-row">
              ${m.hasFile ? `<button class="attach-btn" id="attach-btn" title="Attach file">${icons.attach}</button>` : ""}
              <textarea id="chat-input" placeholder="${m.hasFile ? "Type a message or drop a file..." : "Type a message..."}" rows="1"></textarea>
              <button class="send-btn" id="send-btn" title="Send">${icons.send}</button>
            </div>
          </div>
          ${m.hasFile ? `<input type="file" id="file-input" style="display:none" />` : ""}
          <div class="input-hint">Enter to send, Shift+Enter for new line</div>
        </div>
      </div>

      <div class="activity-panel collapsed" id="activity-panel">
        <div class="panel-header">
          <span class="panel-header-icon">${icons.activity}</span>
          <span class="panel-title">Activity</span>
          <span class="panel-count" id="panel-count"></span>
          <button class="panel-close" id="panel-close" title="Close panel">${icons.xSmall}</button>
        </div>
        <div class="panel-body" id="panel-body">
          <div class="panel-empty" id="panel-empty">No activity yet</div>
        </div>
      </div>
    `;

    const messagesOuter = shadow.getElementById("messages")!;
    const messagesInner = shadow.getElementById("messages-inner")!;
    const welcome = shadow.getElementById("welcome")!;
    const input = shadow.getElementById("chat-input") as HTMLTextAreaElement;
    const sendBtn = shadow.getElementById("send-btn") as HTMLButtonElement;
    const fileChip = shadow.getElementById("file-chip")!;
    const fileName = shadow.getElementById("file-name")!;
    const fileSize = shadow.getElementById("file-size")!;
    const fileRemove = shadow.getElementById("file-remove")!;
    const typing = shadow.getElementById("typing")!;
    const dropOverlay = shadow.getElementById("drop-overlay")!;
    const suggestionsEl = shadow.getElementById("suggestions")!;
    const attachBtn = m.hasFile ? shadow.getElementById("attach-btn") : null;
    const fileInput = m.hasFile ? (shadow.getElementById("file-input") as HTMLInputElement) : null;

    const activityPanel = shadow.getElementById("activity-panel")!;
    const panelBody = shadow.getElementById("panel-body")!;
    const panelEmpty = shadow.getElementById("panel-empty")!;
    const panelCount = shadow.getElementById("panel-count")!;
    const panelClose = shadow.getElementById("panel-close")!;
    const activityToggle = shadow.getElementById("activity-toggle")!;
    const toggleBadge = shadow.getElementById("toggle-badge")!;

    const sidebar = shadow.getElementById("sidebar")!;
    const sidebarList = shadow.getElementById("sidebar-list")!;
    const sidebarEmpty = shadow.getElementById("sidebar-empty")!;
    const newChatBtn = shadow.getElementById("new-chat-btn")!;
    const sidebarCloseBtn = shadow.getElementById("sidebar-close-btn")!;
    const sidebarOpenBtn = shadow.getElementById("sidebar-open-btn")!;

    let panelOpen = false;
    let runningCount = 0;
    let totalCount = 0;

    function togglePanel(open?: boolean) {
      panelOpen = open !== undefined ? open : !panelOpen;
      activityPanel.classList.toggle("collapsed", !panelOpen);
    }

    function updateBadge() {
      if (runningCount > 0) {
        toggleBadge.textContent = String(runningCount);
        toggleBadge.classList.add("visible");
      } else {
        toggleBadge.classList.remove("visible");
      }
      panelCount.textContent = totalCount > 0 ? String(totalCount) : "";
    }

    activityToggle.addEventListener("click", () => togglePanel());
    panelClose.addEventListener("click", () => togglePanel(false));

    let sidebarOpen = true;

    const toggleSidebar = (open?: boolean) => {
      sidebarOpen = open !== undefined ? open : !sidebarOpen;
      sidebar.classList.toggle("collapsed", !sidebarOpen);
      sidebarOpenBtn.classList.toggle("visible", !sidebarOpen);
    };

    sidebarCloseBtn.addEventListener("click", () => toggleSidebar(false));
    sidebarOpenBtn.addEventListener("click", () => toggleSidebar(true));

    const self = this;

    const refreshSidebar = async () => {
      try {
        const resp = await fetch(`/_api/chats?workflow=${encodeURIComponent(m.path)}`);
        if (!resp.ok) return;
        const sessions: ChatSessionSummary[] = await resp.json();
        if (!sessions || sessions.length === 0) {
          sidebarEmpty.style.display = "";
          sidebarList.querySelectorAll(".session-item").forEach((el) => el.remove());
          return;
        }
        sidebarEmpty.style.display = "none";
        sidebarList.querySelectorAll(".session-item").forEach((el) => el.remove());

        for (const sess of sessions) {
          const item = document.createElement("div");
          item.className = `session-item${sess.id === self.sessionId ? " active" : ""}`;
          item.innerHTML = `
            <span class="session-icon">${icons.chat}</span>
            <span class="session-title">${esc(sess.title || "New chat")}</span>
            <button class="session-delete" title="Delete">${icons.trash}</button>
          `;
          item.addEventListener("click", (e) => {
            if ((e.target as HTMLElement).closest(".session-delete")) return;
            self.switchSession(sess.id);
          });
          item.querySelector(".session-delete")!.addEventListener("click", async (e) => {
            e.stopPropagation();
            await fetch(`/_api/chats/${sess.id}`, { method: "DELETE" });
            if (sess.id === self.sessionId) {
              self.startNewChat();
            }
            refreshSidebar();
          });
          sidebarList.appendChild(item);
        }
      } catch {
        // Silently fail
      }
    };

    newChatBtn.addEventListener("click", () => {
      self.startNewChat();
    });

    // Suggestions
    const defaultSuggestions = this.getSuggestions();
    for (const text of defaultSuggestions) {
      const btn = document.createElement("button");
      btn.className = "suggestion";
      btn.textContent = text;
      btn.addEventListener("click", () => {
        input.value = text;
        send();
      });
      suggestionsEl.appendChild(btn);
    }

    // File attach
    if (attachBtn && fileInput) {
      attachBtn.addEventListener("click", () => fileInput.click());
      fileInput.addEventListener("change", () => {
        if (fileInput.files && fileInput.files[0]) {
          this.setFile(fileInput.files[0], fileChip, fileName, fileSize);
        }
      });
      fileRemove.addEventListener("click", () => {
        this.clearFile(fileChip, fileInput);
      });

      let dragCounter = 0;
      shadow.addEventListener("dragenter", (e) => {
        e.preventDefault();
        dragCounter++;
        dropOverlay.classList.add("visible");
      });
      shadow.addEventListener("dragleave", (e) => {
        e.preventDefault();
        dragCounter--;
        if (dragCounter <= 0) {
          dragCounter = 0;
          dropOverlay.classList.remove("visible");
        }
      });
      shadow.addEventListener("dragover", (e) => {
        e.preventDefault();
      });
      shadow.addEventListener("drop", (e) => {
        e.preventDefault();
        dragCounter = 0;
        dropOverlay.classList.remove("visible");
        const dt = (e as DragEvent).dataTransfer;
        if (dt && dt.files && dt.files[0]) {
          self.setFile(dt.files[0], fileChip, fileName, fileSize);
        }
      });
    }

    // Textarea auto-resize
    input.addEventListener("input", () => {
      input.style.height = "auto";
      input.style.height = `${Math.min(input.scrollHeight, 200)}px`;
    });

    input.addEventListener("keydown", (e) => {
      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault();
        send();
      }
    });

    sendBtn.addEventListener("click", send);

    shadow.addEventListener("haira-chat-input", (e: Event) => {
      const text = (e as CustomEvent).detail?.text;
      if (text && !sendBtn.disabled) {
        input.value = text;
        send();
      }
    });

    requestAnimationFrame(() => input.focus());

    const avatarValue = m.avatar || "H";

    function addMessage(role: string, content: string, file?: string): HairaMessage {
      const msg = document.createElement("haira-message") as HairaMessage;
      msg.setAttribute("role", role);
      msg.setAttribute("content", content);
      if (file) msg.setAttribute("file", file);
      if (role === "assistant") msg.setAttribute("avatar", avatarValue);
      messagesInner.insertBefore(msg, typing);
      messagesOuter.scrollTop = messagesOuter.scrollHeight;
      return msg;
    }

    function clearMessages() {
      const toRemove = messagesInner.querySelectorAll("haira-message, haira-tool-card, haira-ui-renderer");
      toRemove.forEach((el) => el.remove());
    }

    async function send() {
      const text = input.value.trim();
      if (!text && !self.attachedFile) return;
      input.value = "";
      input.style.height = "auto";

      self.streamAbort?.abort();
      self.streamAbort = new AbortController();

      welcome.classList.add("hidden");
      messagesOuter.classList.add("active");

      const fileLabel = self.attachedFile ? self.attachedFile.name : undefined;
      addMessage("user", text, fileLabel);

      sendBtn.disabled = true;
      typing.classList.add("visible");
      messagesOuter.scrollTop = messagesOuter.scrollHeight;

      const chatParam = m.chatParam || "message";
      let formData: FormData | undefined;
      const body: Record<string, unknown> = {};
      body[chatParam] = text;
      body["session_id"] = self.sessionId;

      if (self.attachedFile) {
        const fileParamName = m.fileParam || "file_path";
        formData = new FormData();
        formData.append(fileParamName, self.attachedFile);
        formData.append(chatParam, text);
        formData.append("session_id", self.sessionId);
      }

      if (fileInput) self.clearFile(fileChip, fileInput);

      const activeTools = new Map<string, HairaToolCard>();
      let assistantMsg: HairaMessage | null = null;
      let fullText = "";

      await streamSSE(
        m.path,
        body,
        {
          onToolStart: (event) => {
            typing.classList.remove("visible");
            const card = document.createElement("haira-tool-card") as HairaToolCard;
            panelEmpty.style.display = "none";
            panelBody.appendChild(card);
            card.setTool(event.tool);
            activeTools.set(event.tool, card);
            panelBody.scrollTop = panelBody.scrollHeight;

            runningCount++;
            totalCount++;
            updateBadge();

            if (!panelOpen) {
              togglePanel(true);
            }
          },
          onToolRender: (event: ToolRenderEvent) => {
            const renderer = document.createElement("haira-ui-renderer") as HairaUIRenderer;
            messagesInner.insertBefore(renderer, typing);
            renderer.render(event);
            messagesOuter.scrollTop = messagesOuter.scrollHeight;
          },
          onToolEnd: (event) => {
            const card = activeTools.get(event.tool);
            if (card) {
              card.complete(event.ok !== false);
              activeTools.delete(event.tool);
            }
            typing.classList.add("visible");

            runningCount = Math.max(0, runningCount - 1);
            updateBadge();
          },
          onDelta: (delta) => {
            typing.classList.remove("visible");
            if (!assistantMsg) {
              assistantMsg = addMessage("assistant", "");
            }
            fullText += delta;
            assistantMsg.updateContent(fullText);
            messagesOuter.scrollTop = messagesOuter.scrollHeight;
          },
          onError: (error) => {
            typing.classList.remove("visible");
            if (!assistantMsg) {
              assistantMsg = addMessage("assistant", "");
            }
            assistantMsg.updateContent(`Error: ${error}`);
            sendBtn.disabled = false;
            input.focus();
          },
          onDone: () => {
            typing.classList.remove("visible");
            if (!assistantMsg && fullText === "") {
              assistantMsg = addMessage("assistant", "");
              assistantMsg.updateContent("No response received. Please check the server logs.");
            }
            sendBtn.disabled = false;
            input.focus();
            refreshSidebar();
          },
        },
        formData,
        self.streamAbort?.signal,
      );
    }

    const loadSession = async (sessionId: string) => {
      try {
        const resp = await fetch(`/_api/chats/${sessionId}`);
        if (!resp.ok) return;
        const detail: ChatSessionDetail = await resp.json();
        if (!detail.messages || detail.messages.length === 0) return;

        welcome.classList.add("hidden");
        messagesOuter.classList.add("active");

        clearMessages();
        const msgs = detail.messages;
        for (let i = 0; i < msgs.length; i++) {
          const msg = msgs[i];
          const el = addMessage(msg.role, msg.content);
          if (msg.role === "assistant") {
            el.updateContent(msg.content);
          }
          if (msg.ui_events && msg.ui_events.length > 0) {
            const hasFollowUp = msgs.slice(i + 1).some((m) => m.role === "user");
            for (const event of msg.ui_events) {
              const renderer = document.createElement("haira-ui-renderer") as HairaUIRenderer;
              if (hasFollowUp) {
                renderer.setAttribute("data-restored", "true");
              }
              messagesInner.insertBefore(renderer, typing);
              renderer.render(event);
            }
          }
        }
        messagesOuter.scrollTop = messagesOuter.scrollHeight;
      } catch {
        // Session doesn't exist yet — fresh chat
      }
    };

    loadSession(this.sessionId);
    refreshSidebar();
  }

  private switchSession(newSessionId: string) {
    this.streamAbort?.abort();
    this.streamAbort = null;

    this.sessionId = newSessionId;
    const url = new URL(window.location.href);
    url.searchParams.set("session", newSessionId);
    window.history.pushState({}, "", url.toString());
    if (this.shadowRoot) {
      this.shadowRoot.innerHTML = "";
    }
    this.renderChat();
  }

  private startNewChat() {
    const newId = crypto.randomUUID();
    this.switchSession(newId);
  }

  private setFile(file: File, chipEl: HTMLElement, nameEl: HTMLElement, sizeEl: HTMLElement) {
    this.attachedFile = file;
    nameEl.textContent = file.name;
    sizeEl.textContent = formatBytes(file.size);
    chipEl.classList.add("visible");
  }

  private clearFile(chipEl: HTMLElement, fileInput: HTMLInputElement) {
    this.attachedFile = null;
    chipEl.classList.remove("visible");
    fileInput.value = "";
  }

  private getSuggestions(): string[] {
    if (this.meta.suggestions && this.meta.suggestions.length > 0) {
      return this.meta.suggestions;
    }
    return ["What can you help me with?", "Hello!"];
  }
}
