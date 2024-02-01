#!/usr/bin/env node

const fs = require("node:fs");
const path = require("node:path");
const zlib = require("node:zlib");
const { stdout } = require("node:process");
const { execFileSync } = require("node:child_process");

const tar = require("tar-fs");
const unzipper = require("unzipper");

const ARCH_MAPPING = {
  ia32: "386",
  x64: "amd64",
  arm: "arm",
};

const PLATFORM_MAPPING = {
  darwin: "darwin",
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

async function extractFile(path) {
  if (path.endsWith(".zip")) {
    const extract = unzipper.Extract({ path: "./bin" });

    fs.createReadStream(path).pipe(extract);
  } else if (path.endsWith(".tar.gz")) {
    const extract = tar.extract("./bin");

    fs.createReadStream(path).pipe(zlib.createGunzip()).pipe(extract);
  }

  return true
}

async function installBinaries() {
  if (!PLATFORM) throw UNSUPPORTED_PLATFORM;
  if (!ARCH) throw UNSUPPORTED_ARCH;

  const release = await getCurrentRelease();
  const regex = new RegExp(`${BIN_NAME}_${PLATFORM}_${ARCH}`);
  const asset = release.assets.filter((a) => regex.test(a.name))[0];
  if (!asset) throw new Error(`Cannot find an asset for ${PLATFORM} - ${ARCH}`);

  try {
    console.log("Downloading binaries from github release...")
    const res = await fetch(asset.browser_download_url);
    fs.mkdirSync(path.resolve(__dirname, "./bin"), { recursive: true });

    const filepath = path.resolve(__dirname, "./bin", asset.name);
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
      console.log("Extracting file...")
      await extractFile(filepath)
      fs.unlinkSync(filepath);


      console.log("Square Cloud CLI successfuly installed, please perform \"squarecloud --help\"")
    });
  } catch (err) {
    console.log("Error when trying to download and extract the file", err);
  }
}

(async () => {
  const argv = process.argv;
  if (argv[2] === "update") {
    installBinaries();
    return;
  }

  const binDir = path.resolve(__dirname, "./bin");
  const execfile = path.resolve(binDir, "squarecloud");

  if (!fs.existsSync(execfile)) {
    await installBinaries();
    console.log("Square Cloud CLI successfuly installed, please perform the command again.")
    return
  }

  stdout.write(execFileSync(execfile, argv.slice(2)));
})();