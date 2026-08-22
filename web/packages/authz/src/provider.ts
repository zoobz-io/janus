/**
 * The janus provider: letters-patent callbacks over the janus public API.
 *
 * Janus login is redirect-based — `login` sends the browser to the hosted
 * flow and the session lands in a cookie on return, so `resolve` is where
 * state is actually established: it reads identity (`/me`) and app-scoped
 * authorization (`/me/authorization/{app}`) with the ambient cookie and
 * hands both to the bridge. Server-side callers supply a fetch/headers that
 * carry the incoming request's cookie.
 */

import type { Authorization, Identity, Options, Session } from "./types";

import { defineProvider } from "letters-patent/kit";
import { settle, loginUrl } from "./util";

export const createProvider = defineProvider<Options, Session>(
  (options, bridge) => {
    const base = options.api.replace(/\/+$/, "");

    const get = (path: string): Promise<Response> => {
      const request = options.fetch ?? globalThis.fetch.bind(globalThis);
      return request(`${base}${path}`, {
        headers: options.headers,
        credentials: "include",
        redirect: "manual",
      });
    };

    return {
      /**
       * Redirect-based: navigate the browser into the hosted flow and leave
       * state untouched — `resolve` picks the session up when the flow
       * returns. Outside a browser this is a no-op; server-side apps
       * navigate to `loginUrl(api)` themselves.
       */
      login: async () => {
        if (typeof window !== "undefined") {
          window.location.assign(loginUrl(options.api));
        }
      },

      logout: async (state) => {
        await get("/auth/logout");
        state.current = null;
      },

      resolve: async (state) => {
        const [identityRes, authorizationRes] = await Promise.all([
          get("/me"),
          get(`/me/authorization/${encodeURIComponent(options.app)}`),
        ]);

        const identity = await settle<Identity>(identityRes);
        const authorization = await settle<Authorization>(authorizationRes);
        if (identity === null || authorization === null) {
          state.current = null;
          return;
        }

        const entitled = authorization.tenants.length > 0;
        if (!entitled && options.requireEntitlement !== false) {
          state.current = null;
          return;
        }

        state.current = bridge({ identity, authorization });
      },
    };
  },
);
