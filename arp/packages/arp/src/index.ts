// @haira/arp — Agent Rendering Protocol client library
// Pure protocol types and transport clients. Zero DOM dependencies.

// Protocol types
export type {
  ArpMessage,
  ArpComponent,
  ArpPatchOp,
  ArpHello,
  ArpCapabilities,
  ArpInputMessage,
  ArpMessageTypeValue,
  ToolRenderEvent,
  // Component props
  StatusCardProps,
  TableProps,
  TabData,
  CodeBlockProps,
  CodeTabData,
  DiffProps,
  KeyValueProps,
  ProgressProps,
  ConfirmProps,
  ChoicesProps,
  FormViewProps,
  ChartDataset,
  ChartProps,
  ProductCardItem,
  ProductCardsProps,
  // Chat session
  ChatSession,
  ChatSessionDetail,
  ChatMessage,
  // Step events
  StepLogEntry,
  StepConfirmPayload,
  StepEvent,
} from "./types.js";

export { ARP_VERSION, ArpMessageType } from "./types.js";

// WebSocket client
export { ArpClient, arpMessageToToolRenderEvent } from "./arp-client.js";
export type { ArpClientOptions, ArpClientCallbacks } from "./arp-client.js";

// SSE client
export { streamSSE, connectSSE } from "./sse-client.js";
export type {
  SSECallbacks,
  StreamSSEOptions,
  ToolEvent,
} from "./sse-client.js";

// Session API
export { createSessionAPI } from "./session.js";
export type { SessionAPI, WorkflowInfo } from "./session.js";

// Component registry
export {
  createComponentRegistry,
  BUILT_IN_COMPONENTS,
} from "./component-registry.js";
export type {
  ComponentRegistry,
  BuiltInComponentType,
} from "./component-registry.js";
