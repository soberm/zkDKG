const fs = require("fs");
const path = require("path");

function exportABI(artifactsPath, contract) {
    const contractArtifacts = path.resolve(
        __dirname,
        artifactsPath,
        `${contract}.sol`
    );

    const contractPath = path.resolve(contractArtifacts, `${contract}.json`);
    const contractAbiPath = path.resolve(contractArtifacts, `${contract}.abi`);
    const contractBinPath = path.resolve(contractArtifacts, `${contract}.bin`);

    const contractFile = JSON.parse(fs.readFileSync(contractPath, "utf8"));

    // eslint-disable-next-line consistent-return
    fs.writeFile(contractAbiPath, JSON.stringify(contractFile.abi), (err) => {
        if (err) {
            return console.error(err);
        }
        console.log(`${contract} ABI written successfully!`);
    });

    fs.writeFile(
        contractBinPath,
        JSON.stringify(contractFile.bytecode),
        (err) => {
            if (err) {
                return console.error(err);
            }
            console.log(`${contract} BIN written successfully!`);
        }
    );
}

exportABI("../artifacts/contracts", "ZKDKG");