<script lang="ts">
import type { MenuGroup, MenuItem } from "@zoobzio/foundation/types/core/menu";

import Button from "@zoobzio/foundation/components/common/button.vue";
import Menu from "@zoobzio/foundation/components/core/menu.vue";

import { navigateTo, useAuth } from "#imports";
</script>

<script setup lang="ts">
const auth = useAuth();

const groups: MenuGroup[] = [
  {
    key: "account",
    items: [{ label: "Settings" }, { label: "Logout" }],
  },
];

const select = async (item: MenuItem) => {
  if (item.label === "Settings") {
    await navigateTo("/settings");
  } else if (item.label === "Logout") {
    await auth.logout();
  }
};
</script>

<template>
  <Menu :groups="groups" side="top" align="start" @select="select">
    <Button class="user-menu-trigger">
      {{ auth.current?.name ?? "Account" }}
    </Button>
  </Menu>
</template>

<style scoped>
.user-menu-trigger {
  width: 100%;
}
</style>
