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

// Generative UI types

export interface ToolRenderEvent {
  tool: string;
  component: string;
  props: Record<string, unknown>;
}
