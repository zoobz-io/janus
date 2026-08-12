/** People, their linked identity accounts, and their sessions. */

import { op } from "../spec.js";

export const users = {
  list: op("get", "/users"),
  create: op("post", "/users"),
  get: op("get", "/users/{user_id}"),
  update: op("patch", "/users/{user_id}"),

  accounts: {
    list: op("get", "/users/{user_id}/accounts"),
    unlink: op("delete", "/users/{user_id}/accounts/{account_id}"),
  },

  sessions: {
    list: op("get", "/users/{user_id}/sessions"),
    revokeAll: op("delete", "/users/{user_id}/sessions"),
    revoke: op("delete", "/users/{user_id}/sessions/{session_id}"),
  },
};
