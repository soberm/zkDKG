import {task, types} from "hardhat/config";
import "@nomiclabs/hardhat-ethers";

task("deploy", "Deploy the ZKDKG contract(s)")
    .addPositionalParam("participants", "the number of participants for the distributed key generation", undefined, types.int, false)
    .setAction(async ({participants}, env, _) => {
        await env.run("compile");

        const KEYVERIFIER = await env.ethers.getContractFactory("KeyVerifier");
        const keyVerifier = await KEYVERIFIER.deploy();

        await keyVerifier.deployed();
        console.log("KeyVerifier deployed to:", keyVerifier.address);

        const SHAREVERIFIER = await env.ethers.getContractFactory("ShareVerifier");
        const shareVerifier = await SHAREVERIFIER.deploy();

        await shareVerifier.deployed();
        console.log("ShareVerifier deployed to:", shareVerifier.address);

        const ZKDKG = await env.ethers.getContractFactory("ZKDKG");

        const zkDKG = await ZKDKG.deploy(
            shareVerifier.address,
            keyVerifier.address,
            participants,
            Math.floor(2 / 3 * (participants + 1)),
            0,
        );

        await zkDKG.deployed();

        console.log("zkDKG deployed to:", zkDKG.address);
    });
