import { useState, useEffect, useRef, useCallback } from "react";
import {
  ArpClient,
  type ArpClientOptions,
  type ArpClientCallbacks,
  type ArpCapabilities,
} from "@haira/arp";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface UseArpOptions {
  /** WebSocket URL (e.g., "ws://localhost:8080/_arp/v1"). */
  url: string;
  /** Session ID for the connection. */
  sessionId: string;
  /** Auto-connect on mount (default: true). */
  autoConnect?: boolean;
  /** Max reconnect attempts (default: 5). */
  maxReconnectAttempts?: number;
  /** Custom WebSocket constructor (for testing/SSR). */
  WebSocket?: ArpClientOptions["WebSocket"];
}

export interface ArpState {
  connected: boolean;
  capabilities: ArpCapabilities | null;
}

export interface UseArpReturn {
  state: ArpState;
  client: ArpClient | null;
  connect: () => void;
  disconnect: () => void;
}

// ---------------------------------------------------------------------------
// Hook
// ---------------------------------------------------------------------------

/**
 * Core ARP connection hook. Manages the WebSocket lifecycle and exposes
 * the raw ArpClient for custom integrations.
 *
 * For chat-specific use cases, prefer `useArpChat` which builds on this.
 */
export function useArp(options: UseArpOptions): UseArpReturn {
  const { url, sessionId, autoConnect = true, maxReconnectAttempts, WebSocket } = options;

  const [state, setState] = useState<ArpState>({
    connected: false,
    capabilities: null,
  });

  const clientRef = useRef<ArpClient | null>(null);

  // Build callbacks that update React state
  const callbacksRef = useRef<ArpClientCallbacks>({
    onConnect: (caps) => {
      setState({ connected: true, capabilities: caps });
    },
    onDisconnect: () => {
      setState((prev) => ({ ...prev, connected: false }));
    },
  });

  // Initialize client on mount or when url/sessionId change
  useEffect(() => {
    const client = new ArpClient(
      {
        url,
        sessionId,
        maxReconnectAttempts,
        WebSocket,
      },
      callbacksRef.current,
    );
    clientRef.current = client;

    if (autoConnect) {
      client.connect();
    }

    return () => {
      client.disconnect();
      clientRef.current = null;
    };
  }, [url, sessionId, autoConnect, maxReconnectAttempts, WebSocket]);

  const connect = useCallback(() => {
    clientRef.current?.connect();
  }, []);

  const disconnect = useCallback(() => {
    clientRef.current?.disconnect();
  }, []);

  return {
    state,
    client: clientRef.current,
    connect,
    disconnect,
  };
}
