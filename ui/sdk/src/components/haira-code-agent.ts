import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state, query } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles, scrollbarStyles } from "../core/styles";
import { iconStrings, logoSvgStr } from "../core/icons";
import { streamSSE } from "../services/sse-client";
import { ArpClient } from "@haira/arp";
import type {
  WorkflowMeta,
  ToolRenderEvent,
  ChatSessionSummary,
  ChatSessionDetail,
} from "../core/types";

// ---------- Local icons ----------

const iconFolder = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/></svg>`;
const iconTerminal = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg>`;
const iconGitBranch = `<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="6" y1="3" x2="6" y2="15"/><circle cx="18" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><path d="M18 9a9 9 0 01-9 9"/></svg>`;

// ---------- Interfaces ----------

interface ChatMessage {
  role: "user" | "assistant";
  content: string;
  uiEvents?: ToolRenderEvent[];
  restored?: boolean;
}

interface ToolCardState {
  name: string;
  displayName: string;
  detail?: string;
  status: "running" | "done" | "failed";
  startTime: number;
  elapsed?: string;
}

interface FileTreeNode {
  name: string;
  path: string;
  isDir: boolean;
  children: FileTreeNode[];
  expanded: boolean;
}

// ---------- Component ----------

@customElement("haira-code-agent")
export class HairaCodeAgent extends LitElement {
  static styles = [
    baseStyles,
    scrollbarStyles,
    css`
      :host {
        display: flex;
        flex-direction: column;
        flex: 1;
        overflow: hidden;
        background: var(--haira-bg);
      }

      /* ---- Topbar ---- */
      .topbar {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.4rem 0.75rem;
        border-bottom: 1px solid var(--haira-border);
        flex-shrink: 0;
        background: var(--haira-bg);
        height: 36px;
        box-sizing: border-box;
      }
      .topbar-toggle {
        background: none;
        border: none;
        color: var(--haira-muted);
        cursor: pointer;
        display: flex;
        align-items: center;
        padding: 0.2rem;
        border-radius: 4px;
        transition: all 0.15s;
      }
      .topbar-toggle:hover {
        color: var(--haira-accent);
        background: var(--haira-accent-dim);
      }
      .topbar-title {
        font-size: 0.82rem;
        font-weight: 600;
        color: var(--haira-text);
        flex: 1;
      }
      .topbar-session-btn {
        background: none;
        border: none;
        color: var(--haira-muted);
        cursor: pointer;
        display: flex;
        align-items: center;
        padding: 0.2rem;
        border-radius: 4px;
        font-size: 0.72rem;
        gap: 0.3rem;
        transition: all 0.15s;
      }
      .topbar-session-btn:hover {
        color: var(--haira-accent);
        background: var(--haira-accent-dim);
      }

      /* ---- Main layout ---- */
      .main-layout {
        display: flex;
        flex: 1;
        overflow: hidden;
      }

      /* ---- File tree sidebar ---- */
      .file-tree {
        width: 180px;
        flex-shrink: 0;
        display: flex;
        flex-direction: column;
        border-right: 1px solid var(--haira-border);
        background: var(--haira-bg);
        overflow: hidden;
        transition: width 0.2s, opacity 0.2s;
      }
      .file-tree.collapsed {
        width: 0;
        opacity: 0;
        pointer-events: none;
      }
      .file-tree-header {
        display: flex;
        align-items: center;
        gap: 0.35rem;
        padding: 0.45rem 0.6rem;
        border-bottom: 1px solid var(--haira-border);
        flex-shrink: 0;
      }
      .file-tree-label {
        font-size: 0.68rem;
        font-weight: 600;
        color: var(--haira-muted);
        text-transform: uppercase;
        letter-spacing: 0.05em;
        flex: 1;
      }
      .file-tree-icon {
        display: flex;
        color: var(--haira-muted);
      }
      .file-tree-body {
        flex: 1;
        overflow-y: auto;
        padding: 0.25rem 0;
        font-size: 0.75rem;
        font-family: var(--haira-mono);
      }
      .file-tree-empty {
        padding: 1rem 0.75rem;
        text-align: center;
        font-size: 0.72rem;
        color: var(--haira-muted);
        opacity: 0.5;
        font-family: var(--haira-font);
      }
      .tree-item {
        display: flex;
        align-items: center;
        gap: 0.3rem;
        padding: 0.2rem 0.5rem 0.2rem calc(0.5rem + var(--depth, 0) * 0.75rem);
        color: var(--haira-text-dim);
        cursor: default;
        transition: background 0.1s;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }
      .tree-item:hover {
        background: var(--haira-bg-card);
      }
      .tree-item.tree-file {
        cursor: pointer;
      }
      .tree-item.tree-file:hover {
        color: var(--haira-accent);
      }
      .tree-item-icon {
        display: flex;
        flex-shrink: 0;
        opacity: 0.6;
      }
      .tree-dir {
        cursor: pointer;
      }
      .tree-dir .tree-item-icon {
        color: var(--haira-accent);
        opacity: 0.8;
      }
      .tree-chevron {
        display: flex;
        flex-shrink: 0;
        transition: transform 0.15s;
      }
      .tree-chevron.open {
        transform: rotate(90deg);
      }

      /* ---- Conversation area ---- */
      .conversation {
        flex: 1;
        display: flex;
        flex-direction: column;
        overflow: hidden;
        min-width: 0;
        position: relative;
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
      }
      .welcome.hidden {
        display: none;
      }
      .welcome-icon {
        opacity: 0.15;
      }
      .welcome-icon img {
        width: 48px;
        height: 48px;
        object-fit: contain;
        opacity: 1;
      }
      .welcome h2 {
        font-size: 1.05rem;
        font-weight: 600;
        color: var(--haira-text);
      }
      .welcome p {
        font-size: 0.82rem;
        color: var(--haira-muted);
        text-align: center;
        max-width: 480px;
        line-height: 1.5;
      }
      .suggestions {
        display: flex;
        flex-wrap: wrap;
        gap: 0.45rem;
        justify-content: center;
        margin-top: 0.5rem;
        max-width: 600px;
      }
      .suggestion {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        color: var(--haira-text-dim);
        padding: 0.4rem 0.75rem;
        border-radius: 16px;
        font-size: 0.75rem;
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
        width: 100%;
        padding: 1rem 0 0;
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
      }
      .messages-inner > haira-message {
        width: 100%;
        padding: 0 1rem;
        box-sizing: border-box;
      }
      .messages-inner > haira-ui-renderer {
        width: 100%;
        padding: 0 1rem 0.5rem;
        box-sizing: border-box;
      }

      /* ---- File preview ---- */
      .file-preview {
        flex: 1;
        display: flex;
        flex-direction: column;
        overflow: hidden;
      }
      .file-preview-header {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 1rem;
        border-bottom: 1px solid var(--haira-border);
        flex-shrink: 0;
        background: var(--haira-bg-card);
      }
      .file-preview-path {
        flex: 1;
        font-size: 0.78rem;
        font-family: var(--haira-mono);
        color: var(--haira-text);
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      .file-preview-meta {
        font-size: 0.68rem;
        color: var(--haira-muted);
        font-family: var(--haira-mono);
      }
      .file-preview-close {
        background: none;
        border: none;
        color: var(--haira-muted);
        cursor: pointer;
        display: flex;
        align-items: center;
        padding: 0.2rem;
        border-radius: 4px;
        transition: all 0.15s;
      }
      .file-preview-close:hover {
        color: var(--haira-text);
        background: var(--haira-bg-elevated);
      }
      .file-preview-body {
        flex: 1;
        overflow: auto;
        padding: 0.75rem 1rem;
      }
      .file-preview-body pre {
        margin: 0;
        font-size: 0.8rem;
        font-family: var(--haira-mono);
        line-height: 1.6;
        color: var(--haira-text);
        white-space: pre;
        tab-size: 4;
      }
      .file-preview-truncated {
        padding: 0.5rem 1rem;
        font-size: 0.7rem;
        color: var(--haira-muted);
        text-align: center;
        border-top: 1px solid var(--haira-border);
        background: var(--haira-bg-card);
      }

      /* ---- Typing indicator ---- */
      .typing {
        display: none;
        padding: 0.25rem 1rem;
        align-items: center;
        gap: 0.4rem;
        font-size: 0.72rem;
        color: var(--haira-muted);
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
        width: 4px;
        height: 4px;
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

      /* ---- Input area ---- */
      .input-area {
        padding: 0.5rem 0.75rem 0.65rem;
        flex-shrink: 0;
        border-top: 1px solid var(--haira-border);
        background: var(--haira-bg);
      }
      .input-card {
        display: flex;
        flex-direction: column;
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        transition: border-color 0.15s;
      }
      .input-card:focus-within {
        border-color: var(--haira-border-focus);
      }
      .input-row {
        display: flex;
        align-items: flex-end;
        gap: 0.35rem;
        padding: 0.3rem;
      }
      textarea {
        flex: 1;
        background: transparent;
        border: none;
        color: var(--haira-text);
        padding: 0.45rem 0.35rem;
        font-size: 0.85rem;
        font-family: var(--haira-mono);
        resize: none;
        min-height: 40px;
        max-height: 180px;
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
        width: 32px;
        height: 32px;
        border-radius: 6px;
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
      .stop-btn {
        background: var(--haira-error, #ef4444);
        color: #fff;
        border: none;
        width: 32px;
        height: 32px;
        border-radius: 6px;
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        transition: all 0.15s;
        flex-shrink: 0;
      }
      .stop-btn:hover {
        opacity: 0.85;
      }
      .input-hint {
        text-align: center;
        font-size: 0.65rem;
        color: var(--haira-muted);
        opacity: 0.4;
        padding-top: 0.25rem;
      }

      /* ---- Activity panel (persistent, not overlay) ---- */
      .activity-panel {
        width: 240px;
        flex-shrink: 0;
        display: flex;
        flex-direction: column;
        border-left: 1px solid var(--haira-border);
        background: var(--haira-bg);
        overflow: hidden;
        transition: width 0.2s, opacity 0.2s;
      }
      .activity-panel.collapsed {
        width: 0;
        opacity: 0;
        pointer-events: none;
      }
      .panel-header {
        display: flex;
        align-items: center;
        gap: 0.4rem;
        padding: 0.45rem 0.6rem;
        border-bottom: 1px solid var(--haira-border);
        flex-shrink: 0;
      }
      .panel-header-icon {
        display: flex;
        color: var(--haira-muted);
      }
      .panel-title {
        font-size: 0.68rem;
        font-weight: 600;
        color: var(--haira-muted);
        text-transform: uppercase;
        letter-spacing: 0.05em;
        flex: 1;
      }
      .panel-count {
        font-size: 0.65rem;
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
        padding: 0.15rem;
        border-radius: 3px;
        transition: all 0.15s;
      }
      .panel-close:hover {
        color: var(--haira-text);
        background: var(--haira-bg-elevated);
      }
      .panel-body {
        flex: 1;
        overflow-y: auto;
        padding: 0.35rem;
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
      }
      .panel-empty {
        display: flex;
        align-items: center;
        justify-content: center;
        flex: 1;
        font-size: 0.72rem;
        color: var(--haira-muted);
        opacity: 0.5;
      }
      .tool-card {
        display: flex;
        align-items: center;
        gap: 0.4rem;
        padding: 0.35rem 0.5rem;
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: 6px;
        animation: fadeSlideUp 0.2s ease-out;
      }
      .tool-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 20px;
        height: 20px;
        border-radius: 4px;
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
        font-size: 0.72rem;
        font-weight: 600;
        color: var(--haira-text);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }
      .tool-status,
      .tool-detail {
        font-size: 0.65rem;
        color: var(--haira-muted);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        max-width: 160px;
        font-family: var(--haira-mono);
      }
      .tool-duration {
        font-family: var(--haira-mono);
        font-size: 0.62rem;
        color: var(--haira-muted);
        flex-shrink: 0;
      }

      /* ---- Status bar ---- */
      .status-bar {
        display: flex;
        align-items: center;
        gap: 0.75rem;
        padding: 0 0.75rem;
        height: 24px;
        flex-shrink: 0;
        border-top: 1px solid var(--haira-border);
        background: var(--haira-bg-card);
        font-size: 0.68rem;
        font-family: var(--haira-mono);
        color: var(--haira-muted);
      }
      .status-item {
        display: flex;
        align-items: center;
        gap: 0.3rem;
      }
      .status-item svg {
        opacity: 0.7;
      }
      .status-spacer {
        flex: 1;
      }

      /* ---- Sessions dropdown ---- */
      .sessions-dropdown {
        position: absolute;
        top: 36px;
        left: 0;
        width: 240px;
        max-height: 300px;
        overflow-y: auto;
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: 8px;
        box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
        z-index: 100;
        padding: 0.25rem;
        display: flex;
        flex-direction: column;
        gap: 1px;
      }
      .sessions-dropdown.hidden {
        display: none;
      }
      .dropdown-item {
        display: flex;
        align-items: center;
        gap: 0.4rem;
        padding: 0.4rem 0.5rem;
        border-radius: 5px;
        cursor: pointer;
        font-size: 0.75rem;
        color: var(--haira-text-dim);
        transition: all 0.1s;
      }
      .dropdown-item:hover {
        background: var(--haira-bg-elevated);
        color: var(--haira-text);
      }
      .dropdown-item.active {
        background: var(--haira-accent-dim);
        color: var(--haira-accent);
      }
      .dropdown-item-title {
        flex: 1;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      .dropdown-delete {
        display: none;
        background: none;
        border: none;
        color: var(--haira-muted);
        cursor: pointer;
        padding: 0.1rem;
        border-radius: 3px;
      }
      .dropdown-item:hover .dropdown-delete {
        display: flex;
      }
      .dropdown-delete:hover {
        color: var(--haira-error);
      }
    `,
  ];

