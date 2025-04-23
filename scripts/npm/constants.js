import pkg from "../package.json" with { type: 'json' };
import { resolve } from "node:path"

export const ARCH_MAPPING = {
  ia32: "386",
  x64: "amd64",
  arm: "arm",
};

export const PLATFORM_MAPPING = {
  linux: "linux",
  win32: "windows",
};

export const PLATFORM = PLATFORM_MAPPING[process.platform];
export const UNSUPPORTED_PLATFORM = new Error("Unsupported Platform: " + PLATFORM);

export const ARCH = ARCH_MAPPING[process.arch];
export const UNSUPPORTED_ARCH = new Error("Unsupported Arch: " + ARCH);

export const REPOSITORY = "squarecloudofc/cli";
export const BIN_NAME = "squarecloud";

export const VERSION = pkg.version

export function getBinDir() {
  return resolve("bin")
}

export function getExecFile() {
  const extension = process.platform === "win32" ? ".exe" : "";
  const binDir = getBinDir();
  const execfile = resolve(binDir, `squarecloud${extension}`);

  return execfile;
}

