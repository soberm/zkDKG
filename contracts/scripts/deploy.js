const hre = require("hardhat");

async function main() {
  const ZKDKG = await hre.ethers.getContractFactory("ZKDKG");
  const zkDKG = await ZKDKG.deploy();

  await zkDKG.deployed();

  console.log("zkDKG deployed to:", zkDKG.address);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
