// ARP (Agent Rendering Protocol) — Type Definitions
// Protocol version 1, Minimal Mode

// ---------------------------------------------------------------------------
// Protocol constants
// ---------------------------------------------------------------------------

export const ARP_VERSION = 1;

export const ArpMessageType = {
  Hello: "hello",
  Delta: "delta",
  ToolStart: "tool_start",
  ToolEnd: "tool_end",
  Render: "render",
  Patch: "patch",
  Error: "error",
  Commit: "commit",
  Input: "input",
} as const;

export type ArpMessageTypeValue =
  (typeof ArpMessageType)[keyof typeof ArpMessageType];

// ---------------------------------------------------------------------------
// Protocol messages (server → client)
// ---------------------------------------------------------------------------

/** Transport-agnostic message envelope for ARP Minimal Mode. */
export interface ArpMessage {
  v: number;
  type: string;
  session_id?: string;
  payload?: Record<string, unknown>;
  /** Render fields (type="render") */
  components?: ArpComponent[];
  /** Patch fields (type="patch") */
  ops?: ArpPatchOp[];
  /** Commit fields (type="commit") */
  final?: boolean;
  /** Tool identifier for render/tool_start/tool_end messages */
  tool_name?: string;
}

/** A typed UI component with props and optional fallback. */
export interface ArpComponent {
  id: string;
  type: string;
  version?: number;
  props: Record<string, unknown>;
  fallback?: ArpComponent;
}

/** An incremental update operation on the component tree. */
export interface ArpPatchOp {
  op: "update" | "insert" | "remove" | "replace" | "reorder";
  target?: string;
  path?: string;
  value?: unknown;
  after?: string;
  component?: ArpComponent;
}

/** Server capabilities handshake (sent on connect). */
export interface ArpHello {
  v: number;
  type: "hello";
  capabilities: ArpCapabilities;
}

/** What the server supports. */
export interface ArpCapabilities {
  components: string[];
  features: string[];
}

// ---------------------------------------------------------------------------
// Protocol messages (client → server)
// ---------------------------------------------------------------------------

/** Input message from client (renderer → agent). */
export interface ArpInputMessage {
  v: number;
  type: "input";
  session_id: string;
  source_component?: string;
  input_type: string;
  data: Record<string, unknown>;
}

// ---------------------------------------------------------------------------
// Generative UI
// ---------------------------------------------------------------------------

/** A UI component render event produced by an agent tool call. */
export interface ToolRenderEvent {
  tool: string;
  component: string;
  props: Record<string, unknown>;
}

// ---------------------------------------------------------------------------
// Component prop interfaces
// ---------------------------------------------------------------------------

export interface StatusCardProps {
  status: "success" | "error" | "warning" | "info";
  title: string;
  message?: string;
  sections?: Array<{ label: string; content: string; style?: string }>;
  _restored?: boolean;
}

export interface TableProps {
  title?: string;
  headers: string[];
  rows: string[][];
  tabs?: TabData[];
  highlight?: number[];
}

export interface TabData {
  name: string;
  headers: string[];
  rows: string[][];
  highlight?: number[];
}

export interface CodeBlockProps {
  title?: string;
  language?: string;
  code?: string;
  tabs?: CodeTabData[];
}

export interface CodeTabData {
  name: string;
  language?: string;
  code: string;
}

export interface DiffProps {
  title?: string;
  before_label?: string;
  after_label?: string;
  before: string;
  after: string;
  language?: string;
}

export interface KeyValueProps {
  title?: string;
  items: Array<{ key: string; value: string; style?: string }>;
}

export interface ProgressProps {
  title?: string;
  steps: Array<{ name: string; status: string; detail?: string }>;
}

export interface ConfirmProps {
  title: string;
  message?: string;
  confirm_label?: string;
  deny_label?: string;
  _restored?: boolean;
}

export interface ChoicesProps {
  title: string;
  options: string[];
  style?: "buttons" | "list";
  _restored?: boolean;
}

export interface FormViewProps {
  title?: string;
  fields: Array<{
    name: string;
    label?: string;
    field_type?: string;
    value?: string;
    required?: boolean;
    options?: string[];
  }>;
  submit_label?: string;
  submit_action?: string;
}

export interface ChartDataset {
  label: string;
  data: number[];
  color?: string;
}

export interface ChartProps {
  type: "line" | "bar" | "pie" | "scatter" | "area";
  title?: string;
  labels: string[];
  datasets: ChartDataset[];
  height?: number;
}

export interface ProductCardItem {
  name: string;
  price: string;
  image?: string;
  brand?: string;
  description?: string;
  badge?: string;
  url?: string;
}

export interface ProductCardsProps {
  title?: string;
  cards: ProductCardItem[];
}

// ---------------------------------------------------------------------------
// Chat session types
// ---------------------------------------------------------------------------

export interface ChatSession {
  id: string;
  workflow_name: string;
  workflow_path: string;
  title: string;
  owner?: string;
  created_at: string;
  updated_at: string;
  message_count: number;
}

export interface ChatSessionDetail extends ChatSession {
  messages: ChatMessage[];
}

export interface ChatMessage {
  role: "user" | "assistant";
  content: string;
  timestamp: string;
  ui_events?: ToolRenderEvent[];
}

// ---------------------------------------------------------------------------
// Step / Pipeline events (for non-streaming workflow tracking)
// ---------------------------------------------------------------------------

export interface StepLogEntry {
  level: "info" | "warn" | "error";
  message: string;
}

export interface StepRenderPayload {
  component: string;
  props: Record<string, unknown>;
}

export interface StepEvent {
  name: string;
  status: "start" | "end" | "failed" | "retry" | "log" | "render";
  duration_ms?: number;
  error?: string;
  attempt?: number;
  delay_ms?: number;
  log?: StepLogEntry;
  render?: StepRenderPayload;
}
