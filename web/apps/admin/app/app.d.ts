/**
 * Manual runtime-config augmentations. Nuxt generates these into
 * .nuxt/types/runtime-config.d.ts but the merge doesn't reach this project's
 * type checking (unresolved — suspected layer/type-plumbing issue), so the
 * custom keys are declared here directly.
 */

import type { AnyEvents as AdapterEvents } from "@zoobzio/foundation/types/data/adapter";
import type { Events as TableEvents } from "@zoobzio/foundation/types/data/table";

/**
 * Same plumbing gap as runtime config: the foundation layer's own app.d.ts
 * registers these hook events, but the augmentation doesn't reach this
 * project's type checking, so they are re-declared here.
 */
declare module "#app" {
  interface RuntimeNuxtHooks extends AdapterEvents, TableEvents {}
}

declare module "nuxt/schema" {
  interface RuntimeConfig {
    janus: {
      authHost: string;
    };
  }
  interface PublicRuntimeConfig {
    janus: {
      authOrigin: string;
    };
  }
}

declare module "@nuxt/schema" {
  interface RuntimeConfig {
    janus: {
      authHost: string;
    };
  }
  interface PublicRuntimeConfig {
    janus: {
      authOrigin: string;
    };
  }
}

export {};
