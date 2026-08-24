import type { DirectoryDefinition } from "@zoobzio/foundation/definitions/directory";
import type { DirectoryItem } from "@zoobzio/foundation/types/core/directory";

import Directory from "@zoobzio/foundation/components/core/directory.vue";

import { defineAdapter } from "@zoobzio/foundation/definitions/adapter";
import { defineDirectory } from "@zoobzio/foundation/definitions/directory";

export const NAV = defineDirectory<DirectoryItem>({
  groups: [
    {
      key: "directory",
      label: "Directory",
      items: [
        {
          key: "applications",
          label: "Applications",
          icon: "applications",
          link: { to: "/applications" },
        },
        { key: "users", label: "Users", icon: "users", link: { to: "/users" } },
        {
          key: "tenants",
          label: "Tenants",
          icon: "tenants",
          link: { to: "/tenants" },
        },
        {
          key: "providers",
          label: "Providers",
          icon: "providers",
          link: { to: "/providers" },
        },
      ],
    },
  ],
});

export const NAV_ADAPTER = defineAdapter<DirectoryDefinition<DirectoryItem>>({
  component: Directory,
  emits: { select: true },
  settings: NAV,
});
