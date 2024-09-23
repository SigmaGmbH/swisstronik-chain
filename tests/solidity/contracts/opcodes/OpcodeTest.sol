pragma solidity ^0.8;

contract OpcodeTest {
    uint256 public storedValue;

    function testSSTORE(uint256 _value) public {
        assembly {
            sstore(0, _value)
        }
    }

    function testMSTORE() public pure returns (uint256) {
        uint256 result;
        assembly {
            mstore(0x80, 42)
            result := mload(0x80)
        }
        return result;
    }

    function testEXTCODESIZE(address _addr) public view returns (uint256) {
        uint256 size;
        assembly {
            size := extcodesize(_addr)
        }
        return size;
    }

    function testAdd(uint256 a, uint256 b) public pure returns (uint256 result) {
        assembly {
            result := add(a, b)
        }
    }

    function testSub(uint256 a, uint256 b) public pure returns (uint256 result) {
        assembly {
            result := sub(a, b)
        }
    }

    function testMul(uint256 a, uint256 b) public pure returns (uint256 result) {
        assembly {
            result := mul(a, b)
        }
    }

    function testDiv(uint256 a, uint256 b) public pure returns (uint256 result) {
        assembly {
            result := div(a, b)
        }
    }

    function testMod(uint256 a, uint256 b) public pure returns (uint256 result) {
        assembly {
            result := mod(a, b)
        }
    }

    function testShl(uint256 a, uint256 b) public pure returns (uint256 result) {
        assembly {
            result := shl(b, a)
        }
    }

    function testShr(uint256 a, uint256 b) public pure returns (uint256 result) {
        assembly {
            result := shr(b, a)
        }
    }

    function testAnd(uint256 a, uint256 b) public pure returns (uint256 result) {
        assembly {
            result := and(a, b)
        }
    }

    function testOr(uint256 a, uint256 b) public pure returns (uint256 result) {
        assembly {
            result := or(a, b)
        }
    }

    function testXor(uint256 a, uint256 b) public pure returns (uint256 result) {
        assembly {
            result := xor(a, b)
        }
    }

    function testNot(uint256 a) public pure returns (uint256 result) {
        assembly {
            result := not(a)
        }
    }

    function testRevert() public {
        assembly{
            revert(0, 0)
        }
    }

    function testInvalid() public {
        assembly{
            invalid()
        }
    }
}