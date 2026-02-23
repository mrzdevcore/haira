// @haira/arp-react — React hooks and components for the Agent Rendering Protocol

// Hooks
export { useArp } from "./hooks/use-arp.js";
export type { UseArpOptions, ArpState, UseArpReturn } from "./hooks/use-arp.js";

export { useArpChat } from "./hooks/use-arp-chat.js";
export type {
  UseArpChatOptions,
  UseArpChatReturn,
  ChatMessage,
  ToolCardState,
} from "./hooks/use-arp-chat.js";

// Components
export { ArpChat } from "./components/ArpChat.js";
export type { ArpChatProps } from "./components/ArpChat.js";

export { ArpMessage } from "./components/ArpMessage.js";
export type { ArpMessageProps } from "./components/ArpMessage.js";

export { ArpRenderer } from "./components/ArpRenderer.js";
export type { ArpRendererProps } from "./components/ArpRenderer.js";

export { ArpActivityPanel } from "./components/ArpActivityPanel.js";
export type { ArpActivityPanelProps } from "./components/ArpActivityPanel.js";

// Context
export { ArpProvider, useArpContext } from "./context/ArpProvider.js";
export type { ArpProviderProps, ArpContextValue } from "./context/ArpProvider.js";

// Re-export core types for convenience
export type {
  ArpCapabilities,
  ToolRenderEvent,
  ChatSession,
  ChatSessionDetail,
} from "@haira/arp";
