const { ethers, upgrades } = require("hardhat");
const fs = require("fs");
const path = require("path");

async function main() {
  console.log("🚀 Starting Developer DAO Smart Contract Deployment...\n");

  const [deployer] = await ethers.getSigners();
  const network = await ethers.provider.getNetwork();
  
  console.log("📋 Deployment Details:");
  console.log("- Network:", network.name, `(Chain ID: ${network.chainId})`);
  console.log("- Deployer:", deployer.address);
  console.log("- Balance:", ethers.utils.formatEther(await deployer.getBalance()), "ETH\n");

  // Deployment configuration
  const config = {
    // Coffee Token address (should be deployed already in the Go Coffee ecosystem)
    coffeeTokenAddress: process.env.COFFEE_TOKEN_ADDRESS || "0x0000000000000000000000000000000000000000",
    
    // DAO configuration
    votingDelay: 7200, // 1 day in blocks (assuming 12s block time)
    votingPeriod: 50400, // 1 week in blocks
    proposalThreshold: ethers.utils.parseEther("10000"), // 10,000 COFFEE tokens
    quorumPercentage: 4, // 4% of total supply
    timelockDelay: 172800, // 2 days in seconds
    
    // Initial addresses
    treasury: deployer.address, // Will be updated to multisig later
    metricsOracle: deployer.address, // Will be updated to oracle service
  };

  console.log("⚙️  Configuration:");
  console.log("- Coffee Token:", config.coffeeTokenAddress);
  console.log("- Voting Delay:", config.votingDelay, "blocks");
  console.log("- Voting Period:", config.votingPeriod, "blocks");
  console.log("- Proposal Threshold:", ethers.utils.formatEther(config.proposalThreshold), "COFFEE");
  console.log("- Quorum:", config.quorumPercentage + "%");
  console.log("- Timelock Delay:", config.timelockDelay, "seconds\n");

  const deployedContracts = {};

  try {
    // 1. Deploy TimelockController first (required for Governor)
    console.log("1️⃣  Deploying TimelockController...");
    const TimelockController = await ethers.getContractFactory("TimelockController");
    const timelock = await TimelockController.deploy(
      config.timelockDelay,
      [], // proposers (will be set to governor)
      [], // executors (will be set to governor)
      deployer.address // admin (will be renounced later)
    );
    await timelock.deployed();
    deployedContracts.timelock = timelock.address;
    console.log("✅ TimelockController deployed to:", timelock.address);

    // 2. Deploy DAOGovernor
    console.log("\n2️⃣  Deploying DAOGovernor...");
    const DAOGovernor = await ethers.getContractFactory("DAOGovernor");
    const governor = await DAOGovernor.deploy(
      config.coffeeTokenAddress,
      timelock.address,
      config.quorumPercentage
    );
    await governor.deployed();
    deployedContracts.governor = governor.address;
    console.log("✅ DAOGovernor deployed to:", governor.address);

    // 3. Configure TimelockController roles
    console.log("\n3️⃣  Configuring TimelockController roles...");
    const PROPOSER_ROLE = await timelock.PROPOSER_ROLE();
    const EXECUTOR_ROLE = await timelock.EXECUTOR_ROLE();
    const TIMELOCK_ADMIN_ROLE = await timelock.TIMELOCK_ADMIN_ROLE();

    // Grant proposer and executor roles to governor
    await timelock.grantRole(PROPOSER_ROLE, governor.address);
    await timelock.grantRole(EXECUTOR_ROLE, governor.address);
    
    // Allow anyone to execute (common pattern)
    await timelock.grantRole(EXECUTOR_ROLE, ethers.constants.AddressZero);
    
    // Renounce admin role (governor will be the admin through proposals)
    await timelock.renounceRole(TIMELOCK_ADMIN_ROLE, deployer.address);
    console.log("✅ TimelockController roles configured");

    // 4. Deploy BountyManager
    console.log("\n4️⃣  Deploying BountyManager...");
    const BountyManager = await ethers.getContractFactory("BountyManager");
    const bountyManager = await BountyManager.deploy(
      config.coffeeTokenAddress,
      config.metricsOracle
    );
    await bountyManager.deployed();
    deployedContracts.bountyManager = bountyManager.address;
    console.log("✅ BountyManager deployed to:", bountyManager.address);

    // 5. Deploy RevenueSharing
    console.log("\n5️⃣  Deploying RevenueSharing...");
    const RevenueSharing = await ethers.getContractFactory("RevenueSharing");
    const revenueSharing = await RevenueSharing.deploy(
      config.coffeeTokenAddress,
      config.metricsOracle,
      config.treasury
    );
    await revenueSharing.deployed();
    deployedContracts.revenueSharing = revenueSharing.address;
    console.log("✅ RevenueSharing deployed to:", revenueSharing.address);

    // 6. Deploy SolutionRegistry
    console.log("\n6️⃣  Deploying SolutionRegistry...");
    const SolutionRegistry = await ethers.getContractFactory("SolutionRegistry");
    const solutionRegistry = await SolutionRegistry.deploy();
    await solutionRegistry.deployed();
    deployedContracts.solutionRegistry = solutionRegistry.address;
    console.log("✅ SolutionRegistry deployed to:", solutionRegistry.address);

    // 7. Configure initial settings
    console.log("\n7️⃣  Configuring initial settings...");
    
    // Add deployer as initial reviewer for SolutionRegistry
    await solutionRegistry.addReviewer(deployer.address);
    console.log("✅ Added initial reviewer to SolutionRegistry");

    // Transfer ownership of contracts to timelock (DAO control)
    if (await bountyManager.owner() === deployer.address) {
      await bountyManager.transferOwnership(timelock.address);
      console.log("✅ BountyManager ownership transferred to Timelock");
    }

    if (await revenueSharing.owner() === deployer.address) {
      await revenueSharing.transferOwnership(timelock.address);
      console.log("✅ RevenueSharing ownership transferred to Timelock");
    }

    if (await solutionRegistry.owner() === deployer.address) {
      await solutionRegistry.transferOwnership(timelock.address);
      console.log("✅ SolutionRegistry ownership transferred to Timelock");
    }

    // 8. Save deployment information
    console.log("\n8️⃣  Saving deployment information...");
    const deploymentInfo = {
      network: network.name,
      chainId: network.chainId,
      deployer: deployer.address,
      timestamp: new Date().toISOString(),
      blockNumber: await ethers.provider.getBlockNumber(),
      contracts: deployedContracts,
      config: config,
      gasUsed: {
        // Will be populated by transaction receipts
      }
    };

    // Calculate total gas used
    let totalGasUsed = ethers.BigNumber.from(0);
    for (const [name, address] of Object.entries(deployedContracts)) {
      try {
        const contract = await ethers.getContractAt("IERC165", address);
        const deployTx = contract.deployTransaction;
        if (deployTx) {
          const receipt = await deployTx.wait();
          deploymentInfo.gasUsed[name] = receipt.gasUsed.toString();
          totalGasUsed = totalGasUsed.add(receipt.gasUsed);
        }
      } catch (error) {
        console.log(`⚠️  Could not get gas usage for ${name}`);
      }
    }
    deploymentInfo.gasUsed.total = totalGasUsed.toString();

    // Save to file
    const deploymentsDir = path.join(__dirname, "../deployments");
    if (!fs.existsSync(deploymentsDir)) {
      fs.mkdirSync(deploymentsDir, { recursive: true });
    }

    const filename = `deployment-${network.name}-${Date.now()}.json`;
    const filepath = path.join(deploymentsDir, filename);
    fs.writeFileSync(filepath, JSON.stringify(deploymentInfo, null, 2));

    // Also save as latest deployment
    const latestPath = path.join(deploymentsDir, `latest-${network.name}.json`);
    fs.writeFileSync(latestPath, JSON.stringify(deploymentInfo, null, 2));

    console.log("✅ Deployment information saved to:", filename);

    // 9. Display summary
    console.log("\n🎉 Deployment Complete!");
    console.log("=" .repeat(60));
    console.log("📋 Contract Addresses:");
    console.log("- TimelockController:", deployedContracts.timelock);
    console.log("- DAOGovernor:", deployedContracts.governor);
    console.log("- BountyManager:", deployedContracts.bountyManager);
    console.log("- RevenueSharing:", deployedContracts.revenueSharing);
    console.log("- SolutionRegistry:", deployedContracts.solutionRegistry);
    console.log("\n⛽ Gas Usage:");
    console.log("- Total Gas Used:", ethers.utils.formatUnits(totalGasUsed, "gwei"), "Gwei");
    console.log("\n📄 Next Steps:");
    console.log("1. Update config.yaml with contract addresses");
    console.log("2. Verify contracts on Etherscan");
    console.log("3. Set up monitoring and alerts");
    console.log("4. Configure metrics oracle");
    console.log("5. Test governance workflow");

    return deployedContracts;

  } catch (error) {
    console.error("\n❌ Deployment failed:", error);
    throw error;
  }
}

// Execute deployment
if (require.main === module) {
  main()
    .then(() => process.exit(0))
    .catch((error) => {
      console.error(error);
      process.exit(1);
    });
}

module.exports = main;
