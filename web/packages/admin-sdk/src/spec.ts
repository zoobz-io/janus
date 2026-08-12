/**
 * Binds openapi-press to the admin API's generated `paths`. Every namespace
 * declares its endpoints with `op`; `client` captures the assembled tree into
 * a Press — the config-accepting client factory. Binding here (once) is what
 * makes the positional path-param inference resolve against the admin spec
 * specifically.
 */

import { definePress } from "openapi-press";

import type { paths } from "./schema.gen.js";

export const { op, client } = definePress<paths>();
