// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract ERC20Token is ERC20 {
    address private owner;

    constructor(string memory name, string memory symbol, uint256 initialSupply) ERC20(name, symbol) {
        _mint(msg.sender, initialSupply);
        owner = msg.sender;
    }

    function private_mint(address to, uint256 amount) private {
        _mint(to, amount);
    }

    function public_mint(address to, uint256 amount) public {
        require(msg.sender == owner, "Only the owner can call public_mint");
        private_mint(to, amount);
    }
}
