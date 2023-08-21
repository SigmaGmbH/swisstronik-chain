pragma solidity ^0.8;

contract Sender {
    function getSender() public view returns (address) {
        return msg.sender;
    }
}