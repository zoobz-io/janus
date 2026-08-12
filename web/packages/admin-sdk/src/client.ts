/**
 * The admin client factory. Assembles the domain namespaces into one tree and
 * captures it into a Press via `client` from the spec binding. The returned
 * client is fully typed against the admin spec; `createAdminClient` is the
 * only entry point.
 */

import { applications } from "./namespaces/applications.js";
import { providers } from "./namespaces/providers.js";
import { tenants } from "./namespaces/tenants.js";
import { users } from "./namespaces/users.js";
import { client } from "./spec.js";

const namespaces = { applications, tenants, users, providers };

/**
 * Creates an admin client. With no config it targets the same origin (relative
 * `baseUrl`), which is what the browser wants when a proxy fronts the admin API.
 */
export const createAdminClient = client(namespaces);

/** A live, fully-typed admin client. */
export type AdminClient = ReturnType<typeof createAdminClient>;
