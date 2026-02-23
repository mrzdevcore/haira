// @haira/arp-vue/ui — Built-in ARP UI components for Vue

import type { Component } from "vue";

export { default as StatusCard } from "./StatusCard.vue";
export { default as Table } from "./Table.vue";
export { default as CodeBlock } from "./CodeBlock.vue";
export { default as Diff } from "./Diff.vue";
export { default as KeyValue } from "./KeyValue.vue";
export { default as Progress } from "./Progress.vue";
export { default as Chart } from "./Chart.vue";
export { default as Form } from "./Form.vue";
export { default as Confirm } from "./Confirm.vue";
export { default as Choices } from "./Choices.vue";
export { default as ProductCards } from "./ProductCards.vue";

import StatusCard from "./StatusCard.vue";
import Table from "./Table.vue";
import CodeBlock from "./CodeBlock.vue";
import Diff from "./Diff.vue";
import KeyValue from "./KeyValue.vue";
import Progress from "./Progress.vue";
import Chart from "./Chart.vue";
import Form from "./Form.vue";
import Confirm from "./Confirm.vue";
import Choices from "./Choices.vue";
import ProductCards from "./ProductCards.vue";

/** Default component registry mapping ARP component type names to Vue components. */
export const DEFAULT_COMPONENTS: Record<string, Component> = {
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