  // ---- Properties ----

  @property({ type: Object }) meta!: WorkflowMeta;

  // ---- State ----

  @state() private _sessionId = "";
  @state() private _sessions: ChatSessionSummary[] = [];
  @state() private _messages: ChatMessage[] = [];
  @state() private _isStreaming = false;
  @state() private _showTyping = false;
  @state() private _showWelcome = true;
  @state() private _toolCards: ToolCardState[] = [];
  @state() private _runningToolCount = 0;
  @state() private _totalToolCount = 0;
  @state() private _panelOpen = true;
  @state() private _fileTreeOpen = true;
  @state() private _touchedFiles: string[] = [];
  @state() private _fileTree: FileTreeNode[] = [];
  @state() private _sessionsDropdownOpen = false;
  @state() private _projectCwd = "";
  @state() private _projectTree: FileTreeNode[] = [];
  @state() private _previewPath = "";
  @state() private _previewContent = "";
  @state() private _previewLang = "text";
  @state() private _previewLines = 0;
  @state() private _previewTruncated = false;
  @state() private _gitBranch = "";

  // ---- DOM refs ----

  @query("#messages-scroll") private _messagesEl!: HTMLDivElement;
  @query("#chat-input") private _inputEl!: HTMLTextAreaElement;
  @query("#panel-body") private _panelBodyEl!: HTMLDivElement;

