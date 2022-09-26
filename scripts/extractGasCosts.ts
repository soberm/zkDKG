import fs from "fs/promises";
import {constants} from "fs";
import events from "events";
import readline from "readline";

const args = process.argv.slice(2);
if (args.some(arg => typeof arg === "undefined")) {
    process.exit(1);
}

const [pipePath, cAdvisorLog, hardhatLog, repetitions] = args;

const dockerStats: readonly [number[], number[]][] = [[[], []], [[], []]];
const gasCosts: Map<string, number[]> = new Map();

(async() => {
    for (let repetition = 0; repetition < Number(repetitions); repetition++) {
        
        let containerIndex = 0;

        const pipe = (await (fs.open(pipePath, constants.O_RDONLY))).createReadStream();
        pipe.on("data", async dockerId => {
        
            const rl = readline.createInterface({
                "input": (await fs.open(cAdvisorLog)).createReadStream(),
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
    
            const [runtimes, memUsages] = dockerStats[containerIndex++];
            runtimes.push(time);
            memUsages.push(Math.round(maxMemUsage / (10 ** 6)));
        });

        await events.once(pipe, "close");
        
        const data = (await fs.readFile(hardhatLog)).toString();
        const regex = /eth_sendRawTransaction.*?Contract call:.*?#(.*?)$.*?Gas used:\s*(\d+)/gms;
        
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
    }
    
    console.log("method,avggascosts,memory_in_mb,time_in_s");
    
    for (const [method, costs] of gasCosts) {
        let memory = 0;
        let time = 0;

        if (method === "defendShare") {
            [time, memory] = getStats(0);
        } else if (method === "submitPublicKey") {
            [time, memory] = getStats(1);
        }

        console.log(`${method},${average(costs)},${memory},${time}`);
    }
})();

function getStats(index: number): [number, number] {
    const [runTimes, memUsages] = dockerStats[index];
    return [average(runTimes), average(memUsages)];
}

function average(arr: number[]): number {
    return Math.round(arr.reduce((p, c) => p + c, 0) / arr.length);
}
