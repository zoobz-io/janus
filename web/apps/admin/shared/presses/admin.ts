/**
 * The admin Press, default-exported for @openapi-press/nuxt. The module
 * registers this under `press.clients.admin` in nuxt.config and `usePress`
 * builds it with environment-appropriate config (SSR hits the host directly
 * with credentials forwarded; the browser goes through the /api/admin proxy).
 */

import { createAdminClient } from "@janus/admin-sdk";

export default createAdminClient;
