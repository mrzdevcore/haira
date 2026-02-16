import {
  baseStyles,
  sharedKeyframes,
  scrollbarStyles,
  logoSvg,
} from "../theme";
import { streamSSE } from "../sse";
import type { WorkflowMeta, ToolRenderEvent } from "../types";
import type { HairaMessage } from "./haira-message";
import type { HairaToolCard } from "./haira-tool-card";
import type { HairaUIRenderer } from "./haira-ui-renderer";

const iconAttach = `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21.44 11.05l-9.19 9.19a6 6 0 01-8.49-8.49l9.19-9.19a4 4 0 015.66 5.66l-9.2 9.19a2 2 0 01-2.83-2.83l8.49-8.48"/></svg>`;

const iconSend = `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>`;

const iconFile = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>`;

const iconX = `<svg width="12" height="12" viewBox="0 0 16 16" fill="none"><path d="M5 5L11 11M11 5L5 11" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>`;

export class HairaChat extends HTMLElement {
  private meta!: WorkflowMeta;
  private sessionId = "";
  private attachedFile: File | null = null;

  connectedCallback() {
    this.meta = JSON.parse(this.getAttribute("data-meta") || "{}");
    this.sessionId =
      sessionStorage.getItem(`haira-session-${this.meta.path}`) ||
      crypto.randomUUID();
    sessionStorage.setItem(`haira-session-${this.meta.path}`, this.sessionId);
    this.render();
  }

  private render() {
    const m = this.meta;
    const shadow = this.attachShadow({ mode: "open" });
    shadow.innerHTML = `
      <style>
        ${baseStyles}
        ${sharedKeyframes}
        :host {
          display: flex;
          flex-direction: column;
          flex: 1;
          overflow: hidden;
        }

        /* Welcome screen */
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
          border-color: var(--haira-gold);
          color: var(--haira-gold);
          background: var(--haira-gold-dim);
        }

        /* Messages area */
        .messages {
          flex: 1;
          overflow-y: auto;
          display: none;
          flex-direction: column;
          ${scrollbarStyles}
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

        /* Typing indicator (inside messages-inner) */
        .typing {
          display: none;
          padding: 0.25rem 0;
          align-items: center;
          gap: 0.4rem;
          font-size: 0.75rem;
          color: var(--haira-muted);
          margin-left: 2.25rem;
        }
        .typing.visible { display: flex; }
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
          background: var(--haira-gold);
          animation: bounce 1.4s ease-in-out infinite;
        }
        .typing-dot:nth-child(2) { animation-delay: 0.2s; }
        .typing-dot:nth-child(3) { animation-delay: 0.4s; }

        /* Drop overlay */
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
          border: 2px dashed var(--haira-gold);
          border-radius: var(--haira-radius);
          margin: 0.5rem;
        }
        .drop-overlay.visible {
          display: flex;
        }
        .drop-overlay-icon {
          color: var(--haira-gold);
          opacity: 0.7;
        }
        .drop-overlay-text {
          color: var(--haira-gold);
          font-size: 0.9rem;
          font-weight: 600;
        }

        /* Input area — sticky bottom */
        .input-area {
          padding: 0.75rem 1rem 1rem;
          flex-shrink: 0;
          background: var(--haira-bg);
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

        /* File chip */
        .file-chip {
          display: none;
          align-items: center;
          gap: 0.4rem;
          padding: 0.4rem 0.6rem 0;
          margin: 0 0.5rem;
        }
        .file-chip.visible { display: flex; }
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
        .file-chip-icon { color: var(--haira-gold); display: flex; }
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
          color: var(--haira-gold);
          background: var(--haira-gold-dim);
        }
        textarea {
          flex: 1;
          background: transparent;
          border: none;
          color: var(--haira-text);
          padding: 0.45rem 0.25rem;
          font-size: 0.88rem;
          font-family: var(--haira-font);
          resize: none;
          min-height: 24px;
          max-height: 140px;
          outline: none;
          line-height: 1.45;
        }
        textarea::placeholder {
          color: var(--haira-muted);
        }
        .send-btn {
          background: var(--haira-gold);
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
          background: var(--haira-gold-light);
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
      </style>

      <div class="welcome" id="welcome">
        <span class="welcome-icon">${logoSvg.replace(/width="22" height="22"/, 'width="56" height="56"')}</span>
        <h2>${this.esc(m.title || m.name || "Chat")}</h2>
        ${m.description ? `<p>${this.esc(m.description)}</p>` : ""}
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
        <span class="drop-overlay-icon">${iconAttach}</span>
        <span class="drop-overlay-text">Drop file to attach</span>
      </div>

      <div class="input-area">
        <div class="input-card" id="input-card">
          <div class="file-chip" id="file-chip">
            <div class="file-chip-inner">
              <span class="file-chip-icon">${iconFile}</span>
              <span class="file-chip-name" id="file-name"></span>
              <span class="file-chip-size" id="file-size"></span>
              <button class="file-chip-remove" id="file-remove" title="Remove file">${iconX}</button>
            </div>
          </div>
          <div class="input-row">
            ${m.hasFile ? `<button class="attach-btn" id="attach-btn" title="Attach file">${iconAttach}</button>` : ""}
            <textarea id="chat-input" placeholder="${m.hasFile ? "Type a message or drop a file..." : "Type a message..."}" rows="1"></textarea>
            <button class="send-btn" id="send-btn" title="Send">${iconSend}</button>
          </div>
        </div>
        ${m.hasFile ? `<input type="file" id="file-input" style="display:none" />` : ""}
        <div class="input-hint">Enter to send, Shift+Enter for new line</div>
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
    const suggestions = shadow.getElementById("suggestions")!;
    const attachBtn = m.hasFile ? shadow.getElementById("attach-btn") : null;
    const fileInput = m.hasFile
      ? (shadow.getElementById("file-input") as HTMLInputElement)
      : null;

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
      suggestions.appendChild(btn);
    }

    // File attach button (only if workflow accepts files)
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

      // Drag and drop
      const host = this;
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
          host.setFile(dt.files[0], fileChip, fileName, fileSize);
        }
      });
    }

    // Textarea auto-resize
    input.addEventListener("input", () => {
      input.style.height = "auto";
      input.style.height = `${Math.min(input.scrollHeight, 140)}px`;
    });

    input.addEventListener("keydown", (e) => {
      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault();
        send();
      }
    });

    sendBtn.addEventListener("click", send);

    requestAnimationFrame(() => input.focus());

    const self = this;

    function addMessage(
      role: string,
      content: string,
      file?: string,
    ): HairaMessage {
      const msg = document.createElement("haira-message") as HairaMessage;
      msg.setAttribute("role", role);
      msg.setAttribute("content", content);
      if (file) msg.setAttribute("file", file);
      messagesInner.insertBefore(msg, typing);
      messagesOuter.scrollTop = messagesOuter.scrollHeight;
      return msg;
    }

    async function send() {
      const text = input.value.trim();
      if (!text && !self.attachedFile) return;
      input.value = "";
      input.style.height = "auto";

      // Hide welcome, show messages
      welcome.classList.add("hidden");
      messagesOuter.classList.add("active");

      // Show user message with file indicator
      const fileLabel = self.attachedFile ? self.attachedFile.name : undefined;
      addMessage("user", text, fileLabel);

      sendBtn.disabled = true;
      typing.classList.add("visible");
      messagesOuter.scrollTop = messagesOuter.scrollHeight;

      // Build request body
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

      // Track active tool cards by name
      const activeTools = new Map<string, HairaToolCard>();
      let assistantMsg: HairaMessage | null = null;
      let fullText = "";

      await streamSSE(
        m.path,
        body,
        {
          onToolStart: (event) => {
            typing.classList.remove("visible");
            const card = document.createElement(
              "haira-tool-card",
            ) as HairaToolCard;
            messagesInner.insertBefore(card, typing);
            card.setTool(event.tool);
            activeTools.set(event.tool, card);
            messagesOuter.scrollTop = messagesOuter.scrollHeight;
          },
          onToolRender: (event: ToolRenderEvent) => {
            const renderer = document.createElement(
              "haira-ui-renderer",
            ) as HairaUIRenderer;
            messagesInner.insertBefore(renderer, typing);
            requestAnimationFrame(() => renderer.render(event));
            messagesOuter.scrollTop = messagesOuter.scrollHeight;
          },
          onToolEnd: (event) => {
            const card = activeTools.get(event.tool);
            if (card) {
              card.complete(event.ok !== false);
              activeTools.delete(event.tool);
            }
            typing.classList.add("visible");
            messagesOuter.scrollTop = messagesOuter.scrollHeight;
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
            // If no response was received at all, show a fallback message
            if (!assistantMsg && fullText === "") {
              assistantMsg = addMessage("assistant", "");
              assistantMsg.updateContent(
                "No response received. Please check the server logs.",
              );
            }
            sendBtn.disabled = false;
            input.focus();
          },
        },
        formData,
      );
    }
  }

  private setFile(
    file: File,
    chipEl: HTMLElement,
    nameEl: HTMLElement,
    sizeEl: HTMLElement,
  ) {
    this.attachedFile = file;
    nameEl.textContent = file.name;
    sizeEl.textContent = this.formatSize(file.size);
    chipEl.classList.add("visible");
  }

  private clearFile(chipEl: HTMLElement, fileInput: HTMLInputElement) {
    this.attachedFile = null;
    chipEl.classList.remove("visible");
    fileInput.value = "";
  }

  private formatSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  }

  private getSuggestions(): string[] {
    const title = (this.meta.title || "").toLowerCase();
    if (
      title.includes("maltimize") ||
      title.includes("config") ||
      title.includes("migration")
    ) {
      return [
        "What can you help me with?",
        "Dry-run my config file",
        "Deploy a configuration",
      ];
    }
    return ["What can you help me with?", "Hello!"];
  }

  private esc(s: string): string {
    return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }
}
