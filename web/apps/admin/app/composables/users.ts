import { useWorkspace } from "@zoobzio/foundation/composables/workspace";
import { useTable } from "@zoobzio/foundation/factories/table";

import { useAdminApi } from "~/composables/api";
import { USERS_WORKSPACE } from "~/constants/users";

export const useUsersPage = () => {
  const api = useAdminApi();

  const users = useTable("users", USERS_WORKSPACE.slots.users.widget, {
    fetch: async ({ page, pageSize, sortField, sortDirection }) => {
      const { users: data, page: meta } = await api.users.search({
        body: {
          page: { number: page, size: pageSize },
          ...(sortField !== null && {
            sort: { field: sortField, order: sortDirection },
          }),
        },
      });
      return {
        data,
        total: meta.total_items,
        pageCount: meta.total_pages,
      };
    },
  });

  return useWorkspace(USERS_WORKSPACE, { users });
};
