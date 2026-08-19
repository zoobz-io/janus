<script lang="ts">
import Workspace from "@zoobzio/foundation/components/system/workspace.vue";

import { definePageMeta } from "#imports";
import { useProvidersPage } from "~/composables/providers";
</script>

<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: "auth",
  auth: { scopes: ["directory:read"] },
});

const workspace = useProvidersPage();
</script>

<template>
  <Workspace :workspace="workspace" />
</template>
