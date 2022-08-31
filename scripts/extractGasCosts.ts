const fs = require("fs");

if (typeof process.argv[2] === "undefined") {
    process.exit(1);
}

const data = fs.readFileSync(process.argv[2]).toString();
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

console.log("method,avggascosts");

for (const [method, costs] of map) {
    const average = costs.reduce((p, c) => p + c, 0) / costs.length;
    console.log(`${method},${average}`);
}
