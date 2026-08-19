import { usePanel } from "@zoobzio/foundation/composables/panel";
import { useAdapter } from "@zoobzio/foundation/factories/adapter";

import { SIDEBAR } from "~/constants/sidebar";

export const useSidebar = () =>
  usePanel({
    header: useAdapter("sidebar-brand", SIDEBAR.header),
    content: useAdapter("sidebar-nav", SIDEBAR.content),
    footer: useAdapter("sidebar-user-menu", SIDEBAR.footer),
  });
