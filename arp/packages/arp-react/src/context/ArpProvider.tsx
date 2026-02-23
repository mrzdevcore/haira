import { createContext, useContext, useMemo, type ReactNode } from "react";
import type { ArpCapabilities } from "@haira/arp";

// ---------------------------------------------------------------------------
// Context
// ---------------------------------------------------------------------------

export interface ArpContextValue {
  /** Whether the ARP WebSocket is connected. */
  connected: boolean;
  /** Server capabilities (null until connected). */
  capabilities: ArpCapabilities | null;
  /** Custom component overrides merged with built-ins. */
  components: Record<string, React.ComponentType<any>>;
}

const ArpContext = createContext<ArpContextValue>({
  connected: false,
  capabilities: null,
  components: {},
});

// ---------------------------------------------------------------------------
// Provider
// ---------------------------------------------------------------------------

export interface ArpProviderProps {
  connected?: boolean;
  capabilities?: ArpCapabilities | null;
  /** Additional or override component mappings. */
  components?: Record<string, React.ComponentType<any>>;
  children: ReactNode;
}

/**
 * Provides ARP state to child components via React context.
 * Used internally by ArpChat, but can be used standalone for custom layouts.
 */
export function ArpProvider({
  connected = false,
  capabilities = null,
  components = {},
  children,
}: ArpProviderProps) {
  const value = useMemo(
    () => ({ connected, capabilities, components }),
    [connected, capabilities, components],
  );

  return <ArpContext.Provider value={value}>{children}</ArpContext.Provider>;
}

/** Access ARP context from child components. */
export function useArpContext(): ArpContextValue {
  return useContext(ArpContext);
}
