/**
 * The app's front door to the admin API. Wraps `usePress` so app-wide wiring
 * (logger, hooks, mapError) has one place to live as the UI grows; callers
 * can still layer their own config on top.
 */

import type { ClientConfig } from "@janus/admin-sdk";
import { usePress } from "#imports";

export const useAdminApi = (config?: ClientConfig) => {
  return usePress("admin", config);
};
