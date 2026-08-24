<script lang="ts">
import type { DetailTab } from "~/components/detail-header.vue";

import {
  computed,
  definePageMeta,
  useAsyncData,
  useRoute,
} from "#imports";
import DetailHeader from "~/components/detail-header.vue";
import { useAdminApi } from "~/composables/api";
</script>

<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: "auth",
  auth: { scopes: ["directory:read"] },
});

const route = useRoute();
const api = useAdminApi();

const appId = computed(() => String(route.params.id));

const { data: application, error } = await useAsyncData(
  () => `application:${appId.value}`,
  () => api.applications.get(appId.value),
  { watch: [appId] },
);

const tabs = computed<DetailTab[]>(() => {
  const base = `/applications/${appId.value}`;
  return [
    { label: "Overview", to: base, icon: "overview" },
    { label: "Scopes", to: `${base}/scopes`, icon: "scopes" },
    { label: "Tiers", to: `${base}/tiers`, icon: "tiers" },
    { label: "Licenses", to: `${base}/licenses`, icon: "licenses" },
    { label: "Grants", to: `${base}/grants`, icon: "grants" },
  ];
});
</script>

<template>
  <div>
    <template v-if="application">
      <DetailHeader
        :title="application.name"
        :subtitle="application.slug"
        :tabs="tabs"
      />
      <NuxtPage />
    </template>
    <p v-else-if="error">Failed to load application: {{ error.message }}</p>
  </div>
</template>
