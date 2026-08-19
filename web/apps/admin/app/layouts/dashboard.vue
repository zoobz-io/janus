<script lang="ts">
import Aside from "@zoobzio/foundation/components/common/aside.vue";
import Group from "@zoobzio/foundation/components/common/group.vue";
import Main from "@zoobzio/foundation/components/common/main.vue";
import Panel from "@zoobzio/foundation/components/system/panel.vue";

import { useSidebar } from "~/composables/nav";
</script>

<script setup lang="ts">
defineSlots<{ default(): unknown }>();

const sidebar = useSidebar();
</script>

<template>
  <Group class="dashboard">
    <Aside class="dashboard-sidebar">
      <Panel :panel="sidebar" />
    </Aside>
    <Main class="dashboard-content">
      <slot />
    </Main>
  </Group>
</template>

<style scoped>
.dashboard {
  display: grid;
  grid-template-columns: 16rem 1fr;
  min-height: 100dvh;
}

.dashboard-content {
  min-width: 0;
}

.dashboard-sidebar :deep(.f-system-panel) {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.dashboard-sidebar :deep(.f-system-panel-content) {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}
</style>
