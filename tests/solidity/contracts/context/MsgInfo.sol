pragma solidity ^0.8;

// Contract if msg.* fields are returned correctly
// Correctness of msg.sender value is checked in Sender.sol and its test
contract MsgInfo {
    // Stores the value sent with the message
    uint256 public storedValue;

    function updateValue() public payable {
        storedValue = msg.value;
    }

    function getMsgData(uint256 encodedParam) public pure returns (bytes memory) {
        return msg.data;
    }

    function getMsgSig() public pure returns (bytes4) {
        return msg.sig;
    }
}
