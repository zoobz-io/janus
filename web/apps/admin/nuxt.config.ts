import { defineNuxtConfig } from "nuxt/config";

import { contract } from "./shared/contract";
import iconSheets from "./config/icon-sheets";

export default defineNuxtConfig({
  compatibilityDate: "2025-11-06",
  extends: ["@zoobzio/foundation"],
  css: ["~/assets/reset.css", "~/assets/admin.css"],
  modules: ["@openapi-press/nuxt", "@letters-patent/nuxt"],
  press: {
    clients: {
      admin: {
        client: "~~/shared/presses/admin",
        host: "http://127.0.0.1:8081",
        prefix: "/api/admin",
      },
    },
  },
  lettersPatent: { contract, login: "/" },
  iconSheets,
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
