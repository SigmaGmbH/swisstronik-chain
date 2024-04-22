// SPDX-License-Identifier: MIT
pragma solidity ^0.8.4;

error SampleError(uint256 provided, uint256 required);

contract TestRevert {
    event Passed();
    
    function testRevert(uint256 value) public {
        require(value < 10, "Expected value >= 10");
        emit Passed();
    }

    function testError(uint256 value) public {
        if (value < 10) {
            revert SampleError({
                provided: value,
                required: 10
            });
        }
        emit Passed();
    }
}
