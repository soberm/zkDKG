import "hardhat-abi-exporter";
import "./contracts/plugins/abigen-exporter";
import "./contracts/plugins/deploy";
import "./contracts/plugins/launcher";

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
