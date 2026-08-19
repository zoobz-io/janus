import { useWorkspace } from "@zoobzio/foundation/composables/workspace";
import { useTable } from "@zoobzio/foundation/factories/table";

import { useAdminApi } from "~/composables/api";
import { PROVIDERS_WORKSPACE } from "~/constants/providers";
import { toPage } from "~/utils/rows";

export const useProvidersPage = () => {
  const api = useAdminApi();

  const providers = useTable(
    "providers",
    PROVIDERS_WORKSPACE.slots.providers.widget,
    {
      fetch: async (params) => {
        const { providers: names } = await api.providers.list();
        return toPage(
          names.map((provider) => ({ provider })),
          params,
        );
      },
    },
  );

  return useWorkspace(PROVIDERS_WORKSPACE, { providers });
};
