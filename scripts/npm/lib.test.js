// Offline tests: node --test scripts/npm/lib.test.js (fetch is mocked).
import { test } from "node:test";
import assert from "node:assert/strict";

import { ARCH, PLATFORM } from "./constants.js";
import { assetName, checksumFor, installBinaries, releaseURL } from "./lib.js";

test("asset names match the goreleaser archives", () => {
  assert.equal(assetName("linux", "amd64"), "squarecloud_linux_amd64.tar.gz");
  assert.equal(assetName("linux", "armv7"), "squarecloud_linux_armv7.tar.gz");
  assert.equal(assetName("darwin", "arm64"), "squarecloud_darwin_arm64.tar.gz");
  assert.equal(assetName("windows", "386"), "squarecloud_windows_386.zip");
});

test("release files come straight from the release, not the API", () => {
  assert.equal(
    releaseURL("1.2.3", "checksums.txt"),
    "https://github.com/squarecloudofc/cli/releases/download/v1.2.3/checksums.txt",
  );
});

test("checksum lookup is exact", () => {
  const a = "a".repeat(64);
  const b = "B".repeat(64);
  const text = `${a}  squarecloud_linux_amd64.tar.gz\r\n${b}  squarecloud_windows_amd64.zip\n\n`;
  assert.equal(checksumFor(text, "squarecloud_linux_amd64.tar.gz"), a);
  assert.equal(checksumFor(text, "squarecloud_windows_amd64.zip"), "b".repeat(64));
  assert.throws(() => checksumFor(text, "squarecloud_linux_386.tar.gz"), /no entry/);
  assert.throws(() => checksumFor(`${a}  x_squarecloud_linux_amd64.tar.gz`, "squarecloud_linux_amd64.tar.gz"), /no entry/);
  assert.throws(() => checksumFor(`not-a-hash  squarecloud_linux_amd64.tar.gz`, "squarecloud_linux_amd64.tar.gz"), /no entry/);
});

test("a failed download rejects instead of being swallowed", async (t) => {
  const urls = [];
  t.mock.method(globalThis, "fetch", async (url) => {
    urls.push(String(url));
    return new Response("rate limited", { status: 403, statusText: "Forbidden" });
  });
  await assert.rejects(installBinaries(), /403 Forbidden/);
  assert.ok(urls.every((u) => u.startsWith("https://github.com/squarecloudofc/cli/releases/download/")), urls.join());
});

test("a checksum mismatch rejects before anything is written", async (t) => {
  const file = assetName(PLATFORM, ARCH);
  t.mock.method(globalThis, "fetch", async (url) =>
    String(url).endsWith("/checksums.txt")
      ? new Response(`${"0".repeat(64)}  ${file}\n`)
      : new Response("tampered archive"),
  );
  await assert.rejects(installBinaries(), /does not match checksums\.txt/);
});
