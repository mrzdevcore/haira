// Core barrel export — single import for all foundation utilities.

export { BaseComponent } from "./base-component";
export { esc, escAttr, classMap, styleMap, formatBytes, formatToolName } from "./html";
export { icons, icon, logoSvg } from "./icons";
export type { IconName } from "./icons";
export {
  themeVars,
  lightThemeVars,
  keyframes,
  baseCSS,
  scrollbarCSS,
  cardCSS,
  animateInCSS,
  methodColor,
  uiTypeColor,
  hexToRgb,
  lighten,
} from "./styles";

// Re-export all types
export type {
  WorkflowParam,
  WorkflowMeta,
  WorkflowListItem,
  StepLogEntry,
  StepEvent,
  StepStatus,
  RunSummary,
  RunDetail,
  ChatSessionSummary,
  ChatSessionDetail,
  ChatMessageEntry,
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
} from "./types";
