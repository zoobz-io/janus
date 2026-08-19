<script lang="ts">
import { loginUrl } from "@janus/crest";

import {
  useAsyncData,
  useAuth,
  useRuntimeConfig,
  definePageMeta,
} from "#imports";
import { useAdminApi } from "~/composables/api";
</script>

<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
});

const auth = useAuth();
const config = useRuntimeConfig();

const login = loginUrl(config.public.janus.authOrigin);

const api = useAdminApi();
const { data: directory, error: directoryError } = await useAsyncData(
  "directory",
  async () => {
    if (!auth.can("directory:read")) {
      return null;
    }
    const [applications, tenants] = await Promise.all([
      api.applications.list(),
      api.tenants.list(),
    ]);
    return {
      applications: applications.applications,
      tenants: tenants.tenants,
    };
  },
);
</script>

<template>
  <section>
    <h1>Janus Admin</h1>

    <template v-if="auth.authenticated">
      <p>
        Signed in as <strong>{{ auth.current?.name }}</strong> ({{
          auth.current?.email
        }}) — {{ auth.current?.meta.tenant }}
      </p>
      <ul>
        <li>operator: {{ auth.is("operator") }}</li>
        <li>directory:read: {{ auth.can("directory:read") }}</li>
        <li>users:manage: {{ auth.can("users:manage") }}</li>
        <li>tenants:manage: {{ auth.can("tenants:manage") }}</li>
        <li>applications:manage: {{ auth.can("applications:manage") }}</li>
      </ul>

      <template v-if="directory">
        <h2>Applications ({{ directory.applications.length }})</h2>
        <ul>
          <li v-for="app in directory.applications" :key="app.id">
            {{ app.name }} <code>{{ app.slug }}</code>
          </li>
        </ul>
        <h2>Tenants ({{ directory.tenants.length }})</h2>
        <ul>
          <li v-for="tenant in directory.tenants" :key="tenant.id">
            {{ tenant.name }} <code>{{ tenant.slug }}</code>
          </li>
        </ul>
      </template>
      <p v-else-if="directoryError">
        Admin API error: {{ directoryError.message }}
      </p>

      <button @click="auth.logout()">Sign out</button>
    </template>

    <template v-else>
      <p>Not signed in.</p>
      <a :href="login">Sign in with Janus</a>
    </template>
  </section>
</template>
