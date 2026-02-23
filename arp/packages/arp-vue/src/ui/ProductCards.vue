<script setup lang="ts">
import { ref, computed } from "vue";

interface ProductCardItem { name: string; price: string; image?: string; brand?: string; description?: string; badge?: string; url?: string }

const props = defineProps<{ title?: string; cards: ProductCardItem[] }>();

const search = ref("");
const showSearch = computed(() => props.cards.length >= 12);
const filtered = computed(() => {
  if (!search.value) return props.cards;
  const q = search.value.toLowerCase();
  return props.cards.filter((c) => c.name.toLowerCase().includes(q) || c.brand?.toLowerCase().includes(q) || c.description?.toLowerCase().includes(q) || c.price.toLowerCase().includes(q));
});

function openUrl(url?: string) { if (url) window.open(url, "_blank"); }
</script>

<template>
  <div class="arp-product-cards" style="background: #111118; border-radius: 8px; overflow: hidden">
    <div style="padding: 12px 16px; display: flex; align-items: center; gap: 12px; border-bottom: 1px solid #1a1a2e">
      <span v-if="title" style="font-weight: 600; font-size: 14px; color: #e0e0e0">{{ title }}</span>
      <input v-if="showSearch" v-model="search" type="text" placeholder="Search..." style="margin-left: auto; background: #1a1a2e; color: #e0e0e0; border: 1px solid #2a2a3e; border-radius: 6px; padding: 4px 10px; font-size: 12px; width: 160px" />
    </div>
    <div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(190px, 1fr)); gap: 12px; padding: 16px; max-height: 520px; overflow: auto">
      <div v-for="(card, i) in filtered" :key="i" :style="{ background: '#1a1a2e', borderRadius: '8px', overflow: 'hidden', cursor: card.url ? 'pointer' : 'default' }" @click="openUrl(card.url)">
        <div v-if="card.image" style="height: 140px; overflow: hidden; background: #0a0a12">
          <img :src="card.image" :alt="card.name" loading="lazy" style="width: 100%; height: 100%; object-fit: cover" />
        </div>
        <div style="padding: 10px 12px">
          <span v-if="card.badge" style="font-size: 10px; background: #e8a317; color: #000; border-radius: 4px; padding: 1px 6px; font-weight: 600; margin-bottom: 4px; display: inline-block">{{ card.badge }}</span>
          <div style="font-weight: 600; font-size: 13px; color: #e0e0e0; margin-top: 2px">{{ card.name }}</div>
          <div v-if="card.brand" style="font-size: 11px; color: #666">{{ card.brand }}</div>
          <div v-if="card.description" style="font-size: 12px; color: #888; margin-top: 4px; line-height: 1.4">{{ card.description }}</div>
          <div style="font-weight: 700; font-size: 14px; color: #e8a317; margin-top: 6px">{{ card.price }}</div>
        </div>
      </div>
    </div>
  </div>
</template>
