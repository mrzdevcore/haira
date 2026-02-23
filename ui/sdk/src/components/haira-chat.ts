import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state, query } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles, scrollbarStyles } from "../core/styles";
import { iconStrings, logoSvgStr } from "../core/icons";
import { formatBytes } from "../core/utils";
import { streamSSE } from "../services/sse-client";
import { ArpClient } from "@haira/arp";
import type {
  WorkflowMeta,
  ToolRenderEvent,
  ChatSessionSummary,
  ChatSessionDetail,
} from "../core/types";

// ---------- Internal interfaces ----------

interface ChatMessage {
  role: "user" | "assistant";
  content: string;
  file?: string;
  uiEvents?: ToolRenderEvent[];
  /** Whether this message was loaded from a saved session */
  restored?: boolean;
}

interface ToolCardState {
  name: string;
  displayName: string;
  status: "running" | "done" | "failed";
  startTime: number;
  elapsed?: string;
}

// ---------- Component ----------

@customElement("haira-chat")
export class HairaChat extends LitElement {
  static styles = [
    baseStyles,
    scrollbarStyles,
    css`
      :host {
        display: flex;
        flex-direction: row;
        flex: 1;
        overflow: hidden;
        position: relative;
      }

      /* ---- Sidebar ---- */
      .sidebar {
        width: 240px;
        flex-shrink: 0;
        display: flex;
        flex-direction: column;
        border-right: 1px solid var(--haira-border);
        background: var(--haira-bg);
        overflow: hidden;
        transition: width 0.2s, opacity 0.2s;
      }
      .sidebar.collapsed {
        width: 0;
        opacity: 0;
        pointer-events: none;
      }
      .sidebar-header {
        display: flex;
        align-items: center;
        gap: 0.4rem;
        padding: 0.55rem 0.65rem;
        border-bottom: 1px solid var(--haira-border);
        flex-shrink: 0;
      }
      .sidebar-title {
        font-size: 0.75rem;
        font-weight: 600;
        color: var(--haira-text-dim);
        flex: 1;
      }
      .sidebar-btn {
        background: none;
        border: none;
        color: var(--haira-muted);
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 0.25rem;
        border-radius: 4px;
        transition: all 0.15s;
      }
      .sidebar-btn:hover {
        color: var(--haira-accent);
        background: var(--haira-accent-dim);
      }
      .sidebar-list {
        flex: 1;
        overflow-y: auto;
        padding: 0.35rem;
        display: flex;
        flex-direction: column;
        gap: 1px;
      }
      .session-item {
        display: flex;
        align-items: center;
        gap: 0.4rem;
        padding: 0.45rem 0.5rem;
        border-radius: 6px;
        cursor: pointer;
        transition: all 0.12s;
        text-decoration: none;
        color: var(--haira-text-dim);
        font-size: 0.78rem;
        line-height: 1.35;
        position: relative;
      }
      .session-item:hover {
        background: var(--haira-bg-card);
        color: var(--haira-text);
      }
      .session-item.active {
        background: var(--haira-accent-dim);
        color: var(--haira-accent);
      }
      .session-icon {
        display: flex;
        flex-shrink: 0;
        opacity: 0.5;
      }
      .session-item.active .session-icon {
        opacity: 1;
      }
      .session-title {
        flex: 1;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        min-width: 0;
      }
      .session-delete {
        display: none;
        background: none;
        border: none;
        color: var(--haira-muted);
        cursor: pointer;
        padding: 0.15rem;
        border-radius: 3px;
        flex-shrink: 0;
        align-items: center;
        justify-content: center;
      }
      .session-item:hover .session-delete {
        display: flex;
      }
      .session-delete:hover {
        color: var(--haira-error);
        background: rgba(239, 68, 68, 0.1);
      }
      .sidebar-empty {
        padding: 1rem;
        text-align: center;
        font-size: 0.75rem;
        color: var(--haira-muted);
        opacity: 0.5;
      }

      .sidebar-toggle {
        position: absolute;
        top: 0.5rem;
        left: 0.5rem;
        z-index: 10;
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        color: var(--haira-muted);
        cursor: pointer;
        display: none;
        align-items: center;
        justify-content: center;
        padding: 0.35rem;
        border-radius: 6px;
        transition: all 0.15s;
      }
      .sidebar-toggle.visible {
        display: flex;
      }
      .sidebar-toggle:hover {
        color: var(--haira-accent);
        border-color: var(--haira-accent);
        background: var(--haira-accent-dim);
      }

      /* ---- Chat main ---- */
      .chat-main {
        flex: 1;
        display: flex;
        flex-direction: column;
        overflow: hidden;
        position: relative;
        min-width: 0;
        height: 100%;
      }

      /* ---- Welcome ---- */
      .welcome {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 1rem;
        padding: 2rem;
        opacity: 1;
        transition: opacity 0.3s;
      }
      .welcome.hidden {
        display: none;
      }
      .welcome-icon {
        opacity: 0.15;
      }
      .welcome-icon img {
        width: 56px;
        height: 56px;
        object-fit: contain;
        opacity: 1;
      }
      .welcome h2 {
        font-size: 1.1rem;
        font-weight: 600;
        color: var(--haira-text);
      }
      .welcome p {
        font-size: 0.85rem;
        color: var(--haira-muted);
        text-align: center;
        max-width: 420px;
        line-height: 1.5;
      }
      .suggestions {
        display: flex;
        flex-wrap: wrap;
        gap: 0.5rem;
        justify-content: center;
        margin-top: 0.5rem;
        max-width: 540px;
      }
      .suggestion {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        color: var(--haira-text-dim);
        padding: 0.45rem 0.85rem;
        border-radius: 20px;
        font-size: 0.78rem;
        font-family: var(--haira-font);
        cursor: pointer;
        transition: all 0.15s;
      }
      .suggestion:hover {
        border-color: var(--haira-accent);
        color: var(--haira-accent);
        background: var(--haira-accent-dim);
      }

      /* ---- Messages ---- */
      .messages {
        flex: 1;
        min-height: 0;
        overflow-y: auto;
        display: none;
        flex-direction: column;
      }
      .messages.active {
        display: flex;
      }
      .messages-inner {
        max-width: 768px;
        width: 100%;
        margin: 0 auto;
        padding: 1.5rem 1.25rem;
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
      }

      /* ---- Typing indicator ---- */
      .typing {
        display: none;
        padding: 0.25rem 0;
        align-items: center;
        gap: 0.4rem;
        font-size: 0.75rem;
        color: var(--haira-muted);
        margin-left: 2.25rem;
      }
      .typing.visible {
        display: flex;
      }
      .typing-dots {
        display: flex;
        gap: 0.2rem;
        align-items: center;
      }
      .typing-dot {
        display: inline-block;
        width: 5px;
        height: 5px;
        border-radius: 50%;
        background: var(--haira-accent);
        animation: bounce 1.4s ease-in-out infinite;
      }
      .typing-dot:nth-child(2) {
        animation-delay: 0.2s;
      }
      .typing-dot:nth-child(3) {
        animation-delay: 0.4s;
      }

      /* ---- Drop overlay ---- */
      .drop-overlay {
        display: none;
        position: absolute;
        inset: 0;
        background: rgba(9, 9, 11, 0.85);
        z-index: 200;
        align-items: center;
        justify-content: center;
        flex-direction: column;
        gap: 0.75rem;
        border: 2px dashed var(--haira-accent);
        border-radius: var(--haira-radius);
        margin: 0.5rem;
      }
      .drop-overlay.visible {
        display: flex;
      }
      .drop-overlay-icon {
        color: var(--haira-accent);
        opacity: 0.7;
      }
      .drop-overlay-text {
        color: var(--haira-accent);
        font-size: 0.9rem;
        font-weight: 600;
      }

      /* ---- Input area ---- */
      .input-area {
        padding: 0.75rem 1rem 1rem;
        flex-shrink: 0;
        background: var(--haira-bg);
        border-top: 1px solid var(--haira-border);
      }
      .input-card {
        display: flex;
        flex-direction: column;
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        transition: border-color 0.15s;
        max-width: 768px;
        margin: 0 auto;
      }
      .input-card:focus-within {
        border-color: var(--haira-border-focus);
      }
      .file-chip {
        display: none;
        align-items: center;
        gap: 0.4rem;
        padding: 0.4rem 0.6rem 0;
        margin: 0 0.5rem;
      }
      .file-chip.visible {
        display: flex;
      }
      .file-chip-inner {
        display: flex;
        align-items: center;
        gap: 0.35rem;
        background: var(--haira-bg-elevated);
        border: 1px solid var(--haira-border);
        border-radius: 6px;
        padding: 0.25rem 0.5rem;
        font-size: 0.75rem;
        color: var(--haira-text-dim);
      }
      .file-chip-icon {
        color: var(--haira-accent);
        display: flex;
      }
      .file-chip-name {
        max-width: 200px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      .file-chip-size {
        color: var(--haira-muted);
        font-size: 0.7rem;
      }
      .file-chip-remove {
        background: none;
        border: none;
        color: var(--haira-muted);
        cursor: pointer;
        display: flex;
        padding: 0.1rem;
        border-radius: 3px;
        transition: all 0.15s;
      }
      .file-chip-remove:hover {
        color: var(--haira-error);
        background: rgba(239, 68, 68, 0.1);
      }
      .input-row {
        display: flex;
        align-items: flex-end;
        gap: 0.35rem;
        padding: 0.35rem;
      }
      .attach-btn {
        background: none;
        border: none;
        color: var(--haira-muted);
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 0.45rem;
        border-radius: 6px;
        transition: all 0.15s;
        flex-shrink: 0;
      }
      .attach-btn:hover {
        color: var(--haira-accent);
        background: var(--haira-accent-dim);
      }
      textarea {
        flex: 1;
        background: transparent;
        border: none;
        color: var(--haira-text);
        padding: 0.5rem 0.35rem;
        font-size: 0.9rem;
        font-family: var(--haira-font);
        resize: none;
        min-height: 44px;
        max-height: 200px;
        outline: none;
        line-height: 1.5;
      }
      textarea::placeholder {
        color: var(--haira-muted);
      }
      .send-btn {
        background: var(--haira-accent);
        color: #1a0e04;
        border: none;
        width: 34px;
        height: 34px;
        border-radius: 8px;
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        transition: all 0.15s;
        flex-shrink: 0;
      }
      .send-btn:hover {
        background: var(--haira-accent-light);
        box-shadow: 0 2px 12px rgba(232, 163, 23, 0.25);
      }
      .send-btn:disabled {
        opacity: 0.35;
        cursor: not-allowed;
        box-shadow: none;
      }
      .input-hint {
        text-align: center;
        font-size: 0.68rem;
        color: var(--haira-muted);
        opacity: 0.5;
        padding-top: 0.35rem;
      }

      /* ---- Activity panel ---- */
      .activity-panel {
        position: absolute;
        right: 0;
        top: 0;
        bottom: 0;
        width: 280px;
        z-index: 50;
        display: flex;
        flex-direction: column;
        border-left: 1px solid var(--haira-border);
        background: var(--haira-bg);
        overflow: hidden;
        box-shadow: -4px 0 24px rgba(0, 0, 0, 0.25);
      }
      .activity-panel.collapsed {
        display: none;
      }
      .panel-header {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.55rem 0.75rem;
        border-bottom: 1px solid var(--haira-border);
        flex-shrink: 0;
      }
      .panel-header-icon {
        display: flex;
        color: var(--haira-muted);
      }
      .panel-title {
        font-size: 0.78rem;
        font-weight: 600;
        color: var(--haira-text-dim);
        flex: 1;
      }
      .panel-count {
        font-size: 0.68rem;
        color: var(--haira-muted);
        font-family: var(--haira-mono);
      }
      .panel-close {
        background: none;
        border: none;
        color: var(--haira-muted);
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 0.2rem;
        border-radius: 4px;
        transition: all 0.15s;
      }
      .panel-close:hover {
        color: var(--haira-text);
        background: var(--haira-bg-elevated);
      }
      .panel-body {
        flex: 1;
        overflow-y: auto;
        padding: 0.5rem;
        display: flex;
        flex-direction: column;
        gap: 0.4rem;
      }
      .panel-empty {
        display: flex;
        align-items: center;
        justify-content: center;
        flex: 1;
        font-size: 0.75rem;
        color: var(--haira-muted);
        opacity: 0.5;
      }

      .activity-toggle {
        position: absolute;
        top: 0.5rem;
        right: 0.5rem;
        z-index: 10;
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        color: var(--haira-muted);
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 0.35rem;
        border-radius: 6px;
        transition: all 0.15s;
        gap: 0.3rem;
      }
      .activity-toggle:hover {
        color: var(--haira-accent);
        border-color: var(--haira-accent);
        background: var(--haira-accent-dim);
      }
      .badge {
        display: none;
        min-width: 16px;
        height: 16px;
        padding: 0 4px;
        border-radius: 8px;
        background: var(--haira-accent);
        color: #1a0e04;
        font-size: 0.62rem;
        font-weight: 700;
        line-height: 16px;
        text-align: center;
      }
      .badge.visible {
        display: inline-block;
      }

      /* ---- Tool card (inline in activity panel) ---- */
      .tool-card {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 0.75rem;
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: 8px;
        animation: fadeSlideUp 0.25s ease-out;
      }
      .tool-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 24px;
        height: 24px;
        border-radius: 6px;
        flex-shrink: 0;
      }
      .tool-icon.running {
        background: rgba(232, 163, 23, 0.1);
        color: var(--haira-accent);
      }
      .tool-icon.done {
        background: rgba(34, 197, 94, 0.1);
        color: var(--haira-success);
      }
      .tool-icon.failed {
        background: rgba(239, 68, 68, 0.1);
        color: var(--haira-error);
      }
      .tool-info {
        flex: 1;
        min-width: 0;
      }
      .tool-name {
        font-size: 0.78rem;
        font-weight: 600;
        color: var(--haira-text);
      }
      .tool-status {
        font-size: 0.7rem;
        color: var(--haira-muted);
        display: flex;
        align-items: center;
        gap: 0.3rem;
      }
      .tool-duration {
        font-family: var(--haira-mono);
        font-size: 0.68rem;
        color: var(--haira-muted);
        flex-shrink: 0;
      }

      /* ---- Responsive ---- */
      @media (max-width: 640px) {
        .sidebar {
          position: absolute;
          left: 0;
          top: 0;
          bottom: 0;
          z-index: 100;
          box-shadow: 4px 0 24px rgba(0, 0, 0, 0.3);
        }
        .messages-inner {
          padding: 1rem 0.75rem;
        }
        .input-area {
          padding: 0.5rem 0.5rem 0.75rem;
          padding-bottom: max(0.75rem, env(safe-area-inset-bottom));
        }
        .welcome {
          padding: 1.5rem 1rem;
        }
        .suggestions {
          max-width: 100%;
        }
      }
    `,
  ];

