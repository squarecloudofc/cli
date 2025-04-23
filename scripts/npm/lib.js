import { createReadStream, mkdirSync, createWriteStream, readFileSync, unlinkSync } from "node:fs";
import { resolve as _resolve } from "node:path";
import { createGunzip } from "node:zlib";

import { extract as _extract } from "tar-fs";
import { createHash } from "node:crypto";
import AdmZip from "adm-zip";
import { REPOSITORY, VERSION, BIN_NAME, ARCH, PLATFORM, UNSUPPORTED_ARCH, UNSUPPORTED_PLATFORM, BIN_DIR } from "./constants.js";

export async function installBinaries() {
  if (!PLATFORM) throw UNSUPPORTED_PLATFORM;
  if (!ARCH) throw UNSUPPORTED_ARCH;

  const destination = BIN_DIR;
  const release = await getRelease(`v${VERSION}`).catch(err => {
    throw new Error(`Unable to find release v${VERSION}`);
  });
  const assetPrefix = `${BIN_NAME}_${PLATFORM}_${ARCH}`
  const asset = release.assets.find((a) => a.name.startsWith(assetPrefix));
  if (!asset) throw new Error(`Cannot find an asset for ${PLATFORM} - ${ARCH} (${assetPrefix})`);

  const checksumAsset = release.assets.find(a => a.name === "checksums.txt");

  try {
    const checksumText = await (await fetch(checksumAsset.browser_download_url)).text();
    const checksumList = checksumText.split('\n').map((c) => c.split(/\s+/));
    const checksumEntry = checksumList.find((c) => asset.browser_download_url.includes(c[1]));
    if (!checksumEntry) throw new Error('Checksum entry not found for the asset');
    const [checksum] = checksumEntry;

    mkdirSync(destination, { recursive: true });
    const filepath = _resolve(destination, asset.name);

    await downloadFile(asset.browser_download_url, filepath);
    await validateChecksum(filepath, checksum);

    console.log(`Extracting file to ${destination}`);
    await extractFile(filepath, destination);
    unlinkSync(filepath);

  } catch (err) {
    console.log(`Unable to install the CLI:`, err)
  }
}

async function validateChecksum(filePath, checksum) {
  console.log('Validating file checksum...');
  const file = readFileSync(filePath);
  const hash = createHash('sha256').update(file).digest('hex');
  if (hash !== checksum) {
    throw new Error('The file checksum does not match with checksums.txt');
  }
}

async function downloadFile(url, destination) {
  return new Promise(async (resolve, reject) => {
    console.log(`Downloading from ${url}`);
    const res = await fetch(url);
    if (!res.ok) throw new Error(`Failed to download file from ${url}`);

    const stream = createWriteStream(destination);
    const reader = res.body.getReader();

    while (true) {
      const { done, value } = await reader.read();
      if (done) {
        break;
      }

      stream.write(value);
    }

    stream.end();

    stream.once("error", (err) => {
      console.log("Error when trying to download the file", err);
      reject(err);
    });

    stream.once("close", () => {
      console.log("Download completed.");
      resolve();
    });
  })
}

async function extractFile(path, destination) {
  return new Promise((resolve, reject) => {
    if (path.endsWith(".zip")) {
      const zip = new AdmZip(path);
      zip.extractAllToAsync(destination, true, false, (err) => (err ? reject(err) : resolve()))
    }

    if (path.endsWith(".tar.gz")) {
      const extract = _extract(destination);
      const stream = createReadStream(path)
        .pipe(createGunzip())
        .pipe(extract);

      stream.on("finish", resolve);
      stream.on("error", reject);
    }
  })
}

async function getRelease(tag) {
  const res = await fetch(
    `https://api.github.com/repos/${REPOSITORY}/releases/tags/${tag}`,
    {
      headers: {
        "Content-Type": "application/json",
        "User-Agent": "Square Cloud CLI",
      },
    },
  );

  return res.json();
}
