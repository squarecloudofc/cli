import { installBinaries } from "./lib.js";

// A failed postinstall must fail `npm install`, not leave a CLI that cannot run.
installBinaries().catch((err) => {
  console.error(`Square Cloud CLI installation failed: ${err.message}`);
  console.error("Check your connection and run npm install -g @squarecloud/cli again.");
  process.exit(1);
});
