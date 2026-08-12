/**
 * @janus/admin-sdk — the typed client for the Janus Admin API.
 *
 * `createAdminClient(config)` returns a resource-namespaced client
 * (`client.tenants.members.list(tenantId)`); methods take positional path params
 * then a trailing options object, return the response body directly, and throw
 * from the openapi-press error hierarchy on failure. Cross-cutting behavior
 * (retry, timeouts, caching, …) is applied per endpoint through `.with`.
 */

export { createAdminClient } from "./client.js";
export type { AdminClient } from "./client.js";

// Re-export openapi-press's shared surface so consumers configure, instrument,
// and catch without a second dependency.
export {
  withCache,
  withCircuitBreaker,
  withDedupe,
  withFallback,
  withReauth,
  withRetry,
  withTimeout,
} from "openapi-press";
export type {
  CallConfig,
  CallMeta,
  CircuitBreakerPolicy,
  ClientConfig,
  ErrorHookContext,
  Hooks,
  Logger,
  RequestHookContext,
  ResponseHookContext,
  RetryPolicy,
  Wrapper,
} from "openapi-press";
export * from "openapi-press/error";

// The generated schema types, for consumers that want to name request/response
// shapes directly (e.g. `components["schemas"]["TenantResponse"]`).
export type { components, paths } from "./schema.gen.js";
