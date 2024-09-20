pragma solidity ^0.8;

contract MsgInfo {
    // Stores the value sent with the message
    uint256 public storedValue;

    function updateValue() public payable {
        storedValue = msg.value;
    }

    function getMsgSender() public view returns (address) {
        return msg.sender;
    }

    function getMsgValue() public payable returns (uint256) {
        return msg.value;
    }

    function getMsgData() public pure returns (bytes memory) {
        return msg.data;
    }

    function getMsgSig() public pure returns (bytes4) {
        return msg.sig;
    }

    function getAllMsgInfo() public payable returns (
        address sender,
        uint256 value,
        bytes memory data,
        bytes4 sig
    ) {
        return (
            msg.sender,
            msg.value,
            msg.data,
            msg.sig
        );
    }
}
