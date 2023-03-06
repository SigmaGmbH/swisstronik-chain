pragma solidity ^0.8.0;

contract OpCodes {
    function test_revert() public {
        //revert
        assembly{revert(0, 0)}
    }

    function test_invalid() public {
        //revert
        assembly{invalid()}
    }
}
