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

run()
