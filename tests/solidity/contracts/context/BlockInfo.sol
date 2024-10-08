pragma solidity ^0.8;

contract BlockInfo {
    function getBlockBaseFee() public view returns (uint256) {
        return block.basefee;
    }

    function getBlockChainId() public view returns (uint256) {
        return block.chainid;
    }

    function getBlockCoinbase() public view returns (address) {
        return block.coinbase;
    }

    function getBlockDifficulty() public view returns (uint256) {
        return block.difficulty;
    }

    function getBlockGasLimit() public view returns (uint256) {
        return block.gaslimit;
    }

    function getBlockNumber() public view returns (uint256) {
        return block.number;
    }

    function getBlockTimestamp() public view returns (uint256) {
        return block.timestamp;
    }

    function getBlockhash(uint256 blockNumber) public view returns (bytes32) {
        return blockhash(blockNumber);
    }
}