  // ---- Public properties ----

  @property({ type: Object }) meta!: WorkflowMeta;

  // ---- Reactive state ----

  @state() private _sessionId = "";
  @state() private _sessions: ChatSessionSummary[] = [];
  @state() private _messages: ChatMessage[] = [];
  @state() private _sidebarOpen = true;
  @state() private _panelOpen = false;
  @state() private _isStreaming = false;
  @state() private _showTyping = false;
  @state() private _showWelcome = true;
  @state() private _dropVisible = false;
  @state() private _attachedFile: File | null = null;
  @state() private _toolCards: ToolCardState[] = [];
  @state() private _runningToolCount = 0;
  @state() private _totalToolCount = 0;

  // ---- DOM references ----

  @query("#messages-scroll") private _messagesEl!: HTMLDivElement;
  @query("#chat-input") private _inputEl!: HTMLTextAreaElement;
  @query("#file-input") private _fileInputEl!: HTMLInputElement;
  @query("#panel-body") private _panelBodyEl!: HTMLDivElement;

  // ---- Internal state (non-reactive) ----

  private _streamAbort: AbortController | null = null;
  private _dragCounter = 0;
  /** Index of the assistant message currently being streamed into */
  private _streamingMsgIndex = -1;
  private _fullStreamText = "";
  /** Map tool name -> index in _toolCards */
  private _activeToolMap = new Map<string, number>();
  /** ARP WebSocket client — null if WebSocket not available */
  private _arpClient: ArpClient | null = null;

