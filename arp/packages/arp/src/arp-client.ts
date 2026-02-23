/**
 * ARP WebSocket Client — Agent Rendering Protocol (Minimal Mode)
 *
 * Pure protocol client with zero DOM dependencies. Manages a WebSocket
 * connection to an ARP endpoint, handles capability negotiation,
 * message dispatch, and auto-reconnect.
 */

import type {
  ArpMessage,
  ArpHello,
  ArpCapabilities,
  ToolRenderEvent,
} from "./types.js";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface ArpClientOptions {
  /** WebSocket URL (e.g., "ws://localhost:8080/_arp/v1"). Required. */
  url: string;
  /** Session ID for the connection. */
  sessionId: string;
  /** Max reconnect attempts (default: 5). */
  maxReconnectAttempts?: number;
  /** Custom WebSocket constructor (for Node.js or testing). */
  WebSocket?: new (url: string) => WebSocket;
}

export interface ArpClientCallbacks {
  onConnect?: (caps: ArpCapabilities) => void;
  onDisconnect?: () => void;
  onDelta?: (text: string) => void;
  onToolStart?: (tool: string, args: string) => void;
  onToolEnd?: (tool: string, ok: boolean) => void;
  onRender?: (event: ToolRenderEvent) => void;
  onError?: (error: string) => void;
  onDone?: () => void;
}

// ---------------------------------------------------------------------------
// ArpClient
// ---------------------------------------------------------------------------

export class ArpClient {
  private ws: WebSocket | null = null;
  private caps: ArpCapabilities | null = null;
  private callbacks: ArpClientCallbacks;
  private sessionId: string;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts: number;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private WSConstructor: new (url: string) => WebSocket;

  constructor(options: ArpClientOptions, callbacks: ArpClientCallbacks) {
    this.url = options.url;
    this.sessionId = options.sessionId;
    this.maxReconnectAttempts = options.maxReconnectAttempts ?? 5;
    this.callbacks = callbacks;
    this.WSConstructor = options.WebSocket ?? globalThis.WebSocket;
  }

  /** Connect to the ARP WebSocket endpoint. */
  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) return;

    this.ws = new this.WSConstructor(this.url);

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
    };

    this.ws.onmessage = (event: MessageEvent) => {
      try {
        const msg = JSON.parse(
          typeof event.data === "string" ? event.data : "",
        );
        this.handleMessage(msg);
      } catch {
        // Ignore malformed messages
      }
    };

    this.ws.onclose = () => {
      this.ws = null;
      this.callbacks.onDisconnect?.();
      this.scheduleReconnect();
    };

    this.ws.onerror = () => {
      // onclose will fire after onerror
    };
  }

  /** Disconnect and stop reconnecting. */
  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.maxReconnectAttempts = 0;
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  /** Send a text message to the agent. */
  sendText(text: string): void {
    this.send({
      v: 1,
      type: "input",
      session_id: this.sessionId,
      input_type: "text",
      data: { text },
    });
  }

  /** Send an action event (button click, confirm/deny, etc). */
  sendAction(
    componentId: string,
    action: string,
    payload?: unknown,
  ): void {
    this.send({
      v: 1,
      type: "input",
      session_id: this.sessionId,
      source_component: componentId,
      input_type: "action",
      data: { action, payload },
    });
  }

  /** Send a form submission. */
  sendFormSubmit(fields: Record<string, unknown>, action?: string): void {
    this.send({
      v: 1,
      type: "input",
      session_id: this.sessionId,
      input_type: "form_submit",
      data: { action, fields },
    });
  }

  /** Whether the WebSocket is currently connected. */
  get connected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  /** The negotiated capabilities from the server. */
  get capabilities(): ArpCapabilities | null {
    return this.caps;
  }

  /** Update the session ID without reconnecting. */
  setSessionId(id: string): void {
    this.sessionId = id;
  }

  // ---------------------------------------------------------------------------
  // Internal
  // ---------------------------------------------------------------------------

  private send(msg: Record<string, unknown>): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg));
    }
  }

  private handleMessage(msg: ArpMessage | ArpHello): void {
    switch (msg.type) {
      case "hello":
        this.caps = (msg as ArpHello).capabilities;
        this.callbacks.onConnect?.(this.caps);
        break;

      case "delta": {
        const delta = (msg.payload as { delta?: string })?.delta;
        if (delta) this.callbacks.onDelta?.(delta);
        break;
      }

      case "tool_start": {
        const p = msg.payload as { tool?: string; args?: string };
        if (p?.tool) this.callbacks.onToolStart?.(p.tool, p.args ?? "");
        break;
      }

      case "tool_end": {
        const p = msg.payload as { tool?: string; ok?: boolean };
        if (p?.tool) this.callbacks.onToolEnd?.(p.tool, p.ok ?? false);
        break;
      }

      case "render": {
        const arpMsg = msg as ArpMessage;
        const components =
          arpMsg.components ?? (arpMsg.payload as any)?.components;
        if (components?.length) {
          const comp = components[0];
          const event: ToolRenderEvent = {
            tool: arpMsg.tool_name ?? (arpMsg.payload as any)?.tool ?? "",
            component: comp.type,
            props: comp.props,
          };
          this.callbacks.onRender?.(event);
        }
        break;
      }

      case "error": {
        const errMsg = (msg.payload as { error?: string })?.error;
        if (errMsg) this.callbacks.onError?.(errMsg);
        break;
      }

      case "commit": {
        const arpMsg = msg as ArpMessage;
        if (arpMsg.final) this.callbacks.onDone?.();
        break;
      }
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) return;
    this.reconnectAttempts++;

    const delay = Math.min(
      1000 * Math.pow(2, this.reconnectAttempts - 1),
      16000,
    );
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, delay);
  }
}

// ---------------------------------------------------------------------------
// Utilities
// ---------------------------------------------------------------------------

/**
 * Converts an ARP render message to a ToolRenderEvent.
 * Returns null if the message is not a render event.
 */
export function arpMessageToToolRenderEvent(
  msg: ArpMessage,
): ToolRenderEvent | null {
  if (msg.type === "render" && msg.components?.length) {
    const comp = msg.components[0];
    return {
      tool: msg.tool_name ?? "",
      component: comp.type,
      props: comp.props,
    };
  }
  return null;
}
