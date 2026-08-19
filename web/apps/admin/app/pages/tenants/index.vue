<script lang="ts">
import Workspace from "@zoobzio/foundation/components/system/workspace.vue";

import { definePageMeta } from "#imports";
import { useTenantsPage } from "~/composables/tenants";
</script>

<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: "auth",
  auth: { scopes: ["directory:read"] },
});

const workspace = useTenantsPage();
</script>

<template>
  <Workspace :workspace="workspace" />
</template>
