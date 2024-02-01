#!/usr/bin/env node

const fs = require("node:fs");
const path = require("node:path");
const zlib = require("node:zlib");
const { stdout } = require("node:process");
const { execFileSync } = require("node:child_process");

const tar = require("tar-fs");
const { createHash } = require("node:crypto");
const AdmZip = require("adm-zip");

const ARCH_MAPPING = {
  ia32: "386",
  x64: "amd64",
  arm: "arm",
};

const PLATFORM_MAPPING = {
  linux: "linux",
  win32: "windows",
};

const PLATFORM = PLATFORM_MAPPING[process.platform];
const UNSUPPORTED_PLATFORM = new Error("Unsupported Platform: " + PLATFORM);

const ARCH = ARCH_MAPPING[process.arch];
const UNSUPPORTED_ARCH = new Error("Unsupported Arch: " + ARCH);

const REPOSITORY = "squarecloudofc/cli";
const BIN_NAME = "squarecloud";

async function getCurrentRelease() {
  const res = await fetch(
    "https://api.github.com/repos/" + REPOSITORY + "/releases/latest",
    {
      headers: {
        "Content-Type": "application/json",
        "User-Agent": "Square Cloud CLI",
      },
    },
  );

  return res.json();
}

async function extractFileZip(path, destination) {
  const zip = new AdmZip(path)
  return new Promise((resolve, reject) => zip.extractAllToAsync(destination, true, false, (err) => (err ? reject(err) : resolve())));
}

async function extractFile(path, destination) {
  if (path.endsWith(".zip")) {
    return extractFileZip(path, destination)  
  } else if (path.endsWith(".tar.gz")) {
    const extract = tar.extract(destination);
    fs.createReadStream(path).pipe(zlib.createGunzip()).pipe(extract);
  }

  return true;
}

async function installBinaries(destination) {
  if (!PLATFORM) throw UNSUPPORTED_PLATFORM;
  if (!ARCH) throw UNSUPPORTED_ARCH;

  const release = await getCurrentRelease();
  const regex = new RegExp(`${BIN_NAME}_${PLATFORM}_${ARCH}`);
  const asset = release.assets.filter((a) => regex.test(a.name))[0];
  if (!asset) throw new Error(`Cannot find an asset for ${PLATFORM} - ${ARCH}`);

  const checksumAsset = release.assets.find(a => a.name === "checksums.txt")
  try {
    console.log("Downloading checksums...")
    const cres = await fetch(checksumAsset.browser_download_url);
    const checksumText = await cres.text()
    const checksumList = checksumText.split(/\n/g).map(c => c.split(/\s+/g))
    const [checksum, checksumFilename] = checksumList.find(c => asset.browser_download_url.includes(c[1]))

    console.log("Downloading binaries from github release...");
    console.log(asset.browser_download_url);

    const res = await fetch(asset.browser_download_url);
    fs.mkdirSync(destination, { recursive: true });

    const filepath = path.resolve(destination, asset.name);
    const stream = fs.createWriteStream(filepath);

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
    });

    stream.once("close", async () => {
      console.log("Validating file...")
      const file = fs.readFileSync(filepath)
      const hash = createHash("sha256").update(file).digest("hex")
      
      if (hash != checksum) {
        throw new Error("The file checksum does not match with checksums.txt")
      }

      console.log("Extracting file...");
      await extractFile(filepath, destination);
      fs.unlinkSync(filepath);


      console.log("Square Cloud CLI successfuly installed, please perform \"squarecloud --help\"");
    });
  } catch (err) {
    console.log("Error when trying to download and extract the file", err);
  }
}

function getBinDir() {
  switch (process.platform) {
    case "linux":
      return path.join(process.env.HOME, ".squarecloud");
    case "win32":
      return path.join(process.env.APPDATA, "squarecloud");
    default:
      throw new Error("Platform Unsupported");
  }
}

function getExecFile() {
  const extension = process.platform === "win32" ? ".exe" : "";
  const binDir = getBinDir();
  const execfile = path.resolve(binDir, `squarecloud${extension}`);

  return execfile;
}

(async () => {
  const binDir = getBinDir();
  const execfile = getExecFile();

  const argv = process.argv;
  if (argv[2] === "update" || !fs.existsSync(execfile)) {
    installBinaries(binDir);
    return;
  }

  stdout.write(execFileSync(execfile, argv.slice(2)));
})();