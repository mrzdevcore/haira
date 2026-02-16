import { baseStyles, sharedKeyframes, scrollbarStyles } from "../theme";
import { streamSSE } from "../sse";
import type { WorkflowMeta } from "../types";
import type { HairaMessage } from "./haira-message";

export class HairaChat extends HTMLElement {
  private meta!: WorkflowMeta;
  private sessionId = "";

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
    const settingsParams = m.settingsParams || [];
    const hasSettings = settingsParams.length > 0;

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
        .settings-bar {
          display: flex;
          align-items: center;
          justify-content: flex-end;
          padding: 0.4rem 1rem;
          border-bottom: 1px solid var(--haira-border);
        }
        .settings-btn {
          background: none;
          border: 1px solid var(--haira-border);
          color: var(--haira-muted);
          padding: 0.25rem 0.6rem;
          border-radius: 4px;
          cursor: pointer;
          font-size: 0.78rem;
          font-family: var(--haira-font);
          transition: all 0.15s;
        }
        .settings-btn:hover {
          border-color: var(--haira-gold);
          color: var(--haira-gold);
        }
        .settings-panel {
          display: flex;
          flex-wrap: wrap;
          gap: 0.75rem;
          padding: 0 1rem;
          border-bottom: 1px solid var(--haira-border);
          max-height: 0;
          overflow: hidden;
          transition: max-height 0.25s ease, padding 0.25s ease;
        }
        .settings-panel.open {
          max-height: 200px;
          padding: 0.6rem 1rem;
        }
        .settings-panel label {
          font-size: 0.78rem;
          color: var(--haira-muted);
          display: flex;
          flex-direction: column;
          gap: 0.2rem;
          min-width: 140px;
        }
        .settings-panel input {
          background: var(--haira-bg-input);
          border: 1px solid var(--haira-border);
          color: var(--haira-text);
          padding: 0.3rem 0.5rem;
          border-radius: 4px;
          font-size: 0.82rem;
          font-family: var(--haira-font);
          outline: none;
          transition: border-color 0.15s;
        }
        .settings-panel input:focus { border-color: var(--haira-gold); }
        .messages {
          flex: 1;
          overflow-y: auto;
          padding: 1rem;
          display: flex;
          flex-direction: column;
          gap: 0.5rem;
          ${scrollbarStyles}
        }
        .typing {
          display: none;
          padding: 0.4rem 1rem;
          align-items: center;
          gap: 0.35rem;
        }
        .typing.visible { display: flex; }
        .dot {
          display: inline-block;
          width: 6px;
          height: 6px;
          border-radius: 50%;
          background: var(--haira-gold);
          animation: bounce 1.4s ease-in-out infinite;
        }
        .dot:nth-child(2) { animation-delay: 0.2s; }
        .dot:nth-child(3) { animation-delay: 0.4s; }
        .input-area {
          padding: 0.75rem 1rem;
          border-top: 1px solid var(--haira-border);
          display: flex;
          gap: 0.5rem;
        }
        textarea {
          flex: 1;
          background: var(--haira-bg-input);
          border: 1px solid var(--haira-border);
          color: var(--haira-text);
          padding: 0.6rem 0.75rem;
          border-radius: var(--haira-radius);
          font-size: 0.88rem;
          font-family: var(--haira-font);
          resize: none;
          min-height: 42px;
          max-height: 120px;
          outline: none;
          transition: border-color 0.15s;
        }
        textarea:focus { border-color: var(--haira-gold); }
        .send-btn {
          background: linear-gradient(135deg, var(--haira-gold), var(--haira-gold-light));
          color: #1a0e04;
          border: none;
          padding: 0 1.25rem;
          border-radius: var(--haira-radius);
          cursor: pointer;
          font-size: 0.88rem;
          font-weight: 600;
          font-family: var(--haira-font);
          transition: all 0.2s;
        }
        .send-btn:hover { box-shadow: 0 2px 16px rgba(232, 163, 23, 0.3); }
        .send-btn:disabled { opacity: 0.5; cursor: not-allowed; box-shadow: none; }
      </style>
      ${
        hasSettings
          ? `
        <div class="settings-bar">
          <button class="settings-btn" id="settings-btn">Settings</button>
        </div>
        <div class="settings-panel" id="settings-panel">
          ${settingsParams
            .map(
              (p) => `
            <label>${p.Name}
              ${
                p.Type === "bool"
                  ? `<input type="checkbox" id="s-${p.Name}" name="${p.Name}">`
                  : p.Type === "int"
                    ? `<input type="number" id="s-${p.Name}" name="${p.Name}" step="1">`
                    : p.Type === "float"
                      ? `<input type="number" id="s-${p.Name}" name="${p.Name}" step="any">`
                      : `<input type="text" id="s-${p.Name}" name="${p.Name}">`
              }
            </label>
          `,
            )
            .join("")}
        </div>
      `
          : ""
      }
      <div class="messages" id="messages"></div>
      <div class="typing" id="typing">
        <span class="dot"></span>
        <span class="dot"></span>
        <span class="dot"></span>
      </div>
      <div class="input-area">
        <textarea id="chat-input" placeholder="Type a message..." rows="1"></textarea>
        <button class="send-btn" id="send-btn">Send</button>
      </div>
    `;

    const messages = shadow.getElementById("messages")!;
    const input = shadow.getElementById("chat-input") as HTMLTextAreaElement;
    const sendBtn = shadow.getElementById("send-btn") as HTMLButtonElement;
    const typing = shadow.getElementById("typing")!;
    const settingsBtn = shadow.getElementById("settings-btn");
    const settingsPanel = shadow.getElementById("settings-panel");

    if (settingsBtn && settingsPanel) {
      settingsBtn.addEventListener("click", () =>
        settingsPanel.classList.toggle("open"),
      );
    }

    input.addEventListener("keydown", (e) => {
      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault();
        send();
      }
    });

    input.addEventListener("input", () => {
      input.style.height = "auto";
      input.style.height = `${Math.min(input.scrollHeight, 120)}px`;
    });

    sendBtn.addEventListener("click", send);

    // Autofocus the input
    requestAnimationFrame(() => input.focus());

    const self = this;

    function addMessage(role: string, content: string): HairaMessage {
      const msg = document.createElement("haira-message") as HairaMessage;
      msg.setAttribute("role", role);
      msg.setAttribute("content", content);
      messages.appendChild(msg);
      messages.scrollTop = messages.scrollHeight;
      return msg;
    }

    function getSettings(): Record<string, unknown> {
      const s: Record<string, unknown> = {};
      if (!settingsPanel) return s;
      const inputs = settingsPanel.querySelectorAll("input");
      for (const inp of inputs) {
        if (inp.type === "checkbox") s[inp.name] = inp.checked;
        else if (inp.type === "number" && inp.value)
          s[inp.name] = Number(inp.value);
        else if (inp.value) s[inp.name] = inp.value;
      }
      return s;
    }

    async function send() {
      const text = input.value.trim();
      if (!text) return;
      input.value = "";
      input.style.height = "auto";
      addMessage("user", text);

      sendBtn.disabled = true;
      typing.classList.add("visible");
      messages.scrollTop = messages.scrollHeight;

      const body: Record<string, unknown> = getSettings();
      const chatParam = m.chatParam || "message";
      body[chatParam] = text;
      body["session_id"] = self.sessionId;

      const assistantMsg = addMessage("assistant", "");
      let fullText = "";

      await streamSSE(m.path, body, {
        onDelta: (delta) => {
          fullText += delta;
          assistantMsg.updateContent(fullText);
          messages.scrollTop = messages.scrollHeight;
        },
        onError: (error) => {
          assistantMsg.updateContent(`*Error: ${error}*`);
        },
        onDone: () => {
          typing.classList.remove("visible");
          sendBtn.disabled = false;
          input.focus();
        },
      });
    }
  }
}
