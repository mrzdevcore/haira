// Haira Console types.
// ARP protocol types and component props are re-exported from @haira/arp.
// Haira-specific types (workflow, run, observe) are defined here.

// --- Re-exports from @haira/arp ---

export type {
  ArpMessage,
  ArpComponent,
  ArpPatchOp,
  ArpHello,
  ArpCapabilities,
  ToolRenderEvent,
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
  StepLogEntry,
  StepEvent,
} from "@haira/arp";

// Chat session types: @haira/arp uses different names, alias them here
// for backward compatibility with Haira console code.
import type {
  ChatSession as _ChatSession,
  ChatSessionDetail as _ChatSessionDetail,
  ChatMessage as _ChatMessage,
} from "@haira/arp";

export type ChatSessionSummary = _ChatSession;
export type ChatSessionDetail = _ChatSessionDetail;
export type ChatMessageEntry = _ChatMessage;

// --- Haira-specific types (not part of ARP protocol) ---

export interface WorkflowParam {
  Name: string;
  Type: string; // "string" | "int" | "float" | "bool" | "file"
}

export interface WorkflowMeta {
  mode: "index" | "form" | "chat" | "orchestrator";
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
  arpUrl?: string;
  backend?: string;
  workflows?: WorkflowListItem[];
  deployments?: DeploymentItem[];
}

export interface DeploymentItem {
  name: string;
  status: string; // "running" | "stopped" | "crashed" | "deploying"
  port: number;
  pid: number;
  url: string;
  restarts: number;
  created_at: string;
  updated_at: string;
}

export interface WorkflowListItem {
  name: string;
  path: string;
  method: string;
  uiType: string;
  title: string;
  description?: string;
  hasFile?: boolean;
  steps?: string[];
  params?: WorkflowParam[];
  chatParam?: string;
  fileParam?: string;
  suggestions?: string[];
  accent?: string;
  logo?: string;
  theme?: string;
  avatar?: string;
  arpUrl?: string;
  backend?: string;
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
  model?: string;
  input_tokens?: number;
  output_tokens?: number;
  latency_ms?: number;
  cost_usd?: number;
  tool_call_count?: number;
  tool_name?: string;
  success?: boolean;
  duration_ms?: number;
}
