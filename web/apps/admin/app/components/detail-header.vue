<script lang="ts">
import type { IconAlias } from "@zoobzio/foundation/types/common/icon";

import Anchor from "@zoobzio/foundation/components/common/anchor.vue";
import Icon from "@zoobzio/foundation/components/common/icon.vue";

/**
 * One tab in a detail header. `to` is a route path: the tab bar is
 * router-driven — each tab is a page under the entity's `[id]/` directory,
 * and the active state comes from the router, so tabs are deep-linkable.
 */
export interface DetailTab {
  label: string;
  to: string;
  icon?: IconAlias;
}
</script>

<script setup lang="ts">
const { title, subtitle, tabs } = defineProps<{
  title: string;
  subtitle?: string;
  tabs: DetailTab[];
}>();
</script>

<template>
  <header class="detail-header">
    <div class="detail-header-identity">
      <h1>{{ title }}</h1>
      <code v-if="subtitle">{{ subtitle }}</code>
    </div>
    <nav class="detail-header-tabs">
      <Anchor
        v-for="tab in tabs"
        :key="tab.to"
        :to="tab.to"
        class="detail-header-tab"
      >
        <Icon v-if="tab.icon" :alias="tab.icon" />
        {{ tab.label }}
      </Anchor>
    </nav>
  </header>
</template>

<style scoped>
.detail-header-identity {
  display: flex;
  align-items: baseline;
  gap: 0.75rem;
}

.detail-header-tabs {
  display: flex;
  gap: 0.25rem;
  border-bottom: 1px solid color-mix(in srgb, currentColor 20%, transparent);
}

.detail-header-tab {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 0.75rem;
  text-decoration: none;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
}

.detail-header-tab.router-link-exact-active {
  border-bottom-color: currentColor;
  font-weight: 600;
}
</style>