  // ---- Internal ----

  private _streamAbort: AbortController | null = null;
  private _streamingMsgIndex = -1;
  private _fullStreamText = "";
  private _activeToolMap = new Map<string, number>();
  private _arpClient: ArpClient | null = null;

  // ---------- Lifecycle ----------

  connectedCallback() {
    super.connectedCallback();
    this._initSession();
    this._loadProjectFiles();
    this._loadGitBranch();
    document.addEventListener("click", this._onDocClick);
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    this._streamAbort?.abort();
    this._arpClient?.disconnect();
    this._arpClient = null;
    document.removeEventListener("click", this._onDocClick);
  }

  private _onDocClick = (e: Event) => {
    if (this._sessionsDropdownOpen) {
      const path = e.composedPath();
      const dropdown = this.shadowRoot?.querySelector(".sessions-dropdown");
      const btn = this.shadowRoot?.querySelector(".topbar-session-btn");
      if (dropdown && !path.includes(dropdown) && btn && !path.includes(btn)) {
        this._sessionsDropdownOpen = false;
      }
    }
  };

  // ---------- Session init ----------

  private async _initSession() {
    const url = new URL(window.location.href);
    const urlSession = url.searchParams.get("session");

    if (urlSession) {
      this._sessionId = urlSession;
    } else {
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
        // ignore
      }
      this._sessionId = crypto.randomUUID();
      this._setSessionUrl(this._sessionId);
    }

