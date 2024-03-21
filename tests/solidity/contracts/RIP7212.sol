// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract RIP7212 {
    /// @dev The address of the pre-compiled p256 verifier contract (following RIP-7212)
    address internal constant PRECOMPILED_P256_VERIFIER = address(1032);

    /// @dev Check if the pre-compiled p256 verifier is available on this chain
    function isPreCompiledP256Available() public view returns (bool) {
        // Test signature data, from https://gist.github.com/ulerdogan/8f1714895e23a54147fc529ea30517eb
        bytes memory testSignatureData =
                    hex"4cee90eb86eaa050036147a12d49004b6b9c72bd725d39d4785011fe190f0b4da73bd4903f0ce3b639bbbf6e8e80d16931ff4bcf5993d58468e8fb19086e8cac36dbcd03009df8c59286b162af3bd7fcc0450c9aa81be5d10d312af6c66b1d604aebd3099c618202fcfe16ae7770b0c49ab5eadf74b754204a3bb6060e44eff37618b065f9832de4ca6ca971a7a1adc826d0f7c00181a5fb2ddf79ae00b4e10e";

        // Perform the static call
        (bool success, bytes memory data) = PRECOMPILED_P256_VERIFIER.staticcall(testSignatureData);
        if (!success || data.length == 0) {
            return false;
        }

        // Decode the result
        uint256 result = abi.decode(data, (uint256));

        // Check it's 1 (valid signature)
        return result == uint256(1);
    }
}