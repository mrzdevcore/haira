<script setup lang="ts">
import { computed } from "vue";

interface ToolCardLocal {
  name: string;
  displayName: string;
  status: "running" | "done" | "failed";
  startTime: number;
  elapsed?: string;
}

const props = defineProps<{
  toolCards: ToolCardLocal[];
  open: boolean;
}>();

const emit = defineEmits<{
  toggle: [];
}>();

const runningCount = computed(
  () => props.toolCards.filter((t) => t.status === "running").length,
);

function statusIcon(status: string): string {
  if (status === "running") return "\u23F3";
  if (status === "done") return "\u2713";
  return "\u2717";
}
</script>

<template>
  <div
    v-if="toolCards.length > 0"
    class="arp-activity-panel"
    :style="{
      borderLeft: '1px solid #2a2a3e',
      background: '#111118',
      overflow: 'hidden',
      transition: 'width 0.2s ease',
      width: open ? '280px' : '0px',
    }"
  >
    <div v-if="open" style="padding: 16px">
      <!-- Header -->
      <div
        style="
          display: flex;
          align-items: center;
          justify-content: space-between;
          margin-bottom: 12px;
        "
      >
        <span style="color: #a0a0b0; font-size: 13px; font-weight: 600">
          Activity
          <span
            v-if="runningCount > 0"
            style="
              margin-left: 8px;
              background: #e8a317;
              color: #000;
              border-radius: 10px;
              padding: 1px 8px;
              font-size: 11px;
            "
          >
            {{ runningCount }}
          </span>
        </span>
        <button
          style="
            background: none;
            border: none;
            color: #a0a0b0;
            cursor: pointer;
            font-size: 16px;
          "
          @click="emit('toggle')"
        >
          &#x2715;
        </button>
      </div>

      <!-- Tool cards -->
      <div style="display: flex; flex-direction: column; gap: 8px">
        <div
          v-for="(card, i) in toolCards"
          :key="i"
          style="
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 8px 10px;
            background: #1a1a2e;
            border-radius: 8px;
            font-size: 13px;
          "
        >
          <span style="font-size: 14px">{{ statusIcon(card.status) }}</span>
          <span
            style="
              flex: 1;
              color: #e0e0e0;
              overflow: hidden;
              text-overflow: ellipsis;
              white-space: nowrap;
            "
          >
            {{ card.displayName }}
          </span>
          <span v-if="card.elapsed" style="color: #666; font-size: 11px">
            {{ card.elapsed }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
