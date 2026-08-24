import { defineNuxtIconSheetsConfig } from "@icon-sheets/nuxt/config";

// The admin portal's own semantic aliases, merged over the foundation
// layer's sheet by Nuxt's layer config merging. Refs resolve at build time
// against the local @iconify-json/lucide collection.
export default defineNuxtIconSheetsConfig({
  id: "janus-admin",
  name: "Janus Admin Icons",
  icons: {
    applications: "lucide:app-window",
    overview: "lucide:layout-dashboard",
    view: "lucide:eye",
    scopes: "lucide:key-round",
    tiers: "lucide:layers",
    licenses: "lucide:scroll-text",
    grants: "lucide:shield-check",
    tenants: "lucide:building-2",
    users: "lucide:users",
    providers: "lucide:plug",
  },
});
