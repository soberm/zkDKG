import "@nomiclabs/hardhat-ethers";
import "hardhat-abi-exporter";
import "./plugins/abigen-exporter";

import type {HardhatUserConfig} from "hardhat/config";

const config: HardhatUserConfig = {
  abiExporter: {
    runOnCompile: true,
    flat: true,
    only: ["ZKDKG"],
  },
  solidity: "0.8.4",
  networks: {
    hardhat: {
      accounts: {
        count: 300,
      },
    },
  },
};

export default config;
