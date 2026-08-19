<script lang="ts">
import Workspace from "@zoobzio/foundation/components/system/workspace.vue";

import { definePageMeta } from "#imports";
import { useScopesPage } from "~/composables/scopes";
</script>

<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: "auth",
  auth: { scopes: ["directory:read"] },
});

const workspace = useScopesPage();
</script>

<template>
  <Workspace :workspace="workspace" />
</template>
