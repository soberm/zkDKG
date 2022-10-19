import fsProm from "fs/promises";
import fs from "fs";
import path from "path";
import events from "events";
import readline from "readline";
import {EOL} from "os";

const args = process.argv.slice(2);
if (args.some(arg => typeof arg === "undefined")) {
    process.exit(1);
}

const [generateOnlyStr, participants, repetitions] = args;

const generateOnly = generateOnlyStr === "true";

const dir = path.resolve(__dirname, `../build/${participants}`);

(async() => {
    const csv = fs.createWriteStream(path.resolve(dir, "report.csv"));
    const columns = ["run"];

    if (!generateOnly) {
        columns.push("gas_share", "gas_dispute", "gas_justify", "gas_derive");
    }

    columns.push("time_in_s_justify", "memory_in_mb_justify", "time_in_s_derive", "memory_in_mb_derive");

    csv.write(`${columns.join(",")}${EOL}`);

    for (let repetition = 1; repetition <= Number(repetitions); repetition++) {

        let containerIndex = 0;
        const dockerStats: [number, number][] = [];

        const pipe = (await (fsProm.open(path.resolve(dir, "container_pipe"), fs.constants.O_RDONLY))).createReadStream();

        await new Promise<void>(resolve => {
            pipe.on("data", async dockerId => {
                const rl = readline.createInterface({
                    "input": (await fsProm.open(path.resolve(dir, "../cadvisor.log"))).createReadStream(),
                });
                const regex = new RegExp(`^cName=${dockerId.toString().trim()}(?=.*timestamp=(\\d+))(?=.*memory_usage=(\\d+))`);
                
                let startTimestamp = "0";
                let endTimestamp = "0";
                let maxMemUsage = 0;
        
                rl.on("line", line => {
                    const result = regex.exec(line);
                    if (result !== null) {
                        const [, timestamp, memUsage] = result;
            
                        if (startTimestamp === "0") {
                            startTimestamp = timestamp;
                        }
            
                        endTimestamp = timestamp;
            
                        maxMemUsage = Math.max(maxMemUsage, Number(memUsage));
                    }
                });
        
                await events.once(rl, "close");
        
                const time = Number((BigInt(endTimestamp) - BigInt(startTimestamp)) / (10n ** 9n));
        
                dockerStats[containerIndex++] = [time, Math.round(maxMemUsage / (10 ** 6))];

                // We collect data from 2 containers
                if (containerIndex == 2) {
                    resolve();
                }
            });
        });
        
        const values = [repetition];

        if (!generateOnly) {
            const data = (await fsProm.readFile(path.resolve(dir, "hardhat.log"))).toString();
            const regex = /eth_sendRawTransaction.*?Contract call:.*?#(.*?)$.*?Gas used:\s*(\d+)/gms;
            
            const gasCosts: Map<string, number[]> = new Map();

            let result;
            while ((result = regex.exec(data)) !== null) {
                const method = result[1];
                const costs = Number(result[2]);
                const entry = gasCosts.get(method);
            
                if (typeof entry !== "undefined") {
                    entry.push(costs);
                } else {
                    gasCosts.set(method, [costs]);
                }
            }

            const averages = ["broadcastShares", "disputeShare", "defendShare", "submitPublicKey"]
                .map(method => average(gasCosts.get(method)))

            values.push(...averages);
        }

        values.push(...dockerStats[0], ...dockerStats[1]);

        csv.write(`${values.join(",")}${EOL}`);
    }
})();

function average(arr: number[] = []): number {
    return Math.round(arr.reduce((p, c) => p + c, 0) / arr.length);
}
