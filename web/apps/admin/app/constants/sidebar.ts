import { definePanel } from "@zoobzio/foundation/definitions/panel";

import { BRAND_ADAPTER } from "~/constants/brand";
import { NAV_ADAPTER } from "~/constants/nav";
import { USER_MENU_ADAPTER } from "~/constants/user-menu";

export const SIDEBAR = definePanel({
  header: BRAND_ADAPTER,
  content: NAV_ADAPTER,
  footer: USER_MENU_ADAPTER,
});
