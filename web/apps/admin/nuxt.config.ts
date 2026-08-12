import { defineNuxtConfig } from "nuxt/config";

export default defineNuxtConfig({
  compatibilityDate: "2025-11-06",

  // Foundation base layer. It lives in a separate repo (~/code/foundation),
  // sibling to this monorepo, so the relative path climbs out of web/apps/admin
  // → web/apps → web → janus → code, then into foundation.
  extends: ["../../../../foundation"],

  modules: ["@openapi-press/nuxt"],

  // The admin API client. SSR calls the host directly (cookie/authorization
  // forwarded); the browser goes through the /api/admin proxy so the upstream
  // host never reaches the client bundle. Host is env-overridable via
  // NUXT_PRESS_CLIENTS_ADMIN_HOST; the default matches `make dev-admin`
  // (127.0.0.1, not localhost — the container publishes on IPv4 only and
  // localhost resolves to ::1 here).
  press: {
    clients: {
      admin: {
        client: "~~/shared/presses/admin",
        host: "http://127.0.0.1:8081",
        prefix: "/api/admin",
      },
    },
  },
});
