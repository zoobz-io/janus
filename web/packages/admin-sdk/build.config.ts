import { execFileSync } from "node:child_process";

import { defineBuildConfig } from "unbuild";

export default defineBuildConfig({
  entries: ["src/index"],
  outDir: ".dist",
  declaration: true,
  rollup: {
    emitCJS: false,
  },
  hooks: {
    // Regenerate the openapi-typescript client types from the committed spec
    // before every build (and stub), so the bundled types can never drift from
    // data/openapi.json. The generated file is gitignored (*.gen.ts).
    "build:before": () => {
      execFileSync(
        "openapi-typescript",
        ["./data/openapi.json", "-o", "./src/schema.ts"],
        { stdio: "inherit" },
      );
    },
  },
});
