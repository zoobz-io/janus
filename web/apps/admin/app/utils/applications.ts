import type { AdminClient } from "@janus/admin-sdk";
import type { components } from "@janus/admin-sdk";

export type Application = components["schemas"]["ApplicationResponse"];

/**
 * Aggregates an application-scoped listing across every application on the
 * mesh: the admin API only lists scopes, tiers, licenses, and grants under
 * one application, so the cross-app pages fan out and flatten, labeling each
 * row with the application it came from.
 */
export const acrossApplications = async <R>(
  api: AdminClient,
  list: (app: Application) => Promise<R[]>,
): Promise<(R & { application: string })[]> => {
  const { applications } = await api.applications.list();
  const results = await Promise.all(
    applications.map(async (app) =>
      (await list(app)).map((row) => ({ ...row, application: app.name })),
    ),
  );
  return results.flat();
};