  // ---------- Lifecycle ----------

  connectedCallback() {
    super.connectedCallback();
    this._initSession();
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    this._streamAbort?.abort();
    this._arpClient?.disconnect();
    this._arpClient = null;
  }

  private async _initSession() {
    const url = new URL(window.location.href);
    const urlSession = url.searchParams.get("session");

    if (urlSession) {
      this._sessionId = urlSession;
    } else {
      // Try to resume the latest session for this workflow
      try {
        const resp = await fetch(
          `/_api/chats?workflow=${encodeURIComponent(this.meta.path)}`
        );
        if (resp.ok) {
          const sessions: ChatSessionSummary[] = await resp.json();
          if (sessions && sessions.length > 0) {
            this._sessionId = sessions[0].id;
            this._setSessionUrl(this._sessionId);
            this._loadSession(this._sessionId);
            this._refreshSidebar();
            this._connectArp();
            return;
          }
        }
      } catch {
        // API unavailable
      }
      this._sessionId = crypto.randomUUID();
      this._setSessionUrl(this._sessionId);
    }

    this._loadSession(this._sessionId);
    this._refreshSidebar();
    this._connectArp();
  }

  /** Update URL with session ID */
  private _setSessionUrl(sessionId: string) {
    const url = new URL(window.location.href);
    url.searchParams.set("session", sessionId);
    window.history.replaceState({}, "", url.toString());
  }

