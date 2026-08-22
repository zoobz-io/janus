import { afterEach, describe, expect, it, vi } from "vitest";

import { defineSchema } from "letters-patent";
import type { Contract } from "letters-patent";
import type { State } from "letters-patent/kit";

import { createProvider } from "../src/provider";
import type { Authorization, Identity, Options, Session } from "../src/types";

/**
 * A contract mirroring the admin portal's, so `schema.parse.user` exercises
 * the same validation the real bridge output is proven against.
 */
const contract = {
  scopes: [
    "directory:read",
    "users:manage",
    "tenants:manage",
    "applications:manage",
  ],
  roles: ["operator", "auditor"],
  meta: { tenant: "string" },
} as const satisfies Contract;

type AdminContract = typeof contract;

const schema = defineSchema(contract);

/** The admin portal's bridge: first entitled tenant wins. */
const bridge = (session: Session) => {
  const tenant = session.authorization.tenants[0];
  return {
    id: session.identity.id,
    email: session.identity.email,
    name: session.identity.display_name,
    scopes: tenant?.scopes ?? [],
    roles: tenant?.roles ?? [],
    meta: { tenant: tenant?.tenant_name ?? "" },
  };
};

const identity: Identity = {
  id: "u_1",
  email: "op@example.com",
  display_name: "Op Erator",
  memberships: [{ tenant_id: "t_1", tenant_name: "Ops", role: "owner" }],
};

const authorization: Authorization = {
  user_id: "u_1",
  application: { id: "a_1", name: "Janus Admin", slug: "janus-admin" },
  tenants: [
    {
      tenant_id: "t_1",
      tenant_name: "Ops",
      role: "owner",
      roles: ["operator"],
      scopes: ["directory:read"],
    },
  ],
};

const json = (body: unknown, status = 200): Response =>
  new Response(JSON.stringify(body), { status });

/** A fetch stub that routes canned responses by request pathname. */
const stubFetch = (routes: Record<string, () => Response>) =>
  vi.fn(
    async (
      input: Parameters<typeof globalThis.fetch>[0],
    ): Promise<Response> => {
      const href =
        typeof input === "string"
          ? input
          : input instanceof URL
            ? input.href
            : input.url;
      const { pathname } = new URL(href, "http://localhost");
      const responder = routes[pathname];
      if (responder === undefined) {
        throw new Error(`unexpected request: ${href}`);
      }
      return responder();
    },
  );

const api = "https://janus.test";

const makeProvider = (
  fetch: ReturnType<typeof stubFetch>,
  overrides: Partial<Options> = {},
) =>
  createProvider(
    schema,
    {
      api,
      app: "janus-admin",
      fetch: fetch as unknown as typeof globalThis.fetch,
      ...overrides,
    },
    bridge,
  );

const freshState = (): State<AdminContract> => ({ current: null });

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("resolve", () => {
  it("maps identity + authorization through the bridge onto state", async () => {
    const fetch = stubFetch({
      "/me": () => json(identity),
      "/me/authorization/janus-admin": () => json(authorization),
    });
    const state = freshState();

    await makeProvider(fetch).resolve(state, schema);

    expect(state.current).toEqual({
      id: "u_1",
      email: "op@example.com",
      name: "Op Erator",
      scopes: ["directory:read"],
      roles: ["operator"],
      meta: { tenant: "Ops" },
    });
  });

  it("sends the ambient cookie with manual redirects on every request", async () => {
    const fetch = stubFetch({
      "/me": () => json(identity),
      "/me/authorization/janus-admin": () => json(authorization),
    });

    await makeProvider(fetch).resolve(freshState(), schema);

    expect(fetch).toHaveBeenCalledWith(
      "https://janus.test/me",
      expect.objectContaining({
        credentials: "include",
        redirect: "manual",
      }),
    );
  });

  it("clears state when the caller isn't signed in (401 on /me)", async () => {
    const fetch = stubFetch({
      "/me": () => new Response(null, { status: 401 }),
      "/me/authorization/janus-admin": () => json(authorization),
    });
    const state = freshState();

    await makeProvider(fetch).resolve(state, schema);

    expect(state.current).toBeNull();
  });

  it("clears state for a valid session with no entitled tenants", async () => {
    const fetch = stubFetch({
      "/me": () => json(identity),
      "/me/authorization/janus-admin": () =>
        json({ ...authorization, tenants: [] }),
    });
    const state = freshState();

    await makeProvider(fetch).resolve(state, schema);

    expect(state.current).toBeNull();
  });

  it("hands an unentitled session to the bridge when requireEntitlement is false", async () => {
    const fetch = stubFetch({
      "/me": () => json(identity),
      "/me/authorization/janus-admin": () =>
        json({ ...authorization, tenants: [] }),
    });
    const state = freshState();

    await makeProvider(fetch, { requireEntitlement: false }).resolve(
      state,
      schema,
    );

    expect(state.current).toEqual({
      id: "u_1",
      email: "op@example.com",
      name: "Op Erator",
      scopes: [],
      roles: [],
      meta: { tenant: "" },
    });
  });

  it("throws on a non-401 error status", async () => {
    const fetch = stubFetch({
      "/me": () => new Response("boom", { status: 500 }),
      "/me/authorization/janus-admin": () => json(authorization),
    });

    await expect(
      makeProvider(fetch).resolve(freshState(), schema),
    ).rejects.toThrow(/500/);
  });

  it("normalizes a trailing-slash base and encodes the app slug", async () => {
    const fetch = stubFetch({
      "/me": () => json(identity),
      "/me/authorization/a%20b": () => json(authorization),
    });

    await makeProvider(fetch, {
      api: "https://janus.test/",
      app: "a b",
    }).resolve(freshState(), schema);

    expect(fetch).toHaveBeenCalledWith(
      "https://janus.test/me",
      expect.anything(),
    );
    expect(fetch).toHaveBeenCalledWith(
      "https://janus.test/me/authorization/a%20b",
      expect.anything(),
    );
  });
});

describe("logout", () => {
  it("calls the logout endpoint and clears state", async () => {
    const fetch = stubFetch({
      "/auth/logout": () => new Response(null, { status: 200 }),
    });
    const state: State<AdminContract> = {
      current: { id: "u_1", scopes: [], roles: [], meta: { tenant: "" } },
    };

    await makeProvider(fetch).logout(state, schema);

    expect(fetch).toHaveBeenCalledWith(
      "https://janus.test/auth/logout",
      expect.objectContaining({ credentials: "include" }),
    );
    expect(state.current).toBeNull();
  });
});

describe("login", () => {
  it("is a no-op outside a browser", async () => {
    const fetch = stubFetch({});

    await expect(
      makeProvider(fetch).login(freshState(), schema),
    ).resolves.toBeUndefined();
    expect(fetch).not.toHaveBeenCalled();
  });

  it("navigates the browser to the hosted flow", async () => {
    const assign = vi.fn();
    vi.stubGlobal("window", { location: { assign } });

    await makeProvider(stubFetch({})).login(freshState(), schema);

    expect(assign).toHaveBeenCalledWith("https://janus.test/auth/login");
  });
});
