import pkg from "./package.json" with { type: 'json' };
import { join, resolve } from "node:path"
import url from "node:url";

const __dirname = url.fileURLToPath(new URL('.', import.meta.url));

// Values are the goreleaser archive arch names (arm is built as armv7).
export const ARCH_MAPPING = {
  ia32: "386",
  x64: "amd64",
  arm: "armv7",
  arm64: "arm64",
};

export const PLATFORM_MAPPING = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

export const PLATFORM = PLATFORM_MAPPING[process.platform];
export const UNSUPPORTED_PLATFORM = new Error("Unsupported Platform: " + process.platform);

export const ARCH = ARCH_MAPPING[process.arch];
export const UNSUPPORTED_ARCH = new Error("Unsupported Arch: " + process.arch);

export const REPOSITORY = "squarecloudofc/cli";
export const BIN_NAME = "squarecloud";

export const VERSION = process.env.SQUARECLOUD_VERSION ?? pkg.version

export const BIN_DIR = join(__dirname, "bin")

export function getExecFile() {
  const extension = process.platform === "win32" ? ".exe" : "";
  const execfile = resolve(BIN_DIR, `squarecloud${extension}`);

  return execfile;
}

