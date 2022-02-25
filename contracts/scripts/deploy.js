const hre = require("hardhat");

async function main() {
  const KEYVERIFIER = await hre.ethers.getContractFactory("KeyVerifier");
  const keyVerifier = await KEYVERIFIER.deploy();

  await keyVerifier.deployed();
  console.log("KeyVerifier deployed to:", keyVerifier.address);

  const SHAREVERIFIER = await hre.ethers.getContractFactory("ShareVerifier");
  const shareVerifier = await SHAREVERIFIER.deploy();

  await shareVerifier.deployed();
  console.log("ShareVerifier deployed to:", shareVerifier.address);


  const ZKDKG = await hre.ethers.getContractFactory("ZKDKG");
  const zkDKG = await ZKDKG.deploy(shareVerifier.address, keyVerifier.address, 3);

  await zkDKG.deployed();

  console.log("zkDKG deployed to:", zkDKG.address);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
