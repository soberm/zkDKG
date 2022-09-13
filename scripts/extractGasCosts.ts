import fs from "fs/promises";
import {constants} from "fs";
import events from "events";
import net from "net";
import readline from "readline";

(async() => {
    if (typeof process.argv[2] === "undefined" || typeof process.argv[3] === "undefined" || typeof process.argv[4] === "undefined") {
        process.exit(1);
    }
    
    const results: [bigint, number][] = [];
    const fd = await fs.open(process.argv[2], constants.O_RDONLY | constants.O_NONBLOCK);
    const pipe = new net.Socket({"fd": fd.fd});
    pipe.on("data", async dockerId => {
    
        const rl = readline.createInterface({
            "input": (await fs.open(process.argv[3])).createReadStream(),
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

        const time = (BigInt(endTimestamp) - BigInt(startTimestamp)) / (10n ** 9n);

        results.push([time, Math.round(maxMemUsage / (10 ** 6))]);
    });

    await events.once(pipe, "close");
    
    const data = (await fs.readFile(process.argv[4])).toString();
    const regex = /eth_sendRawTransaction.*?Contract call:.*?#(.*?)$.*?Gas used:\s*(\d+)/gms;
    
    const map: Map<string, number[]> = new Map();
    let result;
    while ((result = regex.exec(data)) !== null) {
        const method = result[1];
        const costs = Number(result[2]);
        const entry = map.get(method);
    
        if (typeof entry !== "undefined") {
            entry.push(costs);
        } else {
            map.set(method, [costs]);
        }
    }
    
    console.log("method,avggascosts,memory_in_mb,time_in_s");
    
    for (const [method, costs] of map) {
        const average = Math.round(costs.reduce((p, c) => p + c, 0) / costs.length);
        let memory = 0;
        let time = 0n;

        if (method === "defendShare") {
            time = results[0][0];
            memory = results[0][1];
        } else if (method === "submitPublicKey") {
            time = results[1][0];
            memory = results[1][1];
        }

        console.log(`${method},${average},${memory},${time}`);
    }
})();
