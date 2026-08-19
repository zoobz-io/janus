<script lang="ts">
import Workspace from "@zoobzio/foundation/components/system/workspace.vue";

import { definePageMeta } from "#imports";
import { useLicensesPage } from "~/composables/licenses";
</script>

<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: "auth",
  auth: { scopes: ["directory:read"] },
});

const workspace = useLicensesPage();
</script>

<template>
  <Workspace :workspace="workspace" />
</template>
