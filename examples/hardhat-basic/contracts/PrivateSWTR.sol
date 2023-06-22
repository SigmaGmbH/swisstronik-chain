// SPDX-License-Identifier: MIT
pragma solidity 0.8.17;

import "./PERC20.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";


contract PrivateSWTR is PERC20 {
    using ECDSA for bytes32;

    constructor() PERC20("PrivateSWTR", "PSWTR") {}

    /// @notice Wraps SWTR to PSWTR.
    receive() external payable {
        _mint(_msgSender(), msg.value);
    }

    /// @notice Regular `balanceOf` function is disabled to force users to use `balanceOfWithSignature` function
    function balanceOf(address account) public view override returns (uint256) {
        revert("PSWTR: Public balanceOf was disabled");
    }

    /// @notice Modified `balanceOf` function with signature check
    function balanceOfWithSignature(address account, bytes memory sig) public view returns (uint256) {
        bool isAccountOwner = keccak256(abi.encodePacked(account))
            .toEthSignedMessageHash()
            .recover(sig) == account;

        require(isAccountOwner, "PSWTR: You can check only your own balance");
        return super.balanceOf(account); 
    }
}