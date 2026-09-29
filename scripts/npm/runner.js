#!/usr/bin/env node

import { getExecFile } from "./constants.js";
import { spawnSync } from "node:child_process"
import { installBinaries } from "./lib.js";
import { existsSync } from "node:fs";


async function run() {
  const binfile = getExecFile();
  if (!existsSync(binfile)) {
    await installBinaries()
  }

  const [, , ...args] = process.argv

  let result = spawnSync(binfile, args, {
    cwd: process.cwd(),
    stdio: "inherit",
  });
  if (result.error)
    console.error(result.error);

  return result.status;
}

// Forward the CLI's exit code; null means it was killed by a signal.
run().then(
  (status) => process.exit(status ?? 1),
  (err) => {
    console.error(`Square Cloud CLI installation failed: ${err.message}`);
    process.exit(1);
  },
);
