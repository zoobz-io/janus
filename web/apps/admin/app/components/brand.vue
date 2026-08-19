<script lang="ts">
import Anchor from "@zoobzio/foundation/components/common/anchor.vue";
</script>

<script setup lang="ts">
const { title } = defineProps<{ title: string }>();
</script>

<template>
  <Anchor to="/" class="brand">{{ title }}</Anchor>
</template>

<style scoped>
.brand {
  display: block;
  font-weight: 600;
  text-decoration: none;
}
</style>
