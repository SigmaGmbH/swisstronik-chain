// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract ED25519VerifyPrecompile {
    address internal constant PRECOMPILED_ED25519_VERIFIER = address(1031);

    function checkPrecompile() public view returns (bool) {
        bytes memory test25519Data =
                    hex"6162636465666768696a6b6c6d6e6f707172737475767778797a313233343536d75a980182b10ab7d54bfed3c964073a0ee172f3daa62325af021a68f707511a66310a048b3a29b2d45b4656e0a74314e529083433d66a606386d6448ba1e9eaf05854f2f02706ee0a14e16eeade37671eee18c137c73ffbcc342e26b3312b01";

        // Perform the static call
        (bool success, bytes memory data) = PRECOMPILED_ED25519_VERIFIER.staticcall(test25519Data);
        if (!success || data.length == 0) {
            return false;
        }

        // Decode the result
        bytes4 result = abi.decode(data, (bytes4));

        // Check it's 0 (valid signature)
        return result == bytes4(0);
    }
}