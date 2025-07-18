// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Votes.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title MockCoffeeToken
 * @dev Mock Coffee Token for testing with voting capabilities
 */
contract MockCoffeeToken is ERC20, ERC20Votes, ERC20Burnable, Ownable {

    constructor()
        ERC20("Mock Coffee Token", "MCOFFEE")
        ERC20Permit("Mock Coffee Token")
        Ownable(msg.sender)
    {
        // Mint initial supply to deployer
        _mint(msg.sender, 1000000000 * 10**18); // 1 billion tokens
    }

    /**
     * @dev Mint tokens to an address (for testing)
     */
    function mint(address to, uint256 amount) external onlyOwner {
        _mint(to, amount);
    }

    /**
     * @dev Burn tokens from an address (for testing)
     */
    function burnFrom(address account, uint256 amount) public override {
        super.burnFrom(account, amount);
    }

    // Required override for ERC20Votes
    function _update(
        address from,
        address to,
        uint256 amount
    ) internal override(ERC20, ERC20Votes) {
        super._update(from, to, amount);
    }

    function nonces(address owner) public view virtual override(ERC20Permit, Nonces) returns (uint256) {
        return super.nonces(owner);
    }
}
