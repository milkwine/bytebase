<template>
  <div
    class="tab-item"
    :class="{
      current: isCurrentTab,
      temp: isTempTab(tab),
      hovering: state.hovering,
      admin: tab.mode === TabMode.Admin,
    }"
    @mousedown="$emit('select', tab, index)"
    @mouseenter="state.hovering = true"
    @mouseleave="state.hovering = false"
  >
    <div class="body">
      <Prefix :tab="tab" :index="index" />
      <Label :tab="tab" :index="index" />
      <Suffix :tab="tab" :index="index" @close="$emit('close', tab, index)" />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, PropType, reactive } from "vue";
import { storeToRefs } from "pinia";

import type { TabInfo } from "@/types";
import { TabMode } from "@/types";
import { isTempTab } from "@/utils";
import { useTabStore } from "@/store";
import Prefix from "./Prefix.vue";
import Label from "./Label.vue";
import Suffix from "./Suffix.vue";

type LocalState = {
  hovering: boolean;
};

const props = defineProps({
  tab: {
    type: Object as PropType<TabInfo>,
    required: true,
  },
  index: {
    type: Number,
    required: true,
  },
});

defineEmits<{
  (e: "select", tab: TabInfo, index: number): void;
  (e: "close", tab: TabInfo, index: number): void;
}>();

const state = reactive<LocalState>({
  hovering: false,
});

const tabStore = useTabStore();
const { currentTabId } = storeToRefs(tabStore);

const isCurrentTab = computed(() => props.tab.id === currentTabId.value);
</script>

<style scoped lang="postcss">
.tab-item {
  @apply cursor-pointer border-r bg-white gap-x-2;
}
.hovering {
  @apply bg-gray-50;
}

.body {
  @apply flex items-center justify-between gap-x-1 pl-2 pr-1 py-1 border-t-2 border-t-transparent;
}
.current .body {
  @apply relative bg-white text-accent border-t-accent;
}
</style>
