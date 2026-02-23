import { ref, computed, onMounted, onUnmounted, watch, type Ref } from "vue";
import {
  ArpClient,
  type ArpClientCallbacks,
  type ArpCapabilities,
  type ToolRenderEvent,
  type ChatSession,
  streamSSE,
  createSessionAPI,
} from "@haira/arp";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface ChatMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  file?: string;
  uiEvents?: ToolRenderEvent[];
  restored?: boolean;
}

export interface ToolCardState {
  name: string;
  displayName: string;
  status: "running" | "done" | "failed";
  startTime: number;
  elapsed?: string;
}

export interface UseArpChatOptions {
  /** WebSocket URL. */
  url: string | Ref<string>;
  /** API base URL for session persistence. */
  apiBaseUrl?: string;
  /** Initial session ID. Auto-generated if not provided. */
  sessionId?: string;
  /** Param name for chat messages (default: "message"). */
  chatParam?: string;
  /** SSE endpoint path for fallback. */
  sseFallbackPath?: string;
  /** Whether the workflow accepts file uploads. */
  hasFileUpload?: boolean;
  /** Param name for file uploads. */
  fileParam?: string;
}

export interface UseArpChatReturn {
  messages: Ref<ChatMessage[]>;
  isStreaming: Ref<boolean>;
  isConnected: Ref<boolean>;
  toolCards: Ref<ToolCardState[]>;
  runningToolCount: Ref<number>;
  sessionId: Ref<string>;
  capabilities: Ref<ArpCapabilities | null>;
  sessions: Ref<ChatSession[]>;
  sendMessage: (text: string, file?: File) => void;
  switchSession: (id: string) => Promise<void>;
  startNewChat: () => void;
  loadSessions: () => Promise<void>;
  deleteSession: (id: string) => Promise<void>;
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

let msgCounter = 0;
function nextId(): string {
  return `msg_${++msgCounter}_${Date.now()}`;
}

function formatToolName(name: string): string {
  return name
    .replace(/^render_/, "")
    .replace(/_/g, " ")
    .replace(/\b\w/g, (c) => c.toUpperCase());
}

// ---------------------------------------------------------------------------
// Composable
// ---------------------------------------------------------------------------

/**
 * Chat-specific ARP composable. Manages messages, streaming, tool activity,
 * and session persistence.
 *
 * ```vue
 * <script setup>
 * const { messages, sendMessage, isStreaming } = useArpChat({
 *   url: "ws://localhost:8080/_arp/v1",
 * });
 * </script>
 * ```
 */
export function useArpChat(options: UseArpChatOptions): UseArpChatReturn {
  const {
    apiBaseUrl,
    sessionId: initialSessionId,
    chatParam = "message",
    sseFallbackPath,
    hasFileUpload = false,
    fileParam = "file",
  } = options;

  // --- State ---
  const messages = ref<ChatMessage[]>([]);
  const isStreaming = ref(false);
  const isConnected = ref(false);
  const toolCards = ref<ToolCardState[]>([]);
  const sessionId = ref(initialSessionId || crypto.randomUUID());
  const sessions = ref<ChatSession[]>([]);
  const capabilities = ref<ArpCapabilities | null>(null);

  const runningToolCount = computed(
    () => toolCards.value.filter((t) => t.status === "running").length,
  );

  // --- Refs for streaming ---
  let client: ArpClient | null = null;
  let abortCtrl: AbortController | null = null;
  let fullText = "";
  let streamingId: string | null = null;

  // --- Session API ---
  const sessionAPI = apiBaseUrl ? createSessionAPI(apiBaseUrl) : null;

  const urlVal = () => typeof options.url === "string" ? options.url : options.url.value;
  const effectiveApiBase = () =>
    apiBaseUrl ?? urlVal().replace(/^ws/, "http").replace(/\/_arp\/v1$/, "");

  // --- Streaming callbacks ---
  function handleDelta(delta: string) {
    fullText += delta;
    const text = fullText;
    const id = streamingId;
    if (!id) return;
    const idx = messages.value.findIndex((m) => m.id === id);
    if (idx !== -1) {
      messages.value[idx] = { ...messages.value[idx], content: text };
    }
  }

  function handleToolStart(tool: string) {
    toolCards.value = [
      ...toolCards.value,
      {
        name: tool,
        displayName: formatToolName(tool),
        status: "running",
        startTime: Date.now(),
      },
    ];
  }

  function handleToolEnd(tool: string, ok: boolean) {
    toolCards.value = toolCards.value.map((t) =>
      t.name === tool && t.status === "running"
        ? {
            ...t,
            status: ok ? "done" : "failed",
            elapsed: `${((Date.now() - t.startTime) / 1000).toFixed(1)}s`,
          }
        : t,
    );
  }

  function handleRender(event: ToolRenderEvent) {
    const id = streamingId;
    if (!id) return;
    const idx = messages.value.findIndex((m) => m.id === id);
    if (idx !== -1) {
      const msg = messages.value[idx];
      messages.value[idx] = {
        ...msg,
        uiEvents: [...(msg.uiEvents ?? []), event],
      };
    }
  }

  function handleError(error: string) {
    messages.value = [
      ...messages.value,
      { id: nextId(), role: "assistant", content: `Error: ${error}` },
    ];
  }

  function handleDone() {
    isStreaming.value = false;
    toolCards.value = [];
    fullText = "";
    streamingId = null;
  }

  // --- Connect ARP WebSocket ---
  function connectARP() {
    client?.disconnect();

    const callbacks: ArpClientCallbacks = {
      onConnect: (caps) => {
        isConnected.value = true;
        capabilities.value = caps;
      },
      onDisconnect: () => {
        isConnected.value = false;
      },
      onDelta: handleDelta,
      onToolStart: (tool, _args) => handleToolStart(tool),
      onToolEnd: (tool, ok) => handleToolEnd(tool, ok),
      onRender: handleRender,
      onError: handleError,
      onDone: handleDone,
    };

    client = new ArpClient(
      { url: urlVal(), sessionId: sessionId.value },
      callbacks,
    );
    client.connect();
  }

  onMounted(() => {
    connectARP();

    // Load session history
    if (sessionAPI) {
      sessionAPI.getSession(sessionId.value).then((detail) => {
        if (!detail?.messages?.length) return;
        messages.value = detail.messages.map((m) => ({
          id: nextId(),
          role: m.role,
          content: m.content,
          uiEvents: m.ui_events,
          restored: true,
        }));
      });
    }
  });

  onUnmounted(() => {
    client?.disconnect();
    client = null;
    abortCtrl?.abort();
  });

  // Reconnect when sessionId changes
  watch(sessionId, () => {
    connectARP();
    // Load new session history
    if (sessionAPI) {
      sessionAPI.getSession(sessionId.value).then((detail) => {
        if (!detail?.messages?.length) return;
        messages.value = detail.messages.map((m) => ({
          id: nextId(),
          role: m.role,
          content: m.content,
          uiEvents: m.ui_events,
          restored: true,
        }));
      });
    }
  });

  // --- Send message ---
  function sendMessage(text: string, file?: File) {
    if ((!text.trim() && !file) || isStreaming.value) return;

    // Add user message
    const userMsg: ChatMessage = {
      id: nextId(),
      role: "user",
      content: text,
      file: file?.name,
    };

    // Add empty assistant message for streaming
    const assistantId = nextId();
    streamingId = assistantId;
    fullText = "";

    messages.value = [
      ...messages.value,
      userMsg,
      { id: assistantId, role: "assistant", content: "" },
    ];
    isStreaming.value = true;
    toolCards.value = [];

    // Try ARP WebSocket
    if (client?.connected && !file) {
      client.sendText(text);
      return;
    }

    // SSE fallback
    if (sseFallbackPath) {
      const sseUrl = `${effectiveApiBase()}${sseFallbackPath}`;
      abortCtrl = new AbortController();

      let body: unknown;
      let formData: FormData | undefined;

      if (file && hasFileUpload) {
        formData = new FormData();
        formData.append(chatParam, text);
        formData.append(fileParam, file);
        formData.append("session_id", sessionId.value);
      } else {
        body = { [chatParam]: text, session_id: sessionId.value };
      }

      streamSSE(sseUrl, body, {
        onDelta: handleDelta,
        onToolStart: (e) => handleToolStart(e.tool),
        onToolEnd: (e) => handleToolEnd(e.tool, e.ok ?? false),
        onToolRender: handleRender,
        onError: handleError,
        onDone: handleDone,
      }, { formData, signal: abortCtrl.signal });
    }
  }

  // --- Session management ---
  async function loadSessions() {
    if (!sessionAPI) return;
    sessions.value = await sessionAPI.listSessions("");
  }

  async function switchSession(newId: string) {
    abortCtrl?.abort();
    isStreaming.value = false;
    toolCards.value = [];
    fullText = "";
    streamingId = null;
    messages.value = [];
    sessionId.value = newId;
  }

  function startNewChat() {
    switchSession(crypto.randomUUID());
  }

  async function deleteSession(id: string) {
    if (!sessionAPI) return;
    await sessionAPI.deleteSession(id);
    sessions.value = sessions.value.filter((s) => s.id !== id);
    if (id === sessionId.value) {
      startNewChat();
    }
  }

  return {
    messages,
    isStreaming,
    isConnected,
    toolCards,
    runningToolCount,
    sessionId,
    capabilities,
    sessions,
    sendMessage,
    switchSession,
    startNewChat,
    loadSessions,
    deleteSession,
  };
}
