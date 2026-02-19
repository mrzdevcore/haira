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
  // Chat-specific
  chatParam?: string;
  fileParam?: string;
  settingsParams?: WorkflowParam[];
  suggestions?: string[];
  // Theming
  accent?: string;
  logo?: string;
  theme?: string;
  avatar?: string;
  // Index-specific
  workflows?: WorkflowListItem[];
}

export interface WorkflowListItem {
  name: string;
  path: string;
  method: string;
  uiType: string;
  title: string;
}

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

// Run history types

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

// Generative UI types

export interface ToolRenderEvent {
  tool: string;
  component: string;
  props: Record<string, unknown>;
}
