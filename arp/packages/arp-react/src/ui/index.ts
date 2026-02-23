// @haira/arp-react/ui — Built-in ARP UI components for React
// Import individually for tree-shaking, or use DEFAULT_COMPONENTS for the full set.

import type { ComponentType } from "react";

export { StatusCard } from "./StatusCard.js";
export { Table } from "./Table.js";
export { CodeBlock } from "./CodeBlock.js";
export { Diff } from "./Diff.js";
export { KeyValue } from "./KeyValue.js";
export { Progress } from "./Progress.js";
export { Chart } from "./Chart.js";
export { Form } from "./Form.js";
export { Confirm } from "./Confirm.js";
export { Choices } from "./Choices.js";
export { ProductCards } from "./ProductCards.js";

// Re-imports for the map
import { StatusCard } from "./StatusCard.js";
import { Table } from "./Table.js";
import { CodeBlock } from "./CodeBlock.js";
import { Diff } from "./Diff.js";
import { KeyValue } from "./KeyValue.js";
import { Progress } from "./Progress.js";
import { Chart } from "./Chart.js";
import { Form } from "./Form.js";
import { Confirm } from "./Confirm.js";
import { Choices } from "./Choices.js";
import { ProductCards } from "./ProductCards.js";

/**
 * Default component registry mapping ARP component type names to React components.
 * Pass this to `<ArpChat components={...}>` or `<ArpRenderer components={...}>`.
 *
 * ```tsx
 * import { DEFAULT_COMPONENTS } from "@haira/arp-react/ui";
 *
 * <ArpChat
 *   url="ws://localhost:8080/_arp/v1"
 *   components={{
 *     ...DEFAULT_COMPONENTS,
 *     "my-widget": MyWidget, // extend with custom components
 *   }}
 * />
 * ```
 */
export const DEFAULT_COMPONENTS: Record<string, ComponentType<any>> = {
  "status-card": StatusCard,
  "table": Table,
  "code-block": CodeBlock,
  "diff": Diff,
  "key-value": KeyValue,
  "progress": Progress,
  "chart": Chart,
  "form": Form,
  "confirm": Confirm,
  "choices": Choices,
  "product-cards": ProductCards,
};