  /** Attempt to connect via ARP WebSocket. Falls back to SSE silently. */
  private _connectArp() {
    try {
      const wsProto = location.protocol === "https:" ? "wss:" : "ws:";
      const wsPath = this.meta.arpUrl ?? "/_arp/v1";
      const wsUrl = `${wsProto}//${location.host}${wsPath}`;
      this._arpClient = new ArpClient(
        { url: wsUrl, sessionId: this._sessionId },
        {
          onDelta: (text) => this._handleDelta(text),
          onToolStart: (tool) => this._handleToolStart(tool),
          onToolEnd: (tool, ok) => this._handleToolEnd(tool, ok),
          onRender: (event) => this._handleToolRender(event),
          onError: (error) => this._handleError(error),
          onDone: () => this._handleDone(),
        },
      );
      this._arpClient.connect();
    } catch {
      this._arpClient = null;
    }
  }

  // ---------- Session management ----------

  private async _refreshSidebar() {
    try {
      const resp = await fetch(
        `/_api/chats?workflow=${encodeURIComponent(this.meta.path)}`
      );
      if (!resp.ok) return;
      const sessions: ChatSessionSummary[] = await resp.json();
      this._sessions = sessions || [];
    } catch {
      // Silently fail
    }
  }

  private async _loadSession(sessionId: string) {
    try {
      const resp = await fetch(`/_api/chats/${sessionId}`);
      if (!resp.ok) return;
      const detail: ChatSessionDetail = await resp.json();
      if (!detail.messages || detail.messages.length === 0) return;

      const loaded: ChatMessage[] = [];
      for (let i = 0; i < detail.messages.length; i++) {
        const msg = detail.messages[i];
        const hasFollowUp = detail.messages
          .slice(i + 1)
          .some((m) => m.role === "user");
        loaded.push({
          role: msg.role,
          content: msg.content,
          uiEvents: msg.ui_events,
          restored: hasFollowUp,
        });
      }

      this._messages = loaded;
      this._showWelcome = false;
      this.updateComplete.then(() => this._scrollToBottom());
    } catch {
      // Session does not exist yet, show welcome
    }
  }

