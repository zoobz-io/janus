import { defineNuxtConfig } from "nuxt/config";

import { contract } from "./shared/contract";

export default defineNuxtConfig({
  compatibilityDate: "2025-11-06",
  extends: ["@zoobzio/foundation"],
  modules: ["@openapi-press/nuxt", "@crestable/nuxt"],
  press: {
    clients: {
      admin: {
        client: "~~/shared/presses/admin",
        host: "http://127.0.0.1:8081",
        prefix: "/api/admin",
      },
    },
  },
  crestable: { contract },
  runtimeConfig: {
    janus: {
      authHost: "http://127.0.0.1:8080",
    },
    public: {
      janus: {
        authOrigin: "http://localhost:8080",
      },
    },
  },
});
