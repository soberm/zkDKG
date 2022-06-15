import {subtask} from "hardhat/config";
import {HardhatPluginError} from "hardhat/plugins";

import {execFileSync} from "child_process";
import path from "path";

const PLUGIN_NAME = "abigen-exporter";

subtask("export-abi-group").setAction(async (args, _, runSuper) => {
    await runSuper(args);

    const script = path.join(__dirname, "../../dkg/scripts/abigen.sh");

    try {
        execFileSync(script);
    } catch (err) {
        throw new HardhatPluginError(PLUGIN_NAME, "abigen script failed", err as Error);
    }
});