  private _switchSession(newSessionId: string) {
    this._streamAbort?.abort();
    this._streamAbort = null;

    this._sessionId = newSessionId;
    this._messages = [];
    this._toolCards = [];
    this._activeToolMap.clear();
    this._runningToolCount = 0;
    this._totalToolCount = 0;
    this._panelOpen = false;
    this._isStreaming = false;
    this._showTyping = false;
    this._showWelcome = true;
    this._attachedFile = null;
    this._streamingMsgIndex = -1;
    this._fullStreamText = "";

    // Reconnect ARP client with new session
    this._arpClient?.disconnect();
    this._arpClient = null;
    this._connectArp();

    const url = new URL(window.location.href);
    url.searchParams.set("session", newSessionId);
    window.history.pushState({}, "", url.toString());

    this._loadSession(newSessionId);
    this._refreshSidebar();
  }

  private _startNewChat() {
    const newId = crypto.randomUUID();
    this._switchSession(newId);
  }

  private async _deleteSession(sessId: string, e: Event) {
    e.stopPropagation();
    try {
      await fetch(`/_api/chats/${sessId}`, { method: "DELETE" });
    } catch {
      // ignore
    }
    if (sessId === this._sessionId) {
      this._startNewChat();
    }
    this._refreshSidebar();
  }

  // ---------- File handling ----------

  private _setFile(file: File) {
    this._attachedFile = file;
  }

  private _clearFile() {
    this._attachedFile = null;
    if (this._fileInputEl) {
      this._fileInputEl.value = "";
    }
  }

  private _onAttachClick() {
    this._fileInputEl?.click();
  }

  private _onFileChange() {
    if (this._fileInputEl?.files?.[0]) {
      this._setFile(this._fileInputEl.files[0]);
    }
  }

  // Drag and drop handlers
  private _onDragEnter(e: DragEvent) {
    e.preventDefault();
    this._dragCounter++;
    this._dropVisible = true;
  }

  private _onDragLeave(e: DragEvent) {
    e.preventDefault();
    this._dragCounter--;
    if (this._dragCounter <= 0) {
      this._dragCounter = 0;
      this._dropVisible = false;
    }
  }

  private _onDragOver(e: DragEvent) {
    e.preventDefault();
  }

  private _onDrop(e: DragEvent) {
    e.preventDefault();
    this._dragCounter = 0;
    this._dropVisible = false;
    if (e.dataTransfer?.files?.[0]) {
      this._setFile(e.dataTransfer.files[0]);
    }
  }

  // ---------- Input handling ----------

  private _onInputChange() {
    const el = this._inputEl;
    if (!el) return;
    el.style.height = "auto";
    el.style.height = `${Math.min(el.scrollHeight, 200)}px`;
  }

