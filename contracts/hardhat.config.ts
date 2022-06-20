import "hardhat-abi-exporter";
import "./plugins/abigen-exporter";
import "./plugins/deploy";
import "./plugins/launcher";

import type {HardhatUserConfig} from "hardhat/config";

const config: HardhatUserConfig = {
  abiExporter: {
    runOnCompile: true,
    flat: true,
    only: ["ZKDKG"],
  },
  solidity: "0.8.4",
};

export default config;
