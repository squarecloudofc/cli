#!/usr/bin/env node

import { getExecFile } from "./constants.js";
import { spawnSync } from "node:child_process"

const binfile = getExecFile();
const result = spawnSync(binfile, process.argv.slice(2), {
  stdio: "inherit",
  cwd: process.cwd()
});

if (result.error) {
  console.error(result.error);
  process.exit(1);
}

process.exit(result.status);