    this._loadSession(this._sessionId);
    this._refreshSidebar();
    this._connectArp();
  }

  private _setSessionUrl(sessionId: string) {
    const url = new URL(window.location.href);
    url.searchParams.set("session", sessionId);
    window.history.replaceState({}, "", url.toString());
  }

  private _connectArp() {
    try {
      const wsProto = location.protocol === "https:" ? "wss:" : "ws:";
      const wsPath = this.meta.arpUrl ?? "/_arp/v1";
      const wsUrl = `${wsProto}//${location.host}${wsPath}`;
      this._arpClient = new ArpClient(
        { url: wsUrl, sessionId: this._sessionId },
        {
          onDelta: (text) => this._handleDelta(text),
          onToolStart: (tool, args) => this._handleToolStart(tool, args),
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
      // ignore
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
      this._restoreActivity();
      this.updateComplete.then(() => this._scrollToBottom());
    } catch {
      // session not found
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
    this._isStreaming = false;
    this._showTyping = false;
    this._showWelcome = true;
    this._streamingMsgIndex = -1;
    this._fullStreamText = "";
    this._touchedFiles = [];
    this._fileTree = [];
    this._sessionsDropdownOpen = false;

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
    this._switchSession(crypto.randomUUID());
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

  // ---------- File tree ----------

  private _addTouchedFile(filePath: string) {
    // Normalize: strip leading ./ or /
    let p = filePath.replace(/^\.\//, "");
    if (p.startsWith("/")) {
      // Try to make relative using meta description (which contains cwd)
      const cwd = this._getCwd();
      if (cwd && p.startsWith(cwd)) {
        p = p.substring(cwd.length + 1);
      }
    }
    if (!p || this._touchedFiles.includes(p)) return;
    this._touchedFiles = [...this._touchedFiles, p];
    this._rebuildFileTree();
  }

  private _getCwd(): string {
    // Try to extract from meta description
    return this.meta.description || "";
  }

  private _rebuildFileTree() {
    const root: FileTreeNode[] = [];
    const dirMap = new Map<string, FileTreeNode>();

    for (const filepath of this._touchedFiles) {
      const parts = filepath.split("/");
      let currentLevel = root;
      let currentPath = "";

      for (let i = 0; i < parts.length; i++) {
        const part = parts[i];
        currentPath = currentPath ? `${currentPath}/${part}` : part;
        const isLast = i === parts.length - 1;

        if (isLast) {
          // File node
          if (!currentLevel.find((n) => n.path === currentPath)) {
            currentLevel.push({
              name: part,
              path: currentPath,
              isDir: false,
              children: [],
              expanded: false,
            });
          }
        } else {
          // Directory node
          let dirNode = dirMap.get(currentPath);
          if (!dirNode) {
            dirNode = {
              name: part,
              path: currentPath,
              isDir: true,
              children: [],
              expanded: true,
            };
            dirMap.set(currentPath, dirNode);
            currentLevel.push(dirNode);
          }
          currentLevel = dirNode.children;
        }
      }
    }

    // Sort: dirs first, then files, alphabetical
    const sortTree = (nodes: FileTreeNode[]) => {
      nodes.sort((a, b) => {
        if (a.isDir !== b.isDir) return a.isDir ? -1 : 1;
        return a.name.localeCompare(b.name);
      });
      for (const n of nodes) {
        if (n.isDir) sortTree(n.children);
      }
    };
    sortTree(root);

    this._fileTree = [...root];
  }

  private async _loadProjectFiles(dirPath = ".") {
    try {
      const resp = await fetch(
        `/_api/files?path=${encodeURIComponent(dirPath)}`
      );
      if (!resp.ok) return;
      const data = await resp.json();

      if (data.cwd) {
        this._projectCwd = data.cwd;
      }

      const files = (data.files || []) as Array<{
        name: string;
        isDir: boolean;
        size: number;
      }>;

      const nodes: FileTreeNode[] = files.map((f) => ({
        name: f.name,
        path: dirPath === "." ? f.name : `${dirPath}/${f.name}`,
        isDir: f.isDir,
        children: [],
        expanded: false,
      }));

      // Sort: dirs first, then files
      nodes.sort((a, b) => {
        if (a.isDir !== b.isDir) return a.isDir ? -1 : 1;
        return a.name.localeCompare(b.name);
      });

      if (dirPath === ".") {
        this._projectTree = nodes;
      } else {
        // Find the parent node and set its children
        const setChildren = (tree: FileTreeNode[]): boolean => {
          for (const n of tree) {
            if (n.path === dirPath) {
              n.children = nodes;
              return true;
            }
            if (n.isDir && setChildren(n.children)) return true;
          }
          return false;
        };
        setChildren(this._projectTree);
        this._projectTree = [...this._projectTree];
      }
    } catch {
      // ignore
    }
  }

  private async _loadGitBranch() {
    try {
      const resp = await fetch("/_api/git/branch");
      if (!resp.ok) return;
      const data = await resp.json();
      if (data.branch) this._gitBranch = data.branch;
    } catch {
      // not a git repo or endpoint unavailable
    }
  }

  private _toggleProjectDir(node: FileTreeNode) {
    node.expanded = !node.expanded;
    // Lazy-load children on first expand
    if (node.expanded && node.children.length === 0) {
      this._loadProjectFiles(node.path);
    }
    this._projectTree = [...this._projectTree];
  }

  private async _onFileClick(filePath: string) {
    try {
      const resp = await fetch(
        `/_api/files/read?path=${encodeURIComponent(filePath)}`
      );
      if (!resp.ok) return;
      const data = await resp.json();
      this._previewPath = data.path || filePath;
      this._previewContent = data.content || "";
      this._previewLang = data.language || "text";
      this._previewLines = data.lines || 0;
      this._previewTruncated = data.truncated || false;
      this._showWelcome = false;
    } catch {
      // ignore
    }
  }

  private _closePreview() {
    this._previewPath = "";
    this._previewContent = "";
  }

  private _toggleDir(node: FileTreeNode) {
    node.expanded = !node.expanded;
    this._fileTree = [...this._fileTree];
  }

  // ---------- Streaming handlers ----------

  private _handleToolStart(tool: string, args?: string) {
    this._showTyping = false;
    const displayName = tool
      .replace(/^render_/, "")
      .replace(/_/g, " ")
      .replace(/\b\w/g, (c) => c.toUpperCase());

    // Extract a short detail string from tool args
    let detail: string | undefined;
    if (args) {
      try {
        const parsed = JSON.parse(args);
        // Show the most useful arg: path, pattern, command, url, query, title
        detail =
          parsed.path ||
          parsed.pattern ||
          parsed.command ||
          parsed.url ||
          parsed.query ||
          parsed.title ||
          undefined;
        if (detail && detail.length > 40) {
          detail = detail.slice(0, 37) + "...";
        }
      } catch {
        /* ignore parse errors */
      }
    }

    const cardState: ToolCardState = {
      name: tool,
      displayName,
      detail,
      status: "running",
      startTime: Date.now(),
    };
    const idx = this._toolCards.length;
    this._toolCards = [...this._toolCards, cardState];
    this._activeToolMap.set(tool, idx);

    this._runningToolCount++;
    this._totalToolCount++;
    this._saveActivity();

    // Scroll panel to top since newest tools are shown first
    this.updateComplete.then(() => {
      if (this._panelBodyEl) {
        this._panelBodyEl.scrollTop = 0;
      }
    });
  }

  private _handleToolRender(event: ToolRenderEvent) {
    if (!event?.component || !event.props) return;

    // Extract file paths from rendered components
    this._extractPathsFromRender(event);

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

  private _extractPathsFromRender(event: ToolRenderEvent) {
    const props = event.props as Record<string, unknown>;

    // Code blocks: title is often a file path
    if (event.component === "code_block") {
      const title = (props.title as string) || "";
      if (title && /\.\w+$/.test(title)) {
        this._addTouchedFile(title);
      }
      // Check tabs for file paths
      const tabs = props.tabs as Array<{ title?: string }> | undefined;
      if (tabs) {
        for (const tab of tabs) {
          if (tab.title && /\.\w+$/.test(tab.title)) {
            this._addTouchedFile(tab.title);
          }
        }
      }
    }

    // Diff: title often contains "Edit: path/to/file.ext"
    if (event.component === "diff") {
      const title = (props.title as string) || "";
      const match = title.match(/(?:Edit:\s*)?(\S+\.\w+)/);
      if (match) this._addTouchedFile(match[1]);
    }

    // Table: rows may contain file paths (directory listings, search results)
    if (event.component === "table") {
      const rows = props.rows as string[][] | undefined;
      if (rows) {
        for (const row of rows) {
          for (const cell of row) {
            if (typeof cell === "string" && /^[a-zA-Z0-9_\-./]+\.\w+$/.test(cell)) {
              this._addTouchedFile(cell);
            }
          }
        }
      }
    }
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
      this._saveActivity();
    }

    // Extract file paths from tool names
    // The tool name is just the function name (e.g., "read_file"), not the args.
    // We rely on render events or delta content for file path extraction.

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

    // Try to extract file paths from streamed content
    this._extractFilePaths(delta);

    const updated = [...this._messages];
    updated[this._streamingMsgIndex] = {
      ...updated[this._streamingMsgIndex],
      content: this._fullStreamText,
    };
    this._messages = updated;
    this.updateComplete.then(() => this._scrollToBottom());
  }

  private _extractFilePaths(text: string) {
    const patterns = [
      /`([a-zA-Z0-9_\-./]+\.\w+)`/g,                   // `path/to/file.ext`
      /\[([a-zA-Z0-9_\-./]+\.\w+)\]/g,                  // [path/to/file.ext]
      /modified:\s+(\S+\.\w+)/g,                         // git status: modified: file.ext
      /new file:\s+(\S+\.\w+)/g,                         // git status: new file: file.ext
      /deleted:\s+(\S+\.\w+)/g,                          // git status: deleted: file.ext
      /renamed:\s+\S+\s+->\s+(\S+\.\w+)/g,              // git status: renamed: a -> b
      /(?:^|\s)((?:[\w.-]+\/)+[\w.-]+\.\w+)(?:\s|$|:)/gm, // bare path/to/file.ext in text
    ];
    for (const pattern of patterns) {
      let match;
      while ((match = pattern.exec(text)) !== null) {
        const p = match[1];
        if (p.length > 2 && p.length < 200 && !p.includes(" ")) {
          this._addTouchedFile(p);
        }
      }
    }
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
          content: "No response received. Check server logs.",
        },
      ];
    }
    this._isStreaming = false;
    this._streamingMsgIndex = -1;
    this._fullStreamText = "";
    this.updateComplete.then(() => this._scrollToBottom());
    this._focusInput();
    this._refreshSidebar();
  }

  // ---------- Input ----------

  private _onInputChange() {
    const el = this._inputEl;
    if (!el) return;
    el.style.height = "auto";
    el.style.height = `${Math.min(el.scrollHeight, 180)}px`;
  }

  private _onKeyDown(e: KeyboardEvent) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      this._send();
    }
  }

  private _onSuggestionClick(text: string) {
    if (this._inputEl) this._inputEl.value = text;
    this._send();
  }

  private async _send() {
    const text = this._inputEl?.value?.trim() || "";
    if (!text) return;

    if (this._inputEl) {
      this._inputEl.value = "";
      this._inputEl.style.height = "auto";
    }

    this._streamAbort?.abort();
    this._streamAbort = new AbortController();

    // Close file preview when sending a message
    this._previewPath = "";
    this._previewContent = "";

    this._showWelcome = false;
    this._messages = [
      ...this._messages,
      { role: "user", content: text },
    ];

    this._isStreaming = true;
    this._showTyping = true;
    this._streamingMsgIndex = -1;
    this._fullStreamText = "";

    await this.updateComplete;
    this._scrollToBottom();

    // ARP WebSocket
    if (this._arpClient?.connected) {
      this._arpClient.sendText(text);
      return;
    }

    // SSE fallback
    const m = this.meta;
    const chatParam = m.chatParam || "message";
    const body: Record<string, unknown> = {};
    body[chatParam] = text;
    body["session_id"] = this._sessionId;

    await streamSSE(
      m.path,
      body,
      {
        onToolStart: (event) => this._handleToolStart(event.tool, event.args),
        onToolRender: (event: ToolRenderEvent) => this._handleToolRender(event),
        onToolEnd: (event) => this._handleToolEnd(event.tool, event.ok !== false),
        onDelta: (delta) => this._handleDelta(delta),
        onError: (error) => this._handleError(error),
        onDone: () => this._handleDone(),
      },
      { signal: this._streamAbort?.signal },
    );
  }

  private _stopStreaming() {
    this._streamAbort?.abort();
    this._streamAbort = null;
    this._arpClient?.disconnect();
    this._arpClient = null;
    this._isStreaming = false;
    this._showTyping = false;
    this._runningToolCount = 0;

    // Mark any running tool cards as failed
    const updated = this._toolCards.map((card) =>
      card.status === "running"
        ? { ...card, status: "failed" as const, elapsed: "stopped" }
        : card
    );
    this._toolCards = updated;

    // Finalize any in-progress message
    if (this._streamingMsgIndex >= 0) {
      this._streamingMsgIndex = -1;
      this._fullStreamText = "";
    }

    this._focusInput();
    // Reconnect ARP for next message
    this._connectArp();
  }

  private _onChatInput(e: Event) {
    const text = (e as CustomEvent).detail?.text;
    if (text && !this._isStreaming) {
      if (this._inputEl) this._inputEl.value = text;
      this._send();
    }
  }

  // ---------- Helpers ----------

  private _scrollToBottom() {
    // Use rAF to ensure the browser has reflowed after DOM updates
    requestAnimationFrame(() => {
      if (this._messagesEl) {
        this._messagesEl.scrollTop = this._messagesEl.scrollHeight;
      }
    });
  }

  private _focusInput() {
    this.updateComplete.then(() => this._inputEl?.focus());
  }

  // ---- Activity persistence ----

  private _activityStorageKey(): string {
    return `haira-activity-${this._sessionId}`;
  }

  private _saveActivity() {
    if (!this._sessionId) return;
    try {
      const data = {
        toolCards: this._toolCards,
        totalToolCount: this._totalToolCount,
      };
      sessionStorage.setItem(this._activityStorageKey(), JSON.stringify(data));
    } catch { /* quota exceeded — ignore */ }
  }

  private _restoreActivity() {
    if (!this._sessionId) return;
    try {
      const raw = sessionStorage.getItem(this._activityStorageKey());
      if (!raw) return;
      const data = JSON.parse(raw);
      if (data.toolCards?.length > 0) {
        this._toolCards = data.toolCards;
        this._totalToolCount = data.totalToolCount || data.toolCards.length;
      }
    } catch { /* parse error — ignore */ }
  }

  private get _avatarValue(): string {
    return this.meta.avatar || ">";
  }

  private _logoHtml48(): string {
    if (this.meta.logo) {
      return `<img src="${this.meta.logo}" alt="" style="width:48px;height:48px;object-fit:contain">`;
    }
    return logoSvgStr
      .replace(/width="22"/, 'width="48"')
      .replace(/height="22"/, 'height="48"');
  }

  private _getSuggestions(): string[] {
    if (this.meta.suggestions && this.meta.suggestions.length > 0) {
      return this.meta.suggestions;
    }
    return ["What files are in this project?", "Show me the project structure"];
  }

  // ---------- Render ----------

  protected render() {
    if (!this.meta) return nothing;

    return html`
      ${this._renderTopbar()}
      <div class="main-layout">
        ${this._renderFileTree()}
        <div class="conversation" @haira-chat-input=${this._onChatInput}>
          ${this._previewPath
            ? this._renderFilePreview()
            : html`
                ${this._renderWelcome()}
                ${this._renderMessages()}
              `}
          ${this._renderInputArea()}
        </div>
        ${this._renderActivityPanel()}
      </div>
      ${this._renderStatusBar()}
    `;
  }

  // ---- Topbar ----

  private _renderTopbar() {
    return html`
      <div class="topbar">
        <button
          class="topbar-toggle"
          title="Toggle file tree"
          @click=${() => (this._fileTreeOpen = !this._fileTreeOpen)}
        >
          ${unsafeHTML(iconStrings.sidebar)}
        </button>
        <span class="topbar-title">${this.meta.title || "Code Agent"}</span>

        <button
          class="topbar-session-btn"
          title="Sessions"
          @click=${(e: Event) => {
            e.stopPropagation();
            this._sessionsDropdownOpen = !this._sessionsDropdownOpen;
            if (this._sessionsDropdownOpen) this._refreshSidebar();
          }}
        >
          ${unsafeHTML(iconStrings.chat)}
          Sessions
        </button>
        <button
          class="topbar-toggle"
          title="New chat"
          @click=${this._startNewChat}
        >
          ${unsafeHTML(iconStrings.plus)}
        </button>
        <button
          class="topbar-toggle"
          title="Toggle activity"
          @click=${() => (this._panelOpen = !this._panelOpen)}
        >
          ${unsafeHTML(iconStrings.activity)}
          ${this._runningToolCount > 0
            ? html`<span
                style="font-size:0.6rem;color:var(--haira-accent);font-weight:700"
                >${this._runningToolCount}</span
              >`
            : nothing}
        </button>

        ${this._renderSessionsDropdown()}
      </div>
    `;
  }

  private _renderSessionsDropdown() {
    return html`
      <div
        class="sessions-dropdown ${this._sessionsDropdownOpen ? "" : "hidden"}"
      >
        ${this._sessions.length === 0
          ? html`<div
              style="padding:0.75rem;text-align:center;font-size:0.72rem;color:var(--haira-muted)"
            >
              No sessions yet
            </div>`
          : this._sessions.map(
              (sess) => html`
                <div
                  class="dropdown-item ${sess.id === this._sessionId
                    ? "active"
                    : ""}"
                  @click=${() => this._switchSession(sess.id)}
                >
                  <span class="dropdown-item-title"
                    >${sess.title || "New chat"}</span
                  >
                  <button
                    class="dropdown-delete"
                    title="Delete"
                    @click=${(e: Event) => this._deleteSession(sess.id, e)}
                  >
                    ${unsafeHTML(iconStrings.xSmall)}
                  </button>
                </div>
              `
            )}
      </div>
    `;
  }

  // ---- File tree ----

  private _renderFileTree() {
    return html`
      <div class="file-tree ${this._fileTreeOpen ? "" : "collapsed"}">
        <div class="file-tree-header">
          <span class="file-tree-icon">${unsafeHTML(iconFolder)}</span>
          <span class="file-tree-label">Explorer</span>
        </div>
        <div class="file-tree-body">
          ${this._projectTree.length === 0
            ? html`<div class="file-tree-empty">Loading...</div>`
            : this._projectTree.map((node) =>
                this._renderProjectNode(node, 0)
              )}
        </div>
      </div>
    `;
  }

  private _renderProjectNode(
    node: FileTreeNode,
    depth: number
  ): ReturnType<typeof html> {
    // Check if this file was touched by the agent
    const touched = this._touchedFiles.includes(node.path);

    if (node.isDir) {
      return html`
        <div
          class="tree-item tree-dir"
          style="--depth: ${depth}"
          @click=${() => this._toggleProjectDir(node)}
        >
          <span class="tree-chevron ${node.expanded ? "open" : ""}">
            ${unsafeHTML(iconStrings.chevron)}
          </span>
          <span class="tree-item-icon">${unsafeHTML(iconFolder)}</span>
          ${node.name}
        </div>
        ${node.expanded
          ? node.children.map((child) =>
              this._renderProjectNode(child, depth + 1)
            )
          : nothing}
      `;
    }

    return html`
      <div
        class="tree-item tree-file"
        style="--depth: ${depth + 0.5}${touched
          ? ";color:var(--haira-accent)"
          : ""}"
        @click=${() => this._onFileClick(node.path)}
      >
        <span class="tree-item-icon">${unsafeHTML(iconStrings.file)}</span>
        ${node.name}
      </div>
    `;
  }

  // ---- File preview ----

  private _renderFilePreview() {
    return html`
      <div class="file-preview">
        <div class="file-preview-header">
          <span class="file-preview-path">${this._previewPath}</span>
          <span class="file-preview-meta">${this._previewLines} lines · ${this._previewLang}</span>
          <button
            class="file-preview-close"
            title="Close preview"
            @click=${this._closePreview}
          >
            ${unsafeHTML(iconStrings.xSmall)}
          </button>
        </div>
        <div class="file-preview-body">
          <pre>${this._previewContent}</pre>
        </div>
        ${this._previewTruncated
          ? html`<div class="file-preview-truncated">
              File truncated (showing first 500KB)
            </div>`
          : nothing}
      </div>
    `;
  }

  // ---- Welcome ----

  private _renderWelcome() {
    if (!this._showWelcome) return nothing;
    const suggestions = this._getSuggestions();

    return html`
      <div class="welcome">
        <span class="welcome-icon">${unsafeHTML(this._logoHtml48())}</span>
        <h2>${this.meta.title || "Code Agent"}</h2>
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

  // ---- Messages ----

  private _renderMessages() {
    const hasMessages = this._messages.length > 0;
    return html`
      <div
        class="messages ${hasMessages && !this._showWelcome ? "active" : ""}"
        id="messages-scroll"
      >
        <div class="messages-inner">
          ${this._messages.map((msg) => this._renderMessageGroup(msg))}

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

  private _renderMessageGroup(msg: ChatMessage) {
    const uiBlock =
      msg.uiEvents && msg.uiEvents.length > 0
        ? msg.uiEvents.map(
            (event) => html`
              <haira-ui-renderer
                .event=${event}
                ?data-restored=${!!msg.restored}
              ></haira-ui-renderer>
            `
          )
        : nothing;

    // Render tool UI before the text so the LLM summary appears after tool output
    return html`
      ${uiBlock}
      <haira-message
        .role=${msg.role}
        .content=${msg.content}
        .avatar=${msg.role === "assistant" ? this._avatarValue : ""}
      ></haira-message>
    `;
  }

  // ---- Input ----

  private _renderInputArea() {
    return html`
      <div class="input-area">
        <div class="input-card">
          <div class="input-row">
            <textarea
              id="chat-input"
              placeholder="Ask about code, run commands, edit files..."
              rows="1"
              @input=${this._onInputChange}
              @keydown=${this._onKeyDown}
            ></textarea>
            ${this._isStreaming
              ? html`<button
                  class="stop-btn"
                  title="Stop"
                  @click=${this._stopStreaming}
                >
                  ${unsafeHTML(iconStrings.x)}
                </button>`
              : html`<button
                  class="send-btn"
                  title="Send"
                  @click=${this._send}
                >
                  ${unsafeHTML(iconStrings.send)}
                </button>`}
          </div>
        </div>
        <div class="input-hint">Enter to send, Shift+Enter for new line</div>
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
            title="Close"
            @click=${() => (this._panelOpen = false)}
          >
            ${unsafeHTML(iconStrings.xSmall)}
          </button>
        </div>
        <div class="panel-body" id="panel-body">
          ${this._toolCards.length === 0
            ? html`<div class="panel-empty">No activity yet</div>`
            : [...this._toolCards].reverse().map(
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
                      <div class="tool-detail">
                        ${card.detail || (card.status === "running"
                          ? "Running..."
                          : card.status === "done"
                            ? "Done"
                            : "Failed")}
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

  // ---- Status bar ----

  private _renderStatusBar() {
    const cwd = this._projectCwd || this.meta.description || "Project";
    const fileCount = this._touchedFiles.length;

    return html`
      <div class="status-bar">
        <span class="status-item">
          ${unsafeHTML(iconTerminal)}
          ${cwd}
        </span>
        <span class="status-spacer"></span>
        ${fileCount > 0
          ? html`<span class="status-item">
              ${unsafeHTML(iconStrings.file)}
              ${fileCount} file${fileCount !== 1 ? "s" : ""} touched
            </span>`
          : nothing}
        ${this._gitBranch
          ? html`<span class="status-item">
              ${unsafeHTML(iconGitBranch)}
              ${this._gitBranch}
            </span>`
          : nothing}
      </div>
    `;
  }

  // ---- Focus ----

  protected firstUpdated() {
    requestAnimationFrame(() => this._inputEl?.focus());
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-code-agent": HairaCodeAgent;
  }
}
