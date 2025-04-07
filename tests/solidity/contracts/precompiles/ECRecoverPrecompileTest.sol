// SPDX-License-Identifier: MIT
pragma solidity ^0.8;

contract ECRecoverPrecompileTest {
    /**
     * @dev Recovers the address of the signer from a signed message.
     * @param messageHash The hash of the message that was signed.
     * @param v The recovery ID (27 or 28 for Ethereum, or adjusted by EIP-155).
     * @param r The first 32 bytes of the signature.
     * @param s The second 32 bytes of the signature.
     * @return The Ethereum address of the signer.
     */
    function recoverAddress(
        bytes32 messageHash,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) public pure returns (address) {
        // Ensure the signature is valid (v must be 27 or 28, or adjusted by EIP-155)
        require(v == 27 || v == 28, "Invalid v value");

        // Use the ecrecover precompile to recover the signer's address
        address recoveredAddress = ecrecover(messageHash, v, r, s);

        // Ensure the recovered address is not zero (invalid signature)
        require(recoveredAddress != address(0), "Invalid signature");

        return recoveredAddress;
    }
}