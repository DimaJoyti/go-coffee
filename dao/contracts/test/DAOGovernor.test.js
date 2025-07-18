const { expect } = require("chai");
const { ethers } = require("hardhat");
const { time, loadFixture } = require("@nomicfoundation/hardhat-network-helpers");

describe("DAOGovernor", function () {
  // Fixture for deploying contracts
  async function deployGovernorFixture() {
    const [owner, proposer, voter1, voter2, voter3] = await ethers.getSigners();

    // Deploy mock Coffee Token (ERC20Votes)
    const MockCoffeeToken = await ethers.getContractFactory("MockCoffeeToken");
    const coffeeToken = await MockCoffeeToken.deploy();
    await coffeeToken.deployed();

    // Mint tokens to voters
    const mintAmount = ethers.utils.parseEther("100000"); // 100k tokens each
    await coffeeToken.mint(proposer.address, mintAmount);
    await coffeeToken.mint(voter1.address, mintAmount);
    await coffeeToken.mint(voter2.address, mintAmount);
    await coffeeToken.mint(voter3.address, mintAmount);

    // Delegate voting power to themselves
    await coffeeToken.connect(proposer).delegate(proposer.address);
    await coffeeToken.connect(voter1).delegate(voter1.address);
    await coffeeToken.connect(voter2).delegate(voter2.address);
    await coffeeToken.connect(voter3).delegate(voter3.address);

    // Deploy TimelockController
    const TimelockController = await ethers.getContractFactory("TimelockController");
    const timelock = await TimelockController.deploy(
      172800, // 2 days
      [], // proposers
      [], // executors
      owner.address // admin
    );
    await timelock.deployed();

    // Deploy DAOGovernor
    const DAOGovernor = await ethers.getContractFactory("DAOGovernor");
    const governor = await DAOGovernor.deploy(
      coffeeToken.address,
      timelock.address,
      4 // 4% quorum
    );
    await governor.deployed();

    // Configure timelock roles
    const PROPOSER_ROLE = await timelock.PROPOSER_ROLE();
    const EXECUTOR_ROLE = await timelock.EXECUTOR_ROLE();
    const TIMELOCK_ADMIN_ROLE = await timelock.TIMELOCK_ADMIN_ROLE();

    await timelock.grantRole(PROPOSER_ROLE, governor.address);
    await timelock.grantRole(EXECUTOR_ROLE, governor.address);
    await timelock.grantRole(EXECUTOR_ROLE, ethers.constants.AddressZero);
    await timelock.renounceRole(TIMELOCK_ADMIN_ROLE, owner.address);

    return {
      governor,
      coffeeToken,
      timelock,
      owner,
      proposer,
      voter1,
      voter2,
      voter3,
    };
  }

  describe("Deployment", function () {
    it("Should deploy with correct initial settings", async function () {
      const { governor, coffeeToken, timelock } = await loadFixture(deployGovernorFixture);

      expect(await governor.name()).to.equal("Developer DAO Governor");
      expect(await governor.token()).to.equal(coffeeToken.address);
      expect(await governor.timelock()).to.equal(timelock.address);
      expect(await governor.votingDelay()).to.equal(7200); // 1 day
      expect(await governor.votingPeriod()).to.equal(50400); // 1 week
      expect(await governor.proposalThreshold()).to.equal(ethers.utils.parseEther("10000"));
      expect(await governor.quorum(await ethers.provider.getBlockNumber())).to.be.gt(0);
    });
  });

  describe("Proposal Creation", function () {
    it("Should allow creating a proposal with sufficient tokens", async function () {
      const { governor, proposer } = await loadFixture(deployGovernorFixture);

      const targets = [ethers.constants.AddressZero];
      const values = [0];
      const calldatas = ["0x"];
      const description = "Test Proposal";

      await expect(
        governor.connect(proposer).proposeWithInfo(
          targets,
          values,
          calldatas,
          description,
          0, // GENERAL category
          "Test Proposal Title",
          Math.floor(Date.now() / 1000) + 86400 // 1 day from now
        )
      ).to.emit(governor, "ProposalCreatedWithInfo");
    });

    it("Should reject proposal from address with insufficient tokens", async function () {
      const { governor, voter1 } = await loadFixture(deployGovernorFixture);

      // Transfer away most tokens to have less than threshold
      const coffeeToken = await ethers.getContractAt("MockCoffeeToken", await governor.token());
      await coffeeToken.connect(voter1).transfer(
        ethers.constants.AddressZero,
        ethers.utils.parseEther("95000")
      );

      const targets = [ethers.constants.AddressZero];
      const values = [0];
      const calldatas = ["0x"];
      const description = "Test Proposal";

      await expect(
        governor.connect(voter1).proposeWithInfo(
          targets,
          values,
          calldatas,
          description,
          0,
          "Test Proposal Title",
          Math.floor(Date.now() / 1000) + 86400
        )
      ).to.be.revertedWith("Insufficient tokens to create proposal");
    });

    it("Should store proposal information correctly", async function () {
      const { governor, proposer } = await loadFixture(deployGovernorFixture);

      const targets = [ethers.constants.AddressZero];
      const values = [0];
      const calldatas = ["0x"];
      const description = "Test Proposal Description";
      const title = "Test Proposal Title";
      const category = 1; // BOUNTY
      const deadline = Math.floor(Date.now() / 1000) + 86400;

      const tx = await governor.connect(proposer).proposeWithInfo(
        targets,
        values,
        calldatas,
        description,
        category,
        title,
        deadline
      );

      const receipt = await tx.wait();
      const event = receipt.events.find(e => e.event === "ProposalCreatedWithInfo");
      const proposalId = event.args.proposalId;

      const proposalInfo = await governor.getProposalInfo(proposalId);
      expect(proposalInfo.category).to.equal(category);
      expect(proposalInfo.title).to.equal(title);
      expect(proposalInfo.description).to.equal(description);
      expect(proposalInfo.proposer).to.equal(proposer.address);
    });
  });

  describe("Voting", function () {
    async function createProposal(governor, proposer) {
      const targets = [ethers.constants.AddressZero];
      const values = [0];
      const calldatas = ["0x"];
      const description = "Test Proposal";

      const tx = await governor.connect(proposer).proposeWithInfo(
        targets,
        values,
        calldatas,
        description,
        0,
        "Test Proposal Title",
        Math.floor(Date.now() / 1000) + 86400
      );

      const receipt = await tx.wait();
      const event = receipt.events.find(e => e.event === "ProposalCreatedWithInfo");
      return event.args.proposalId;
    }

    it("Should allow voting on active proposals", async function () {
      const { governor, proposer, voter1 } = await loadFixture(deployGovernorFixture);

      const proposalId = await createProposal(governor, proposer);

      // Wait for voting delay
      await time.increase(7201); // 1 day + 1 second

      // Vote FOR
      await expect(governor.connect(voter1).castVote(proposalId, 1))
        .to.emit(governor, "VoteCast");

      // Check if vote was recorded
      expect(await governor.hasVoted(proposalId, voter1.address)).to.be.true;
    });

    it("Should allow voting with reason", async function () {
      const { governor, proposer, voter1 } = await loadFixture(deployGovernorFixture);

      const proposalId = await createProposal(governor, proposer);
      await time.increase(7201);

      const reason = "I support this proposal because...";
      await expect(
        governor.connect(voter1).castVoteWithReason(proposalId, 1, reason)
      ).to.emit(governor, "VoteCast");
    });

    it("Should prevent double voting", async function () {
      const { governor, proposer, voter1 } = await loadFixture(deployGovernorFixture);

      const proposalId = await createProposal(governor, proposer);
      await time.increase(7201);

      // First vote
      await governor.connect(voter1).castVote(proposalId, 1);

      // Second vote should fail
      await expect(governor.connect(voter1).castVote(proposalId, 0))
        .to.be.revertedWith("GovernorVotingSimple: vote already cast");
    });

    it("Should calculate voting power correctly", async function () {
      const { governor, voter1 } = await loadFixture(deployGovernorFixture);

      const votingPower = await governor.getVotes(voter1.address, await ethers.provider.getBlockNumber() - 1);
      expect(votingPower).to.equal(ethers.utils.parseEther("100000"));
    });
  });

  describe("Proposal Execution", function () {
    it("Should execute successful proposals after timelock delay", async function () {
      const { governor, proposer, voter1, voter2, voter3 } = await loadFixture(deployGovernorFixture);

      // Create a proposal that transfers tokens from timelock
      const targets = [await governor.token()];
      const values = [0];
      const transferAmount = ethers.utils.parseEther("1000");
      const iface = new ethers.utils.Interface(["function transfer(address,uint256)"]);
      const calldatas = [iface.encodeFunctionData("transfer", [proposer.address, transferAmount])];
      const description = "Transfer tokens from treasury";

      const tx = await governor.connect(proposer).proposeWithInfo(
        targets,
        values,
        calldatas,
        description,
        2, // TREASURY category
        "Treasury Transfer",
        Math.floor(Date.now() / 1000) + 86400
      );

      const receipt = await tx.wait();
      const event = receipt.events.find(e => e.event === "ProposalCreatedWithInfo");
      const proposalId = event.args.proposalId;

      // Wait for voting delay
      await time.increase(7201);

      // Vote (need majority)
      await governor.connect(voter1).castVote(proposalId, 1); // FOR
      await governor.connect(voter2).castVote(proposalId, 1); // FOR
      await governor.connect(voter3).castVote(proposalId, 1); // FOR

      // Wait for voting period to end
      await time.increase(50401);

      // Queue the proposal
      const descriptionHash = ethers.utils.keccak256(ethers.utils.toUtf8Bytes(description));
      await governor.queue(targets, values, calldatas, descriptionHash);

      // Wait for timelock delay
      await time.increase(172801); // 2 days + 1 second

      // Execute the proposal
      await expect(governor.execute(targets, values, calldatas, descriptionHash))
        .to.emit(governor, "ProposalExecuted");
    });
  });

  describe("Governance Statistics", function () {
    it("Should track proposal counts correctly", async function () {
      const { governor, proposer } = await loadFixture(deployGovernorFixture);

      const initialCount = await governor.totalProposals();
      
      await createProposal(governor, proposer);
      
      expect(await governor.totalProposals()).to.equal(initialCount.add(1));
      expect(await governor.proposalCount(proposer.address)).to.equal(1);
    });
  });
});
