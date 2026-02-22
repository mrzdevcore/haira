/**
 * ARP WebSocket Client — Agentic Rendering Protocol (Minimal Mode)
 *
 * Manages a WebSocket connection to the Haira ARP endpoint (/_arp/v1).
 * Handles capability negotiation, message dispatch, and auto-reconnect.
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
  private maxReconnectAttempts = 5;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;

  constructor(
    sessionId: string,
    callbacks: ArpClientCallbacks,
    url?: string,
  ) {
    this.sessionId = sessionId;
    this.callbacks = callbacks;

    // Derive WebSocket URL from current page location
    if (url) {
      this.url = url;
    } else {
      const proto = location.protocol === "https:" ? "wss:" : "ws:";
      this.url = `${proto}//${location.host}/_arp/v1`;
    }
  }

  /** Connect to the ARP WebSocket endpoint. */
  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) return;

    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
    };

    this.ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
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
    this.maxReconnectAttempts = 0; // Prevent reconnect
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
        if (arpMsg.components?.length) {
          const comp = arpMsg.components[0];
          const event: ToolRenderEvent = {
            tool: arpMsg.tool_name ?? "",
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

    // Exponential backoff: 1s, 2s, 4s, 8s, 16s
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts - 1), 16000);
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, delay);
  }
}

// ---------------------------------------------------------------------------
// Adapter: ArpMessage → ToolRenderEvent
// ---------------------------------------------------------------------------

/**
 * Converts an ARP render message to the existing ToolRenderEvent interface
 * used by the UI renderer. This allows the existing rendering pipeline
 * (haira-ui-renderer.ts) to work unchanged with ARP messages.
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
