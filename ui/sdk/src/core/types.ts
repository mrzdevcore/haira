// Strict prop interfaces for all UI components.

// --- Workflow metadata (from Go server) ---

export interface WorkflowParam {
  Name: string;
  Type: string; // "string" | "int" | "float" | "bool" | "file"
}

export interface WorkflowMeta {
  mode: "index" | "form" | "chat";
  name: string;
  method: string;
  path: string;
  params: WorkflowParam[];
  steps: string[];
  title: string;
  description: string;
  hasFile: boolean;
  chatParam?: string;
  fileParam?: string;
  settingsParams?: WorkflowParam[];
  suggestions?: string[];
  accent?: string;
  logo?: string;
  theme?: string;
  avatar?: string;
  workflows?: WorkflowListItem[];
}

export interface WorkflowListItem {
  name: string;
  path: string;
  method: string;
  uiType: string;
  title: string;
  description?: string;
  hasFile?: boolean;
  params?: WorkflowParam[];
  chatParam?: string;
  fileParam?: string;
  suggestions?: string[];
  accent?: string;
  logo?: string;
  theme?: string;
  avatar?: string;
}

// --- Step / Pipeline events ---

export interface StepLogEntry {
  level: "info" | "warn" | "error";
  message: string;
}

export interface StepEvent {
  name: string;
  status: "start" | "end" | "failed" | "retry" | "log";
  duration_ms?: number;
  error?: string;
  attempt?: number;
  delay_ms?: number;
  log?: StepLogEntry;
}

export type StepStatus =
  | "pending"
  | "running"
  | "done"
  | "failed"
  | "retrying"
  | "skipped";

// --- Run history ---

export interface RunSummary {
  id: string;
  workflow_name: string;
  workflow_path: string;
  status: "running" | "completed" | "failed";
  started_at: string;
  finished_at?: string;
  step_count: number;
}

export interface RunDetail {
  id: string;
  workflow_name: string;
  workflow_path: string;
  status: "running" | "completed" | "failed";
  params?: Record<string, unknown>;
  steps: StepEvent[];
  result?: unknown;
  error?: string;
  started_at: string;
  finished_at?: string;
}

// --- Chat sessions ---

export interface ChatSessionSummary {
  id: string;
  workflow_name: string;
  workflow_path: string;
  title: string;
  owner?: string;
  created_at: string;
  updated_at: string;
  message_count: number;
}

export interface ChatSessionDetail extends ChatSessionSummary {
  messages: ChatMessageEntry[];
}

export interface ChatMessageEntry {
  role: "user" | "assistant";
  content: string;
  timestamp: string;
  ui_events?: ToolRenderEvent[];
}

// --- Generative UI ---

export interface ToolRenderEvent {
  tool: string;
  component: string;
  props: Record<string, unknown>;
}

// --- Component prop interfaces ---

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

// --- ARP Protocol Types (Minimal Mode) ---

export interface ArpMessage {
  v: number;
  type: string;
  session_id?: string;
  payload?: Record<string, unknown>;
  components?: ArpComponent[];
  ops?: ArpPatchOp[];
  final?: boolean;
  tool_name?: string;
}

export interface ArpComponent {
  id: string;
  type: string;
  version?: number;
  props: Record<string, unknown>;
  fallback?: ArpComponent;
}

export interface ArpPatchOp {
  op: "update" | "insert" | "remove" | "replace" | "reorder";
  target?: string;
  path?: string;
  value?: unknown;
  after?: string;
  component?: ArpComponent;
}

export interface ArpHello {
  v: number;
  type: "hello";
  capabilities: ArpCapabilities;
}

export interface ArpCapabilities {
  components: string[];
  features: string[];
}

// --- Observe types ---

export interface ObserveUsage {
  total_tokens: number;
  input_tokens: number;
  output_tokens: number;
  llm_calls: number;
  tool_calls: number;
  total_latency_ms: number;
  estimated_cost_usd: number;
}

export interface ObserveEvent {
  type: "generation" | "tool";
  timestamp: string;
  agent: string;
  session_id?: string;
  // generation fields
  model?: string;
  input_tokens?: number;
  output_tokens?: number;
  latency_ms?: number;
  cost_usd?: number;
  tool_call_count?: number;
  // tool fields
  tool_name?: string;
  success?: boolean;
  duration_ms?: number;
}
