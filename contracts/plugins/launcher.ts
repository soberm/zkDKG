import {task, types} from "hardhat/config";
import {HardhatPluginError} from "hardhat/plugins";
import type {HardhatNetworkAccountsConfig, HardhatNetworkHDAccountsConfig} from "hardhat/types";

const PLUGIN_NAME = "launcher";

task("launch", "Start the local Hardhat network with a given amount of accounts")
    .addPositionalParam("accounts", "the number of accounts to generate", undefined, types.int, false)
    .setAction(async (args, env, _) => {
        const accounts = env.config.networks.hardhat.accounts;
        if (!isHdConfig(accounts)) {
            throw new HardhatPluginError(PLUGIN_NAME, "Expected HD wallet configuration");
        }

        accounts.count = args.accounts;

        return env.run("node");
    })

function isHdConfig(config: HardhatNetworkAccountsConfig): config is HardhatNetworkHDAccountsConfig {
    return "count" in config;
}
