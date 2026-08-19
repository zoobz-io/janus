import { useWorkspace } from "@zoobzio/foundation/composables/workspace";
import { useTable } from "@zoobzio/foundation/factories/table";

import { useAdminApi } from "~/composables/api";
import { USERS_WORKSPACE } from "~/constants/users";

export const useUsersPage = () => {
  const api = useAdminApi();

  const users = useTable("users", USERS_WORKSPACE.slots.users.widget, {
    fetch: async ({ page, pageSize }) => {
      const { users: data, total } = await api.users.list({
        query: {
          limit: String(pageSize),
          offset: String((page - 1) * pageSize),
        },
      });
      return {
        data,
        total,
        pageCount: Math.max(1, Math.ceil(total / pageSize)),
      };
    },
  });

  return useWorkspace(USERS_WORKSPACE, { users });
};
