import { ref, onMounted, onUnmounted, watch, type Ref } from "vue";
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
  url: string | Ref<string>;
  /** Session ID for the connection. */
  sessionId: string | Ref<string>;
  /** Auto-connect on mount (default: true). */
  autoConnect?: boolean;
  /** Max reconnect attempts (default: 5). */
  maxReconnectAttempts?: number;
  /** Custom WebSocket constructor. */
  WebSocket?: ArpClientOptions["WebSocket"];
}

export interface UseArpReturn {
  connected: Ref<boolean>;
  capabilities: Ref<ArpCapabilities | null>;
  client: Ref<ArpClient | null>;
  connect: () => void;
  disconnect: () => void;
}

// ---------------------------------------------------------------------------
// Composable
// ---------------------------------------------------------------------------

/**
 * Core ARP connection composable. Manages the WebSocket lifecycle.
 * For chat-specific use cases, prefer `useArpChat`.
 */
export function useArp(options: UseArpOptions): UseArpReturn {
  const connected = ref(false);
  const capabilities = ref<ArpCapabilities | null>(null);
  const client = ref<ArpClient | null>(null);

  const {
    autoConnect = true,
    maxReconnectAttempts,
    WebSocket: WSConstructor,
  } = options;

  function createClient() {
    const urlVal = typeof options.url === "string" ? options.url : options.url.value;
    const sessionVal = typeof options.sessionId === "string" ? options.sessionId : options.sessionId.value;

    // Disconnect previous
    client.value?.disconnect();

    const callbacks: ArpClientCallbacks = {
      onConnect: (caps) => {
        connected.value = true;
        capabilities.value = caps;
      },
      onDisconnect: () => {
        connected.value = false;
      },
    };

    const c = new ArpClient(
      { url: urlVal, sessionId: sessionVal, maxReconnectAttempts, WebSocket: WSConstructor },
      callbacks,
    );
    client.value = c;

    if (autoConnect) {
      c.connect();
    }
  }

  onMounted(() => {
    createClient();
  });

  onUnmounted(() => {
    client.value?.disconnect();
    client.value = null;
  });

  // Reconnect when url or sessionId changes
  watch(
    () => [
      typeof options.url === "string" ? options.url : options.url.value,
      typeof options.sessionId === "string" ? options.sessionId : options.sessionId.value,
    ],
    () => createClient(),
  );

  function connect() {
    client.value?.connect();
  }

  function disconnect() {
    client.value?.disconnect();
  }

  return {
    connected,
    capabilities,
    client: client as Ref<ArpClient | null>,
    connect,
    disconnect,
  };
}
