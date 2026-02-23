// @haira/arp-vue — Vue composables and components for the Agent Rendering Protocol

// Composables
export { useArp, type UseArpOptions, type UseArpReturn } from "./composables/use-arp";
export {
  useArpChat,
  type UseArpChatOptions,
  type UseArpChatReturn,
  type ChatMessage,
  type ToolCardState,
} from "./composables/use-arp-chat";

// Components
export { default as ArpChat } from "./components/ArpChat.vue";
export { default as ArpMessage } from "./components/ArpMessage.vue";
export { default as ArpRenderer } from "./components/ArpRenderer.vue";
export { default as ArpActivityPanel } from "./components/ArpActivityPanel.vue";

// Provide/inject keys
export { ArpComponentsKey, ArpConnectedKey, ArpCapabilitiesKey } from "./provide-keys";
