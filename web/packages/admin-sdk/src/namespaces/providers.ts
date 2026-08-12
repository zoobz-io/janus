/** Supported external identity providers. */

import { op } from "../spec.js";

export const providers = {
  list: op("get", "/providers"),
};
