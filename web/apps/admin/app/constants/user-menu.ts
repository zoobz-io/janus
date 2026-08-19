import UserMenu from "~/components/user-menu.vue";

import { defineAdapter } from "@zoobzio/foundation/definitions/adapter";

export const USER_MENU_ADAPTER = defineAdapter({
  component: UserMenu,
  emits: {},
});
