import { describe, expect, it } from "vitest";

import { createAdminClient } from "../src/index.js";

/**
 * Smoke test that the namespace tree assembles into a live client. This mostly
 * exercises the build/spec wiring — a green run here means `op`/`client` bound
 * against the generated `paths` and the namespaces resolved end-to-end.
 */
describe("createAdminClient", () => {
  const client = createAdminClient();

  it("exposes the top-level resource namespaces", () => {
    expect(client.tenants).toBeDefined();
    expect(client.users).toBeDefined();
    expect(client.applications).toBeDefined();
    expect(client.providers).toBeDefined();
  });

  it("exposes namespace operations as callable methods", () => {
    expect(typeof client.tenants.list).toBe("function");
    expect(typeof client.tenants.get).toBe("function");
  });

  it("exposes the search operations", () => {
    expect(typeof client.applications.search).toBe("function");
    expect(typeof client.tenants.search).toBe("function");
    expect(typeof client.users.search).toBe("function");
  });

  it("exposes nested sub-namespaces", () => {
    expect(typeof client.tenants.members.list).toBe("function");
    expect(typeof client.tenants.members.add).toBe("function");
  });
});
