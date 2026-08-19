import Brand from "~/components/brand.vue";

import { defineAdapter } from "@zoobzio/foundation/definitions/adapter";

export const BRAND_ADAPTER = defineAdapter({
  component: Brand,
  emits: {},
  settings: { title: "Janus Admin" },
});
