import type { DirectoryDefinition } from "@zoobzio/foundation/definitions/directory";
import type { DirectoryItem } from "@zoobzio/foundation/types/core/directory";

import Directory from "@zoobzio/foundation/components/core/directory.vue";

import { defineAdapter } from "@zoobzio/foundation/definitions/adapter";
import { defineDirectory } from "@zoobzio/foundation/definitions/directory";

export const NAV = defineDirectory<DirectoryItem>({
  groups: [
    {
      key: "catalog",
      label: "Catalog",
      items: [
        {
          key: "applications",
          label: "Applications",
          icon: "applications",
          link: { to: "/applications" },
        },
        {
          key: "scopes",
          label: "Scopes",
          icon: "scopes",
          link: { to: "/scopes" },
        },
        {
          key: "tiers",
          label: "Tiers",
          icon: "tiers",
          link: { to: "/tiers" },
        },
        {
          key: "licenses",
          label: "Licenses",
          icon: "licenses",
          link: { to: "/licenses" },
        },
        {
          key: "grants",
          label: "Grants",
          icon: "grants",
          link: { to: "/grants" },
        },
      ],
    },
    {
      key: "directory",
      label: "Directory",
      items: [
        {
          key: "tenants",
          label: "Tenants",
          icon: "tenants",
          link: { to: "/tenants" },
        },
        { key: "users", label: "Users", icon: "users", link: { to: "/users" } },
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
