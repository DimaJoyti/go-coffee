const { expect } = require("chai");
const { ethers } = require("hardhat");
const { time, loadFixture } = require("@nomicfoundation/hardhat-network-helpers");

describe("BountyManager", function () {
  async function deployBountyManagerFixture() {
    const [owner, creator, developer1, developer2, oracle] = await ethers.getSigners();

    // Deploy mock Coffee Token
    const MockCoffeeToken = await ethers.getContractFactory("MockCoffeeToken");
    const coffeeToken = await MockCoffeeToken.deploy();
    await coffeeToken.deployed();

    // Deploy BountyManager
    const BountyManager = await ethers.getContractFactory("BountyManager");
    const bountyManager = await BountyManager.deploy(
      coffeeToken.address,
      oracle.address
    );
    await bountyManager.deployed();

    // Mint tokens to creator and approve bounty manager
    const mintAmount = ethers.utils.parseEther("1000000"); // 1M tokens
    await coffeeToken.mint(creator.address, mintAmount);
    await coffeeToken.connect(creator).approve(bountyManager.address, mintAmount);

    return {
      bountyManager,
      coffeeToken,
      owner,
      creator,
      developer1,
      developer2,
      oracle,
    };
  }

  describe("Deployment", function () {
    it("Should deploy with correct initial settings", async function () {
      const { bountyManager, coffeeToken, oracle } = await loadFixture(deployBountyManagerFixture);

      expect(await bountyManager.coffeeToken()).to.equal(coffeeToken.address);
      expect(await bountyManager.metricsOracle()).to.equal(oracle.address);
      expect(await bountyManager.MIN_BOUNTY_REWARD()).to.equal(ethers.utils.parseEther("100"));
    });
  });

  describe("Bounty Creation", function () {
    it("Should create a bounty with milestones", async function () {
      const { bountyManager, creator } = await loadFixture(deployBountyManagerFixture);

      const title = "Implement DEX Integration";
      const description = "Integrate with Uniswap V3 for token swaps";
      const category = 0; // TVL_GROWTH
      const totalReward = ethers.utils.parseEther("1000");
      const deadline = Math.floor(Date.now() / 1000) + 86400 * 30; // 30 days

      const milestoneDescriptions = ["Design Phase", "Implementation", "Testing"];
      const milestoneRewards = [
        ethers.utils.parseEther("300"),
        ethers.utils.parseEther("500"),
        ethers.utils.parseEther("200")
      ];
      const milestoneDeadlines = [
        Math.floor(Date.now() / 1000) + 86400 * 7,  // 7 days
        Math.floor(Date.now() / 1000) + 86400 * 21, // 21 days
        Math.floor(Date.now() / 1000) + 86400 * 30  // 30 days
      ];

      await expect(
        bountyManager.connect(creator).createBounty(
          title,
          description,
          category,
          totalReward,
          deadline,
          milestoneDescriptions,
          milestoneRewards,
          milestoneDeadlines
        )
      ).to.emit(bountyManager, "BountyCreated");

      const bountyId = 1;
      const bounty = await bountyManager.getBounty(bountyId);
      expect(bounty.title).to.equal(title);
      expect(bounty.description).to.equal(description);
      expect(bounty.category).to.equal(category);
      expect(bounty.totalReward).to.equal(totalReward);
      expect(bounty.creator).to.equal(creator.address);
    });

    it("Should reject bounty with insufficient reward", async function () {
      const { bountyManager, creator } = await loadFixture(deployBountyManagerFixture);

      const totalReward = ethers.utils.parseEther("50"); // Below minimum
      const deadline = Math.floor(Date.now() / 1000) + 86400;

      await expect(
        bountyManager.connect(creator).createBounty(
          "Test Bounty",
          "Description",
          0,
          totalReward,
          deadline,
          ["Milestone 1"],
          [totalReward],
          [deadline]
        )
      ).to.be.revertedWith("Reward below minimum");
    });

    it("Should reject bounty with mismatched milestone rewards", async function () {
      const { bountyManager, creator } = await loadFixture(deployBountyManagerFixture);

      const totalReward = ethers.utils.parseEther("1000");
      const deadline = Math.floor(Date.now() / 1000) + 86400;

      await expect(
        bountyManager.connect(creator).createBounty(
          "Test Bounty",
          "Description",
          0,
          totalReward,
          deadline,
          ["Milestone 1"],
          [ethers.utils.parseEther("500")], // Doesn't match total
          [deadline]
        )
      ).to.be.revertedWith("Milestone rewards don't match total");
    });
  });

  describe("Bounty Application", function () {
    async function createTestBounty(bountyManager, creator) {
      const totalReward = ethers.utils.parseEther("1000");
      const deadline = Math.floor(Date.now() / 1000) + 86400 * 30;

      await bountyManager.connect(creator).createBounty(
        "Test Bounty",
        "Test Description",
        0, // TVL_GROWTH
        totalReward,
        deadline,
        ["Milestone 1", "Milestone 2"],
        [ethers.utils.parseEther("600"), ethers.utils.parseEther("400")],
        [deadline - 86400 * 10, deadline]
      );

      return 1; // First bounty ID
    }

    it("Should allow developers to apply for bounties", async function () {
      const { bountyManager, creator, developer1 } = await loadFixture(deployBountyManagerFixture);

      const bountyId = await createTestBounty(bountyManager, creator);

      await expect(bountyManager.connect(developer1).applyForBounty(bountyId))
        .to.not.be.reverted;
    });

    it("Should prevent duplicate applications", async function () {
      const { bountyManager, creator, developer1 } = await loadFixture(deployBountyManagerFixture);

      const bountyId = await createTestBounty(bountyManager, creator);

      await bountyManager.connect(developer1).applyForBounty(bountyId);

      await expect(bountyManager.connect(developer1).applyForBounty(bountyId))
        .to.be.revertedWith("Already applied");
    });

    it("Should prevent creator from applying to their own bounty", async function () {
      const { bountyManager, creator } = await loadFixture(deployBountyManagerFixture);

      const bountyId = await createTestBounty(bountyManager, creator);

      await expect(bountyManager.connect(creator).applyForBounty(bountyId))
        .to.be.revertedWith("Creator cannot apply");
    });
  });

  describe("Bounty Assignment", function () {
    async function createAndApplyToBounty(bountyManager, creator, developer) {
      const bountyId = await createTestBounty(bountyManager, creator);
      await bountyManager.connect(developer).applyForBounty(bountyId);
      return bountyId;
    }

    async function createTestBounty(bountyManager, creator) {
      const totalReward = ethers.utils.parseEther("1000");
      const deadline = Math.floor(Date.now() / 1000) + 86400 * 30;

      await bountyManager.connect(creator).createBounty(
        "Test Bounty",
        "Test Description",
        0,
        totalReward,
        deadline,
        ["Milestone 1", "Milestone 2"],
        [ethers.utils.parseEther("600"), ethers.utils.parseEther("400")],
        [deadline - 86400 * 10, deadline]
      );

      return 1;
    }

    it("Should allow creator to assign bounty to applicant", async function () {
      const { bountyManager, creator, developer1 } = await loadFixture(deployBountyManagerFixture);

      const bountyId = await createAndApplyToBounty(bountyManager, creator, developer1);

      await expect(bountyManager.connect(creator).assignBounty(bountyId, developer1.address))
        .to.emit(bountyManager, "BountyAssigned")
        .withArgs(bountyId, developer1.address);

      const bounty = await bountyManager.getBounty(bountyId);
      expect(bounty.assignee).to.equal(developer1.address);
      expect(bounty.status).to.equal(1); // ASSIGNED
    });

    it("Should prevent assignment to non-applicant", async function () {
      const { bountyManager, creator, developer1, developer2 } = await loadFixture(deployBountyManagerFixture);

      const bountyId = await createAndApplyToBounty(bountyManager, creator, developer1);

      await expect(bountyManager.connect(creator).assignBounty(bountyId, developer2.address))
        .to.be.revertedWith("Developer didn't apply");
    });
  });

  describe("Milestone Completion", function () {
    async function createAssignedBounty(bountyManager, creator, developer) {
      const bountyId = await createTestBounty(bountyManager, creator);
      await bountyManager.connect(developer).applyForBounty(bountyId);
      await bountyManager.connect(creator).assignBounty(bountyId, developer.address);
      await bountyManager.connect(developer).startBounty(bountyId);
      return bountyId;
    }

    async function createTestBounty(bountyManager, creator) {
      const totalReward = ethers.utils.parseEther("1000");
      const deadline = Math.floor(Date.now() / 1000) + 86400 * 30;

      await bountyManager.connect(creator).createBounty(
        "Test Bounty",
        "Test Description",
        0,
        totalReward,
        deadline,
        ["Milestone 1", "Milestone 2"],
        [ethers.utils.parseEther("600"), ethers.utils.parseEther("400")],
        [deadline - 86400 * 10, deadline]
      );

      return 1;
    }

    it("Should allow creator to complete milestones", async function () {
      const { bountyManager, coffeeToken, creator, developer1 } = await loadFixture(deployBountyManagerFixture);

      const bountyId = await createAssignedBounty(bountyManager, creator, developer1);
      const initialBalance = await coffeeToken.balanceOf(developer1.address);

      await expect(bountyManager.connect(creator).completeMilestone(bountyId, 0))
        .to.emit(bountyManager, "MilestoneCompleted")
        .withArgs(bountyId, 0, ethers.utils.parseEther("600"));

      const finalBalance = await coffeeToken.balanceOf(developer1.address);
      expect(finalBalance.sub(initialBalance)).to.equal(ethers.utils.parseEther("600"));

      // Check milestone status
      const milestones = await bountyManager.getBountyMilestones(bountyId);
      expect(milestones[0].completed).to.be.true;
      expect(milestones[0].paid).to.be.true;
    });

    it("Should complete bounty when all milestones are done", async function () {
      const { bountyManager, creator, developer1 } = await loadFixture(deployBountyManagerFixture);

      const bountyId = await createAssignedBounty(bountyManager, creator, developer1);

      // Complete first milestone
      await bountyManager.connect(creator).completeMilestone(bountyId, 0);

      // Complete second milestone
      await expect(bountyManager.connect(creator).completeMilestone(bountyId, 1))
        .to.emit(bountyManager, "BountyCompleted");

      const bounty = await bountyManager.getBounty(bountyId);
      expect(bounty.status).to.equal(4); // COMPLETED
    });
  });

  describe("Performance Verification", function () {
    it("Should allow oracle to verify performance", async function () {
      const { bountyManager, creator, developer1, oracle } = await loadFixture(deployBountyManagerFixture);

      const bountyId = await createAssignedBounty(bountyManager, creator, developer1);
      const tvlImpact = ethers.utils.parseEther("1000000"); // $1M TVL
      const mauImpact = 1500; // 1500 MAU

      await expect(
        bountyManager.connect(oracle).verifyPerformance(bountyId, tvlImpact, mauImpact)
      ).to.emit(bountyManager, "PerformanceVerified")
        .withArgs(bountyId, tvlImpact, mauImpact);
    });

    it("Should give bonus for exceptional performance", async function () {
      const { bountyManager, coffeeToken, creator, developer1, oracle } = await loadFixture(deployBountyManagerFixture);

      const bountyId = await createAssignedBounty(bountyManager, creator, developer1);
      const initialBalance = await coffeeToken.balanceOf(developer1.address);
      
      // Verify exceptional performance (>$1M TVL)
      const tvlImpact = ethers.utils.parseEther("2000000"); // $2M TVL
      const mauImpact = 2000; // 2000 MAU

      await bountyManager.connect(oracle).verifyPerformance(bountyId, tvlImpact, mauImpact);

      const finalBalance = await coffeeToken.balanceOf(developer1.address);
      const bonus = ethers.utils.parseEther("100"); // 10% of 1000 total reward
      expect(finalBalance.sub(initialBalance)).to.equal(bonus);
    });

    it("Should prevent non-oracle from verifying performance", async function () {
      const { bountyManager, creator, developer1 } = await loadFixture(deployBountyManagerFixture);

      const bountyId = await createAssignedBounty(bountyManager, creator, developer1);

      await expect(
        bountyManager.connect(creator).verifyPerformance(bountyId, 1000, 100)
      ).to.be.revertedWith("Only metrics oracle");
    });
  });

  describe("Developer Reputation", function () {
    it("Should increase reputation score on milestone completion", async function () {
      const { bountyManager, creator, developer1 } = await loadFixture(deployBountyManagerFixture);

      const bountyId = await createAssignedBounty(bountyManager, creator, developer1);
      const initialReputation = await bountyManager.developerReputationScore(developer1.address);

      await bountyManager.connect(creator).completeMilestone(bountyId, 0);

      const finalReputation = await bountyManager.developerReputationScore(developer1.address);
      expect(finalReputation.sub(initialReputation)).to.equal(10); // REPUTATION_MULTIPLIER
    });
  });

  async function createAssignedBounty(bountyManager, creator, developer) {
    const bountyId = await createTestBounty(bountyManager, creator);
    await bountyManager.connect(developer).applyForBounty(bountyId);
    await bountyManager.connect(creator).assignBounty(bountyId, developer.address);
    await bountyManager.connect(developer).startBounty(bountyId);
    return bountyId;
  }

  async function createTestBounty(bountyManager, creator) {
    const totalReward = ethers.utils.parseEther("1000");
    const deadline = Math.floor(Date.now() / 1000) + 86400 * 30;

    await bountyManager.connect(creator).createBounty(
      "Test Bounty",
      "Test Description",
      0,
      totalReward,
      deadline,
      ["Milestone 1", "Milestone 2"],
      [ethers.utils.parseEther("600"), ethers.utils.parseEther("400")],
      [deadline - 86400 * 10, deadline]
    );

    return 1;
  }
});