  private _onKeyDown(e: KeyboardEvent) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      this._send();
    }
  }

  private _onSuggestionClick(text: string) {
    if (this._inputEl) {
      this._inputEl.value = text;
    }
    this._send();
  }

  // ---------- Send / streaming (ARP WebSocket or SSE fallback) ----------

  private _handleToolStart(tool: string) {
    this._showTyping = false;
    const displayName = tool
      .replace(/^render_/, "")
      .replace(/_/g, " ")
      .replace(/\b\w/g, (c) => c.toUpperCase());

    const cardState: ToolCardState = {
      name: tool,
      displayName,
      status: "running",
      startTime: Date.now(),
    };
    const idx = this._toolCards.length;
    this._toolCards = [...this._toolCards, cardState];
    this._activeToolMap.set(tool, idx);

    this._runningToolCount++;
    this._totalToolCount++;

    if (!this._panelOpen) {
      this._panelOpen = true;
    }

    this.updateComplete.then(() => {
      if (this._panelBodyEl) {
        this._panelBodyEl.scrollTop = this._panelBodyEl.scrollHeight;
      }
    });
  }

  private _handleToolRender(event: ToolRenderEvent) {
    // Validate the event has the minimum required fields
    if (!event?.component || !event.props) {
      console.warn("[Haira] Ignoring invalid tool render event:", event);
      return;
    }

    if (this._streamingMsgIndex >= 0) {
      const msg = this._messages[this._streamingMsgIndex];
      const uiEvents = [...(msg.uiEvents || []), event];
      const updated = [...this._messages];
      updated[this._streamingMsgIndex] = { ...msg, uiEvents };
      this._messages = updated;
    } else {
      this._messages = [
        ...this._messages,
        { role: "assistant", content: "", uiEvents: [event] },
      ];
      this._streamingMsgIndex = this._messages.length - 1;
    }
    this.updateComplete.then(() => this._scrollToBottom());
  }

  private _handleToolEnd(tool: string, ok: boolean) {
    const idx = this._activeToolMap.get(tool);
    if (idx !== undefined) {
      const card = this._toolCards[idx];
      const elapsed = ((Date.now() - card.startTime) / 1000).toFixed(1);
      const updated = [...this._toolCards];
      updated[idx] = {
        ...card,
        status: ok ? "done" : "failed",
        elapsed: `${elapsed}s`,
      };
      this._toolCards = updated;
      this._activeToolMap.delete(tool);
    }
    this._showTyping = true;
    this._runningToolCount = Math.max(0, this._runningToolCount - 1);
  }

  private _handleDelta(delta: string) {
    this._showTyping = false;
    if (this._streamingMsgIndex < 0) {
      this._messages = [
        ...this._messages,
        { role: "assistant", content: "" },
      ];
      this._streamingMsgIndex = this._messages.length - 1;
    }
    this._fullStreamText += delta;
    const updated = [...this._messages];
    updated[this._streamingMsgIndex] = {
      ...updated[this._streamingMsgIndex],
      content: this._fullStreamText,
    };
    this._messages = updated;
    this.updateComplete.then(() => this._scrollToBottom());
  }

  private _handleError(error: string) {
    this._showTyping = false;
    if (this._streamingMsgIndex < 0) {
      this._messages = [
        ...this._messages,
        { role: "assistant", content: `Error: ${error}` },
      ];
      this._streamingMsgIndex = this._messages.length - 1;
    } else {
      const updated = [...this._messages];
      updated[this._streamingMsgIndex] = {
        ...updated[this._streamingMsgIndex],
        content: `Error: ${error}`,
      };
      this._messages = updated;
    }
    this._isStreaming = false;
    this._focusInput();
  }

  private _handleDone() {
    this._showTyping = false;
    if (this._streamingMsgIndex < 0 && this._fullStreamText === "") {
      this._messages = [
        ...this._messages,
        {
          role: "assistant",
          content: "No response received. Please check the server logs.",
        },
      ];
    }
    this._isStreaming = false;
    this._streamingMsgIndex = -1;
    this._fullStreamText = "";
    this._focusInput();
    this._refreshSidebar();
  }

  private async _send() {
    const text = this._inputEl?.value?.trim() || "";
    if (!text && !this._attachedFile) return;

    if (this._inputEl) {
      this._inputEl.value = "";
      this._inputEl.style.height = "auto";
    }

    this._streamAbort?.abort();
    this._streamAbort = new AbortController();

    // Show messages area
    this._showWelcome = false;

    // Add user message
    const fileLabel = this._attachedFile ? this._attachedFile.name : undefined;
    this._messages = [
      ...this._messages,
      { role: "user", content: text, file: fileLabel },
    ];

    this._isStreaming = true;
    this._showTyping = true;
    this._streamingMsgIndex = -1;
    this._fullStreamText = "";

    await this.updateComplete;
    this._scrollToBottom();

    // Use ARP WebSocket when connected and no file attachment (WS doesn't support file upload)
    if (this._arpClient?.connected && !this._attachedFile) {
      this._clearFile();
      this._arpClient.sendText(text);
      return;
    }

    // Fallback to SSE
    const m = this.meta;
    const chatParam = m.chatParam || "message";
    let formData: FormData | undefined;
    const body: Record<string, unknown> = {};
    body[chatParam] = text;
    body["session_id"] = this._sessionId;

    if (this._attachedFile) {
      const fileParamName = m.fileParam || "file_path";
      formData = new FormData();
      formData.append(fileParamName, this._attachedFile);
      formData.append(chatParam, text);
      formData.append("session_id", this._sessionId);
    }

    this._clearFile();

    await streamSSE(
      m.path,
      body,
      {
        onToolStart: (event) => this._handleToolStart(event.tool),
        onToolRender: (event: ToolRenderEvent) => this._handleToolRender(event),
        onToolEnd: (event) => this._handleToolEnd(event.tool, event.ok !== false),
        onDelta: (delta) => this._handleDelta(delta),
        onError: (error) => this._handleError(error),
        onDone: () => this._handleDone(),
      },
      { formData, signal: this._streamAbort?.signal },
    );
  }

  // Handle custom haira-chat-input event (from UI components like confirm/choices)
  private _onChatInput(e: Event) {
    const text = (e as CustomEvent).detail?.text;
    if (text && !this._isStreaming) {
      if (this._inputEl) {
        this._inputEl.value = text;
      }
      this._send();
    }
  }

  // ---------- Helpers ----------

  private _getSuggestions(): string[] {
    if (this.meta.suggestions && this.meta.suggestions.length > 0) {
      return this.meta.suggestions;
    }
    return ["What can you help me with?", "Hello!"];
  }

  private _scrollToBottom() {
    if (this._messagesEl) {
      this._messagesEl.scrollTop = this._messagesEl.scrollHeight;
    }
  }

  private _focusInput() {
    this.updateComplete.then(() => {
      this._inputEl?.focus();
    });
  }

  private get _avatarValue(): string {
    return this.meta.avatar || "H";
  }

  private _logoHtml56(): string {
    if (this.meta.logo) {
      return `<img src="${this.meta.logo}" alt="" style="width:56px;height:56px;object-fit:contain">`;
    }
    return logoSvgStr
      .replace(/width="22"/, 'width="56"')
      .replace(/height="22"/, 'height="56"');
  }

  // ---------- Render ----------

  protected render() {
    const m = this.meta;
    if (!m) return nothing;

    return html`
      ${this._renderSidebar()}

      <div
        class="chat-main"
        @dragenter=${m.hasFile ? this._onDragEnter : nothing}
        @dragleave=${m.hasFile ? this._onDragLeave : nothing}
        @dragover=${m.hasFile ? this._onDragOver : nothing}
        @drop=${m.hasFile ? this._onDrop : nothing}
        @haira-chat-input=${this._onChatInput}
      >
        <!-- Sidebar open toggle (when collapsed) -->
        <button
          class="sidebar-toggle ${this._sidebarOpen ? "" : "visible"}"
          title="Show chats"
          @click=${() => (this._sidebarOpen = true)}
        >
          ${unsafeHTML(iconStrings.sidebar)}
        </button>

        ${this._renderWelcome()} ${this._renderMessages()}
        ${this._renderDropOverlay()} ${this._renderActivityToggle()}
        ${this._renderInputArea()} ${this._renderActivityPanel()}
      </div>
    `;
  }

  // ---- Sidebar ----

  private _renderSidebar() {
    return html`
      <div class="sidebar ${this._sidebarOpen ? "" : "collapsed"}">
        <div class="sidebar-header">
          <span class="sidebar-title">Chats</span>
          <button
            class="sidebar-btn"
            title="New chat"
            @click=${this._startNewChat}
          >
            ${unsafeHTML(iconStrings.plus)}
          </button>
          <button
            class="sidebar-btn"
            title="Close sidebar"
            @click=${() => (this._sidebarOpen = false)}
          >
            ${unsafeHTML(iconStrings.chevronLeft)}
          </button>
        </div>
        <div class="sidebar-list">
          ${this._sessions.length === 0
            ? html`<div class="sidebar-empty">No chats yet</div>`
            : this._sessions.map(
                (sess) => html`
                  <div
                    class="session-item ${sess.id === this._sessionId
                      ? "active"
                      : ""}"
                    @click=${() => this._switchSession(sess.id)}
                  >
                    <span class="session-icon"
                      >${unsafeHTML(iconStrings.chat)}</span
                    >
                    <span class="session-title"
                      >${sess.title || "New chat"}</span
                    >
                    <button
                      class="session-delete"
                      title="Delete"
                      @click=${(e: Event) => this._deleteSession(sess.id, e)}
                    >
                      ${unsafeHTML(iconStrings.trash)}
                    </button>
                  </div>
                `
              )}
        </div>
      </div>
    `;
  }

  // ---- Welcome screen ----

  private _renderWelcome() {
    if (!this._showWelcome) return nothing;

    const suggestions = this._getSuggestions();

    return html`
      <div class="welcome">
        <span class="welcome-icon">${unsafeHTML(this._logoHtml56())}</span>
        <h2>${this.meta.title || this.meta.name || "Chat"}</h2>
        ${this.meta.description
          ? html`<p>${this.meta.description}</p>`
          : nothing}
        <div class="suggestions">
          ${suggestions.map(
            (text) => html`
              <button
                class="suggestion"
                @click=${() => this._onSuggestionClick(text)}
              >
                ${text}
              </button>
            `
          )}
        </div>
      </div>
    `;
  }

  // ---- Messages area ----

  private _renderMessages() {
    const hasMessages = this._messages.length > 0;
    return html`
      <div
        class="messages ${hasMessages && !this._showWelcome ? "active" : ""}"
        id="messages-scroll"
      >
        <div class="messages-inner">
          ${this._messages.map((msg, idx) =>
            this._renderMessageGroup(msg, idx)
          )}

          <div class="typing ${this._showTyping ? "visible" : ""}">
            <div class="typing-dots">
              <span class="typing-dot"></span>
              <span class="typing-dot"></span>
              <span class="typing-dot"></span>
            </div>
            <span>Thinking...</span>
          </div>
        </div>
      </div>
    `;
  }

  private _renderMessageGroup(msg: ChatMessage, _idx: number) {
    return html`
      <haira-message
        .role=${msg.role}
        .content=${msg.content}
        .file=${msg.file || ""}
        .avatar=${msg.role === "assistant" ? this._avatarValue : ""}
      ></haira-message>
      ${msg.uiEvents && msg.uiEvents.length > 0
        ? msg.uiEvents.map(
            (event) => html`
              <haira-ui-renderer
                .event=${event}
                ?data-restored=${!!msg.restored}
              ></haira-ui-renderer>
            `
          )
        : nothing}
    `;
  }

  // ---- Drop overlay ----

  private _renderDropOverlay() {
    if (!this.meta.hasFile) return nothing;
    return html`
      <div class="drop-overlay ${this._dropVisible ? "visible" : ""}">
        <span class="drop-overlay-icon">${unsafeHTML(iconStrings.attach)}</span>
        <span class="drop-overlay-text">Drop file to attach</span>
      </div>
    `;
  }

  // ---- Activity toggle button ----

  private _renderActivityToggle() {
    return html`
      <button
        class="activity-toggle"
        title="Toggle activity panel"
        @click=${() => (this._panelOpen = !this._panelOpen)}
      >
        ${unsafeHTML(iconStrings.activity)}
        <span
          class="badge ${this._runningToolCount > 0 ? "visible" : ""}"
          >${this._runningToolCount}</span
        >
      </button>
    `;
  }

  // ---- Input area ----

  private _renderInputArea() {
    const m = this.meta;
    const placeholder = m.hasFile
      ? "Type a message or drop a file..."
      : "Type a message...";

    return html`
      <div class="input-area">
        <div class="input-card">
          ${this._renderFileChip()}
          <div class="input-row">
            ${m.hasFile
              ? html`
                  <button
                    class="attach-btn"
                    title="Attach file"
                    @click=${this._onAttachClick}
                  >
                    ${unsafeHTML(iconStrings.attach)}
                  </button>
                `
              : nothing}
            <textarea
              id="chat-input"
              placeholder=${placeholder}
              rows="1"
              @input=${this._onInputChange}
              @keydown=${this._onKeyDown}
            ></textarea>
            <button
              class="send-btn"
              title="Send"
              ?disabled=${this._isStreaming}
              @click=${this._send}
            >
              ${unsafeHTML(iconStrings.send)}
            </button>
          </div>
        </div>
        ${m.hasFile
          ? html`<input
              type="file"
              id="file-input"
              style="display:none"
              @change=${this._onFileChange}
            />`
          : nothing}
        <div class="input-hint">Enter to send, Shift+Enter for new line</div>
      </div>
    `;
  }

  private _renderFileChip() {
    if (!this._attachedFile) return nothing;

    return html`
      <div class="file-chip visible">
        <div class="file-chip-inner">
          <span class="file-chip-icon"
            >${unsafeHTML(iconStrings.file)}</span
          >
          <span class="file-chip-name">${this._attachedFile.name}</span>
          <span class="file-chip-size"
            >${formatBytes(this._attachedFile.size)}</span
          >
          <button
            class="file-chip-remove"
            title="Remove file"
            @click=${this._clearFile}
          >
            ${unsafeHTML(iconStrings.xSmall)}
          </button>
        </div>
      </div>
    `;
  }

  // ---- Activity panel ----

  private _renderActivityPanel() {
    return html`
      <div class="activity-panel ${this._panelOpen ? "" : "collapsed"}">
        <div class="panel-header">
          <span class="panel-header-icon"
            >${unsafeHTML(iconStrings.activity)}</span
          >
          <span class="panel-title">Activity</span>
          <span class="panel-count"
            >${this._totalToolCount > 0
              ? String(this._totalToolCount)
              : ""}</span
          >
          <button
            class="panel-close"
            title="Close panel"
            @click=${() => (this._panelOpen = false)}
          >
            ${unsafeHTML(iconStrings.xSmall)}
          </button>
        </div>
        <div class="panel-body" id="panel-body">
          ${this._toolCards.length === 0
            ? html`<div class="panel-empty">No activity yet</div>`
            : this._toolCards.map(
                (card) => html`
                  <div class="tool-card">
                    <div class="tool-icon ${card.status}">
                      ${card.status === "running"
                        ? unsafeHTML(iconStrings.spinner)
                        : card.status === "done"
                          ? unsafeHTML(iconStrings.check)
                          : unsafeHTML(iconStrings.x)}
                    </div>
                    <div class="tool-info">
                      <div class="tool-name">${card.displayName}</div>
                      <div class="tool-status">
                        ${card.status === "running"
                          ? "Running..."
                          : card.status === "done"
                            ? "Completed"
                            : "Failed"}
                      </div>
                    </div>
                    ${card.elapsed
                      ? html`<span class="tool-duration">${card.elapsed}</span>`
                      : nothing}
                  </div>
                `
              )}
        </div>
      </div>
    `;
  }

  // ---- Focus input after first render ----

  protected firstUpdated() {
    requestAnimationFrame(() => this._inputEl?.focus());
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-chat": HairaChat;
  }
}
