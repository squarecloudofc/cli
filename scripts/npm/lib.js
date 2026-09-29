import { createReadStream, mkdirSync, unlinkSync, writeFileSync } from "node:fs";
import { resolve as _resolve } from "node:path";
import { pipeline } from "node:stream/promises";
import { createGunzip } from "node:zlib";
import { createHash } from "node:crypto";

import { REPOSITORY, VERSION, BIN_NAME, ARCH, PLATFORM, UNSUPPORTED_ARCH, UNSUPPORTED_PLATFORM, BIN_DIR } from "./constants.js";

/** The goreleaser archive for a platform and arch (see goreleaser.yaml). */
export function assetName(platform, arch) {
  return `${BIN_NAME}_${platform}_${arch}${platform === "windows" ? ".zip" : ".tar.gz"}`;
}

/** A file of the tagged release. A direct download: no GitHub API call, so no API rate limit. */
export function releaseURL(version, file) {
  return `https://github.com/${REPOSITORY}/releases/download/v${version}/${file}`;
}

/** The SHA-256 of file in a goreleaser checksums.txt ("<hex>  <name>" lines). */
export function checksumFor(text, file) {
  for (const line of text.split("\n")) {
    const [hash, name] = line.trim().split(/\s+/);
    if (name === file && /^[0-9a-f]{64}$/i.test(hash)) return hash.toLowerCase();
  }
  throw new Error(`checksums.txt has no entry for ${file}`);
}

/** Downloads, verifies and extracts the CLI binary. Rejects on any failure. */
export async function installBinaries() {
  if (!PLATFORM) throw UNSUPPORTED_PLATFORM;
  if (!ARCH) throw UNSUPPORTED_ARCH;

  const file = assetName(PLATFORM, ARCH);
  const expected = checksumFor((await download(releaseURL(VERSION, "checksums.txt"))).toString("utf8"), file);

  const archive = await download(releaseURL(VERSION, file));
  const actual = createHash("sha256").update(archive).digest("hex");
  if (actual !== expected) {
    throw new Error(`${file} does not match checksums.txt (expected ${expected}, got ${actual})`);
  }

  mkdirSync(BIN_DIR, { recursive: true });
  const filepath = _resolve(BIN_DIR, file);
  writeFileSync(filepath, archive);
  try {
    console.log(`Extracting ${file} to ${BIN_DIR}`);
    await extractFile(filepath, BIN_DIR);
  } finally {
    unlinkSync(filepath);
  }
}

async function download(url) {
  console.log(`Downloading ${url}`);
  const res = await fetch(url);
  if (!res.ok) throw new Error(`GET ${url} failed: ${res.status} ${res.statusText}`);
  return Buffer.from(await res.arrayBuffer());
}

// The archive libraries load on use, so the helpers above run (and are
// tested) without node_modules.
async function extractFile(path, destination) {
  if (path.endsWith(".zip")) {
    const { default: AdmZip } = await import("adm-zip");
    return new Promise((resolve, reject) => {
      new AdmZip(path).extractAllToAsync(destination, true, false, (err) => (err ? reject(err) : resolve()));
    });
  }

  const { extract } = await import("tar-fs");
  await pipeline(createReadStream(path), createGunzip(), extract(destination));
}
